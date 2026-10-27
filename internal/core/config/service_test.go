package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// 配置编辑以 base_generation 判定陈旧提交，Pending 随 Active 变化失效；Patch 叠加在有效 Pending
// 或 Active 配置上，保留未知键、显式空值与超出 2^53 的数字（CONTRACTS §5.1、V-23、V-38）。
func TestPendingConfigEdits(t *testing.T) {
	ctx := context.Background()

	t.Run("base generation must equal Active", func(t *testing.T) {
		store, dir := activatedStore(t, `{"from":"active"}`)
		service := New(store, dir)
		edits := []struct {
			name string
			call func(base string) (Pending, error)
		}{
			{"Replace", func(base string) (Pending, error) { return service.Replace(ctx, "demo", base, []byte(`{}`)) }},
			{"Patch", func(base string) (Pending, error) { return service.Patch(ctx, "demo", base, []byte(`{}`)) }},
		}
		for _, edit := range edits {
			for _, base := range []string{"", "stale"} {
				if _, err := edit.call(base); !hasCode(err, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION) {
					t.Fatalf("%s with base %q should be FAILED_PRECONDITION, got %v", edit.name, base, err)
				}
			}
			if _, err := edit.call("g1"); err != nil {
				t.Fatalf("%s with the current Active should succeed: %v", edit.name, err)
			}
		}
	})

	t.Run("Active change expires Pending", func(t *testing.T) {
		store, dir := activatedStore(t, `{"from":"g1"}`)
		service := New(store, dir)
		if _, err := service.Replace(ctx, "demo", "g1", []byte(`{"edited":true}`)); err != nil {
			t.Fatal(err)
		}
		if p, found, err := service.Pending(ctx, "demo"); err != nil || !found || string(p.ConfigJSON) != `{"edited":true}` {
			t.Fatalf("pending before the new generation: %+v %v %v", p, found, err)
		}
		if err := store.AddGeneration(ctx, "demo", "g2", `{"from":"g2"}`, 1); err != nil {
			t.Fatal(err)
		}
		if err := store.CommitActive(ctx, "demo", "g1", "g2"); err != nil {
			t.Fatal(err)
		}
		if _, found, err := service.Pending(ctx, "demo"); err != nil || found {
			t.Fatalf("Pending based on the old Active should not be returned: found=%v err=%v", found, err)
		}
		if stored, err := store.GetPendingConfig(ctx, "demo"); err != nil || stored.BaseGeneration != "g1" {
			t.Fatalf("the expired row should stay until it is discarded or overwritten: %+v %v", stored, err)
		}
	})

	t.Run("Patch layers over Pending", func(t *testing.T) {
		store, dir := unactivatedStore(t)
		service := New(store, dir)
		// 尚无 Active 时，Replace 与 Patch 以空串为基础。
		if _, err := service.Replace(ctx, "demo", "", []byte(`{"a":1,"unknown":{"nested":false},"big":9007199254740993}`)); err != nil {
			t.Fatal(err)
		}
		patched, err := service.Patch(ctx, "demo", "", []byte(`{"a":2,"empty":[],"removed":null}`))
		if err != nil {
			t.Fatal(err)
		}
		const want = `{"a":2,"big":9007199254740993,"empty":[],"unknown":{"nested":false}}`
		if string(patched.ConfigJSON) != want {
			t.Fatalf("Patch over Pending = %s, want %s", patched.ConfigJSON, want)
		}
	})

	t.Run("Patch layers over Active config", func(t *testing.T) {
		store, dir := activatedStore(t, `{"from":"active","unknown":{"keep":true},"zero":0}`)
		service := New(store, dir)
		patched, err := service.Patch(ctx, "demo", "g1", []byte(`{"flag":false,"empty":{},"listed":[1,2]}`))
		if err != nil {
			t.Fatal(err)
		}
		const want = `{"empty":{},"flag":false,"from":"active","listed":[1,2],"unknown":{"keep":true},"zero":0}`
		if string(patched.ConfigJSON) != want {
			t.Fatalf("Patch over Active = %s, want %s", patched.ConfigJSON, want)
		}
		if p, found, err := service.Pending(ctx, "demo"); err != nil || !found || string(p.ConfigJSON) != want || p.BaseGeneration != "g1" {
			t.Fatalf("stored Pending = %+v %v %v", p, found, err)
		}
	})

	t.Run("ImportSnapshot does not change Active", func(t *testing.T) {
		store, dir := activatedStore(t, `{"from":"active"}`)
		service := New(store, dir)
		if _, err := service.ImportSnapshot(ctx, "demo"); !hasCode(err, pb.ErrorCode_ERROR_CODE_NOT_FOUND) {
			t.Fatalf("a missing config.json should be NOT_FOUND, got %v", err)
		}
		path := filepath.Join(dir, "plugins", "demo", "config.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		const canonical = `{"big":9007199254740993,"edited":true,"unknown":{"x":null}}`
		if err := os.WriteFile(path, []byte(`{"edited":true,"unknown":{"x":null},"big":9007199254740993}`), 0o600); err != nil {
			t.Fatal(err)
		}
		imported, err := service.ImportSnapshot(ctx, "demo")
		if err != nil {
			t.Fatal(err)
		}
		if string(imported.ConfigJSON) != canonical || imported.BaseGeneration != "g1" {
			t.Fatalf("imported Pending = %+v", imported)
		}
		if active, err := store.Active(ctx, "demo"); err != nil || active != "g1" {
			t.Fatalf("import must not change Active: %q %v", active, err)
		}

		// 非对象快照被拒，已有 Pending 保持不变。
		if err := os.WriteFile(path, []byte(`[1,2]`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := service.ImportSnapshot(ctx, "demo"); !hasCode(err, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT) {
			t.Fatalf("a non-object config.json should be INVALID_ARGUMENT, got %v", err)
		}
		if p, found, err := service.Pending(ctx, "demo"); err != nil || !found || string(p.ConfigJSON) != canonical {
			t.Fatalf("a rejected import must leave Pending unchanged: %+v %v %v", p, found, err)
		}
	})
}

// unactivatedStore 返回还没有 Active 代次的插件状态库与 data_dir。
func unactivatedStore(t *testing.T) (*state.Store, string) {
	t.Helper()
	store, dir := openStore(t)
	installPlugin(t, store, "demo")
	return store, dir
}

// activatedStore 以 config 建立代次 g1 并提交为 Active。
func activatedStore(t *testing.T, config string) (*state.Store, string) {
	t.Helper()
	store, dir := unactivatedStore(t)
	ctx := context.Background()
	if err := store.AddGeneration(ctx, "demo", "g1", config, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.CommitActive(ctx, "demo", "", "g1"); err != nil {
		t.Fatal(err)
	}
	return store, dir
}

// openStore 在临时 data_dir 打开真实状态库，测试结束时关闭。
func openStore(t *testing.T) (*state.Store, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := state.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, dir
}

// installPlugin 登记一个最小安装记录，让配置编辑有可写入的插件。
func installPlugin(t *testing.T, store *state.Store, pluginID string) {
	t.Helper()
	if err := store.AddInstall(context.Background(), state.Install{
		PluginID:    pluginID,
		Source:      "dev-local",
		Entrypoint:  "bin/" + pluginID,
		BuildDigest: []byte{1},
	}); err != nil {
		t.Fatal(err)
	}
}

// hasCode 报告 err 是否为带指定 Code 的 *validate.Error。
func hasCode(err error, code pb.ErrorCode) bool {
	var failure *validate.Error
	return errors.As(err, &failure) && failure.Code == code
}
