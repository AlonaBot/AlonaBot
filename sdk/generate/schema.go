package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

const configSchemaFileName = "config_schema.json"
const translationsDirectory = "i18n"

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
	// 2. 只提取 config 根节点核对字段与模板，翻译目录的其他检查由其所有者负责。
	directory := filepath.Join(s.dir, translationsDirectory)
	files, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config translations: %w", err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		relative := filepath.ToSlash(filepath.Join(translationsDirectory, file.Name()))
		raw, err := os.ReadFile(filepath.Join(directory, file.Name()))
		if err != nil {
			return fmt.Errorf("read config translation %s: %w", relative, err)
		}
		var translation struct {
			Config json.RawMessage `json:"config"`
		}
		if err := json.Unmarshal(raw, &translation); err != nil {
			s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{relative, err.Error()})
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
