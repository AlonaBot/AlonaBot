package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/jsonschema-go/jsonschema"
)

// schemaImporter 对扫描包、SDK 声明所在包及回引它们的包检查源码，其他包读取导出数据。
type schemaImporter struct {
	scan     *scan
	compiled types.Importer
	sizes    types.Sizes
	checked  map[string]*types.Package
	info     *types.Info
}

func (i *schemaImporter) Import(path string) (*types.Package, error) {
	if pkg := i.checked[path]; pkg != nil {
		return pkg, nil
	}
	syntax, local := i.scan.syntax[path]
	if !local {
		return i.compiled.Import(path)
	}
	options := &types.Config{
		Importer: i, Sizes: i.sizes,
		Error: func(err error) {
			problem := err.(types.Error)
			i.scan.diagnostic(problem.Pos, problem.Msg)
		},
	}
	pkg, err := options.Check(path, i.scan.files, syntax, i.info)
	if err != nil {
		return nil, err
	}
	i.checked[path] = pkg
	return pkg, nil
}

func (s *scan) schemaImporter(ctx context.Context, list *exec.Cmd) (*schemaImporter, error) {
	// 1. 在相同 GOOS/GOARCH、tags 与 overlay 下取得依赖构建视图和目标位宽。
	external := map[string]bool{"encoding": true, "encoding/json": true}
	if s.adapterProject {
		external[adapterPath] = true
	}
	for _, files := range s.syntax {
		for _, file := range files {
			for _, spec := range file.Imports {
				path, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					continue // 语法诊断已由 readFile 保留。
				}
				if _, local := s.syntax[path]; !local {
					external[path] = true
				}
			}
		}
	}
	var paths []string
	for path := range external {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	args := append(slices.Clone(list.Args[1:len(list.Args)-1]), "-export", "-compiled")
	args = append(args, paths...)
	command := exec.CommandContext(ctx, list.Path, args...)
	command.Dir, command.Env = s.dir, list.Env
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go list schema imports: %w\n%s", err, stderr.String())
	}
	exports := make(map[string]string)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode schema imports: %w", err)
		}
		exports[pkg.ImportPath] = pkg.Export
		if _, source := s.syntax[pkg.ImportPath]; source {
			continue
		}
		// go list -deps 按依赖优先输出；回引源码包的整条链共用同一份类型身份。
		if !slices.Contains(pkg.Imports, sdkPath) && !slices.ContainsFunc(pkg.Imports, func(path string) bool { _, source := s.syntax[path]; return source }) {
			continue
		}
		for _, name := range pkg.CompiledGoFiles {
			path := name
			if !filepath.IsAbs(path) {
				path = filepath.Join(pkg.Dir, path)
			}
			file, err := s.parseFile(path)
			if err != nil {
				return nil, err
			}
			s.syntax[pkg.ImportPath] = append(s.syntax[pkg.ImportPath], file)
		}
	}
	command = exec.CommandContext(ctx, list.Path, "env", "GOARCH")
	command.Dir, command.Env = s.dir, list.Env
	output, err = command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go env GOARCH: %w\n%s", err, output)
	}
	i := &schemaImporter{
		info: &types.Info{
			Instances: make(map[*ast.Ident]types.Instance), Uses: make(map[*ast.Ident]types.Object),
			Types: make(map[ast.Expr]types.TypeAndValue), Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}, scan: s, sizes: types.SizesFor("gc", strings.TrimSpace(string(output))), checked: make(map[string]*types.Package),
		compiled: importer.ForCompiler(s.files, "gc", func(path string) (io.ReadCloser, error) {
			file, err := os.Open(exports[path])
			if err != nil {
				return nil, fmt.Errorf("open export data for %s: %w", path, err)
			}
			return file, nil
		}),
	}
	return i, nil
}

const serviceSchemaDraft = "https://json-schema.org/draft/2020-12/schema"

// schemas 共用构建视图检查函数体，为本模块及其引用的 JSON 声明生成 Service Schema。
func (s *scan) schemas(ctx context.Context, list *exec.Cmd) error {
	paths := make([]string, 0, len(s.syntax))
	for path := range s.syntax {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	importer, err := s.schemaImporter(ctx, list)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if _, err := importer.Import(path); err != nil {
			return nil
		}
	}
	if s.adapterProject {
		if err := s.checkAdapterType(importer); err != nil {
			return err
		}
	} else {
		s.checkConstructor(importer.checked[s.entry.ImportPath])
	}
	s.migrationCalls(importer.info, paths)
	s.translationCalls(importer.info, paths)
	return s.serviceSchemasFrom(importer, paths)
}

