package settings_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestInstanceConfigurationFiles 通过公开入口验证真实文件迁移、备份与失败保留。
func TestInstanceConfigurationFiles(t *testing.T) {
	t.Setenv("INSTANCE_TOKEN", "local-token")
	schema, diagnostics := configschema.Parse([]byte(`{
		"$version":2,"$migrations":[{"to":2,"rename":{"old_token":"token"}}],
		"token":{"type":"string","secret":true,"required":true},
		"api_base_url":{"type":"string","default":"https://api.telegram.org"}
	}`))
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	for _, test := range []struct {
		name, original string
		problems       int
		ready          int
		loadError      bool
	}{
		{name: "rename and backup", original: `{"$version":1,"adapters":{"telegram":{"instances":{"main":{"old_token":"${INSTANCE_TOKEN}","retained":{"large":9007199254740993}}}}}}`, ready: 1},
		{name: "validation failure", original: `{"$version":1,"adapters":{"telegram":{"$version":1,"instances":{"main":{"old_token":"${INSTANCE_TOKEN}"},"invalid":{"old_token":""}}}}}`, ready: 1, problems: 1},
		{name: "adapter downgrade", original: `{"$version":1,"adapters":{"telegram":{"$version":3,"instances":{"main":{"token":"${INSTANCE_TOKEN}"},"other":{"token":"other-token"}}}}}`, problems: 1},
		{name: "core downgrade", original: `{"$version":2,"adapters":{"telegram":{"instances":{"main":{"token":"${INSTANCE_TOKEN}"}}}}}`, loadError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			name := filepath.Join(dir, settings.FileName)
			original := []byte(test.original)
			if err := os.WriteFile(name, original, 0o600); err != nil {
				t.Fatal(err)
			}
			file, err := settings.Load(dir)
			if test.loadError {
				if err == nil || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "version 2") || !strings.Contains(err.Error(), "version 1") {
					t.Fatalf("incomplete downgrade error: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				ready, problems, err := file.PrepareInstances("telegram", schema)
				if err != nil || len(problems) != test.problems || len(ready) != test.ready {
					t.Fatalf("PrepareInstances = %v, %v, %v", ready, problems, err)
				}
				if test.problems == 0 {
					if file.Settings().Adapters["telegram"].Version != 2 {
						t.Fatal("migration did not publish the new adapter version")
					}
				} else if test.name == "validation failure" {
					if problems[0].InstanceID != "invalid" || !strings.HasSuffix(problems[0].Path, ".token") {
						t.Fatalf("wrong invalid instance path: %+v", problems[0])
					}
				} else if !strings.Contains(problems[0].Reason, "version 3") || !strings.Contains(problems[0].Reason, "version 2") {
					t.Fatalf("missing downgrade versions: %+v", problems)
				}
				if test.ready > 0 {
					var complete map[string]any
					if err := json.Unmarshal(ready["main"], &complete); err != nil {
						t.Fatal(err)
					}
					if complete["token"] != "local-token" || complete["api_base_url"] != "https://api.telegram.org" {
						t.Fatalf("incomplete running configuration: %s", ready["main"])
					}
					if _, exists := complete["old_token"]; exists {
						t.Fatal("rename left its source in running configuration")
					}
				}
			}
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			backupDir := filepath.Join(dir, settings.BackupDir)
			if test.loadError || test.problems > 0 {
				if !bytes.Equal(content, original) {
					t.Fatal("failed migration changed the original file")
				}
				if _, err := os.Stat(backupDir); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("failed migration changed the backup directory: %v", err)
				}
				return
			}
			var saved struct {
				Adapters map[string]struct {
					Version int `json:"$version"`
				} `json:"adapters"`
			}
			if err := json.Unmarshal(content, &saved); err != nil {
				t.Fatal(err)
			}
			if saved.Adapters["telegram"].Version != 2 || bytes.Contains(content, []byte(`"old_token"`)) || !bytes.Contains(content, []byte(`${INSTANCE_TOKEN}`)) || bytes.Contains(content, []byte("local-token")) || !bytes.Contains(content, []byte("9007199254740993")) {
				t.Fatalf("migration lost configuration or expanded the stored secret: %s", content)
			}
			entries, err := os.ReadDir(backupDir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("backup entries = %v, %v", entries, err)
			}
			backup, err := os.ReadFile(filepath.Join(backupDir, entries[0].Name()))
			if err != nil || !bytes.Equal(backup, original) {
				t.Fatalf("backup differs from original: %s, %v", backup, err)
			}
		})
	}
}
