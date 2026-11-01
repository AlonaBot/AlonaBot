// Package generate 扫描插件构建入口及其本 module 依赖，生成静态路由注册代码。
//
// 紧贴包级函数声明的注释行支持以下标记：
//
//	//alonabot:command <路径 token...> [alias=<token[/token...]>]... [priority=<int32>]
//	//alonabot:event <route_id> <event_type> [priority=<int32>]
//
// alias 可以重复，使用 / 分隔路径 token。方法、类型和变量上的标记产生诊断；
// 实例方法通过 Commands/Events 显式注册。生成器用嵌套 Group 注册多 token 命令。
// 扫描遵守 GOOS/GOARCH 和 build tags，排除外部依赖、vendor 与测试文件。
// 注册规则见 CONTRACTS §9、§2.2。
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
	"strconv"
	"strings"
)

const sdkPath = "github.com/AlonaBot/AlonaBot/sdk"
const configPath = sdkPath + "/config"

// FileName 指定在构建入口包目录中写入的生成文件名。
const FileName = "alonabot_gen.go"

// Options 指定构建入口和目标环境。
type Options struct {
	Dir          string // 构建入口包目录
	GOOS, GOARCH string // 空值沿用 go 命令的当前环境
	Tags         []string
}

// Diagnostic 携带一条源码问题的位置和说明。
type Diagnostic struct {
	Pos     string // 相对 Dir 的文件:行:列
	Message string
}

// Command 描述带标记的命令函数与完整路径。
type Command struct {
	Path          []string
	Aliases       [][]string
	Priority      int32
	Package, Func string
}

// Event 描述带标记的事件函数。
type Event struct {
	RouteID, EventType string
	Priority           int32
	Package, Func      string
}

// Result 保存扫描描述、生成源码和全部源码诊断。
type Result struct {
	RootType    string    // 导入路径.类型名
	ConfigType  string    // 没有 Config 时为空
	Commands    []Command // 按规范路径排序
	Events      []Event   // 按 route_id 排序
	Source      []byte    // 完整 gofmt 源码；有诊断时为 nil
	Diagnostics []Diagnostic
}

type listedPackage struct {
	Dir, ImportPath, Name string
	GoFiles               []string
	CgoFiles              []string // 仅在 GoFiles 为空时定位缺根诊断
	Module                *struct{ Main bool }
}

type declaration struct {
	name string
	pos  token.Pos
	pkg  string
}

type scan struct {
	result         Result
	files          *token.FileSet
	dir            string
	entry          listedPackage
	roots, configs []declaration
	paths, routes  map[string]string // 键对应首次声明的位置
}