func (s *scan) serviceSchemasFrom(importer *schemaImporter, paths []string) error {
	b := &serviceBuilder{scan: s, sizes: importer.sizes, active: make(map[types.Type]bool), bounds: make(map[*jsonschema.Schema][2]json.Number)}
	for _, entry := range []struct{ path, name string }{{"encoding/json", "Marshaler"}, {"encoding/json", "Unmarshaler"}, {"encoding", "TextMarshaler"}, {"encoding", "TextUnmarshaler"}} {
		pkg, err := importer.Import(entry.path)
		if err != nil {
			return fmt.Errorf("import service codec: %w", err)
		}
		b.codecs = append(b.codecs, serviceCodec{entry.path + "." + entry.name, pkg.Scope().Lookup(entry.name).Type().Underlying().(*types.Interface)})
	}
	ids := make(map[string]types.Type)
	localFiles := make(map[string]bool)
	for _, path := range paths {
		for _, file := range s.syntax[path] {
			localFiles[s.files.Position(file.Pos()).Filename] = true
		}
	}
	// 被引用的 Service 值保留类型参数，用它们匹配依赖模块声明处的 JSON 实例。
	var referenced []types.Type
	for expr, value := range importer.info.Types {
		if !value.IsValue() {
			continue
		}
		if !localFiles[s.files.Position(expr.Pos()).Filename] {
			// 包装函数中的变量引用也需要 Schema，未使用声明的初始化表达式不参与。
			switch expr.(type) {
			case *ast.Ident, *ast.SelectorExpr:
			default:
				continue
			}
		}
		named, ok := types.Unalias(value.Type).(*types.Named)
		if ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == sdkPath && named.Obj().Name() == "Service" {
			referenced = append(referenced, named)
		}
	}
	var instances []*ast.Ident
	for ident := range importer.info.Instances {
		instances = append(instances, ident)
	}
	slices.SortFunc(instances, func(a, b *ast.Ident) int { return int(a.Pos() - b.Pos()) })
	for _, ident := range instances {
		object := importer.info.Uses[ident]
		fn, ok := object.(*types.Func)
		if !ok || fn.Pkg() == nil || fn.Pkg().Path() != sdkPath || !slices.Contains([]string{"JSON", "Proto", "BusEvent"}, fn.Name()) {
			continue
		}
		if !localFiles[s.files.Position(ident.Pos()).Filename] {
			if fn.Name() != "JSON" {
				continue
			}
			result := importer.info.Instances[ident].Type.(*types.Signature).Results().At(0).Type()
			if !slices.ContainsFunc(referenced, func(typ types.Type) bool { return types.Identical(typ, result) }) {
				continue
			}
		}
		args := importer.info.Instances[ident].TypeArgs
		for index := range args.Len() {
			typ := types.Unalias(args.At(index))
			// Protobuf 的指针参数绑定包级消息类型，JSON 与 BusEvent 仍要求值类型。
			if pointer, ok := typ.(*types.Pointer); ok && fn.Name() == "Proto" {
				typ = types.Unalias(pointer.Elem())
			}
			named, ok := typ.(*types.Named)
			if !ok || named.Obj().Pkg() == nil || named.Obj().Parent() != named.Obj().Pkg().Scope() || named.TypeParams().Len() != 0 || named.TypeArgs().Len() != 0 {
				s.diagnostic(ident.Pos(), "service type must be a package-level named non-generic type: "+typ.String())
				continue
			}
			// 声明处共用类型检查，只有 JSON 声明需要生成 Schema。
			if fn.Name() != "JSON" {
				continue
			}
			pkgPath := named.Obj().Pkg().Path()
			if named.Obj().Pkg().Name() == "main" {
				pkgPath = "main"
			}
			id := pkgPath + "." + named.Obj().Name()
			if previous := ids[id]; previous != nil {
				if !types.Identical(previous, typ) {
					s.diagnostic(ident.Pos(), "conflicting service schema_id "+id)
				}
				continue
			}
			ids[id] = typ
			schema := b.schema(typ, ident.Pos())
			if schema == nil {
				continue
			}
			schema.Schema = serviceSchemaDraft
			if _, err := schema.Resolve(nil); err != nil {
				s.diagnostic(ident.Pos(), "invalid service schema: "+err.Error())
				continue
			}
			encoded, err := b.marshal(schema)
			if err != nil {
				return err
			}
			if s.serviceSchemas == nil {
				s.serviceSchemas = make(map[string][]byte)
			}
			s.serviceSchemas[id] = encoded
		}
	}
	return nil
}

