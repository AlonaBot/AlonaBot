// Package pluginkit 创建、检查、构建和打包组件工程，复用 SDK 生成器与 Go 工具链。
//
// BuildDev 通过 overlay 注入资源与薄入口，保留作者目录原样；构建成功后才替换可执行文件。
// 组件运行与热部署由调用方装配。
//
// 包的组成：[Init] 创建作者工程，[Check] 只读检查，[Compile] 发布正式构建；
// [BuildDev] 提供只读开发构建，[Pack] 输出归档与发布清单。
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

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

const (
	projectFile      = "plugin.yaml"
	registrationFile = "alonabot_resources.go"
	entryFile        = "alonabot_main.go"
	entryDir         = ".alonabot-dev"
	sdkPath          = "github.com/AlonaBot/AlonaBot/sdk"
	privateMode      = 0o700
	fileMode         = 0o600
)

// Build 保存可执行文件、工程 manifest 原文与组件身份。
type Build struct {
	Executable          string
	Manifest            []byte
	PackageManifest     []byte // 包内构建元数据，交给根包 Executable。
	Digest              []byte // 与独立进程 Hello 绑定的可执行文件摘要。
	Role, ID, AdapterID string
}

// BuildDev 构建 dir 中的组件到 outDir，失败时保留编译器原文与上一次成功的产物。
func BuildDev(ctx context.Context, dir, outDir string) (Build, error) {
	return buildComponent(ctx, generate.Options{Dir: dir}, outDir, true, nil)
}

// Compile 生成目标资源并发布独立组件与包内 manifest；outDir 为空时使用工程 bin/。
func Compile(ctx context.Context, options generate.Options, outDir string) (Build, error) {
	return buildComponent(ctx, options, outDir, false, nil)
}

