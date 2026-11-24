package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// TestPendingConfigEdits 验证 Pending 随 Active 变化失效，以及配置叠加与文件导入。
func TestPendingConfigEdits(t *testing.T) {
	ctx := context.Background()

	t.Run("Active change expires Pending", func(t *testing.T) {
		store, dir := activatedStore(t, `{"from":"g1"}`)
		service := New(store, dir)
		if _, err := service.Replace(ctx, "demo", "g1", []byte(`{"edited":true}`)); err != nil {
			t.Fatal(err)
		}
		if p, found, err := service.Pending(ctx, "demo"); err != nil || !found || string(p.ConfigJSON) != `{"edited":true}` {
			t.Fatalf("pending before the new generation: %+v %v %v", p, found, err)
		}
		if err := store.AddGeneration(ctx, "demo", "g2", `{"from":"g2"}`); err != nil {
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
		if _, err := service.Replace(ctx, "demo", "", []byte(`{"a":1,"unknown":{"nested":false}}`)); err != nil {
			t.Fatal(err)
		}
		patched, err := service.Patch(ctx, "demo", "", []byte(`{"a":2,"empty":[],"removed":null}`))
		if err != nil {
			t.Fatal(err)
		}
		const want = `{"a":2,"empty":[],"unknown":{"nested":false}}`
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
		if pending, err := service.ImportSnapshot(ctx, "demo"); err != nil || string(pending.ConfigJSON) != `{"from":"active"}` {
			t.Fatalf("a missing config.json should be restored from Active, got %+v %v", pending, err)
		}
		path := filepath.Join(dir, "plugins", "demo", "config.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		const canonical = `{"edited":true,"unknown":{"x":null}}`
		if err := os.WriteFile(path, []byte(`{"edited":true,"unknown":{"x":null}}`), 0o600); err != nil {
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
	if err := store.AddGeneration(ctx, "demo", "g1", config); err != nil {
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
	if err := store.AddInstall(context.Background(), state.Install{Role: state.RolePlugin,
		PluginID:    pluginID,
		Source:      "dev-local",
		Entrypoint:  "bin/" + pluginID,
		BuildDigest: []byte{1},
	}); err != nil {
		t.Fatal(err)
	}
}