type serviceCodec struct {
	name     string
	contract *types.Interface
}

// serviceBuilder 仅按 Go 类型与 json 标签转换；active 记录当前递归路径。
type serviceBuilder struct {
	scan   *scan
	sizes  types.Sizes
	active map[types.Type]bool
	codecs []serviceCodec
	bounds map[*jsonschema.Schema][2]json.Number // 序列化时恢复精确整数端点。
}

func (b *serviceBuilder) schema(typ types.Type, pos token.Pos) *jsonschema.Schema {
	typ = types.Unalias(typ)
	for _, codec := range b.codecs {
		if types.Implements(typ, codec.contract) || types.Implements(types.NewPointer(typ), codec.contract) {
			b.scan.diagnostic(pos, fmt.Sprintf("service type %s implements %s", typ, codec.name))
			return nil
		}
	}
	if b.active[typ] {
		b.scan.diagnostic(pos, "recursive service type "+typ.String())
		return nil
	}
	if named, ok := typ.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "encoding/json" && named.Obj().Name() == "Number" {
		b.scan.diagnostic(pos, "unsupported service type "+typ.String())
		return nil
	}
	b.active[typ] = true
	defer delete(b.active, typ)
	switch u := typ.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsString != 0:
			return &jsonschema.Schema{Type: "string"}
		case u.Info()&types.IsBoolean != 0:
			return &jsonschema.Schema{Type: "boolean"}
		case u.Info()&types.IsInteger != 0:
			bits := int(b.sizes.Sizeof(typ) * 8)
			minimum, maximum := int64(0), ^uint64(0)>>(64-bits)
			if u.Info()&types.IsUnsigned == 0 {
				maximum >>= 1
				minimum = -int64(maximum) - 1
			}
			lower, upper := float64(minimum), float64(maximum)
			schema := &jsonschema.Schema{Type: "integer", Minimum: &lower, Maximum: &upper}
			b.bounds[schema] = [2]json.Number{json.Number(strconv.FormatInt(minimum, 10)), json.Number(strconv.FormatUint(maximum, 10))}
			return schema
		case u.Info()&types.IsFloat != 0:
			return &jsonschema.Schema{Type: "number"}
		}
	case *types.Pointer:
		schema := b.schema(u.Elem(), pos)
		if schema != nil && schema.Type != "" {
			schema.Types, schema.Type = []string{schema.Type, "null"}, ""
		}
		return schema
	case *types.Slice:
		if element, ok := u.Elem().Underlying().(*types.Basic); ok && element.Kind() == types.Uint8 {
			break
		}
		if items := b.schema(u.Elem(), pos); items != nil {
			return &jsonschema.Schema{Types: []string{"array", "null"}, Items: items}
		}
		return nil
	case *types.Map:
		key, ok := u.Key().Underlying().(*types.Basic)
		if !ok || key.Kind() != types.String {
			b.scan.diagnostic(pos, "service map key must be string: "+typ.String())
			return nil
		}
		for _, codec := range b.codecs {
			if types.Implements(u.Key(), codec.contract) || types.Implements(types.NewPointer(u.Key()), codec.contract) {
				b.scan.diagnostic(pos, "service map key has custom codec: "+typ.String())
				return nil
			}
		}
		if value := b.schema(u.Elem(), pos); value != nil {
			return &jsonschema.Schema{Types: []string{"object", "null"}, AdditionalProperties: value}
		}
		return nil
	case *types.Struct:
		return b.structure(u)
	}
	b.scan.diagnostic(pos, "unsupported service type "+typ.String())
	return nil
}

