package router

import (
	"context"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// RegisterCommandList 将当前路由器的命令列表服务接入固定服务分流器。
func (r *Router) RegisterCommandList(m *Mux) {
	m.Handle(wire.ServiceCommandList, r.listCommands)
}

func (r *Router) listCommands(ctx context.Context, _ Caller, in *wire.Inbound) wire.Response {
	// 1. Session 已校验父请求仍在途；身份只从 Core 保存的父附件取得。
	parent, ok := in.ParentAttach.(*Invocation)
	if in.ParentCallID == 0 || !ok || parent == nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "command list requires an active invocation"}}
	}
	request := &pb.CommandListRequest{}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != handlerSchemaVersion {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "expected CommandListRequest protobuf schema version 1"}}
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "decode command list request: " + err.Error()}}
	}

	// 2. 索引与路由在同一临界区取快照，策略、翻译与编码在锁外执行。
	r.mu.Lock()
	index := r.commands
	var routes []commandRoute
	for _, cached := range r.matches {
		routes = append(routes, cached...)
	}
	r.mu.Unlock()
	actor := parent.Event.GetActor()
	result := &pb.CommandListResponse{}
	for _, route := range routes {
		command := route.route.GetCommand()
		if command.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH || !route.plugin.Activated() {
			continue
		}
		names := index.names[route.pluginID][route.route.GetRouteId()]
		if len(names) == 0 {
			continue
		}
		winner, _, ok := index.Lookup(names[0])
		identity := CommandEntry{PluginID: route.pluginID, RouteID: route.route.GetRouteId()}
		if !ok || !sameRoute(winner, identity) {
			continue
		}
		_, enabled, err := r.Profile(ctx, route.pluginID, parent.DefaultTarget)
		if err != nil {
			return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
		}
		if !enabled {
			continue
		}
		if r.policy != nil {
			if failure := r.policy.CheckCommand(ctx, route.pluginID, route.route, actor, parent.DefaultTarget); failure != nil {
				switch failure.GetCode() {
				case pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION:
					continue
				default:
					return wire.Response{Err: failure}
				}
			}
		} else if !route.route.GetDefaultEnabled() || !r.permitted(route.candidate, parent.Event) {
			continue
		}
		entry := &pb.CommandInfo{PluginId: route.pluginID, RouteId: route.route.GetRouteId(), Path: names[0]}
		for _, name := range names[1:] {
			if winner, _, ok := index.Lookup(name); ok && sameRoute(winner, identity) {
				entry.Aliases = append(entry.Aliases, &pb.CommandAlias{Path: name})
			}
		}
		if source := command.GetDescription(); source != "" {
			description := source
			if catalog := route.plugin.Messages(); catalog != nil {
				description = catalog.CommandDescription([]string{actor.GetLocale(), route.plugin.DefaultLocale()}, command.GetPath(), source)
			}
			entry.Description = proto.String(description)
		}
		result.Commands = append(result.Commands, entry)
	}

	// 3. 别名保留配置顺序，目录按生效主名排序后统一编码。
	slices.SortFunc(result.Commands, func(a, b *pb.CommandInfo) int {
		return strings.Compare(strings.Join(a.GetPath(), " "), strings.Join(b.GetPath(), " "))
	})
	encoded, err := proto.Marshal(result)
	if err != nil {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "encode command list response: " + err.Error()}}
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(result), SchemaVersion: handlerSchemaVersion, Bytes: encoded}}
}
