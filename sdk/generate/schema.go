package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

const configSchemaFileName = "config_schema.json"

func (s *scan) configSchema() error {
	// 1. 未声明 schema 时使用空字段集核对翻译，生成项保持 nil。
	raw, err := os.ReadFile(filepath.Join(s.dir, configSchemaFileName))
	present := err == nil
	if errors.Is(err, os.ErrNotExist) {
		raw = []byte(`{}`)
	} else if err != nil {
		return fmt.Errorf("read config schema: %w", err)
	}
	schema, problems := configschema.Parse(raw)
	s.schemaProblems(configSchemaFileName, problems)
	if len(problems) != 0 {
		return nil
	}
	if present {
		s.result.ConfigSchema = schema.Canonical()
	}
	// 2. 复用已读原文核对 config 字段与模板；JSON 解析问题由 readI18n 报告。
	for _, file := range slices.Sorted(maps.Keys(s.i18nFiles)) {
		relative := filepath.ToSlash(filepath.Join(i18nDirectory, file))
		var translation struct {
			Config json.RawMessage `json:"config"`
		}
		if err := json.Unmarshal(s.i18nFiles[file], &translation); err != nil {
			continue
		}
		s.schemaProblems(relative, schema.CheckI18n(translation.Config))
	}
	return nil
}

func (s *scan) schemaProblems(file string, problems []configschema.Problem) {
	for _, problem := range problems {
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{file + ":" + problem.Path, problem.Message})
	}
}