func (b *serviceBuilder) structure(typ *types.Struct) *jsonschema.Schema {
	schema := &jsonschema.Schema{Type: "object", Properties: make(map[string]*jsonschema.Schema)}
	for index := range typ.NumFields() {
		field := typ.Field(index)
		underlying := types.Unalias(field.Type()).Underlying()
		pointer, indirect := underlying.(*types.Pointer)
		if indirect {
			underlying = pointer.Elem().Underlying()
		}
		_, embeddedStruct := underlying.(*types.Struct)
		if !field.Exported() && (!field.Embedded() || !embeddedStruct) {
			continue
		}
		tag := reflect.StructTag(typ.Tag(index)).Get("json")
		if tag == "-" {
			continue
		}
		name, options, _ := strings.Cut(tag, ",")
		if strings.IndexFunc(name, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", r)
		}) >= 0 {
			name = ""
		}
		if slices.Contains(strings.Split(options, ","), "string") {
			if basic, ok := underlying.(*types.Basic); ok && basic.Info()&(types.IsString|types.IsBoolean|types.IsInteger|types.IsFloat) != 0 {
				b.scan.diagnostic(field.Pos(), "json string option is unsupported for service type "+field.Type().String())
				continue
			}
		}
		child := b.schema(field.Type(), field.Pos())
		if child == nil {
			continue
		}
		if field.Embedded() && embeddedStruct && name == "" {
			for _, key := range child.PropertyOrder {
				b.property(schema, key, child.Properties[key], field.Pos())
			}
			// nil 匿名指针会省略整个展开字段集合。
			if !indirect {
				schema.Required = append(schema.Required, child.Required...)
			}
			continue
		}
		if name == "" {
			name = field.Name()
		}
		b.property(schema, name, child, field.Pos())
		opts := strings.Split(options, ",")
		if !slices.Contains(opts, "omitempty") && !slices.Contains(opts, "omitzero") {
			schema.Required = append(schema.Required, name)
		}
	}
	return schema
}

func (b *serviceBuilder) property(schema *jsonschema.Schema, name string, child *jsonschema.Schema, pos token.Pos) {
	if _, exists := schema.Properties[name]; exists {
		b.scan.diagnostic(pos, "duplicate service JSON name "+strconv.Quote(name))
		return
	}
	schema.Properties[name] = child
	schema.PropertyOrder = append(schema.PropertyOrder, name)
}

// marshal 排序所有对象键，并修正库中 float64 无法精确表示的整数边界。
func (b *serviceBuilder) marshal(schema *jsonschema.Schema) ([]byte, error) {
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal service schema: %w", err)
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		return nil, fmt.Errorf("decode service schema: %w", err)
	}
	b.exactBounds(schema, object)
	return json.Marshal(object)
}

func (b *serviceBuilder) exactBounds(schema *jsonschema.Schema, object map[string]any) {
	if bounds, ok := b.bounds[schema]; ok {
		object["minimum"], object["maximum"] = bounds[0], bounds[1]
	}
	if properties, ok := object["properties"].(map[string]any); ok {
		for key, child := range properties {
			b.exactBounds(schema.Properties[key], child.(map[string]any))
		}
	}
	if child, ok := object["items"].(map[string]any); ok {
		b.exactBounds(schema.Items, child)
	}
	if child, ok := object["additionalProperties"].(map[string]any); ok {
		b.exactBounds(schema.AdditionalProperties, child)
	}
}

func (s *scan) checkConstructor(pkg *types.Package) {
	object := pkg.Scope().Lookup("New")
	fn, ok := object.(*types.Func)
	valid := ok
	if ok {
		sig := fn.Type().(*types.Signature)
		valid = sig.Recv() == nil && sig.Params().Len() == 0 && !sig.Variadic() && sig.Results().Len() == 1 && sig.TypeParams().Len() == 0
		if valid {
			pointer, ok := types.Unalias(sig.Results().At(0).Type()).(*types.Pointer)
			valid = ok
			if ok {
				named, ok := types.Unalias(pointer.Elem()).(*types.Named)
				valid = ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == sdkPath && named.Obj().Name() == "Plugin"
			}
		}
	}
	if !valid {
		s.diagnostic(s.constructors[0], "entry constructor must have signature func New() *sdk.Plugin")
	}
}

func (s *scan) migrationCalls(info *types.Info, paths []string) {
	var self, versioned []token.Pos
	for _, path := range paths {
		for _, file := range s.syntax[path] {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				var ident *ast.Ident
				switch fun := ast.Unparen(call.Fun).(type) {
				case *ast.SelectorExpr:
					ident = fun.Sel
				case *ast.Ident:
					ident = fun
				}
				fn, ok := info.Uses[ident].(*types.Func)
				if !ok || fn.Pkg() == nil || (fn.Pkg().Path() != sdkPath+"/sqlitedb" && fn.Pkg().Path() != sdkPath+"/gormdb") {
					return true
				}
				switch fn.Name() {
				case "Migrate":
					self = append(self, call.Pos())
				case "Migrations":
					versioned = append(versioned, call.Pos())
				}
				return true
			})
		}
	}
	if len(self) > 0 && len(versioned) > 0 {
		s.diagnostic(versioned[0], "Migrate and Migrations registrations are mutually exclusive (Migrate at "+s.position(self[0])+")")
	}
}
