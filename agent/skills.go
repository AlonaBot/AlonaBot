package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/agent/provider"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"go.yaml.in/yaml/v3"
)

const (
	skillFileName         = "SKILL.md"
	skillFileLimit        = 256 << 10
	skillNameLimit        = 64
	skillDescriptionLimit = 1024 // Unicode 码点数。
	skillDelimiter        = "---\n"
	loadSkillToolName     = "load_skill"
)

type skill struct {
	name, description, body string
}

func readSkills(directory string) []skill {
	if directory == "" {
		return nil
	}
	directory = filepath.Clean(directory)
	// 1. 目录项按名字读取，只处理真实的一级子目录。
	info, err := os.Lstat(directory)
	if err != nil {
		slog.Warn("skills directory unavailable", "path", directory, "error", err)
		return nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		slog.Warn("skills directory unavailable", "path", directory, "error", err)
		return nil
	}
	var skills []skill
	names := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name(), skillFileName)
		// 先匹配目录项原名，避免大小写不敏感文件系统把 skill.md 当成 SKILL.md。
		files, err := os.ReadDir(filepath.Dir(path))
		if err == nil && !slices.ContainsFunc(files, func(file os.DirEntry) bool { return file.Name() == skillFileName }) {
			continue
		}
		var info os.FileInfo
		if err == nil {
			info, err = os.Lstat(path)
		}
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		var loaded skill
		if err == nil {
			loaded, err = readSkillFile(path, info)
		}
		if err == nil && names[loaded.name] {
			err = fmt.Errorf("duplicate skill name %q", loaded.name)
		}
		if err != nil {
			slog.Warn("invalid skill file", "path", path, "error", err)
			continue
		}
		names[loaded.name] = true
		skills = append(skills, loaded)
	}
	// 2. 提示顺序只取决于技能名称，重复名称的优先级已由目录顺序确定。
	sort.Slice(skills, func(i, j int) bool { return skills[i].name < skills[j].name })
	return skills
}

func readSkillFile(path string, info os.FileInfo) (skill, error) {
	if !info.Mode().IsRegular() {
		return skill{}, fmt.Errorf("skill must be a regular file")
	}
	if info.Size() > skillFileLimit {
		return skill{}, fmt.Errorf("skill exceeds %d bytes", skillFileLimit)
	}
	file, err := os.Open(path)
	if err != nil {
		return skill{}, err
	}
	defer file.Close()
	// 文件在检查后增长时仍只读取上限加一字节，以识别超限而非截断成功。
	content, err := io.ReadAll(io.LimitReader(file, skillFileLimit+1))
	if err != nil {
		return skill{}, err
	}
	if len(content) > skillFileLimit {
		return skill{}, fmt.Errorf("skill exceeds %d bytes", skillFileLimit)
	}
	return parseSkill(content)
}

func parseSkill(content []byte) (skill, error) {
	// 1. 分隔符必须独占一行；正文保留分隔符之后的原文。
	frontmatter, ok := bytes.CutPrefix(content, []byte(skillDelimiter))
	if !ok {
		return skill{}, fmt.Errorf("missing frontmatter")
	}
	end := bytes.Index(frontmatter, []byte("\n"+skillDelimiter))
	if end < 0 {
		return skill{}, fmt.Errorf("missing closing frontmatter delimiter")
	}
	var metadata struct {
		Name        yaml.Node `yaml:"name"`
		Description yaml.Node `yaml:"description"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(frontmatter[:end+1]))
	if err := decoder.Decode(&metadata); err != nil {
		return skill{}, fmt.Errorf("decode frontmatter: %w", err)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return skill{}, fmt.Errorf("decode frontmatter: %w", err)
		}
		return skill{}, fmt.Errorf("frontmatter must contain one YAML document")
	}
	// 2. Node 保留 YAML 类型，避免数字和布尔值被自动转成字符串。
	name := metadata.Name.Value
	if metadata.Name.Tag != "!!str" || len(name) == 0 || len(name) > skillNameLimit {
		return skill{}, fmt.Errorf("name must match [a-z0-9_-]{1,%d}", skillNameLimit)
	}
	for _, ch := range name {
		if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-') {
			return skill{}, fmt.Errorf("name must match [a-z0-9_-]{1,%d}", skillNameLimit)
		}
	}
	description := metadata.Description.Value
	if metadata.Description.Tag != "!!str" || strings.TrimSpace(description) == "" || utf8.RuneCountInString(description) > skillDescriptionLimit {
		return skill{}, fmt.Errorf("description must contain 1 to %d characters", skillDescriptionLimit)
	}
	return skill{name: name, description: description, body: string(frontmatter[end+1+len(skillDelimiter):])}, nil
}

func (a *component) skillPrompt(header string) string {
	var prompt strings.Builder
	prompt.WriteString(header)
	for _, skill := range a.skills {
		prompt.WriteByte('\n')
		prompt.WriteString(skill.name + ": " + skill.description)
	}
	return prompt.String()
}

func (a *component) addSkillTool(catalog *toolCatalog) {
	if len(a.skills) == 0 {
		return
	}
	tool := roundTool{local: a.loadSkill, definition: provider.Tool{
		Name: loadSkillToolName, Description: "Load the full body of an available skill by name",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`),
	}}
	catalog.byName[loadSkillToolName] = tool
	catalog.definitions = append(catalog.definitions, tool.definition)
}

func (a *component) loadSkill(arguments json.RawMessage) string {
	var input struct {
		Name *string `json:"name"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return toolResult("", errcode.InvalidArgument.New("invalid skill arguments: "+err.Error()))
	}
	if input.Name == nil {
		return toolResult("", errcode.InvalidArgument.New("name must be a string"))
	}
	for _, skill := range a.skills {
		if skill.name == *input.Name {
			return skill.body
		}
	}
	return "error: NOT_FOUND: unknown skill"
}
