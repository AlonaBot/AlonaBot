// Package pluginkit 构建开发组件，复用 SDK 生成器与 Go 工具链。
//
// BuildDev 通过 overlay 注入资源与薄入口，保留作者目录原样；构建成功后才替换可执行文件。
// 组件运行与热部署由调用方装配，开发运行规则见 CONTRACTS §15.4。
package pluginkit

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/generate"
)

const (
	projectFile = "plugin.yaml"
	entryFile   = "alonabot_main.go"
	entryDir    = ".alonabot-dev"
	sdkPath     = "github.com/AlonaBot/AlonaBot/sdk"
	privateMode = 0o700
	fileMode    = 0o600
)

// Build 保存开发可执行文件、工程 manifest 原文与组件身份。
type Build struct {
	Executable          string
	Manifest            []byte
	PackageManifest     []byte // 包内构建元数据，交给根包 Executable。
	Digest              []byte // 与独立进程 Hello 绑定的可执行文件摘要。
	Role, ID, AdapterID string
}

// BuildDev 构建 dir 中的组件到 outDir，失败时保留编译器原文与上一次成功的产物。
func BuildDev(ctx context.Context, dir, outDir string) (build Build, err error) {
	// 1. 身份与生成诊断先于产物发布，工程目录始终只读。
	dir, err = filepath.Abs(dir)
	if err != nil {
		return build, err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return build, err
	}
	build.Manifest, err = os.ReadFile(filepath.Join(dir, projectFile))
	if err != nil {
		return build, fmt.Errorf("read project manifest: %w", err)
	}
	project, err := manifest.ParseProject(build.Manifest)
	if err != nil {
		return build, err
	}
	build.Role, build.ID, build.AdapterID = project.Manifest.Role, project.Manifest.ID, project.Manifest.AdapterID
	if build.Role == "webui" {
		return build, errors.New("webui projects have no development executable")
	}
	result, err := generate.Scan(ctx, generate.Options{Dir: dir})
	if err != nil {
		return build, err
	}
	if len(result.Diagnostics) > 0 {
		var diagnostics []string
		for _, diagnostic := range result.Diagnostics {
			diagnostics = append(diagnostics, diagnostic.Pos+": "+diagnostic.Message)
		}
		return build, errors.New(strings.Join(diagnostics, "\n"))
	}
	// 2. 生成文件放在数据目录；overlay 的虚拟路径保持工程内的导入语义。
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return build, err
	}
	if err := os.MkdirAll(outDir, privateMode); err != nil {
		return build, err
	}
	staging, err := os.MkdirTemp(filepath.Dir(outDir), ".alonabot-build-*")
	if err != nil {
		return build, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()
	replacements := make(map[string]string)
	write := func(virtual, name string, content []byte) error {
		path := filepath.Join(staging, name)
		if err := os.WriteFile(path, content, fileMode); err != nil {
			return err
		}
		replacements[virtual] = path
		return nil
	}
	if err := write(filepath.Join(dir, generate.FileName), generate.FileName, result.Source); err != nil {
		return build, err
	}
	overlayPath := filepath.Join(staging, "overlay.json")
	writeOverlay := func() error {
		overlay, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			return err
		}
		return os.WriteFile(overlayPath, overlay, fileMode)
	}
	if err := writeOverlay(); err != nil {
		return build, err
	}
	var entry struct {
		Name, ImportPath string
		GoFiles          []string
		Module           struct{ GoMod string }
	}
	listed, err := goCommand(ctx, dir, "list", "-overlay", overlayPath, "-json", ".")
	if err != nil {
		return build, err
	}
	if err := json.Unmarshal(listed, &entry); err != nil {
		return build, fmt.Errorf("decode entry package: %w", err)
	}
	// 薄入口额外导入 SDK；允许 Go 在临时 module 文件中补齐其间接依赖。
	modFile := filepath.Join(staging, "project.mod")
	for _, pair := range [][2]string{{entry.Module.GoMod, modFile}, {strings.TrimSuffix(entry.Module.GoMod, ".mod") + ".sum", filepath.Join(staging, "project.sum")}} {
		content, err := os.ReadFile(pair[0])
		if errors.Is(err, os.ErrNotExist) && strings.HasSuffix(pair[0], ".sum") {
			continue
		}
		if err != nil {
			return build, err
		}
		if err := os.WriteFile(pair[1], content, fileMode); err != nil {
			return build, err
		}
	}
	mode, effects, err := migrationMode(ctx, dir, overlayPath, modFile)
	if err != nil {
		return build, err
	}
	target := "."
	if entry.Name != "main" {
		call := "Run"
		if build.Role == "adapter" {
			call = "RunAdapter"
		}
		source := fmt.Sprintf("package main\nimport (\"log/slog\"; \"os\"; sdk %q; component %q)\nfunc main() { if err := sdk.%s(component.New()); err != nil { slog.Error(\"component stopped\", \"error\", err); os.Exit(1) } }\n", sdkPath, entry.ImportPath, call)
		if err := write(filepath.Join(dir, entryDir, entryFile), entryFile, []byte(source)); err != nil {
			return build, err
		}
		target = "./" + entryDir
	} else {
		hasMain, err := declaresMain(dir, entry.GoFiles)
		if err != nil {
			return build, err
		}
		if !hasMain {
			source := fmt.Sprintf("package main\nimport (\"log/slog\"; \"os\"; sdk %q)\nfunc main() { if err := sdk.Run(New()); err != nil { slog.Error(\"component stopped\", \"error\", err); os.Exit(1) } }\n", sdkPath)
			if err := write(filepath.Join(dir, entryFile), entryFile, []byte(source)); err != nil {
				return build, err
			}
		}
	}
	if err := writeOverlay(); err != nil {
		return build, err
	}

	// 3. 链接到临时路径，避免编译失败或运行中的旧二进制受到写入影响。
	name := build.ID
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	version := uint32(1)
	if len(result.ConfigSchema) != 0 {
		schema, _ := configschema.Parse(result.ConfigSchema)
		version = uint32(schema.Version())
	}
	major := uint32(validate.ProtocolMajor)
	capabilities := []manifest.Capability{}
	build.PackageManifest, err = manifest.Generate(project.Manifest, manifest.BuildInfo{
		ManifestVersion: 1, ProtocolMinMajor: &major, ProtocolMaxMajor: &major,
		RequiredCapabilities: &capabilities, OS: runtime.GOOS, Arch: runtime.GOARCH,
		Entrypoint: name, ConfigVersion: &version, MigrationMode: mode, PrepareEffects: effects,
	})
	if err != nil {
		return build, err
	}
	resources := struct {
		Manifest, ConfigSchema, AdapterConfigSchema []byte
		I18n                                        map[string][]byte
	}{Manifest: build.PackageManifest, ConfigSchema: result.ConfigSchema, I18n: map[string][]byte{}}
	resources.AdapterConfigSchema, err = os.ReadFile(filepath.Join(dir, "adapter_config_schema.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return build, err
	}
	catalogs, err := filepath.Glob(filepath.Join(dir, "i18n", "*.json"))
	if err != nil {
		return build, err
	}
	for _, path := range catalogs {
		raw, err := os.ReadFile(path)
		if err != nil {
			return build, err
		}
		resources.I18n[filepath.Base(path)] = raw
	}
	metadata, err := json.Marshal(resources)
	if err != nil {
		return build, err
	}
	metadataPath := filepath.Join(staging, "resources.json")
	if err := os.WriteFile(metadataPath, metadata, fileMode); err != nil {
		return build, err
	}
	executable := filepath.Join(staging, name)
	if _, err := goCommand(ctx, dir, "build", "-modfile", modFile, "-mod=mod", "-overlay", overlayPath, "-o", executable, target); err != nil {
		return build, err
	}
	build.Executable = filepath.Join(outDir, name)
	content, err := os.ReadFile(executable)
	if err != nil {
		return build, err
	}
	digest := sha512.Sum512_256(content)
	build.Digest = digest[:]
	// 资源按摘要保存，旧进程使用的资源在新构建发布期间仍然可读。
	if err := os.Rename(metadataPath, fmt.Sprintf("%s.%x.json", build.Executable, build.Digest)); err != nil {
		return build, err
	}
	if err := os.Rename(executable, build.Executable); err != nil {
		return build, fmt.Errorf("publish development executable: %w", err)
	}
	return build, nil
}

