package pluginkit

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
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
	project   *manifest.Result
	generated *generate.Result
	manifest  []byte
	files     map[string][]byte
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
	files, err := projectResources(options.Dir, project.Manifest)
	if err != nil {
		return nil, err
	}
	snapshot := &projectScan{project: project, manifest: raw, files: files}
	if project.Manifest.Role == "webui" {
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
		files["config_schema.json"] = result.ConfigSchema
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

// projectResources 在工程文件系统边界检查打包资源，所有角色共用这份白名单。
func projectResources(dir string, project *manifest.Manifest) (map[string][]byte, error) {
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
	} else {
		if project.Role == "adapter" {
			if _, err := os.Stat(filepath.Join(dir, "adapter_config_schema.json")); err != nil {
				return nil, fmt.Errorf("adapter_config_schema.json:1:1: %w", err)
			}
		}
		for _, name := range []string{"config_schema.json", "adapter_config_schema.json", "i18n"} {
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
	for _, name := range []string{changelogFile, changelogDirectory} {
		if _, err := os.Lstat(filepath.Join(dir, name)); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		if err := walk(name); err != nil {
			return nil, err
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