func buildComponent(ctx context.Context, options generate.Options, outDir string, development bool, snapshot *projectScan) (build Build, err error) {
	dir := options.Dir
	// 1. 身份与生成诊断先于产物发布；开发构建保持工程目录只读。
	dir, err = filepath.Abs(dir)
	if err != nil {
		return build, err
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return build, err
	}
	options.Dir = dir
	if development {
		options.GOOS = runtime.GOOS
		options.GOARCH = runtime.GOARCH
	} else {
		options.CGOEnabled = "0"
		options.RequireAdapterConstructor = true
	}
	targetEnv, err := goCommand(ctx, dir, "env", "GOOS", "GOARCH")
	if err != nil {
		return build, err
	}
	targetParts := strings.Fields(string(targetEnv))
	if options.GOOS == "" {
		options.GOOS = targetParts[0]
	}
	if options.GOARCH == "" {
		options.GOARCH = targetParts[1]
	}
	if snapshot == nil {
		snapshot, err = scanProject(ctx, options)
		if err != nil {
			return build, err
		}
	}
	project, result := snapshot.project, snapshot.generated
	build.Manifest = snapshot.manifest
	build.Role, build.ID, build.AdapterID = project.Manifest.Role, project.Manifest.ID, project.Manifest.AdapterID
	if outDir == "" {
		outDir = filepath.Join(dir, "bin")
	}
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return build, err
	}
	outputProject := filepath.Join(outDir, projectFile)
	if info, statErr := os.Stat(outputProject); statErr == nil {
		author, statErr := os.Stat(filepath.Join(dir, projectFile))
		if statErr != nil {
			return build, statErr
		}
		if !development && os.SameFile(info, author) {
			return build, errors.New("output directory would overwrite the author manifest")
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return build, statErr
	}
	if build.Role == "webui" && result == nil {
		if development {
			return build, errors.New("webui projects have no development executable")
		}
		build.PackageManifest = snapshot.staticManifest
		if err = os.MkdirAll(outDir, privateMode); err != nil {
			return build, err
		}
		if err = writeProjectResources(outDir, snapshot.files); err != nil {
			return build, err
		}
		err = os.WriteFile(filepath.Join(outDir, projectFile), build.PackageManifest, fileMode)
		return build, err
	}
	runGo := func(args ...string) ([]byte, error) { return targetGoCommand(ctx, options, development, args...) }
	// 2. 生成文件放在数据目录；overlay 的虚拟路径保持工程内的导入语义。
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
	if err := write(filepath.Join(dir, registrationFile), registrationFile, result.Source); err != nil {
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
	goModRaw, err := goCommand(ctx, dir, "env", "GOMOD")
	if err != nil {
		return build, err
	}
	goMod := strings.TrimSpace(string(goModRaw))
	// 薄入口额外导入 SDK；允许 Go 在临时 module 文件中补齐其间接依赖。
	modFile := filepath.Join(staging, "project.mod")
	for _, pair := range [][2]string{{goMod, modFile}, {strings.TrimSuffix(goMod, ".mod") + ".sum", filepath.Join(staging, "project.sum")}} {
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
	var entry struct {
		Name, ImportPath string
		GoFiles          []string
		Module           struct{ GoMod string }
	}
	listed, err := runGo("list", "-modfile", modFile, "-mod=mod", "-overlay", overlayPath, "-json", ".")
	if err != nil {
		return build, err
	}
	if err := json.Unmarshal(listed, &entry); err != nil {
		return build, fmt.Errorf("decode entry package: %w", err)
	}
	mode, effects, err := migrationMode(ctx, options, overlayPath, modFile, development)
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
			call := "Run"
			if build.Role == "adapter" {
				call = "RunAdapter"
			}
			source := fmt.Sprintf("package main\nimport (\"log/slog\"; \"os\"; sdk %q)\nfunc main() { if err := sdk.%s(New()); err != nil { slog.Error(\"component stopped\", \"error\", err); os.Exit(1) } }\n", sdkPath, call)
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
	if !development {
		name = "component"
	}
	if options.GOOS == "windows" {
		name += ".exe"
	}
	major := uint32(validate.ProtocolMajor)
	capabilities, err := requiredCapabilities(ctx, options, overlayPath, modFile, development, result.HTTP)
	if err != nil {
		return build, err
	}
	build.PackageManifest, err = manifest.Generate(project.Manifest, manifest.BuildInfo{
		ManifestVersion: 1, ProtocolMinMajor: &major, ProtocolMaxMajor: &major,
		RequiredCapabilities: &capabilities, OS: options.GOOS, Arch: options.GOARCH,
		Entrypoint: name, MigrationMode: mode, PrepareEffects: effects,
	})
	if err != nil {
		return build, err
	}
	resources := Resources{Manifest: build.PackageManifest, ConfigSchema: result.ConfigSchema, I18n: map[string][]byte{}}
	resources.AdapterConfigSchema = snapshot.files["adapter_config_schema.json"]
	for name, raw := range snapshot.files {
		if strings.HasPrefix(name, "i18n/") && filepath.Dir(name) == "i18n" && strings.HasSuffix(name, ".json") {
			resources.I18n[filepath.Base(name)] = raw
		}
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
	if _, err := runGo("build", "-trimpath", "-buildvcs=false", "-modfile", modFile, "-mod=mod", "-overlay", overlayPath, "-o", executable, target); err != nil {
		return build, err
	}
	if !development {
		if build.Role == "webui" {
			if err := writeProjectResources(outDir, snapshot.files); err != nil {
				return build, err
			}
		}
		if err := os.WriteFile(filepath.Join(outDir, projectFile), build.PackageManifest, fileMode); err != nil {
			return build, err
		}
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

func writeProjectResources(dir string, files map[string][]byte) error {
	for name, raw := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), privateMode); err != nil {
			return err
		}
		if err := os.WriteFile(path, raw, fileMode); err != nil {
			return err
		}
	}
	return nil
}

// migrationMode 检查参与构建的本 module 中的 SDK 函数引用，包含别名和传参。
func migrationMode(ctx context.Context, options generate.Options, overlay, modFile string, development bool) (string, string, error) {
	listed, err := targetGoCommand(ctx, options, development, "list", "-modfile", modFile, "-mod=mod", "-overlay", overlay, "-deps", "-json", ".")
	if err != nil {
		return "", "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(listed))
	mode, effects := manifest.MigrationNone, manifest.EffectsNone
	var self, versioned bool
	for {
		var pkg struct {
			ImportPath string
			Dir        string
			GoFiles    []string
			Module     *struct{ Main bool }
		}
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return "", "", err
		}
		if !development && (pkg.ImportPath == modulePath || strings.HasPrefix(pkg.ImportPath, modulePath+"/internal/core/")) {
			return "", "", fmt.Errorf("independent component imports Core package %q", pkg.ImportPath)
		}
		if pkg.ImportPath == sdkPath+"/kv" {
			effects = manifest.EffectsManagedDB
		}
		if pkg.Module == nil || !pkg.Module.Main {
			continue
		}
		for _, name := range pkg.GoFiles {
			if name == registrationFile {
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
		if name == registrationFile {
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

func targetGoCommand(ctx context.Context, options generate.Options, development bool, args ...string) ([]byte, error) {
	if options.CGOEnabled == "" && !development {
		options.CGOEnabled = "0"
	}
	command := targetGo(ctx, options, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return output, fmt.Errorf("go %s: %w\n%s%s", strings.Join(args, " "), err, output, stderr.String())
	}
	return output, nil
}

func targetGo(ctx context.Context, options generate.Options, args ...string) *exec.Cmd {
	if len(options.Tags) > 0 {
		args = append([]string{args[0], "-tags", strings.Join(options.Tags, ",")}, args[1:]...)
	}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = options.Dir
	command.Env = command.Environ()
	if options.GOOS != "" {
		command.Env = append(command.Env, "GOOS="+options.GOOS)
	}
	if options.GOARCH != "" {
		command.Env = append(command.Env, "GOARCH="+options.GOARCH)
	}
	if options.CGOEnabled != "" {
		command.Env = append(command.Env, "CGO_ENABLED="+options.CGOEnabled)
	}
	return command
}
