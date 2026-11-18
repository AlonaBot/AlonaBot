package settings_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestInstanceConfigurationFiles 验证实例校正只作用于运行副本，失败实例独立返回。
func TestInstanceConfigurationFiles(t *testing.T) {
	t.Setenv("INSTANCE_TOKEN", "local-token")
	schema, problems := configschema.Parse([]byte(`{"token":{"type":"string","secret":true,"required":true},"api_base_url":{"type":"string","default":"https://api.telegram.org"},"limit":{"type":"int","default":10}}`))
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	dir := t.TempDir()
	name := filepath.Join(dir, settings.FileName)
	original := []byte(`{"adapters":{"telegram":{"instances":{"main":{"token":"${INSTANCE_TOKEN}","limit":"bad","retained":{"large":9007199254740993}},"invalid":{"token":""}}}}}`)
	if err := os.WriteFile(name, original, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := settings.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	ready, failures, err := file.PrepareInstances("telegram", schema)
	if err != nil || len(ready) != 1 || len(failures) != 1 {
		t.Fatalf("PrepareInstances = %v, %v, %v", ready, failures, err)
	}
	if failures[0].InstanceID != "invalid" || !strings.HasSuffix(failures[0].Path, ".token") {
		t.Fatalf("wrong invalid path: %+v", failures[0])
	}
	var complete struct {
		Token      string
		APIBaseURL string `json:"api_base_url"`
		Limit      int
	}
	if err := json.Unmarshal(ready["main"], &complete); err != nil {
		t.Fatal(err)
	}
	if complete.Token != "local-token" || complete.APIBaseURL != "https://api.telegram.org" || complete.Limit != 10 {
		t.Fatalf("incomplete runtime config: %s", ready["main"])
	}
	if !bytes.Contains(ready["main"], []byte("9007199254740993")) {
		t.Fatalf("unknown value lost: %s", ready["main"])
	}
	content, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(content, original) {
		t.Fatalf("preparation changed file: %s, %v", content, err)
	}
	group := file.Settings().Adapters["telegram"]
	if len(group.Corrections["main"]) != 1 || group.Corrections["main"][0].Reason != "reset" || len(group.Unknown["main"]) != 1 {
		t.Fatalf("missing correction/unknown: %+v", group)
	}
}
