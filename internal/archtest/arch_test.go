// Package archtest 机械执行 ARCHITECTURE.md「依赖方向」：A -> B 表示 A 的 Go 源码 import B。
// 同一模块内部的 import 不受限制；跨模块的边必须在 allowed 中列出；不属于任何已登记模块的包视为违规。
// 生产代码与测试代码（TestImports/XTestImports）都检查。import 环由 Go 编译器保证不会出现。
package archtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// modules 与 ARCHITECTURE.md「模块与职责」一致，按路径前缀归属，较长前缀优先。
var modules = []string{
	"contracts",
	"internal/wire",
	"internal/core",
	"internal/adapters",
	"internal/archtest",
	"sdk",
	"cmd/alonabot",
	"examples",
}

// allowed[from] 是 from 模块可以 import 的其他模块路径前缀（与 ARCHITECTURE.md 的允许边表一致）。
var allowed = map[string][]string{
	"contracts":         {},
	"internal/wire":     {"contracts"},
	"sdk":               {"contracts", "internal/wire"},
	"internal/core":     {"contracts", "internal/wire"},
	"internal/adapters": {"internal/core/adapter", "contracts"},
	"internal/archtest": {},
	"cmd/alonabot":      {"internal/core", "internal/adapters", "contracts", "sdk/generate"},
	"examples":          {"sdk", "contracts"},
}

type pkg struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
	Module       *struct{ Path string }
}

// hasPrefix 按路径段判断前缀，避免 "sdkx" 被当成 "sdk"。
func hasPrefix(p, prefix string) bool {
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

func moduleOf(rel string) string {
	best := ""
	for _, m := range modules {
		if hasPrefix(rel, m) && len(m) > len(best) {
			best = m
		}
	}
	return best
}

// check 在 dir 下运行 go list，返回所有违规描述（已排序）。
func check(dir string) ([]string, error) {
	// import 分析不需要 VCS 信息；关闭后在无 git 或取不到 VCS 状态的环境（例如 Rosetta 下的 darwin/amd64）也能运行。
	cmd := exec.Command("go", "list", "-buildvcs=false", "-json", "./...")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list 失败: %v: %s", err, stderr.String())
	}
	var violations []string
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var p pkg
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if p.Module == nil {
			return nil, fmt.Errorf("%s 不属于任何 module", p.ImportPath)
		}
		root := p.Module.Path
		rel := strings.TrimPrefix(strings.TrimPrefix(p.ImportPath, root), "/")
		from := moduleOf(rel)
		if from == "" {
			violations = append(violations, fmt.Sprintf("%s：不属于 ARCHITECTURE.md 中任何已登记的模块；新增顶层包时先更新依赖表", rel))
			continue
		}
		imports := append(append(append([]string{}, p.Imports...), p.TestImports...), p.XTestImports...)
		seen := map[string]bool{}
		for _, imp := range imports {
			if seen[imp] || !hasPrefix(imp, root) || imp == root {
				continue // 标准库与第三方依赖不受此表限制
			}
			seen[imp] = true
			target := strings.TrimPrefix(imp, root+"/")
			if hasPrefix(target, from) {
				continue // 同一模块内部
			}
			ok := false
			for _, a := range allowed[from] {
				if hasPrefix(target, a) {
					ok = true
					break
				}
			}
			if !ok {
				violations = append(violations, fmt.Sprintf("%s -> %s：模块 %s 不允许 import %s（见 ARCHITECTURE.md「依赖方向」；需要这条边时先修改依赖表）",
					rel, target, from, target))
			}
		}
	}
	sort.Strings(violations)
	return violations, nil
}

func TestModuleImportsFollowArchitecture(t *testing.T) {
	violations, err := check("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		t.Error(v)
	}
}

// 负例：fixture module 中 contracts/bad -> sdk 被禁止、tools/x 未登记；sdk -> contracts 合法。
func TestGuardRejectsForbiddenEdges(t *testing.T) {
	violations, err := check("testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"contracts/bad -> sdk：", "tools/x：不属于"}
	if len(violations) != len(want) {
		t.Fatalf("违规数 = %d，期望 %d：%q", len(violations), len(want), violations)
	}
	for i, w := range want {
		if !strings.HasPrefix(violations[i], w) {
			t.Errorf("第 %d 条违规 = %q，期望以 %q 开头", i, violations[i], w)
		}
	}
}
