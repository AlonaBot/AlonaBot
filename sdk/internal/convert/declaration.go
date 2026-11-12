package convert

import (
	"encoding/json"
	"fmt"
	"reflect"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
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
	return out, nil
}
