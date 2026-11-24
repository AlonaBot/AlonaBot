package settings_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// TestInstanceConfigurationFiles 验证实例配置解析环境变量、补齐缺省值并保留源文件。
func TestInstanceConfigurationFiles(t *testing.T) {
	t.Setenv("INSTANCE_TOKEN", "local-token")
	schema, problems := configschema.Parse([]byte(`{"token":{"type":"string","secret":true,"required":true},"api_base_url":{"type":"string","default":"https://api.telegram.org"},"limit":{"type":"int","default":10}}`))
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	dir := t.TempDir()
	name := filepath.Join(dir, settings.FileName)
	original := []byte(`{"adapters":{"telegram":{"instances":{"main":{"token":"${INSTANCE_TOKEN}","limit":12}}}}}`)
	if err := os.WriteFile(name, original, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := settings.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	ready, failures, err := file.PrepareInstances("telegram", schema)
	if err != nil || len(ready) != 1 || len(failures) != 0 {
		t.Fatalf("PrepareInstances = %v, %v, %v", ready, failures, err)
	}
	var complete struct {
		Token      string
		APIBaseURL string `json:"api_base_url"`
		Limit      int
	}
	if err := json.Unmarshal(ready["main"], &complete); err != nil {
		t.Fatal(err)
	}
	if complete.Token != "local-token" || complete.APIBaseURL != "https://api.telegram.org" || complete.Limit != 12 {
		t.Fatalf("incomplete runtime config: %s", ready["main"])
	}
	content, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(content, original) {
		t.Fatalf("preparation changed file: %s, %v", content, err)
	}
}
