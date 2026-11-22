package pluginkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// Check 只读检查作者工程、入口和目标平台的资源。
func Check(ctx context.Context, options generate.Options) error {
	options.CGOEnabled = "0"
	options.RequireAdapterConstructor = true
	_, err := scanProject(ctx, options)
	return err
}

// projectScan 持有同一次校验所使用的作者 manifest、资源与生成源码。
type projectScan struct {
	project        *manifest.Result
	generated      *generate.Result
	manifest       []byte
	staticManifest []byte // 静态 WebUI 的包内声明，在 Check 时完成与构建相同的校验。
	files          map[string][]byte
}

func scanProject(ctx context.Context, options generate.Options) (*projectScan, error) {
	raw, err := os.ReadFile(filepath.Join(options.Dir, projectFile))
	if err != nil {
		return nil, err
	}
	project, err := manifest.ParseProject(raw)
	if err != nil {
		return nil, fmt.Errorf("plugin.yaml:1:1: %w", err)
	}
	process := project.Manifest.Role != "webui"
	if !process {
		process, err = webUIHasSource(ctx, options)
		if err != nil {
			return nil, err
		}
	}
	files, err := projectResources(options.Dir, project.Manifest, process)
	if err != nil {
		return nil, err
	}
	snapshot := &projectScan{project: project, manifest: raw, files: files}
	if !process {
		snapshot.staticManifest, err = manifest.Generate(project.Manifest, manifest.BuildInfo{ManifestVersion: 1, OS: "any", Arch: "any"})
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}
	options.Resources = maps.Clone(files)
	options.Resources[projectFile] = raw
	result, err := generate.Scan(ctx, options)
	if err != nil {
		return nil, err
	}
	if len(result.Diagnostics) != 0 {
		diagnostics := make([]string, 0, len(result.Diagnostics))
		for _, d := range result.Diagnostics {
			diagnostics = append(diagnostics, d.Pos+": "+d.Message)
		}
		return nil, errors.New(strings.Join(diagnostics, "\n"))
	}
	if len(result.Source) == 0 {
		return nil, errors.New("entry package must declare a component constructor")
	}
	snapshot.generated = result
	// 包资源与生成代码均采用同一份规范 schema。
	if len(result.ConfigSchema) > 0 {
		files["plugin_config_schema.json"] = result.ConfigSchema
	}
	if raw, ok := files["adapter_config_schema.json"]; ok {
		schema, problems := configschema.Parse(raw)
		if len(problems) > 0 {
			return nil, fmt.Errorf("adapter schema: %v", problems)
		}
		files["adapter_config_schema.json"] = schema.Canonical()
	}
	return snapshot, nil
}

// webUIHasSource 使用目标工具链选文件；静态工程可以没有 Go module。
func webUIHasSource(ctx context.Context, options generate.Options) (bool, error) {
	cmd := targetGo(ctx, options, "list", "-e", "-find", "-json", ".")
	cmd.Env = append(cmd.Env, "GO111MODULE=off")
	raw, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("list WebUI source: %w", err)
	}
	var pkg struct {
		GoFiles, CgoFiles, InvalidGoFiles []string
		Error                             *struct{ Pos string }
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return false, fmt.Errorf("decode WebUI source list: %w", err)
	}
	// 选中的错误源码仍交给生成器诊断，不当成静态资源发布。
	return len(pkg.GoFiles)+len(pkg.CgoFiles)+len(pkg.InvalidGoFiles) != 0 || pkg.Error != nil && pkg.Error.Pos != "", nil
}

// projectResources 在工程文件系统边界检查打包资源，所有角色共用这份白名单。
func projectResources(dir string, project *manifest.Manifest, process bool) (map[string][]byte, error) {
	if _, err := os.Lstat(filepath.Join(dir, "config_schema.json")); err == nil {
		return nil, fmt.Errorf("component %s: rename config_schema.json to plugin_config_schema.json and rebuild", project.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	files := map[string][]byte{}
	names := map[string]string{}
	add := func(name string, info fs.FileInfo) error {
		if err := manifest.ValidateRelPath(name); err != nil {
			return err
		}
		if err := checkHardlink(filepath.Join(dir, filepath.FromSlash(name)), info); err != nil {
			return err
		}
		folded := strings.ToLower(name)
		if previous, ok := names[folded]; ok && previous != name {
			return fmt.Errorf("case-conflicting resource paths %q and %q", previous, name)
		}
		names[folded] = name
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("resource %q must be a regular file or directory", name)
		}
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		files[name] = raw
		return nil
	}
	walk := func(name string) error {
		if err := manifest.ValidateRelPath(name); err != nil {
			return err
		}
		for parent := name; parent != "."; parent = filepath.ToSlash(filepath.Dir(parent)) {
			info, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(parent)))
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("resource path %q contains a symbolic link", name)
			}
		}
		return filepath.Walk(filepath.Join(dir, filepath.FromSlash(name)), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			return add(filepath.ToSlash(relative), info)
		})
	}
	if project.Role == "webui" {
		if err := walk(project.Assets); err != nil {
			return nil, err
		}
		if _, ok := files[project.Assets+"/index.html"]; !ok {
			return nil, fmt.Errorf("%s/index.html is missing", project.Assets)
		}
	}
	if process {
		if project.Role == "adapter" {
			if _, err := os.Stat(filepath.Join(dir, "adapter_config_schema.json")); err != nil {
				return nil, fmt.Errorf("adapter_config_schema.json:1:1: %w", err)
			}
		}
		for _, name := range []string{"plugin_config_schema.json", "adapter_config_schema.json", "i18n"} {
			if _, err := os.Lstat(filepath.Join(dir, name)); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return nil, err
			}
			if err := walk(name); err != nil {
				return nil, err
			}
			if name == "i18n" {
				if _, ok := files["i18n/en-US.json"]; !ok {
					return nil, errors.New("i18n/en-US.json: source catalog is missing")
				}
			}
		}
	}
	if err := checkChangelog(dir, project.Version); err != nil {
		return nil, err
	}
	for _, name := range []string{"LICENSE", changelogFile, changelogDirectory} {
		if _, err := os.Lstat(filepath.Join(dir, name)); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		if err := walk(name); err != nil {
			return nil, err
		}
	}
	for _, schemaName := range []string{"plugin_config_schema.json", "adapter_config_schema.json"} {
		raw, exists := files[schemaName]
		if !exists {
			continue
		}
		schema, problems := configschema.Parse(raw)
		if len(problems) != 0 {
			return nil, fmt.Errorf("%s: %v", schemaName, problems)
		}
		for _, name := range schema.ExampleImages() {
			if err := walk(name); err != nil {
				return nil, fmt.Errorf("%s:%s: %w", schemaName, exampleImageLocation(raw, name), err)
			}
			content, exists := files[name]
			if !exists {
				return nil, fmt.Errorf("%s:%s: image %q must be a regular file", schemaName, exampleImageLocation(raw, name), name)
			}
			if expected := exampleContentType(name); http.DetectContentType(content) != expected {
				return nil, fmt.Errorf("%s:%s: image %q content does not match its extension", schemaName, exampleImageLocation(raw, name), name)
			}
		}
	}
	if project.Icon != "" {
		if err := walk(project.Icon); err != nil {
			return nil, err
		}
		if _, ok := files[project.Icon]; !ok {
			return nil, fmt.Errorf("icon %q must be a regular file", project.Icon)
		}
	}
	return files, nil
}
