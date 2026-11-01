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
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
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
	b := &configBuilder{
		scan: s, sizes: i.sizes, active: make(map[types.Type]bool),
		fields: make(map[*jsonschema.Schema]types.Type), bounds: make(map[*jsonschema.Schema][2]json.Number),
	}
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
	schema := b.schema(pkg.Scope().Lookup(declaration.name).Type(), declaration.pos, "config")
	if schema == nil || len(s.result.Diagnostics) != 0 {
		return nil
	}
	schema.Schema = configSchemaDraft
	// 3. Schema 的结构与默认值检查通过后，再生成规范字节。
	if _, err := schema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true}); err != nil {
		s.diagnostic(declaration.pos, "invalid config schema: "+err.Error())
		return nil
	}
	s.result.ConfigSchema, err = b.marshalSchema(schema)
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
	fields map[*jsonschema.Schema]types.Type     // 默认值按源码字段类型检查，包含匿名展开后的属性。
	bounds map[*jsonschema.Schema][2]json.Number // 库的 float64 字段无法保存的精确整数端点。
}

func (b *configBuilder) schema(fieldType types.Type, pos token.Pos, path string) *jsonschema.Schema {
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
			minimum, maximum := int64(0), ^uint64(0)>>(64-bits)
			if underlying.Info()&types.IsUnsigned == 0 {
				maximum >>= 1
				minimum = -int64(maximum) - 1
			}
			lower, upper := float64(minimum), float64(maximum)
			schema := &jsonschema.Schema{Type: "integer", Minimum: &lower, Maximum: &upper}
			b.bounds[schema] = [2]json.Number{json.Number(strconv.FormatInt(minimum, 10)), json.Number(strconv.FormatUint(maximum, 10))}
			return schema
		case underlying.Info()&types.IsFloat != 0:
			return &jsonschema.Schema{Type: "number"}
		}
	case *types.Pointer:
		schema := b.schema(underlying.Elem(), pos, path)
		if schema == nil {
			return nil
		}
		if schema.Type != "" {
			schema.Types, schema.Type = []string{schema.Type, "null"}, ""
		}
		return schema
	case *types.Slice:
		items := b.schema(underlying.Elem(), pos, path)
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
		values := b.schema(underlying.Elem(), pos, path)
		if values != nil {
			return &jsonschema.Schema{Type: "object", AdditionalProperties: values}
		}
		return nil
	case *types.Struct:
		return b.structure(underlying, path)
	}
	b.scan.diagnostic(pos, "unsupported config type "+fieldType.String())
	return nil
}

// marshalSchema 在规范 JSON 中恢复整数端点，库的浮点表示仍供 Resolve 使用。
func (b *configBuilder) marshalSchema(schema *jsonschema.Schema) ([]byte, error) {
	raw, err := config.MarshalSchema(schema)
	if err != nil {
		return nil, err
	}
	value, err := configJSON(string(raw))
	if err != nil {
		return nil, fmt.Errorf("decode config schema: %w", err)
	}
	b.exactBounds(schema, value.(configObject))
	return json.Marshal(value)
}

