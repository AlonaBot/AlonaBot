// Package generate 按组件入口包扫描所属 module 的源码，生成资源登记代码。
//
// 命令、文本匹配与事件登记在插件对象上。
// 入口包按调用方 module 的构建列表解析，支持模块缓存中的第三方组件。
// 扫描遵守 GOOS/GOARCH 和 build tags，仅收集入口所属 module，排除 vendor 与测试文件。
// 配置从 plugin_config_schema.json 检查并嵌入，翻译覆盖按声明核对。
// 工程 plugin.yaml 声明 Adapter 时按入口包登记身份与实例 Schema；旧实例声明目录仍只检查。
// JSON Service 类型经 go/types 检查生成 Schema。
// 构建入口的 i18n/*.json 按声明检查并将原文嵌入生成代码。
// commands 仅检查结构，覆盖路径与实际注册命令的对应由 SDK 在 Run 收集注册时检查。
//
// 包的组成：[Scan] 接收 [Options]，返回描述、源码与 [Diagnostic] 汇总的 [Result]。
package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/generated"
)

const sdkPath = "github.com/AlonaBot/AlonaBot/sdk"

// Options 指定构建入口和目标环境。
type Options struct {
	Dir                       string // module 构建上下文目录
	Package                   string // 入口包导入路径；为空时扫描 Dir。
	GOOS, GOARCH              string // 空值沿用 go 命令的当前环境
	Tags                      []string
	Resources                 map[string][]byte // 可选的工程资源快照，键为相对路径；nil 时读取工程文件。
	CGOEnabled                string            // 空值沿用 Go 环境；正式构建传 "0"，与编译视图一致。
	RequireAdapterConstructor bool              // 正式工程要求 New 返回 adapter.Adapter；开发构建保留原入口约定。
	Toolchain                 Toolchain         // 零值使用 PATH 中的 go 与当前进程环境。
}

// Toolchain 指定运行 go 命令的可执行文件、构建环境与日志输出。
type Toolchain struct {
	Go            string        // go 可执行文件路径；空串从 PATH 查找。
	Env           []string      // 非空时仅继承 PATH 和代理变量，并启用受管构建监控。
	Stderr        io.Writer     // 非空时同时接收 go 命令的标准错误与受管构建提示。
	StallInterval time.Duration // 受管 Go 命令的无输出提示间隔。
}

