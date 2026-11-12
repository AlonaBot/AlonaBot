package convert

import (
	"encoding/json"
	"fmt"
	"reflect"

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
	schema, err := jsonschema.For[T](nil)
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
	schema, err := jsonschema.For[R](nil)
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
func ActionRequestToProto(to message.Session, name string, version uint32, params any, via string) (*pb.ActionRequest, error) {
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
