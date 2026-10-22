// 本文件测试 plugin.yaml 解析、插件 ID、包内相对路径与 PostgreSQL 标识符映射。
package manifest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fixture 读取 contracts/testdata/manifest 下的样例文件。
func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", "manifest", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// 市场 ID 无警告通过；不带市场前缀的本地 ID 通过并只给出缺前缀警告。
func TestParseFixtures(t *testing.T) {
	r, err := Parse([]byte(fixture(t, "market.yaml")))
	if err != nil {
		t.Fatalf("市场 manifest 应通过: %v", err)
	}
	if len(r.Warnings) != 0 || r.Manifest.ID != "alonabot_plugin_rss_feed" || *r.Manifest.ConfigVersion != 2 {
		t.Fatalf("解析结果不符: %+v warnings=%v", r.Manifest, r.Warnings)
	}

	r, err = Parse([]byte(fixture(t, "local_no_prefix.yaml")))
	if err != nil {
		t.Fatalf("本地 manifest 应通过: %v", err)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "市场前缀") {
		t.Fatalf("本地 ID 应只产生缺前缀警告，实际 %v", r.Warnings)
	}
}

// 对合法 fixture 做单处改动，每种非法输入都以对应的错误拒绝。
func TestParseRejects(t *testing.T) {
	base := fixture(t, "market.yaml")
	cases := []struct {
		// name 是子测试名。
		name string
		// old 是 fixture 中被替换的原文。
		old string
		// new 是替换后的文本。
		new string
		// wantErr 是错误信息中应包含的片段。
		wantErr string
	}{
		{"缺失 ID", "id: alonabot_plugin_rss_feed\n", "", "缺少插件 ID"},
		{"大写 ID", "id: alonabot_plugin_rss_feed", "id: My-Tool", "大写"},
		{"路径穿越 ID", "id: alonabot_plugin_rss_feed", "id: ../etc", "安全 slug"},
		{"Windows 保留名", "id: alonabot_plugin_rss_feed", "id: com1", "保留设备名"},
		{"市场前缀但格式错误", "id: alonabot_plugin_rss_feed", "id: alonabot_plugin_", "市场 ID"},
		{"超长 ID", "id: alonabot_plugin_rss_feed", "id: alonabot_plugin_" + strings.Repeat("a", 60), "超过"},
		{"未知字段", "source:", "unexpected: 1\nsource:", "not found"},
		{"manifest_version 不是 1", "manifest_version: 1", "manifest_version: 2", "manifest_version"},
		{"非法 SemVer", "version: 1.2.0-rc.1+build.5", "version: v1.2", "SemVer"},
		{"不支持的 os", "os: linux", "os: freebsd", "不支持的平台"},
		{"不支持的 arch", "arch: amd64", "arch: 386", "不支持的平台"},
		{"entrypoint 路径穿越", "entrypoint: bin/rss", "entrypoint: ../rss", ".."},
		{"entrypoint 绝对路径", "entrypoint: bin/rss", "entrypoint: /usr/bin/rss", "绝对路径"},
		{"entrypoint 反斜杠", "entrypoint: bin/rss", `entrypoint: 'bin\rss'`, "反斜杠"},
		{"未知 migration_mode", "migration_mode: versioned", "migration_mode: auto", "migration_mode"},
		{"未知 prepare_effects", "prepare_effects: managed_db", "prepare_effects: safe", "prepare_effects"},
		{"缺 config_version", "config_version: 2\n", "", "config_version"},
		{"缺 required_capabilities", "required_capabilities:\n  - name: core.send\n    version: 1\n", "", "required_capabilities"},
		{"协议范围倒置", "protocol_min_major: 1", "protocol_min_major: 2", "协议主版本范围"},
		{"多文档", "prepare_effects: managed_db\n", "prepare_effects: managed_db\n---\nid: x\n", "一个 YAML 文档"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := strings.Replace(base, c.old, c.new, 1)
			if src == base {
				t.Fatalf("用例没有改动 fixture（old=%q）", c.old)
			}
			_, err := Parse([]byte(src))
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("期望包含 %q 的错误，实际 %v", c.wantErr, err)
			}
		})
	}
}

// 缺少 prepare_effects 时仍可解析，但警告该包不允许自动升级。
func TestMissingDeclarationsWarn(t *testing.T) {
	src := strings.Replace(fixture(t, "market.yaml"), "prepare_effects: managed_db\n", "", 1)
	r, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "不允许自动升级") {
		t.Fatalf("缺少 prepare_effects 应警告不允许自动升级，实际 %v", r.Warnings)
	}
}

// 规范的包内相对路径通过；绝对路径、..、非规范形式、盘符与反斜杠被拒绝。
func TestValidateRelPath(t *testing.T) {
	for _, ok := range []string{"bin/rss", "rss.exe", "a/b/c"} {
		if err := ValidateRelPath(ok); err != nil {
			t.Errorf("%q 应通过: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "/abs", "a/../b", "./a", "a/", "a//b", "C:/x", `a\b`, "..", "a/./b"} {
		if err := ValidateRelPath(bad); err == nil {
			t.Errorf("%q 应被拒绝", bad)
		}
	}
}

// 映射结果不撞名、不超过 63 字节、只含合法字符且是确定性的。
func TestPostgresIdentifier(t *testing.T) {
	a, b := PostgresIdentifier("ab_", "my-tool"), PostgresIdentifier("ab_", "my_tool")
	if a == b {
		t.Fatalf("my-tool 与 my_tool 不应撞名: %s", a)
	}
	long := PostgresIdentifier("ab_", "alonabot_plugin_"+strings.Repeat("x", 48))
	if len(long) > 63 {
		t.Fatalf("标识符长度 %d 超过 63", len(long))
	}
	ok := regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)
	for _, id := range []string{"my-tool", "alonabot_plugin_rss", long} {
		if got := PostgresIdentifier("ab_", id); !ok.MatchString(got) {
			t.Errorf("%q 映射为非法标识符 %q", id, got)
		}
	}
	if PostgresIdentifier("ab_", "x") != PostgresIdentifier("ab_", "x") {
		t.Fatal("映射必须是确定性的")
	}
}
