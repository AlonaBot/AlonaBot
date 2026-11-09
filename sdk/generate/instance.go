package generate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

const instanceSchemaFileName = "instance_schema.json"

func (s *scan) instanceSchema() (bool, error) {
	raw, err := os.ReadFile(filepath.Join(s.dir, instanceSchemaFileName))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read instance schema: %w", err)
	}
	_, problems := configschema.Parse(raw)
	s.schemaProblems(instanceSchemaFileName, problems)
	return true, nil
}

// hasBuildSource 让只有实例声明的目录直接完成检查；有源码时仍由 Go 工具链选择文件。
func hasBuildSource(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read instance schema directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && name != FileName {
			return true, nil
		}
	}
	return false, nil
}
