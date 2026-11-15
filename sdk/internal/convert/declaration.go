package convert

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/google/jsonschema-go/jsonschema"
)

type declaration struct {
	name    string
	version uint32
	schema  string
	err     error
}

// ExtensionDeclaration 保存类型推导的 extension 声明，字段只由构造入口设置。
type ExtensionDeclaration struct{ declaration }

// EventDeclaration 保存类型推导的 Adapter 事件声明。
type EventDeclaration struct{ declaration }

// ActionDeclaration 封装动作参数和结果的类型推导与声明错误。
type ActionDeclaration struct {
	params declaration
	result declaration
	scope  int
}

// PermissionDeclaration 保存平台权限及其隐含关系和文案键。
type PermissionDeclaration struct {
	Name     string
	Implies  []string
	LabelKey string
}

// Descriptor 保存 Adapter 的公共声明，供 adapter 包复用。
type Descriptor struct {
	AdapterID    string
	Capabilities []message.Capability
	Extensions   []ExtensionDeclaration
	Permissions  []PermissionDeclaration
	Events       []EventDeclaration
	Actions      []ActionDeclaration
}

func declare[T any](name string, version uint32) declaration {
	d := declaration{name: name, version: version}
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		d.err = fmt.Errorf("declaration %q payload must be a struct", name)
		return d
	}
	schema, err := declarationSchema[T]()
	if err != nil {
		d.err = fmt.Errorf("declaration %q schema: %w", name, err)
		return d
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		d.err = fmt.Errorf("declaration %q schema: %w", name, err)
		return d
	}
	d.schema = string(raw)
	return d
}

// DeclareExtension 推导 extension 载荷的 JSON Schema。
func DeclareExtension[T any](name string, version uint32) ExtensionDeclaration {
	return ExtensionDeclaration{declare[T](name, version)}
}

// DeclareEvent 推导扩展事件载荷的 JSON Schema。
func DeclareEvent[T any](name string, version uint32) EventDeclaration {
	return EventDeclaration{declare[T](name, version)}
}

// DeclareAction 推导动作 Schema；参数须为结构体，结果允许额外对象属性。
func DeclareAction[P, R any](name string, version uint32, scope int) ActionDeclaration {
	d := ActionDeclaration{params: declare[P](name, version), result: declaration{name: name, version: version}, scope: scope}
	schema, err := declarationSchema[R]()
	if err != nil {
		d.result.err = fmt.Errorf("action %q result schema: %w", name, err)
		return d
	}
	allowResultProperties(schema)
	raw, err := json.Marshal(schema)
	if err != nil {
		d.result.err = fmt.Errorf("action %q result schema: %w", name, err)
		return d
	}
	d.result.schema = string(raw)
	return d
}

func declarationSchema[T any]() (*jsonschema.Schema, error) {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		return nil, err
	}
	if err := matchJSONEncoding(reflect.TypeFor[T](), schema); err != nil {
		return nil, err
	}
	return schema, nil
}

// matchJSONEncoding 补齐推导库未处理的 JSON 编码规则；schema 已由同一类型成功推导。
func matchJSONEncoding(t reflect.Type, schema *jsonschema.Schema) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// 推导库已为这两个标准类型提供与其自定义编码一致的 Schema。
	if t == reflect.TypeFor[time.Time]() || t == reflect.TypeFor[slog.Level]() {
		return nil
	}
	for _, encoder := range []reflect.Type{reflect.TypeFor[json.Marshaler](), reflect.TypeFor[encoding.TextMarshaler]()} {
		if t.Implements(encoder) || reflect.PointerTo(t).Implements(encoder) {
			return fmt.Errorf("cannot infer JSON encoding for %v implementing %v", t, encoder)
		}
	}
	switch t.Kind() {
	case reflect.Struct:
		for _, field := range reflect.VisibleFields(t) {
			if field.Anonymous || !field.IsExported() || field.Tag.Get("json") == "-" {
				continue
			}
			name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "" {
				name = field.Name
			}
			property := schema.Properties[name]
			if err := matchJSONEncoding(field.Type, property); err != nil {
				return fmt.Errorf("field %v.%s: %w", t, field.Name, err)
			}
			quoted := field.Type
			if quoted.Name() == "" && quoted.Kind() == reflect.Pointer {
				quoted = quoted.Elem()
			}
			if slices.Contains(strings.Split(options, ","), "string") {
				switch quoted.Kind() {
				case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
					reflect.Float32, reflect.Float64, reflect.String:
					encodedStringSchema(property, field.Type.Kind() == reflect.Pointer)
				}
			}
		}
	case reflect.Slice, reflect.Array:
		if err := matchJSONEncoding(t.Elem(), schema.Items); err != nil {
			return err
		}
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			encodedStringSchema(schema, true)
		}
	case reflect.Map:
		// nil map 与 nil slice 一样编码为 null。
		schema.Type = ""
		schema.Types = []string{"null", "object"}
		return matchJSONEncoding(t.Elem(), schema.AdditionalProperties)
	}
	return nil
}

func encodedStringSchema(schema *jsonschema.Schema, nullable bool) {
	*schema = jsonschema.Schema{Type: "string", Description: schema.Description}
	if nullable {
		schema.Type = ""
		schema.Types = []string{"null", "string"}
	}
}

// allowResultProperties 只移除推导给结构体的额外属性禁令，保留 map 值的类型约束。
func allowResultProperties(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Not != nil {
		schema.AdditionalProperties = nil
	}
	for _, property := range schema.Properties {
		allowResultProperties(property)
	}
	allowResultProperties(schema.Items)
	allowResultProperties(schema.AdditionalProperties)
}

// DescriptorToProto 在 REGISTER 前报告声明错误并编码 Schema。
func DescriptorToProto(d *Descriptor) (*pb.AdapterDescriptor, error) {
	out := &pb.AdapterDescriptor{AdapterId: d.AdapterID, Capabilities: CapabilitiesToProto(d.Capabilities)}
	for _, e := range d.Extensions {
		if e.err != nil {
			return nil, e.err
		}
		out.Extensions = append(out.Extensions, &pb.ExtensionDeclaration{Name: e.name, Version: e.version, SchemaJson: e.schema})
	}
	for _, e := range d.Events {
		if e.err != nil {
			return nil, e.err
		}
		out.Events = append(out.Events, &pb.EventDeclaration{Name: e.name, Version: e.version, SchemaJson: e.schema})
	}
	for _, p := range d.Permissions {
		out.Permissions = append(out.Permissions, &pb.PermissionDeclaration{Name: p.Name, Implies: p.Implies, LabelKey: p.LabelKey})
	}
	for _, a := range d.Actions {
		if a.params.err != nil {
			return nil, a.params.err
		}
		if a.result.err != nil {
			return nil, a.result.err
		}
		out.Actions = append(out.Actions, &pb.ActionDeclaration{Name: a.params.name, Version: a.params.version, Scope: pb.ActionScope(a.scope), ParamsSchemaJson: a.params.schema, ResultSchemaJson: a.result.schema})
	}
	return out, nil
}

// ActionRequestToProto 编码动作参数，按会话是否为空把 Via 放入相应目标字段。
func ActionRequestToProto(to string, name string, version uint32, params any, via string) (*pb.ActionRequest, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, errcode.InvalidArgument.New("encode action params: " + err.Error())
	}
	request := &pb.ActionRequest{Name: name, Version: version, ParamsJson: string(raw)}
	if to == "" {
		request.AdapterInstanceId = via
		return request, nil
	}
	request.To, err = TargetToProto(to, via)
	return request, err
}
