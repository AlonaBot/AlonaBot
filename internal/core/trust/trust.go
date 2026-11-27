// Package trust 定义官方组件的仓库来源。
//
// 清单维护在同目录的 official.yaml，构建时嵌入二进制。
package trust

import (
	_ "embed"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

//go:embed official.yaml
var officialYAML []byte

// official 是组件 ID 到规范来源的映射，包初始化时由 official.yaml 生成。
var official = mustParse(officialYAML)

func mustParse(raw []byte) map[string]string {
	var entries []struct {
		ID     string `yaml:"id"`
		Source string `yaml:"source"`
	}
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		panic(fmt.Sprintf("trust: parse official.yaml: %v", err))
	}
	sources := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.ID == "" || entry.Source == "" {
			panic(fmt.Sprintf("trust: official.yaml entry %q has an empty id or source", entry.ID))
		}
		if _, ok := sources[entry.ID]; ok {
			panic(fmt.Sprintf("trust: official.yaml lists %q twice", entry.ID))
		}
		sources[entry.ID] = entry.Source
	}
	return sources
}

// Source 返回官方组件独立仓库的规范来源；未知组件返回空字符串。
func Source(id string) string {
	return official[id]
}

// MatchesSource 同时绑定组件身份与确切仓库来源。
func MatchesSource(id, source string) bool {
	expected := Source(id)
	return expected != "" && strings.TrimPrefix(source, "https://") == expected
}
