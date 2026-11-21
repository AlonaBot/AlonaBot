// Package pluginindex 解析插件源仓库中的 index.json。
//
// [Index] 与 [Plugin] 描述索引及其条目；[Parse] 校验输入并报告 JSON 路径。
package pluginindex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const indexVersion = 1

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// Index 保存源格式版本与组件仓库条目。
type Index struct {
	Version int      `json:"version"`
	Plugins []Plugin `json:"plugins"`
}

// Plugin 指向组件的 GitHub 仓库，展示元数据由消费方从 Release 读取。
type Plugin struct {
	Repo string   `json:"repo"` // owner/name，与安装来源使用同一格式。
	Role string   `json:"role"`
	Tags []string `json:"tags,omitempty"`
}

// Parse 严格解析索引，拒绝未知字段与尾随内容；错误包含 index.json 与 JSON 路径。
func Parse(raw []byte) (*Index, error) {
	// 1. 先确认索引版本与条目数组，再按条目保留错误位置。
	var document struct {
		Version int               `json:"version"`
		Plugins []json.RawMessage `json:"plugins"`
	}
	if err := decode(raw, "$", &document); err != nil {
		return nil, err
	}
	if document.Version != indexVersion {
		return nil, fmt.Errorf("index.json:$.version: expected %d, got %d", indexVersion, document.Version)
	}
	if document.Plugins == nil {
		return nil, errors.New("index.json:$.plugins: expected an array")
	}
	index := &Index{Version: document.Version, Plugins: make([]Plugin, 0, len(document.Plugins))}
	// 2. 逐条检查仓库、角色与可选标签，返回值始终包含完整的合法索引。
	for i, raw := range document.Plugins {
		path := fmt.Sprintf("$.plugins[%d]", i)
		var entry struct {
			Repo string          `json:"repo"`
			Role string          `json:"role"`
			Tags json.RawMessage `json:"tags"`
		}
		if err := decode(raw, path, &entry); err != nil {
			return nil, err
		}
		if !repositoryName.MatchString(entry.Repo) {
			return nil, fmt.Errorf("index.json:%s.repo: expected owner/name", path)
		}
		for _, part := range strings.Split(entry.Repo, "/") {
			if part == "." || part == ".." {
				return nil, fmt.Errorf("index.json:%s.repo: expected owner/name", path)
			}
		}
		switch entry.Role {
		case "plugin", "adapter", "agent", "webui":
		default:
			return nil, fmt.Errorf("index.json:%s.role: unknown role %q", path, entry.Role)
		}
		plugin := Plugin{Repo: entry.Repo, Role: entry.Role}
		if len(entry.Tags) != 0 {
			var tags []json.RawMessage
			if err := decode(entry.Tags, path+".tags", &tags); err != nil {
				return nil, err
			}
			for j, raw := range tags {
				var tag string
				if err := decode(raw, fmt.Sprintf("%s.tags[%d]", path, j), &tag); err != nil {
					return nil, err
				}
				plugin.Tags = append(plugin.Tags, tag)
			}
		}
		index.Plugins = append(index.Plugins, plugin)
	}
	return index, nil
}

func decode(raw []byte, path string, destination any) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("index.json:%s: unexpected null", path)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		var mismatch *json.UnmarshalTypeError
		if errors.As(err, &mismatch) && mismatch.Field != "" {
			path += "." + mismatch.Field
		}
		return fmt.Errorf("index.json:%s: %w", path, err)
	}
	if _, err := decoder.Token(); err == nil {
		return fmt.Errorf("index.json:%s: unexpected trailing content", path)
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("index.json:%s: unexpected trailing content: %w", path, err)
	}
	return nil
}
