package pluginkit

import (
	"context"
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"go.yaml.in/yaml/v3"
)

const modulePath = "github.com/AlonaBot/AlonaBot"
const initialVersion = "0.1.0"

// Init 在空目录创建组件工程，返回是否需要作者补齐开发版 SDK 依赖。
func Init(ctx context.Context, dir, id, role, adapterID, module string) (needsSDK bool, err error) {
	// 1. 作者字段与目录状态先于写入检查。
	dir, err = filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if len(entries) != 0 {
		return false, fmt.Errorf("directory %s is not empty", dir)
	}
	defaultID := id == ""
	if defaultID {
		id = filepath.Base(dir)
	}
	if role == "adapter" && adapterID == "" {
		adapterID = id
	}
	fields := map[string]any{"id": id, "name": id, "version": initialVersion, "source": "github.com/CHANGE-ME/CHANGE-ME", "role": role}
	if adapterID != "" {
		fields["adapter_id"] = adapterID
	}
	if role == "webui" {
		fields["assets"] = "assets"
		fields["admin_api_min_major"] = 1
		fields["admin_api_max_major"] = 1
	}
	raw, err := yaml.Marshal(fields)
	if err != nil {
		return false, err
	}
	raw = append([]byte("# Change source to your component repository before publishing.\n"), raw...)
	project, err := manifest.ParseProject(raw)
	if err != nil {
		if defaultID {
			return false, fmt.Errorf("directory name %q cannot be used as component ID; specify --id explicitly: %w", filepath.Base(dir), err)
		}
		return false, err
	}
	for _, warning := range project.Warnings {
		slog.Warn(warning)
	}
	within, err := withinModule(dir)
	if err != nil {
		return false, err
	}
	if !within && module == "" {
		module = filepath.Base(dir)
	}
	files := map[string][]byte{projectFile: raw, ".gitignore": []byte("data/\nbin/\n")}
	switch role {
	case "plugin", "agent":
		files["plugin.go"] = []byte(pluginTemplate)
		files["config_schema.json"] = []byte("{}\n")
		files["i18n/en-US.json"] = []byte("{\"messages\":{\"hello\":\"Hello!\"}}\n")
	case "adapter":
		files["adapter.go"] = []byte(fmt.Sprintf(adapterTemplate, adapterID))
		files["adapter_config_schema.json"] = []byte("{}\n")
	case "webui":
		files["assets/index.html"] = []byte("<!doctype html>\n<html lang=\"en\"><meta charset=\"utf-8\"><title>AlonaBot WebUI</title><main><h1>AlonaBot WebUI</h1></main></html>\n")
	}
	// 2. 模板与 Go module 在临时目录准备完成，再发布到空目录。
	staging, err := os.MkdirTemp("", "alonabot-init-*")
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()
	if !within {
		if _, err := goCommand(ctx, staging, "mod", "init", module); err != nil {
			return false, err
		}
		version := sdkVersion()
		if version == "(devel)" {
			needsSDK = role != "webui"
		} else if role != "webui" {
			if _, err := goCommand(ctx, staging, "mod", "edit", "-require="+modulePath+"@"+version); err != nil {
				return false, err
			}
		}
		files["go.mod"], err = os.ReadFile(filepath.Join(staging, "go.mod"))
		if err != nil {
			return false, err
		}
	}
	for name, content := range files {
		if filepath.Ext(name) == ".go" {
			content, err = format.Source(content)
			if err != nil {
				return false, err
			}
		}
		path := filepath.Join(staging, name)
		if err := os.MkdirAll(filepath.Dir(path), privateMode); err != nil {
			return false, err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return false, err
		}
	}
	if err := os.MkdirAll(dir, privateMode); err != nil {
		return false, err
	}
	// 再读一次，避免准备模板期间覆盖作者刚写入的文件。
	entries, err = os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	if len(entries) != 0 {
		return false, fmt.Errorf("directory %s is not empty", dir)
	}
	entries, err = os.ReadDir(staging)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(staging, entry.Name()), filepath.Join(dir, entry.Name())); err != nil {
			return false, err
		}
	}
	return needsSDK, nil
}

func withinModule(dir string) (bool, error) {
	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false, nil
		}
		dir = parent
	}
}

func sdkVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	if info.Main.Path == modulePath && info.Main.Version != "" {
		return info.Main.Version
	}
	for _, dependency := range info.Deps {
		if dependency.Path == modulePath {
			if dependency.Replace != nil {
				dependency = dependency.Replace
			}
			if dependency.Version != "" {
				return dependency.Version
			}
			return "(devel)"
		}
	}
	return "(devel)"
}

const pluginTemplate = `// Package plugin 提供组件的示例命令。
package plugin

import "github.com/AlonaBot/AlonaBot/sdk"

// New 创建独立插件并登记路由。
func New() *sdk.Plugin {
 p := sdk.NewPlugin()
 p.Command("hello").Description("Show a greeting").Handle(func(c *sdk.Context) error {
  return c.Reply(c.T("hello"))
 })
 return p
}
`

const adapterTemplate = `// Package component 提供 Adapter 的可导入入口。
package component

import (
 "context"
 "github.com/AlonaBot/AlonaBot/sdk/adapter"
 "github.com/AlonaBot/AlonaBot/sdk/errcode"
)

type platform struct{}

// New 创建 Adapter；平台实例由作者实现 Start 接入。
func New() adapter.Adapter {return &platform{}}
func (*platform) Descriptor() *adapter.Descriptor {return &adapter.Descriptor{AdapterID:%q}}
func (*platform) Start(context.Context,adapter.StartRequest,adapter.Core)(adapter.Instance,error){
 return nil,errcode.Unsupported.New("implement platform instance startup")
}
`