// migrationMode 检查参与构建的本 module 中的 SDK 函数引用，包含别名和传参。
func migrationMode(ctx context.Context, dir, overlay, modFile string) (string, string, error) {
	listed, err := goCommand(ctx, dir, "list", "-modfile", modFile, "-mod=mod", "-overlay", overlay, "-deps", "-json", ".")
	if err != nil {
		return "", "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(listed))
	mode, effects := manifest.MigrationNone, manifest.EffectsNone
	var self, versioned bool
	for {
		var pkg struct {
			Dir     string
			GoFiles []string
			Module  *struct{ Main bool }
		}
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return "", "", err
		}
		if pkg.Module == nil || !pkg.Module.Main {
			continue
		}
		for _, name := range pkg.GoFiles {
			if name == generate.FileName {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(pkg.Dir, name), nil, 0)
			if err != nil {
				return "", "", err
			}
			aliases := map[string]bool{}
			for _, imported := range file.Imports {
				path, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return "", "", err
				}
				if path == sdkPath+"/sqlitedb" || path == sdkPath+"/gormdb" {
					alias := filepath.Base(path)
					if imported.Name != nil {
						alias = imported.Name.Name
					}
					aliases[alias] = true
				}
			}
			ast.Inspect(file, func(node ast.Node) bool {
				var name string
				switch fun := node.(type) {
				case *ast.SelectorExpr:
					if owner, ok := fun.X.(*ast.Ident); ok && aliases[owner.Name] {
						name = fun.Sel.Name
					}
				case *ast.Ident:
					if aliases["."] {
						name = fun.Name
					}
				}
				switch name {
				case "Migrate":
					self = true
					mode, effects = manifest.MigrationSelf, manifest.EffectsManagedDB
				case "Migrations":
					versioned = true
					mode, effects = manifest.MigrationVersioned, manifest.EffectsManagedDB
				case "Open":
					effects = manifest.EffectsManagedDB
				}
				return true
			})
		}
	}
	if self && versioned {
		return "", "", errors.New("Migrate and Migrations registrations are mutually exclusive")
	}
	return mode, effects, nil
}

func goCommand(ctx context.Context, dir string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = dir
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return output, fmt.Errorf("go %s: %w\n%s%s", strings.Join(args, " "), err, output, stderr.String())
	}
	return output, nil
}

func declaresMain(dir string, files []string) (bool, error) {
	for _, name := range files {
		if name == generate.FileName {
			continue // 资源登记由 overlay 提供，入口只检查作者源码。
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, 0)
		if err != nil {
			return false, err
		}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && function.Name.Name == "main" {
				return true, nil
			}
		}
	}
	return false, nil
}
