package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"math"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/google/jsonschema-go/jsonschema"
)

const configSchemaDraft = "https://json-schema.org/draft/2020-12/schema"

// configImporter 对扫描到的本模块包检查源码，其他包读取同一构建环境的导出数据。
type configImporter struct {
	scan     *scan
	compiled types.Importer
	sizes    types.Sizes
	checked  map[string]*types.Package
}

func (i *configImporter) Import(path string) (*types.Package, error) {
	if pkg := i.checked[path]; pkg != nil {
		return pkg, nil
	}
	syntax, local := i.scan.syntax[path]
	if !local {
		return i.compiled.Import(path)
	}
	options := &types.Config{
		Importer: i, Sizes: i.sizes, IgnoreFuncBodies: true,
		Error: func(err error) {
			problem := err.(types.Error)
			i.scan.diagnostic(problem.Pos, problem.Msg)
		},
	}
	pkg, err := options.Check(path, i.scan.files, syntax, nil)
	if err != nil {
		return nil, err
	}
	i.checked[path] = pkg
	return pkg, nil
}

func (s *scan) configSchema(ctx context.Context, list *exec.Cmd) error {
	// 1. 在相同 GOOS/GOARCH、tags 与 overlay 下取得外部包导出数据和目标位宽。
	external := map[string]bool{"encoding": true, "encoding/json": true}
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
	args := append(slices.Clone(list.Args[1:len(list.Args)-1]), "-export")
	args = append(args, paths...)
	command := exec.CommandContext(ctx, list.Path, args...)
	command.Dir, command.Env = s.dir, list.Env
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("go list config imports: %w\n%s", err, stderr.String())
	}
	exports := make(map[string]string)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("decode config imports: %w", err)
		}
		exports[pkg.ImportPath] = pkg.Export
	}
	command = exec.CommandContext(ctx, list.Path, "env", "GOARCH")
	command.Dir, command.Env = s.dir, list.Env
	output, err = command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go env GOARCH: %w\n%s", err, output)
	}
	i := &configImporter{
		scan: s, sizes: types.SizesFor("gc", strings.TrimSpace(string(output))), checked: make(map[string]*types.Package),
		compiled: importer.ForCompiler(s.files, "gc", func(path string) (io.ReadCloser, error) {
			file, err := os.Open(exports[path])
			if err != nil {
				return nil, fmt.Errorf("open export data for %s: %w", path, err)
			}
			return file, nil
		}),
	}
	// 2. 复用扫描 AST，按真实方法集识别自定义编解码，并从 Config 声明构造 Schema。
	declaration := s.configs[0]
	pkg, err := i.Import(declaration.pkg)
	if err != nil {
		return nil // go/types 的位置诊断已完整收集。
	}
	b := &configBuilder{scan: s, sizes: i.sizes, active: make(map[types.Type]bool)}
	for _, entry := range []struct{ path, name string }{
		{"encoding/json", "Marshaler"}, {"encoding/json", "Unmarshaler"},
		{"encoding", "TextMarshaler"}, {"encoding", "TextUnmarshaler"},
	} {
		dependency, err := i.Import(entry.path)
		if err != nil {
			return fmt.Errorf("import config codec %s.%s: %w", entry.path, entry.name, err)
		}
		b.codecs = append(b.codecs, configCodec{entry.path + "." + entry.name, dependency.Scope().Lookup(entry.name).Type().Underlying().(*types.Interface)})
	}
	schema := b.schema(pkg.Scope().Lookup(declaration.name).Type(), declaration.pos)
	if schema == nil || len(s.result.Diagnostics) != 0 {
		return nil
	}
	schema.Schema = configSchemaDraft
	// 3. Schema 的结构与默认值检查通过后，再生成规范字节。
	if _, err := schema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true}); err != nil {
		s.diagnostic(declaration.pos, "invalid config schema: "+err.Error())
		return nil
	}
	s.result.ConfigSchema, err = config.MarshalSchema(schema)
	return err
}

type configCodec struct {
	name     string
	contract *types.Interface
}

// configBuilder 为每个字段创建独立 Schema 节点；active 只记录当前路径，拒绝递归类型。
type configBuilder struct {
	scan   *scan
	sizes  types.Sizes
	active map[types.Type]bool
	codecs []configCodec
}