// Scan 按目标环境扫描源码并生成注册代码。源码问题记入 Diagnostics；
// 工具链或 IO 失败返回 error，go list 失败时保留 stderr。
func Scan(ctx context.Context, opt Options) (*Result, error) {
	// 1. 由 Go 工具链确定依赖与实际参与构建的文件。
	dir, err := filepath.Abs(opt.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin directory: %w", err)
	}
	// macOS 的临时目录可经 /var 链接到 /private/var，统一后与 go list 的 Dir 比较。
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve plugin directory links: %w", err)
	}
	// 生成文件中的旧 import 不参与依赖闭包，文件本身保持原样。
	overlay, err := os.CreateTemp("", "alonabot-scan-*.json")
	if err != nil {
		return nil, fmt.Errorf("create scan overlay: %w", err)
	}
	defer os.Remove(overlay.Name())
	encodeErr := json.NewEncoder(overlay).Encode(map[string]map[string]string{
		"Replace": {filepath.Join(dir, FileName): ""},
	})
	closeErr := overlay.Close()
	if encodeErr != nil {
		return nil, fmt.Errorf("write scan overlay: %w", encodeErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close scan overlay: %w", closeErr)
	}
	args := []string{"list", "-deps", "-json", "-overlay", overlay.Name()}
	if len(opt.Tags) > 0 {
		args = append(args, "-tags", strings.Join(opt.Tags, ","))
	}
	args = append(args, ".")
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = cmd.Environ() // 连同 PWD 一起匹配 Dir，避免继承调用进程的路径别名。
	if opt.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+opt.GOOS)
	}
	if opt.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+opt.GOARCH)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w\n%s", err, stderr.String())
	}
	var packages []listedPackage
	names := make(map[string]string)
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
		names[pkg.ImportPath] = pkg.Name
		if pkg.Module != nil && pkg.Module.Main && !strings.Contains("/"+pkg.ImportPath+"/", "/vendor/") {
			packages = append(packages, pkg)
		}
	}
	s := &scan{files: token.NewFileSet(), dir: dir, paths: make(map[string]string), routes: make(map[string]string)}
	for _, pkg := range packages {
		if filepath.Clean(pkg.Dir) == dir {
			s.entry = pkg
		}
	}
	if s.entry.ImportPath == "" {
		return nil, fmt.Errorf("go list did not return an entry package in %s", dir)
	}
	slices.SortFunc(packages, func(a, b listedPackage) int { return strings.Compare(a.ImportPath, b.ImportPath) })
	// 2. 解析源码、聚合标记和类型声明；每个文件的问题独立保留。
	for _, pkg := range packages {
		slices.Sort(pkg.GoFiles)
		for _, name := range pkg.GoFiles {
			if err := s.readFile(pkg, filepath.Join(pkg.Dir, name), names); err != nil {
				return nil, err
			}
		}
	}
	s.checkTypes()
	slices.SortFunc(s.result.Commands, func(a, b Command) int { return strings.Compare(strings.Join(a.Path, " "), strings.Join(b.Path, " ")) })
	slices.SortFunc(s.result.Events, func(a, b Event) int { return strings.Compare(a.RouteID, b.RouteID) })
	slices.SortFunc(s.result.Diagnostics, func(a, b Diagnostic) int {
		if n := strings.Compare(a.Pos, b.Pos); n != 0 {
			return n
		}
		return strings.Compare(a.Message, b.Message)
	})
	// 3. 诊断阻止生成；成功时统一格式化完整注册源码。
	if len(s.result.Diagnostics) == 0 {
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

func (s *scan) readFile(pkg listedPackage, path string, names map[string]string) error {
	// 1. 在文件系统边界读取文件，解析器统一使用相对入口目录的文件名。
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	relative, err := filepath.Rel(s.dir, path)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	file, err := parser.ParseFile(s.files, filepath.ToSlash(relative), source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		if list, ok := err.(scanner.ErrorList); ok {
			for _, problem := range list {
				s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{fmt.Sprintf("%s:%d:%d", problem.Pos.Filename, problem.Pos.Line, problem.Pos.Column), problem.Msg})
			}
		} else {
			return fmt.Errorf("parse %s: %w", path, err)
		}
	}
	if file == nil {
		return nil
	}
	imports := make(map[string]string)
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		} // parser 已报告非法字符串。
		name := names[path]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		imports[name] = path
	}
	// 2. 识别根与配置的嵌入类型，同时记录包级函数的文档注释。
	handlers := make(map[*ast.CommentGroup]*ast.FuncDecl)
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Doc != nil {
			handlers[fn.Doc] = fn
		}
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typ, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structure, ok := typ.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structure.Fields.List {
				if len(field.Names) > 0 {
					continue
				}
				embedded := field.Type
				if star, ok := embedded.(*ast.StarExpr); ok {
					embedded = star.X
				}
				selector, ok := embedded.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Base" {
					continue
				}
				name, ok := selector.X.(*ast.Ident)
				if !ok {
					continue
				}
				d := declaration{typ.Name.Name, typ.Name.Pos(), pkg.ImportPath}
				switch imports[name.Name] {
				case sdkPath:
					s.roots = append(s.roots, d)
				case configPath:
					s.configs = append(s.configs, d)
				}
			}
		}
	}
	// 3. 所有标记都参与检查，包含误写在方法、类型或变量上的标记。
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if !strings.HasPrefix(comment.Text, "//alonabot:") {
				continue
			}
			fn := handlers[group]
			if fn == nil {
				s.diagnostic(comment.Pos(), "alonabot marker requires a package-level function")
				continue
			}
			if pkg.ImportPath != s.entry.ImportPath && !ast.IsExported(fn.Name.Name) {
				s.diagnostic(comment.Pos(), "marked function outside entry package must be exported")
				continue
			}
			s.marker(comment, fn.Name.Name, pkg.ImportPath)
		}
	}
	return nil
}

func (s *scan) checkTypes() {
	if len(s.roots) != 1 {
		var positions []string
		for _, root := range s.roots {
			positions = append(positions, s.position(root.pos))
		}
		pos := token.NoPos
		if len(s.roots) > 0 {
			pos = s.roots[0].pos
		}
		// 无根类型时定位到入口文件起点，仍保持诊断位置可用。
		message := fmt.Sprintf("expected exactly one root type, found %d", len(s.roots))
		if len(positions) > 0 {
			message += " at " + strings.Join(positions, ", ")
		}
		if pos == token.NoPos {
			files := s.entry.GoFiles
			if len(files) == 0 {
				// go list 成功的纯 cgo 包在 CgoFiles 中提供源码文件名。
				files = s.entry.CgoFiles
			}
			s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{filepath.ToSlash(files[0]) + ":1:1", message})
		} else {
			s.diagnostic(pos, message)
		}
	} else {
		root := s.roots[0]
		s.result.RootType = root.pkg + "." + root.name
		if root.pkg != s.entry.ImportPath {
			s.diagnostic(root.pos, "root type must be declared in entry package")
		}
	}
	if len(s.configs) > 1 {
		var positions []string
		for _, config := range s.configs {
			positions = append(positions, s.position(config.pos))
		}
		s.diagnostic(s.configs[0].pos, "expected at most one Config type, found "+strconv.Itoa(len(s.configs))+" at "+strings.Join(positions, ", "))
	} else if len(s.configs) == 1 {
		s.result.ConfigType = s.configs[0].pkg + "." + s.configs[0].name
	}
}