// Command 构造在 dir 中运行的 go 命令，标准错误写入 stderr 并按需复制到 Stderr。
// 开发命令沿用进程环境，并将 PWD 与 dir 对齐。
func (t Toolchain) Command(ctx context.Context, dir string, stderr *bytes.Buffer, args ...string) *exec.Cmd {
	name := t.Go
	if name == "" {
		name = "go"
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(t.Env) == 0 {
		cmd.Env = cmd.Environ()
	} else {
		for _, name := range []string{"PATH", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
			if value, exists := os.LookupEnv(name); exists {
				cmd.Env = append(cmd.Env, name+"="+value)
			}
		}
		cmd.Env = append(cmd.Env, t.Env...)
	}
	cmd.Stderr = stderr
	if t.Stderr != nil {
		cmd.Stderr = io.MultiWriter(stderr, t.Stderr)
	}
	return cmd
}

// Diagnostic 携带一条源码或组件资源问题的位置和说明。
type Diagnostic struct {
	Pos     string // 相对 Dir 的文件:行:列；JSON 资源使用文件:JSON 路径。
	Message string
}

// Result 保存扫描描述、生成源码和全部构建诊断。
type Result struct {
	PluginConstructor string              // 插件的导入路径.New；Adapter 或只有实例声明时为空。
	AdapterType       string              // Adapter 的导入路径.具体类型名；插件或只有实例声明时为空。
	HTTP              bool                // 入口工程在所选目标下登记 HTTP 路由；构建据此向包内 required_capabilities 写入 http_listener。
	ConfigSchema      []byte              // plugin_config_schema.json 的规范 JSON；没有该文件时为 nil
	Source            []byte              // 完整 gofmt 源码；有诊断或只有实例声明时为 nil。
	Generated         generated.Resources // 本次扫描取得的全部只读资源。
	Diagnostics       []Diagnostic
}

type listedPackage struct {
	Dir, ImportPath, Name string
	GoFiles               []string
	CgoFiles              []string // 仅在 GoFiles 为空时定位缺根诊断
	CompiledGoFiles       []string // -compiled 选出的源码，包含 cgo 生成文件。
	Imports               []string
	Export                string
	Module                *struct {
		Main bool
		Dir  string
		Path string
	}
}

type declaration struct {
	name string
	pos  token.Pos
	pkg  string
}

type scan struct {
	resources                 map[string][]byte
	result                    Result
	files                     *token.FileSet
	dir                       string
	entry                     listedPackage
	constructors              []token.Pos
	adapterProject            bool
	requireAdapterConstructor bool
	adapterRoot               declaration
	projectManifest           []byte
	instanceJSON              []byte
	paramNames                map[string][]string
	serviceSchemas            map[string][]byte
	i18nFiles                 map[string][]byte // 已读文案原文；全量诊断通过后才嵌入，没有资源时为 nil。
	syntax                    map[string][]*ast.File
}

// Scan 按目标环境扫描源码与文案资源并生成嵌入代码。输入问题记入 Diagnostics；
// 工具链或 IO 失败返回 error，go list 失败时保留 stderr。
// 有实例声明且没有插件 New 时仅检查实例声明，返回的 PluginConstructor 为空、Source 为 nil。
func Scan(ctx context.Context, opt Options) (*Result, error) {
	// 1. 由调用方的 module 构建列表定位入口，再读取该入口的资源。
	dir, err := filepath.Abs(opt.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin directory: %w", err)
	}
	// macOS 的临时目录可经 /var 链接到 /private/var，统一后与 go list 的 Dir 比较。
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin directory links: %w", err)
	}
	s := &scan{files: token.NewFileSet(), dir: dir, resources: opt.Resources, requireAdapterConstructor: opt.RequireAdapterConstructor, syntax: make(map[string][]*ast.File)}
	if opt.Package == "" {
		hasSource, err := hasBuildSource(dir)
		if err != nil {
			return nil, err
		}
		if !hasSource {
			if err := s.readProject(); err != nil {
				return nil, err
			}
			hasSchema, err := s.instanceSchema()
			if err != nil {
				return nil, err
			}
			if hasSchema || len(s.result.Diagnostics) != 0 {
				if s.adapterProject {
					s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{projectFileName + ":1:1", "expected exactly one concrete Adapter type, found 0"})
				}
				return &s.result, nil
			}
		}
	}
	args := []string{"list", "-deps", "-json"}
	if len(opt.Tags) > 0 {
		args = append(args, "-tags", strings.Join(opt.Tags, ","))
	}
	target := opt.Package
	if target == "" {
		target = "."
	}
	args = append(args, target)
	var stderr bytes.Buffer
	cmd := opt.Toolchain.Command(ctx, dir, &stderr, args...)
	if opt.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+opt.GOOS)
	}
	if opt.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+opt.GOARCH)
	}
	if opt.CGOEnabled != "" {
		cmd.Env = append(cmd.Env, "CGO_ENABLED="+opt.CGOEnabled)
	}
	output, err := opt.Toolchain.Output(ctx, cmd, &stderr)
	if err != nil {
		return nil, fmt.Errorf("go list: %w\n%s", err, stderr.String())
	}
	var packages []listedPackage
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg listedPackage
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		packages = append(packages, pkg)
	}
	for _, pkg := range packages {
		if (opt.Package != "" && pkg.ImportPath == opt.Package) || (opt.Package == "" && filepath.Clean(pkg.Dir) == dir) {
			s.entry = pkg
		}
	}
	if s.entry.ImportPath == "" {
		return nil, fmt.Errorf("go list did not return an entry package in %s", dir)
	}
	s.dir = s.entry.Dir
	dir = s.dir
	if err := s.readProject(); err != nil {
		return nil, err
	}
	if len(s.result.Diagnostics) != 0 {
		return &s.result, nil
	}
	instanceSchema, err := s.instanceSchema()
	if err != nil {
		return nil, err
	}
	if s.adapterProject && !instanceSchema {
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{instanceSchemaFileName + ":1:1", "adapter project requires an instance schema"})
		return &s.result, nil
	}
	if instanceSchema {
		hasSource, err := hasBuildSource(dir)
		if err != nil {
			return nil, err
		}
		if !hasSource {
			if s.adapterProject {
				s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{projectFileName + ":1:1", "expected exactly one concrete Adapter type, found 0"})
			}
			return &s.result, nil
		}
	}
	entryModule := s.entry.Module.Path
	packages = slices.DeleteFunc(packages, func(pkg listedPackage) bool {
		return pkg.Module == nil || pkg.Module.Path != entryModule || strings.Contains("/"+pkg.ImportPath+"/", "/vendor/")
	})
	slices.SortFunc(packages, func(a, b listedPackage) int { return strings.Compare(a.ImportPath, b.ImportPath) })
	// 2. 解析源码、聚合类型声明；每个文件的问题独立保留。
	for _, pkg := range packages {
		slices.Sort(pkg.GoFiles)
		for _, name := range pkg.GoFiles {
			if err := s.readFile(pkg, filepath.Join(pkg.Dir, name)); err != nil {
				return nil, err
			}
		}
	}
	if instanceSchema && len(s.constructors) == 0 && !s.adapterProject {
		return &s.result, nil
	}
	if !s.adapterProject {
		s.checkTypes()
	}

	if err := s.readI18n(); err != nil {
		return nil, err
	}
	if err := s.configSchema(); err != nil {
		return nil, err
	}
	if len(s.result.Diagnostics) == 0 {
		if err := s.schemas(ctx, opt.Toolchain, cmd); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(s.result.Diagnostics, func(a, b Diagnostic) int {
		if n := strings.Compare(a.Pos, b.Pos); n != 0 {
			return n
		}
		return strings.Compare(a.Message, b.Message)
	})
	// 3. 诊断阻止生成；成功时统一格式化完整嵌入源码。
	if len(s.result.Diagnostics) == 0 {
		s.result.Generated = generated.Resources{Manifest: s.projectManifest, ConfigSchema: s.result.ConfigSchema, AdapterConfigSchema: s.instanceJSON, ServiceSchemas: s.serviceSchemas, I18n: s.i18nFiles, ParamNames: s.paramNames}
		s.result.Source, err = s.source()
		if err != nil {
			return nil, err
		}
	}
	return &s.result, nil
}

