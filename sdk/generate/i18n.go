package generate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
)

const i18nDirectory = "i18n"
const i18nExtension = ".json"

func (s *scan) readI18n() error {
	directory := filepath.Join(s.dir, i18nDirectory)
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read i18n directory: %w", err)
	}
	files := make(map[string][]byte)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), i18nExtension) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return fmt.Errorf("read i18n file: %w", err)
		}
		files[entry.Name()] = data
	}
	// 存在 i18n 目录时必须提供源语言；Parse 的空输入留给无资源的组件。
	if len(files) == 0 {
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{
			Pos: i18nDirectory + "/en-US.json", Message: "source catalog is missing",
		})
		return nil
	}
	_, problems := i18n.Parse(files)
	for _, problem := range problems {
		position := i18nDirectory + "/" + problem.File
		if problem.Key != "" {
			position += ":" + problem.Key
		}
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{Pos: position, Message: problem.Message})
	}
	if len(problems) == 0 {
		s.i18nFiles = files
	}
	return nil
}
