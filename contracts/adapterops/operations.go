// Package adapterops 定义 Adapter 标准操作的载荷与目录，供 SDK 发现和 Core 分发共用。
package adapterops

import (
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// 标准操作复用已发行的 Adapter RPC 名称。
const (
	// SendRPC 发送消息链。
	SendRPC = "adapter.send"
	// EditRPC 编辑一条平台消息。
	EditRPC = "adapter.edit"
	// DeleteRPC 按顺序撤回消息。
	DeleteRPC = "adapter.delete"
	// MediaReadRPC 读取平台文件分块。
	MediaReadRPC = "adapter.media.read"
	// NodesRPC 读取合并转发节点。
	NodesRPC = "adapter.message.read_nodes"
	// SyncRPC 替换命令菜单。
	SyncRPC = "adapter.commands.sync"
	// MenuRPC 管理一个菜单位置。
	MenuRPC = "adapter.command.menu"
)

// Operation 将公开名称、能力与现有 RPC 载荷绑定在同一处。
type Operation struct {
	RPC        string
	Name       string
	Capability string
	Scope      pb.ActionScope
	Request    func() proto.Message
	Response   func() proto.Message
}

// All 返回标准操作；参数采用相应 protobuf 载荷的 JSON 表示，实例及目标由调用入口绑定。
// 标准名称使用 sdk/ 前缀，与平台 Action 的点分名称保持独立。
func All() []Operation {
	return []Operation{
		{SendRPC, "sdk/message.send", "", pb.ActionScope_ACTION_SCOPE_CHAT, func() proto.Message { return &pb.AdapterSendRequest{} }, func() proto.Message { return &pb.SendResult{} }},
		{EditRPC, "sdk/message.edit", "message.edit", pb.ActionScope_ACTION_SCOPE_CHAT, func() proto.Message { return &pb.AdapterEditRequest{} }, func() proto.Message { return &pb.SendResult{} }},
		{DeleteRPC, "sdk/message.delete", "message.delete", pb.ActionScope_ACTION_SCOPE_CHAT, func() proto.Message { return &pb.AdapterDeleteRequest{} }, func() proto.Message { return &pb.DeleteResult{} }},
		{MediaReadRPC, "sdk/media.read", "", pb.ActionScope_ACTION_SCOPE_ACCOUNT, func() proto.Message { return &pb.AdapterMediaReadRequest{} }, func() proto.Message { return &pb.AdapterMediaReadResponse{} }},
		{NodesRPC, "sdk/message.read_nodes", "", pb.ActionScope_ACTION_SCOPE_CHAT, func() proto.Message { return &pb.AdapterReadNodesRequest{} }, func() proto.Message { return &pb.ReadNodesResponse{} }},
		{SyncRPC, "sdk/commands.sync", "commands.sync", pb.ActionScope_ACTION_SCOPE_ACCOUNT, func() proto.Message { return &pb.AdapterCommandsSyncRequest{} }, func() proto.Message { return &pb.AdapterCommandsSyncResponse{} }},
		{MenuRPC, "sdk/command.menu", "commands.sync", pb.ActionScope_ACTION_SCOPE_ACCOUNT, func() proto.Message { return &pb.CommandMenuRequest{} }, func() proto.Message { return &pb.CommandMenuResponse{} }},
	}
}

// Find 查找标准操作，保留平台动作使用自身声明的路径。
func Find(name string) (Operation, bool) {
	for _, op := range All() {
		if op.Name == name {
			return op, true
		}
	}
	return Operation{}, false
}

// Declaration 从实际 RPC 类型生成操作的参数与结果 Schema。
func (o Operation) Declaration() *pb.ActionDeclaration {
	return &pb.ActionDeclaration{Name: o.Name, Version: 1, Scope: o.Scope, ParamsSchemaJson: schema(o.Request(), true), ResultSchemaJson: schema(o.Response(), false)}
}

func schema(m proto.Message, params bool) string {
	defs := map[string]any{}
	var object func(protoreflect.MessageDescriptor) map[string]any
	var field func(protoreflect.FieldDescriptor) map[string]any
	field = func(f protoreflect.FieldDescriptor) map[string]any {
		var s map[string]any
		switch f.Kind() {
		case protoreflect.MessageKind:
			s = object(f.Message())
		case protoreflect.BoolKind:
			s = map[string]any{"type": "boolean"}
		case protoreflect.StringKind, protoreflect.BytesKind:
			s = map[string]any{"type": "string"}
		case protoreflect.EnumKind:
			s = map[string]any{"type": []string{"string", "integer"}}
		case protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
			s = map[string]any{"type": []string{"string", "integer"}}
		default:
			s = map[string]any{"type": "number"}
		}
		if f.IsMap() {
			return map[string]any{"type": "object", "additionalProperties": field(f.MapValue())}
		}
		if f.IsList() {
			return map[string]any{"type": "array", "items": s}
		}
		return s
	}
	object = func(d protoreflect.MessageDescriptor) map[string]any {
		name := string(d.FullName())
		ref := map[string]any{"$ref": "#/$defs/" + name}
		if _, ok := defs[name]; ok {
			return ref
		}
		props := map[string]any{}
		defs[name] = map[string]any{"type": "object", "properties": props, "additionalProperties": false}
		for i := 0; i < d.Fields().Len(); i++ {
			f := d.Fields().Get(i)
			props[f.JSONName()] = field(f)
		}
		return ref
	}
	root := object(m.ProtoReflect().Descriptor())
	if params {
		d := defs[string(m.ProtoReflect().Descriptor().FullName())].(map[string]any)
		props := d["properties"].(map[string]any)
		delete(props, "instanceId")
		delete(props, "target")
	}
	root["$defs"] = defs
	raw, _ := json.Marshal(root)
	return string(raw)
}