func (s *scan) position(pos token.Pos) string {
	p := s.files.PositionFor(pos, false)
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
func (s *scan) diagnostic(pos token.Pos, message string) {
	s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{s.position(pos), message})
}

func (s *scan) parseFile(path string) (*ast.File, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	relative, err := filepath.Rel(s.dir, path)
	if err != nil {
		return nil, fmt.Errorf("resolve source path: %w", err)
	}
	file, err := parser.ParseFile(s.files, filepath.ToSlash(relative), source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		if list, ok := err.(scanner.ErrorList); ok {
			for _, problem := range list {
				s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{fmt.Sprintf("%s:%d:%d", problem.Pos.Filename, problem.Pos.Line, problem.Pos.Column), problem.Msg})
			}
		} else {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	}
	return file, nil
}

func (s *scan) readFile(pkg listedPackage, path string) error {
	// 1. 在文件系统边界读取文件，解析器统一使用相对入口目录的文件名。
	file, err := s.parseFile(path)
	if err != nil {
		return err
	}
	if file == nil {
		return nil
	}
	s.syntax[pkg.ImportPath] = append(s.syntax[pkg.ImportPath], file)
	// 只识别入口包自己声明的 New；完整签名由 go/types 核对。
	if pkg.ImportPath == s.entry.ImportPath {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "New" {
				s.constructors = append(s.constructors, fn.Name.Pos())
			}
		}
	}
	return nil
}

func (s *scan) checkTypes() {
	if len(s.constructors) != 1 {
		files := s.entry.GoFiles
		if len(files) == 0 {
			files = s.entry.CgoFiles
		}
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{filepath.ToSlash(files[0]) + ":1:1", "entry package must declare func New() *sdk.Plugin"})
		return
	}
	s.result.PluginConstructor = s.entry.ImportPath + ".New"

}

// resourceFile 优先读取调用方提供的资源快照，缺失项按文件不存在处理。
func (s *scan) resourceFile(name string) ([]byte, error) {
	if s.resources == nil {
		return os.ReadFile(filepath.Join(s.dir, name))
	}
	if raw, ok := s.resources[name]; ok {
		return raw, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}
