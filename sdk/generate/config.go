package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"io"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/google/jsonschema-go/jsonschema"
)

const configSchemaDraft = "https://json-schema.org/draft/2020-12/schema"

func (s *scan) configSchema(i *schemaImporter) error {
	// 1. 复用扫描 AST，按真实方法集识别自定义编解码，并从 Config 声明构造 Schema。
	declaration := s.configs[0]
	pkg, err := i.Import(declaration.pkg)
	if err != nil {
		return nil // go/types 的位置诊断已完整收集。
	}
	b := &configBuilder{
		scan: s, sizes: i.sizes, active: make(map[types.Type]bool),
		fields: make(map[*jsonschema.Schema]types.Type), bounds: make(map[*jsonschema.Schema][2]json.Number),
		unexported: make(map[*jsonschema.Schema]types.Type), positions: make(map[*jsonschema.Schema]token.Pos),
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
	schema := b.schema(pkg.Scope().Lookup(declaration.name).Type(), declaration.pos)
	if schema == nil {
		return nil
	}
	b.conditions(schema)
	if len(s.result.Diagnostics) != 0 {
		return nil
	}
	schema.Schema = configSchemaDraft
	// 2. Schema 的结构与默认值检查通过后，再生成规范字节。
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
	scan       *scan
	sizes      types.Sizes
	active     map[types.Type]bool
	codecs     []configCodec
	fields     map[*jsonschema.Schema]types.Type     // 默认值按源码字段类型检查，包含匿名展开后的属性。
	bounds     map[*jsonschema.Schema][2]json.Number // 库的 float64 字段无法保存的精确整数端点。
	unexported map[*jsonschema.Schema]types.Type     // 默认对象访问属性时无法设置的匿名指针。
	positions  map[*jsonschema.Schema]token.Pos      // 显示条件在整棵树完成后校验，保留原字段位置。
}

func (b *configBuilder) schema(fieldType types.Type, pos token.Pos) *jsonschema.Schema {
	fieldType = types.Unalias(fieldType)
	// json.Number 由标准库特殊编解码，底层 string 无法描述它的 JSON 数值语义。
	if named, ok := fieldType.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "encoding/json" && named.Obj().Name() == "Number" {
		b.scan.diagnostic(pos, "unsupported config type "+fieldType.String())
		return nil
	}
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
		schema := b.schema(underlying.Elem(), pos)
		if schema == nil {
			return nil
		}
		if schema.Type != "" {
			schema.Types, schema.Type = []string{schema.Type, "null"}, ""
		}
		return schema
	case *types.Slice:
		// encoding/json 将字节切片编码为 base64 字符串，数组 Schema 无法描述该语义。
		if element, ok := underlying.Elem().Underlying().(*types.Basic); ok && element.Kind() == types.Uint8 {
			b.scan.diagnostic(pos, "unsupported config type "+fieldType.String())
			return nil
		}
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

func (b *configBuilder) structure(structure *types.Struct) *jsonschema.Schema {
	schema := &jsonschema.Schema{Type: "object", Properties: make(map[string]*jsonschema.Schema)}
	for index := range structure.NumFields() {
		field := structure.Field(index)
		fieldType := field.Type()
		baseType := types.Unalias(fieldType)
		pointer, indirect := baseType.(*types.Pointer)
		if indirect {
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
		for _, key := range []string{"form", "description", "secret", "hidden"} {
			if _, exists := tags[key]; exists {
				b.scan.diagnostic(field.Pos(), fmt.Sprintf("standalone %s tag is unsupported; use the config tag", key))
			}
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
			if _, structure := baseType.Underlying().(*types.Struct); structure {
				child := b.schema(fieldType, field.Pos())
				if child != nil {
					// 展开前检查匿名字段本身的标签；输出沿用子字段的元数据。
					b.metadata(child, fieldType, tags, field.Pos())
					for _, key := range child.PropertyOrder {
						if !field.Exported() && indirect {
							b.unexported[child.Properties[key]] = baseType
						}
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
		child := b.schema(fieldType, field.Pos())
		if child != nil {
			b.fields[child] = fieldType
			if !field.Exported() && indirect {
				b.unexported[child] = baseType
			}
			if b.metadata(child, fieldType, tags, field.Pos()) {
				schema.Required = append(schema.Required, name)
			}
			b.property(schema, name, child, field.Pos())
		}
	}
	return schema
}

func (b *configBuilder) property(schema *jsonschema.Schema, name string, child *jsonschema.Schema, pos token.Pos) {
	switch name {
	case "label", "description", "labels", "subtitles":
		b.scan.diagnostic(pos, fmt.Sprintf("config JSON name %q conflicts with translation override keys", name))
	}
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

// configForm 用 JSON 解码器读取 default，保留字符串和容器内的分号与冒号。
func configForm(tag string) (map[string]string, error) {
	form := make(map[string]string)
	for tag != "" {
		key, rest, hasValue := strings.Cut(tag, ":")
		if end := strings.IndexByte(tag, ';'); end >= 0 && (!hasValue || end < len(key)) {
			key, rest, hasValue = tag[:end], tag[end+1:], false
		} else if !hasValue {
			rest = ""
		}
		if _, exists := form[key]; exists {
			return nil, fmt.Errorf("duplicate config key %q", key)
		}
		switch key {
		case "required", "secret", "hidden", "obvious":
			if hasValue {
				return nil, fmt.Errorf("config %s must be a flag", key)
			}
			form[key], tag = "", rest
			continue
		case "label", "description", "widget", "options", "labels", "subtitles", "min", "max", "step", "language", "when", "default":
			if !hasValue || rest == "" || strings.HasPrefix(rest, ";") {
				return nil, fmt.Errorf("config %s requires a value", key)
			}
		default:
			return nil, fmt.Errorf("unknown config key %q", key)
		}
		if key == "default" {
			decoder := json.NewDecoder(strings.NewReader(rest))
			var value json.RawMessage
			if err := decoder.Decode(&value); err != nil {
				return nil, fmt.Errorf("invalid config default: %w", err)
			}
			form[key] = string(value)
			tag = strings.TrimLeft(rest[decoder.InputOffset():], " \t\r\n")
			if tag != "" && tag[0] != ';' {
				return nil, fmt.Errorf("expected semicolon after config default: %q", tag)
			}
			tag = strings.TrimPrefix(tag, ";")
		} else {
			form[key], tag, _ = strings.Cut(rest, ";")
		}
	}
	return form, nil
}

func (b *configBuilder) metadata(schema *jsonschema.Schema, fieldType types.Type, tags map[string]string, pos token.Pos) bool {
	// 1. 只输出显式元数据；翻译路径由表单从 Schema 的嵌套位置推导。
	values, err := configForm(tags["config"])
	if err != nil {
		b.scan.diagnostic(pos, err.Error())
		return false
	}
	form := make(map[string]any)
	schema.Extra = map[string]any{config.KeywordForm: form}
	schema.Description = values["description"]
	for _, entry := range []struct{ tag, keyword string }{{"secret", config.KeywordSecret}, {"hidden", config.KeywordHidden}} {
		if _, exists := values[entry.tag]; exists {
			if entry.tag == "secret" && !configSecretType(fieldType) {
				b.scan.diagnostic(pos, "secret requires string or *string")
			}
			schema.Extra[entry.keyword] = true
		}
	}
	for _, key := range []string{"label", "widget", "language", "when"} {
		if value, exists := values[key]; exists {
			form[key] = value
		}
	}
	if _, exists := values["obvious"]; exists {
		form["obvious"] = true
	}
	// 多选的取值类型与 Schema 来自 slice 元素，控件和数字边界仍属于字段本身。
	optionType, optionSchema := fieldType, schema
	for {
		pointer, ok := optionType.Underlying().(*types.Pointer)
		if !ok {
			break
		}
		optionType = pointer.Elem()
	}
	if slice, ok := optionType.Underlying().(*types.Slice); ok {
		optionType, optionSchema = slice.Elem(), schema.Items
	}
	basic := configBasic(fieldType)
	if widget, ok := values["widget"]; ok {
		valid := false
		switch widget {
		case "select":
			choice := configBasic(optionType)
			valid = choice != nil && choice.Info()&(types.IsString|types.IsInteger|types.IsFloat) != 0
			if _, exists := values["options"]; !exists {
				b.scan.diagnostic(pos, "select widget requires options")
			}
		case "text", "password", "textarea", "code":
			valid = basic != nil && basic.Info()&types.IsString != 0
		case "number", "slider":
			valid = basic != nil && basic.Info()&(types.IsInteger|types.IsFloat) != 0
		case "switch":
			valid = basic != nil && basic.Info()&types.IsBoolean != 0
		}
		if !valid {
			b.scan.diagnostic(pos, fmt.Sprintf("widget %q is incompatible with %s", widget, fieldType))
		}
		if widget == "slider" {
			for _, key := range []string{"min", "max"} {
				if _, exists := values[key]; !exists {
					b.scan.diagnostic(pos, fmt.Sprintf("slider widget requires %s", key))
				}
			}
		}
		if widget == "code" {
			if _, exists := values["language"]; !exists {
				form["language"] = "json"
			}
		}
	}
	if _, exists := values["language"]; exists && values["widget"] != "code" {
		b.scan.diagnostic(pos, "language requires the code widget")
	}
	// 2. 表单选项与边界保留数值原文；默认值按十进制位数和有效数字精确比较。
	for _, key := range []string{"min", "max", "step"} {
		if raw, ok := values[key]; ok {
			value, err := configJSON(raw)
			number, numeric := value.(json.Number)
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
		choice := configBasic(optionType)
		if choice == nil || choice.Info()&(types.IsString|types.IsInteger|types.IsFloat) == 0 {
			b.scan.diagnostic(pos, "form options require a string or numeric field or slice element")
		} else {
			var choices []any
			for _, option := range strings.Split(options, "|") {
				value, err := b.literal(optionType, optionSchema, option)
				if err != nil {
					b.scan.diagnostic(pos, fmt.Sprintf("invalid form option %q: %v", option, err))
					continue
				}
				choices = append(choices, value)
			}
			form["options"] = choices
		}
	}
	for _, key := range []string{"labels", "subtitles"} {
		if raw, ok := values[key]; ok {
			labels := strings.Split(raw, "|")
			options, _ := form["options"].([]any)
			if len(labels) > len(options) {
				b.scan.diagnostic(pos, fmt.Sprintf("config %s has more entries than options", key))
			}
			form[key] = labels
		}
	}
	if len(form) == 0 {
		delete(schema.Extra, config.KeywordForm)
	}
	b.positions[schema] = pos
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

// literal 按源码标量类型解析标签值；字符串原样保留，数字与布尔使用 JSON 字面量。
func (b *configBuilder) literal(fieldType types.Type, schema *jsonschema.Schema, raw string) (any, error) {
	basic := configBasic(fieldType)
	if basic == nil {
		return nil, fmt.Errorf("expected a scalar config field, got %s", fieldType)
	}
	var value any = raw
	var err error
	if basic.Info()&types.IsString == 0 {
		value, err = configJSON(raw)
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, fmt.Errorf("want %s, got null", fieldType)
		}
	}
	if err := b.defaultValue(fieldType, schema, value); err != nil {
		return nil, err
	}
	return value, nil
}

// conditions 在完整 JSON 对象上解析显示条件，因此支持后声明和匿名展开的兄弟字段。
func (b *configBuilder) conditions(schema *jsonschema.Schema) {
	for _, name := range schema.PropertyOrder {
		child := schema.Properties[name]
		form, _ := child.Extra[config.KeywordForm].(map[string]any)
		if raw, ok := form["when"].(string); ok {
			conditions := make(map[string]any)
			for _, condition := range strings.Split(raw, ",") {
				key, value, paired := strings.Cut(condition, "=")
				if !paired {
					b.scan.diagnostic(b.positions[child], fmt.Sprintf("invalid config when condition %q: expected field=value", condition))
					continue
				}
				sibling, exists := schema.Properties[key]
				if !exists {
					b.scan.diagnostic(b.positions[child], fmt.Sprintf("config when references unknown sibling field %q", key))
					continue
				}
				// when 仅比较取值类型，不套用兄弟字段的表单选项或数字边界。
				parsed, err := b.literal(b.fields[sibling], &jsonschema.Schema{}, value)
				if err != nil {
					b.scan.diagnostic(b.positions[child], fmt.Sprintf("invalid config when value for %q: %v", key, err))
					continue
				}
				conditions[key] = parsed
			}
			form["when"] = conditions
		}
		b.conditions(child)
	}
	if schema.Items != nil {
		b.conditions(schema.Items)
	}
	if schema.AdditionalProperties != nil {
		b.conditions(schema.AdditionalProperties)
	}
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
		form, _ := schema.Extra[config.KeywordForm].(map[string]any)
		for index, item := range items {
			if err := b.defaultValue(underlying.Elem(), schema.Items, item); err != nil {
				return fmt.Errorf("item %d: %w", index, err)
			}
			if choices, ok := form["options"].([]any); ok {
				if err := configChoice(item, choices); err != nil {
					return fmt.Errorf("item %d: %w", index, err)
				}
			}
		}
		return nil
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
					if embedded := b.unexported[child]; embedded != nil {
						return fmt.Errorf("field %q: cannot set embedded pointer to unexported struct type %s", item.name, embedded)
					}
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
		return configChoice(value, choices)
	}
	return nil
}

func configChoice(value any, choices []any) error {
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