func (b *configBuilder) schema(fieldType types.Type, pos token.Pos) *jsonschema.Schema {
	fieldType = types.Unalias(fieldType)
	if b.customCodec(fieldType, pos) {
		return nil
	}
	if b.active[fieldType] {
		b.scan.diagnostic(pos, "recursive config type "+fieldType.String())
		return nil
	}
	b.active[fieldType] = true
	defer delete(b.active, fieldType)
	if named, ok := fieldType.(*types.Named); ok && named.TypeParams().Len() > named.TypeArgs().Len() {
		b.scan.diagnostic(pos, "uninstantiated config type "+fieldType.String())
		return nil
	}
	switch underlying := fieldType.Underlying().(type) {
	case *types.Basic:
		switch {
		case underlying.Info()&types.IsString != 0:
			return &jsonschema.Schema{Type: "string"}
		case underlying.Info()&types.IsBoolean != 0:
			return &jsonschema.Schema{Type: "boolean"}
		case underlying.Info()&types.IsInteger != 0:
			bits := int(b.sizes.Sizeof(fieldType) * 8)
			minimum, maximum := 0.0, math.Ldexp(1, bits)-1
			if underlying.Info()&types.IsUnsigned == 0 {
				minimum, maximum = -math.Ldexp(1, bits-1), math.Ldexp(1, bits-1)-1
			}
			return &jsonschema.Schema{Type: "integer", Minimum: &minimum, Maximum: &maximum}
		case underlying.Info()&types.IsFloat != 0:
			return &jsonschema.Schema{Type: "number"}
		}
	case *types.Pointer:
		schema := b.schema(underlying.Elem(), pos)
		if schema == nil {
			return nil
		}
		if schema.Type != "" {
			schema.Types, schema.Type = []string{schema.Type, "null"}, ""
		}
		return schema
	case *types.Slice:
		items := b.schema(underlying.Elem(), pos)
		if items != nil {
			return &jsonschema.Schema{Type: "array", Items: items}
		}
		return nil
	case *types.Map:
		if key, ok := underlying.Key().Underlying().(*types.Basic); !ok || key.Kind() != types.String {
			b.scan.diagnostic(pos, "config map key must be string: "+fieldType.String())
			return nil
		}
		if b.customCodec(underlying.Key(), pos) {
			return nil
		}
		values := b.schema(underlying.Elem(), pos)
		if values != nil {
			return &jsonschema.Schema{Type: "object", AdditionalProperties: values}
		}
		return nil
	case *types.Struct:
		return b.structure(underlying)
	}
	b.scan.diagnostic(pos, "unsupported config type "+fieldType.String())
	return nil
}

func (b *configBuilder) customCodec(fieldType types.Type, pos token.Pos) bool {
	for _, codec := range b.codecs {
		if types.Implements(fieldType, codec.contract) || types.Implements(types.NewPointer(fieldType), codec.contract) {
			b.scan.diagnostic(pos, fmt.Sprintf("config type %s implements %s", fieldType, codec.name))
			return true
		}
	}
	return false
}

func (b *configBuilder) structure(structure *types.Struct) *jsonschema.Schema {
	schema := &jsonschema.Schema{Type: "object", Properties: make(map[string]*jsonschema.Schema)}
	for index := range structure.NumFields() {
		field := structure.Field(index)
		fieldType := field.Type()
		baseType := types.Unalias(fieldType)
		if pointer, ok := baseType.(*types.Pointer); ok {
			baseType = types.Unalias(pointer.Elem())
		}
		if !field.Exported() {
			if _, structure := baseType.Underlying().(*types.Struct); !field.Embedded() || !structure {
				continue
			}
		}
		tags, err := configTags(structure.Tag(index))
		if err != nil {
			b.scan.diagnostic(field.Pos(), err.Error())
			continue
		}
		if tags["json"] == "-" {
			continue
		}
		name := strings.SplitN(tags["json"], ",", 2)[0]
		if !configJSONName(name) {
			name = ""
		}
		if field.Embedded() && name == "" {
			if named, ok := baseType.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == configPath && named.Obj().Name() == "Base" {
				continue
			}
			if _, structure := baseType.Underlying().(*types.Struct); structure {
				child := b.schema(fieldType, field.Pos())
				if child != nil {
					for _, key := range child.PropertyOrder {
						b.property(schema, key, child.Properties[key], field.Pos())
					}
				}
				continue
			}
		}
		if name == "" {
			name = field.Name()
		}
		child := b.schema(fieldType, field.Pos())
		if child != nil {
			b.property(schema, name, child, field.Pos())
		}
	}
	return schema
}

func (b *configBuilder) property(schema *jsonschema.Schema, name string, child *jsonschema.Schema, pos token.Pos) {
	if _, exists := schema.Properties[name]; exists {
		b.scan.diagnostic(pos, "duplicate config JSON name "+strconv.Quote(name))
		return
	}
	schema.Properties[name] = child
	schema.PropertyOrder = append(schema.PropertyOrder, name)
}

func configTags(tag string) (map[string]string, error) {
	tags := make(map[string]string)
	for tag = strings.TrimLeft(tag, " "); tag != ""; tag = strings.TrimLeft(tag, " ") {
		end := strings.IndexByte(tag, ':')
		if end <= 0 || strings.ContainsAny(tag[:end], "\t\r\n\"\\ ") || len(tag) <= end+1 || tag[end+1] != '"' {
			return nil, fmt.Errorf("invalid config struct tag %q", tag)
		}
		key := tag[:end]
		tag = tag[end+1:]
		quoted, err := strconv.QuotedPrefix(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid config struct tag: %w", err)
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return nil, fmt.Errorf("invalid config struct tag: %w", err)
		}
		if _, exists := tags[key]; !exists {
			tags[key] = value
		}
		tag = tag[len(quoted):]
	}
	return tags, nil
}

// configJSONName 遵循 encoding/json 对字段标签名的字符规则。
func configJSONName(name string) bool {
	for _, character := range name {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", character) {
			return false
		}
	}
	return name != ""
}