// exactBounds 只沿生成器创建的 Schema 边遍历，保留 default 和扩展关键字中的普通数据。
func (b *configBuilder) exactBounds(schema *jsonschema.Schema, object configObject) {
	for index := range object {
		member := &object[index]
		switch member.name {
		case "minimum":
			member.value = b.bounds[schema][0]
		case "maximum":
			member.value = b.bounds[schema][1]
		case "properties":
			for _, property := range member.value.(configObject) {
				b.exactBounds(schema.Properties[property.name], property.value.(configObject))
			}
		case "items":
			b.exactBounds(schema.Items, member.value.(configObject))
		case "additionalProperties":
			b.exactBounds(schema.AdditionalProperties, member.value.(configObject))
		}
	}
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

func (b *configBuilder) structure(structure *types.Struct, path string) *jsonschema.Schema {
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
		name, options, _ := strings.Cut(tags["json"], ",")
		if slices.Contains(strings.Split(options, ","), "string") {
			// encoding/json 的 string 选项只跟随一层指针，包含命名指针类型。
			quotedType := fieldType.Underlying()
			if pointer, ok := quotedType.(*types.Pointer); ok {
				quotedType = pointer.Elem().Underlying()
			}
			if basic, ok := quotedType.(*types.Basic); ok && basic.Info()&(types.IsString|types.IsBoolean|types.IsInteger|types.IsFloat) != 0 {
				b.scan.diagnostic(field.Pos(), fmt.Sprintf("json string option is unsupported for config type %s", fieldType))
				continue
			}
		}
		if !configJSONName(name) {
			name = ""
		}
		if field.Embedded() && name == "" {
			if named, ok := baseType.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == configPath && named.Obj().Name() == "Base" {
				continue
			}
			if _, structure := baseType.Underlying().(*types.Struct); structure {
				child := b.schema(fieldType, field.Pos(), path)
				if child != nil {
					for _, key := range child.PropertyOrder {
						b.property(schema, key, child.Properties[key], field.Pos())
					}
					schema.Required = append(schema.Required, child.Required...)
				}
				continue
			}
		}
		if name == "" {
			name = field.Name()
		}
		child := b.schema(fieldType, field.Pos(), path+"."+name)
		if child != nil {
			b.fields[child] = fieldType
			if b.metadata(child, fieldType, tags, path+"."+name, field.Pos()) {
				schema.Required = append(schema.Required, name)
			}
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

// configForm 用 JSON 解码器读取 default，保留字符串和容器内的分号。
func configForm(tag string) (map[string]string, error) {
	form := make(map[string]string)
	for tag != "" {
		key, rest, hasValue := strings.Cut(tag, "=")
		if end := strings.IndexByte(tag, ';'); end >= 0 && (!hasValue || end < len(key)) {
			key, rest, hasValue = tag[:end], tag[end+1:], false
		} else if !hasValue {
			rest = ""
		}
		switch key {
		case "required":
			if hasValue {
				return nil, fmt.Errorf("form required must be a flag")
			}
			form[key], tag = "", rest
			continue
		case "label", "widget", "options", "min", "max", "default":
			if !hasValue {
				return nil, fmt.Errorf("form %s requires a value", key)
			}
		default:
			return nil, fmt.Errorf("unknown form key %q", key)
		}
		if key == "default" {
			decoder := json.NewDecoder(strings.NewReader(rest))
			var value json.RawMessage
			if err := decoder.Decode(&value); err != nil {
				return nil, fmt.Errorf("invalid form default: %w", err)
			}
			form[key] = string(value)
			tag = strings.TrimLeft(rest[decoder.InputOffset():], " \t\r\n")
			if tag != "" && tag[0] != ';' {
				return nil, fmt.Errorf("expected semicolon after form default: %q", tag)
			}
			tag = strings.TrimPrefix(tag, ";")
		} else {
			form[key], tag, _ = strings.Cut(rest, ";")
		}
	}
	return form, nil
}

func (b *configBuilder) metadata(schema *jsonschema.Schema, fieldType types.Type, tags map[string]string, path string, pos token.Pos) bool {
	// 1. 每个 JSON 字段都有稳定的翻译路径，标签只覆盖显式声明的元数据。
	form := map[string]any{"label": path}
	schema.Description = tags["description"]
	schema.Extra = map[string]any{config.KeywordI18n: path, config.KeywordForm: form}
	for _, entry := range []struct{ tag, keyword string }{{"secret", config.KeywordSecret}, {"hidden", config.KeywordHidden}} {
		switch tags[entry.tag] {
		case "", "false":
		case "true":
			if entry.tag == "secret" && !configSecretType(fieldType) {
				b.scan.diagnostic(pos, "secret requires string or *string")
			}
			schema.Extra[entry.keyword] = true
		default:
			b.scan.diagnostic(pos, fmt.Sprintf("invalid %s tag %q: want true or false", entry.tag, tags[entry.tag]))
		}
	}
	values, err := configForm(tags["form"])
	if err != nil {
		b.scan.diagnostic(pos, err.Error())
		return false
	}
	if label, ok := values["label"]; ok {
		form["label"] = label
	}
	if widget, ok := values["widget"]; ok {
		basic := configBasic(fieldType)
		valid := false
		if basic != nil {
			switch widget {
			case "text", "password", "textarea":
				valid = basic.Info()&types.IsString != 0
			case "number":
				valid = basic.Info()&(types.IsInteger|types.IsFloat) != 0
			case "switch":
				valid = basic.Info()&types.IsBoolean != 0
			case "select":
				valid = basic.Info()&(types.IsString|types.IsInteger|types.IsFloat) != 0
			}
		}
		if !valid {
			b.scan.diagnostic(pos, fmt.Sprintf("widget %q is incompatible with %s", widget, fieldType))
		}
		form["widget"] = widget
	}
	// 2. 表单选项与边界保留数值原文；默认值按十进制位数和有效数字精确比较。
	for _, key := range []string{"min", "max"} {
		if raw, ok := values[key]; ok {
			value, err := configJSON(raw)
			number, numeric := value.(json.Number)
			basic := configBasic(fieldType)
			if err != nil {
				b.scan.diagnostic(pos, fmt.Sprintf("form %s must be a JSON number: %q: %v", key, raw, err))
				continue
			}
			if !numeric {
				b.scan.diagnostic(pos, fmt.Sprintf("form %s must be a JSON number: %q", key, raw))
				continue
			}
			if _, err := number.Float64(); err != nil {
				b.scan.diagnostic(pos, fmt.Sprintf("form %s: %v", key, err))
				continue
			}
			if basic == nil || basic.Info()&(types.IsInteger|types.IsFloat) == 0 {
				b.scan.diagnostic(pos, fmt.Sprintf("form %s requires a numeric field", key))
			}
			form[key] = number
		}
	}
	minimum, hasMinimum := form["min"].(json.Number)
	maximum, hasMaximum := form["max"].(json.Number)
	if hasMinimum && hasMaximum && configCompare(minimum, maximum) > 0 {
		b.scan.diagnostic(pos, "form min exceeds max")
	}
	if options, ok := values["options"]; ok {
		basic := configBasic(fieldType)
		if basic == nil || basic.Info()&(types.IsString|types.IsInteger|types.IsFloat) == 0 {
			b.scan.diagnostic(pos, "form options require a string or numeric field")
		} else {
			var choices []any
			icons := make(map[string]string)
			for _, option := range strings.Split(options, "|") {
				value, icon, hasIcon := strings.Cut(option, "@")
				var choice any = value
				if basic.Info()&types.IsString == 0 {
					choice, err = configJSON(value)
					if _, numeric := choice.(json.Number); err == nil && !numeric {
						err = fmt.Errorf("want a JSON number")
					}
					if err == nil {
						err = b.defaultValue(fieldType, schema, choice)
					}
					if err != nil {
						b.scan.diagnostic(pos, fmt.Sprintf("invalid numeric form option %q: %v", value, err))
						continue
					}
				}
				choices = append(choices, choice)
				if hasIcon {
					if err := b.icon(icon); err != nil {
						b.scan.diagnostic(pos, err.Error())
					} else {
						icons[value] = icon
					}
				}
			}
			form["options"] = choices
			if len(icons) != 0 {
				form["option_icons"] = icons
			}
		}
	}
	// 3. 源码类型检查补上库的 float64 默认值校验无法识别的整数越界。
	if raw, ok := values["default"]; ok {
		value, err := configJSON(raw)
		if err == nil {
			err = b.defaultValue(fieldType, schema, value)
		}
		if err != nil {
			b.scan.diagnostic(pos, fmt.Sprintf("invalid config default: %v", err))
		} else {
			schema.Default = json.RawMessage(raw)
		}
	}
	_, required := values["required"]
	return required
}

func configBasic(fieldType types.Type) *types.Basic {
	for {
		if pointer, ok := fieldType.Underlying().(*types.Pointer); ok {
			fieldType = pointer.Elem()
			continue
		}
		basic, _ := fieldType.Underlying().(*types.Basic)
		return basic
	}
}

func configSecretType(fieldType types.Type) bool {
	if pointer, ok := fieldType.Underlying().(*types.Pointer); ok {
		fieldType = pointer.Elem()
	}
	basic, ok := fieldType.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsString != 0
}

func (b *configBuilder) icon(path string) error {
	if !filepath.IsLocal(path) {
		return fmt.Errorf("option icon must be relative to the module directory: %q", path)
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".svg", ".png", ".webp":
	default:
		return fmt.Errorf("unsupported option icon extension: %q", path)
	}
	root, err := filepath.EvalSymlinks(b.scan.entry.Module.Dir)
	if err != nil {
		return fmt.Errorf("resolve option icon module: %w", err)
	}
	real, err := filepath.EvalSymlinks(filepath.Join(root, path))
	if err != nil {
		return fmt.Errorf("resolve option icon %q: %w", path, err)
	}
	relative, err := filepath.Rel(root, real)
	if err != nil {
		return fmt.Errorf("resolve option icon relative path %q: %w", path, err)
	}
	if !filepath.IsLocal(relative) {
		return fmt.Errorf("option icon escapes module directory: %q", path)
	}
	file, err := os.Stat(real)
	if err != nil {
		return fmt.Errorf("stat option icon %q: %w", path, err)
	}
	if !file.Mode().IsRegular() {
		return fmt.Errorf("option icon must be a regular file: %q", path)
	}
	return nil
}

func configJSON(raw string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	value, err := configJSONValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("multiple JSON values")
	}
	return value, nil
}

// configObject 保留 JSON 的键顺序和重复键，供默认值检查及 Schema 精确序列化使用。
type configObject []configMember

type configMember struct {
	name  string
	value any
}

// MarshalJSON 保留已有规范键序，包括 properties 的声明顺序。
func (object configObject) MarshalJSON() ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte('{')
	for index, member := range object {
		if index > 0 {
			out.WriteByte(',')
		}
		key, _ := json.Marshal(member.name) // 字符串编码没有失败分支。
		out.Write(key)
		out.WriteByte(':')
		value, err := json.Marshal(member.value)
		if err != nil {
			return nil, err
		}
		out.Write(value)
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}

func configJSONValue(decoder *json.Decoder) (any, error) {
	value, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch value {
	case json.Delim('{'):
		var object configObject
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			item, err := configJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			object = append(object, configMember{key.(string), item})
		}
		_, err = decoder.Token()
		return object, err
	case json.Delim('['):
		items := []any{} // 空数组在重新编码时保持 []。
		for decoder.More() {
			item, err := configJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		_, err = decoder.Token()
		return items, err
	default:
		return value, nil
	}
}

// configCompare 的两端均来自已解码的 JSON number；比较指数而不展开幂，兼顾精度与极小数。
func configCompare(left, right json.Number) int {
	a, aOrder, aSign := configDecimal(left)
	b, bOrder, bSign := configDecimal(right)
	if aSign != bSign {
		return aSign - bSign
	}
	if aSign == 0 {
		return 0
	}
	if order := aOrder.Cmp(bOrder); order != 0 {
		return aSign * order
	}
	return aSign * strings.Compare(a, b)
}

func configDecimal(number json.Number) (digits string, order *big.Int, sign int) {
	mantissa, exponent, _ := strings.Cut(strings.ToLower(string(number)), "e")
	order = new(big.Int)
	if exponent != "" {
		order.SetString(exponent, 10)
	}
	sign = 1
	if strings.HasPrefix(mantissa, "-") {
		sign, mantissa = -1, mantissa[1:]
	}
	whole, fraction, _ := strings.Cut(mantissa, ".")
	digits = strings.TrimLeft(whole+fraction, "0")
	if digits == "" {
		return "", order, 0
	}
	order.Add(order, big.NewInt(int64(len(digits)-len(fraction))))
	return strings.TrimRight(digits, "0"), order, sign
}

func (b *configBuilder) defaultValue(fieldType types.Type, schema *jsonschema.Schema, value any) error {
	if pointer, ok := fieldType.Underlying().(*types.Pointer); ok {
		if value == nil {
			return nil
		}
		return b.defaultValue(pointer.Elem(), schema, value)
	}
	if value == nil {
		return fmt.Errorf("null requires a pointer, want %s", fieldType)
	}
	switch underlying := fieldType.Underlying().(type) {
	case *types.Basic:
		switch {
		case underlying.Info()&types.IsString != 0:
			if _, ok := value.(string); !ok {
				return fmt.Errorf("want %s, got %T", fieldType, value)
			}
		case underlying.Info()&types.IsBoolean != 0:
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("want %s, got %T", fieldType, value)
			}
		default:
			number, ok := value.(json.Number)
			if !ok {
				return fmt.Errorf("want %s, got %T", fieldType, value)
			}
			bits := int(b.sizes.Sizeof(fieldType) * 8)
			var err error
			switch {
			case underlying.Info()&types.IsUnsigned != 0:
				_, err = strconv.ParseUint(string(number), 10, bits)
			case underlying.Info()&types.IsInteger != 0:
				_, err = strconv.ParseInt(string(number), 10, bits)
			case underlying.Info()&types.IsFloat != 0:
				_, err = strconv.ParseFloat(string(number), bits)
			}
			if err != nil {
				return fmt.Errorf("decode %s: %w", fieldType, err)
			}
		}
	case *types.Slice:
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("want %s, got %T", fieldType, value)
		}
		for index, item := range items {
			if err := b.defaultValue(underlying.Elem(), schema.Items, item); err != nil {
				return fmt.Errorf("item %d: %w", index, err)
			}
		}
	case *types.Map, *types.Struct:
		object, ok := value.(configObject)
		if !ok {
			return fmt.Errorf("want %s, got %T", fieldType, value)
		}
		if mapping, ok := underlying.(*types.Map); ok {
			for _, item := range object {
				if err := b.defaultValue(mapping.Elem(), schema.AdditionalProperties, item.value); err != nil {
					return fmt.Errorf("key %q: %w", item.name, err)
				}
			}
		} else {
			provided := make(map[string]bool)
			for _, item := range object {
				provided[item.name] = true
				child := schema.Properties[item.name]
				if child == nil {
					// encoding/json 优先精确匹配，然后按声明顺序忽略大小写匹配。
					for _, key := range schema.PropertyOrder {
						if strings.EqualFold(key, item.name) {
							child = schema.Properties[key]
							break
						}
					}
				}
				if child != nil {
					if err := b.defaultValue(b.fields[child], child, item.value); err != nil {
						return fmt.Errorf("field %q: %w", item.name, err)
					}
				}
			}
			for _, key := range schema.Required {
				if !provided[key] {
					return fmt.Errorf("missing required field %q", key)
				}
			}
		}
	}
	form, _ := schema.Extra[config.KeywordForm].(map[string]any)
	if number, ok := value.(json.Number); ok {
		if minimum, ok := form["min"].(json.Number); ok && configCompare(number, minimum) < 0 {
			return fmt.Errorf("%s is below form min %s", number, minimum)
		}
		if maximum, ok := form["max"].(json.Number); ok && configCompare(number, maximum) > 0 {
			return fmt.Errorf("%s exceeds form max %s", number, maximum)
		}
	}
	if choices, ok := form["options"].([]any); ok {
		for _, choice := range choices {
			if number, ok := value.(json.Number); ok {
				if configCompare(number, choice.(json.Number)) == 0 {
					return nil
				}
			} else if value == choice {
				return nil
			}
		}
		return fmt.Errorf("%v is not in form options", value)
	}
	return nil
}
