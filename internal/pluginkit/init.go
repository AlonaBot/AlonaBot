package pluginkit

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"go.yaml.in/yaml/v3"
)

const modulePath = "github.com/AlonaBot/AlonaBot"
const initialVersion = "0.1.0"

// Init 在空目录创建组件工程，返回是否需要作者补齐开发版 SDK 依赖。
func Init(ctx context.Context, dir, id, role, adapterID, module, version string) (needsSDK bool, err error) {
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
	files, err := templateFiles(role, templateData{ID: id, AdapterID: adapterID})
	if err != nil {
		return false, err
	}
	files[projectFile] = raw
	// 2. 在目标目录内暂存模板与 Go module，确保发布时 Rename 位于同一文件系统。
	if err := os.MkdirAll(dir, privateMode); err != nil {
		return false, err
	}
	staging, err := os.MkdirTemp(dir, ".alonabot-init-*")
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()
	tidySDK := false
	if !within {
		if _, err := goCommand(ctx, generate.Toolchain{}, staging, "mod", "init", module); err != nil {
			return false, err
		}
		version := sdkVersion(version)
		if version == "(devel)" {
			needsSDK = role != "webui"
		} else if role != "webui" {
			if _, err := goCommand(ctx, generate.Toolchain{}, staging, "mod", "edit", "-require="+modulePath+"@"+version); err != nil {
				return false, err
			}
			tidySDK = true
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
	// 再读一次，只允许本次暂存目录，避免覆盖准备模板期间作者刚写入的文件。
	entries, err = os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(staging) {
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
	// 3. 在已发布的工程中解析 SDK 依赖；失败时保留作者可修复的全部文件。
	if tidySDK {
		if _, err := goCommand(ctx, generate.Toolchain{}, dir, "mod", "tidy"); err != nil {
			_, outputErr := fmt.Fprintf(os.Stderr, "Project files were kept at %s; run: go -C %q mod tidy\n", dir, dir)
			return false, errors.Join(err, outputErr)
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

// pseudoVersion 匹配 Go 伪版本末尾的 14 位时间戳与 12 位提交哈希。
var pseudoVersion = regexp.MustCompile(`[-.]\d{14}-[0-9a-f]{12}$`)

func sdkVersion(version string) string {
	// Go 会给仓库 checkout 的构建盖上伪版本；它不是发布的版本，按开发版处理。
	if version == "" || version == "(devel)" || strings.HasSuffix(version, "+dirty") || pseudoVersion.MatchString(version) {
		return "(devel)"
	}
	normalized := strings.TrimPrefix(version, "v")
	if !manifest.ValidVersion(normalized) {
		slog.Warn("Invalid CLI version; treating this as a development CLI", "version", version)
		return "(devel)"
	}
	return "v" + normalized
}

// templates 是 alonabot init 的工程骨架，文件规则见 templates/README.md。
//
//go:embed all:templates/common templates/plugin templates/adapter templates/webui
var templates embed.FS

// templateData 是 .tmpl 模板可用的字段。
type templateData struct {
	ID        string
	AdapterID string
}

// templateFiles 返回角色对应的模板文件，键为工程内的相对路径。
func templateFiles(role string, data templateData) (map[string][]byte, error) {
	dirs := []string{"templates/common"}
	switch role {
	case "plugin", "agent":
		dirs = append(dirs, "templates/plugin")
	case "adapter", "webui":
		dirs = append(dirs, "templates/"+role)
	}
	files := map[string][]byte{}
	for _, root := range dirs {
		err := fs.WalkDir(templates, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			content, err := templates.ReadFile(name)
			if err != nil {
				return err
			}
			relative := strings.TrimPrefix(name, root+"/")
			dir, base := path.Split(relative)
			if rest, ok := strings.CutPrefix(base, "_"); ok {
				base = "." + rest
			}
			if rest, ok := strings.CutSuffix(base, ".tmpl"); ok {
				base = rest
				parsed, err := template.New(name).Option("missingkey=error").Parse(string(content))
				if err != nil {
					return err
				}
				var rendered bytes.Buffer
				if err := parsed.Execute(&rendered, data); err != nil {
					return err
				}
				content = rendered.Bytes()
			}
			files[dir+base] = content
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}
