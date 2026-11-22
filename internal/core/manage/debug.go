package manage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/adapterops"
	"github.com/AlonaBot/AlonaBot/contracts/debugapi"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const debugTraceBytes = 16

func (s *Service) debugRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/debug/events", func(w http.ResponseWriter, r *http.Request) {
		var request debugapi.EventRequest
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		event := &pb.Event{}
		if err := protojson.Unmarshal(request.Event, event); err != nil {
			respond(w, nil, invalid(err.Error()))
			return
		}
		if s.opt.DebugEvent == nil {
			respond(w, nil, precondition("debug event service is unavailable"))
			return
		}
		id := make([]byte, debugTraceBytes)
		rand.Read(id)
		event.SourceEventId = proto.String("debug:" + hex.EncodeToString(id))
		result, err := s.opt.DebugEvent(r.Context(), event)
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/debug/adapters/{instance}/operations", func(w http.ResponseWriter, r *http.Request) {
		_, ops, err := s.debugOperations(r.Context(), r.PathValue("instance"))
		respond(w, ops, err)
	})
	mux.HandleFunc("POST /api/v1/debug/adapters/call", func(w http.ResponseWriter, r *http.Request) {
		var request debugapi.Call
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		result, err := s.debugCall(r.Context(), request)
		respond(w, result, err)
	})
}

func (s *Service) debugOperations(ctx context.Context, id string) (adapter.Instance, []debugapi.Operation, error) {
	ctx, instance, ok := s.opt.Adapters.Pin(ctx, id)
	if !ok {
		return nil, nil, precondition("adapter instance is not running")
	}
	catalog, ok := instance.(adapter.OperationCatalog)
	if !ok {
		return nil, nil, precondition("adapter instance has no SDK operation catalog")
	}
	response, err := catalog.Operations(ctx)
	if err != nil {
		return nil, nil, err
	}
	_, ok = s.opt.Adapters.Resolve(ctx, id)
	if !ok {
		return nil, nil, precondition("adapter instance changed during discovery")
	}
	result := make([]debugapi.Operation, 0, len(response.Operations))
	for _, op := range response.Operations {
		result = append(result, debugapi.Operation{Name: op.Name, Version: op.Version, Scope: strings.TrimPrefix(op.Scope.String(), "ACTION_SCOPE_"), ParamsSchema: json.RawMessage(op.ParamsSchemaJson), ResultSchema: json.RawMessage(op.ResultSchemaJson)})
	}
	return instance, result, nil
}

func (s *Service) debugCall(ctx context.Context, request debugapi.Call) (debugapi.Result, error) {
	ctx, _, pinned := s.opt.Adapters.Pin(ctx, request.Instance)
	if !pinned {
		return debugapi.Result{}, precondition("adapter instance is not running")
	}
	// 1. 目录从当前 Session 读取，操作与目标均在产生副作用前绑定。
	instance, operations, err := s.debugOperations(ctx, request.Instance)
	if err != nil {
		return debugapi.Result{}, err
	}
	var operation *debugapi.Operation
	for _, op := range operations {
		if op.Name == request.Name && op.Version == request.Version {
			operation = &op
			break
		}
	}
	if operation == nil {
		return debugapi.Result{}, invalid("operation name or version is not available")
	}
	if !json.Valid(request.Params) {
		return debugapi.Result{}, invalid("params must contain one JSON value")
	}
	target := &pb.Target{AdapterId: instance.AdapterID(), AdapterInstanceId: instance.InstanceID(), AccountId: instance.AccountID()}
	if len(request.To) > 0 {
		if err := protojson.Unmarshal(request.To, target); err != nil {
			return debugapi.Result{}, invalid(err.Error())
		}
		if target.AdapterInstanceId != "" && target.AdapterInstanceId != request.Instance || target.AdapterId != "" && target.AdapterId != instance.AdapterID() || target.AccountId != "" && target.AccountId != instance.AccountID() {
			return debugapi.Result{}, invalid("target does not match instance")
		}
		target.AdapterId = instance.AdapterID()
		target.AdapterInstanceId = instance.InstanceID()
		target.AccountId = instance.AccountID()
	}
	if operation.Scope == "CHAT" && target.ChatId == "" {
		return debugapi.Result{}, invalid("operation requires to with chatId")
	}
	// 2. 管理认证授予本次目标；沿用业务服务的会话禁止、Hook 和媒体授权。
	trace := make([]byte, debugTraceBytes)
	rand.Read(trace)
	ctx = wire.WithTraceID(ctx, trace)
	caller := router.NewCaller("@debug", "")
	invocation := &router.Invocation{PluginID: caller.PluginID, DefaultTarget: target, Event: &pb.Event{AdapterInstanceId: proto.String(request.Instance)}, TraceID: trace}
	release := invocation.Begin()
	defer release()
	if s.opt.DebugRelease != nil {
		defer s.opt.DebugRelease(caller)
	}
	var response wire.Response
	var output proto.Message = &pb.ActionResponse{}
	standard, ok := adapterops.Find(request.Name)
	if ok {
		params := standard.Request()
		output = standard.Response()
		if err := protojson.Unmarshal(request.Params, params); err != nil {
			return debugapi.Result{}, invalid(err.Error())
		}
		// 实例和目标只有管理入口可以绑定，参数中的同名字段拒绝覆盖。
		fields := params.ProtoReflect().Descriptor().Fields()
		for _, name := range []string{"instance_id", "target"} {
			if f := fields.ByName(protoreflect.Name(name)); f != nil && params.ProtoReflect().Has(f) {
				return debugapi.Result{}, invalid("instance and target must be provided outside params")
			}
		}
		var service string
		var message proto.Message
		switch p := params.(type) {
		case *pb.AdapterSendRequest:
			service = wire.ServiceSend
			message = &pb.SendRequest{To: target, Chain: p.Chain}
		case *pb.AdapterEditRequest:
			service = wire.ServiceEdit
			message = &pb.EditRequest{To: target, PlatformMessageId: p.PlatformMessageId, Chain: p.Chain}
		case *pb.AdapterDeleteRequest:
			service = wire.ServiceDelete
			message = &pb.DeleteRequest{To: target, PlatformMessageIds: p.PlatformMessageIds}
		case *pb.AdapterReadNodesRequest:
			service = wire.ServiceReadNodes
			message = &pb.ReadNodesRequest{To: target, Segment: p.Segment}
		case *pb.CommandMenuRequest:
			p.InstanceId = request.Instance
			if s.opt.DebugMenu == nil {
				return debugapi.Result{}, precondition("debug command menu service is unavailable")
			}
			response = s.opt.DebugMenu(ctx, instance, p)
		case *pb.AdapterCommandsSyncRequest:
			sync, ok := instance.(adapter.CommandSynchronizer)
			if !ok {
				return debugapi.Result{}, precondition("command synchronizer is unavailable")
			}
			response.Err = sync.SyncCommands(ctx, p.Commands)
			response.Payload = debugPayload(&pb.AdapterCommandsSyncResponse{})
		case *pb.AdapterMediaReadRequest:
			if p.Length == 0 || p.Length > wire.MaxMediaChunk {
				return debugapi.Result{}, invalid("invalid media read length")
			}
			reader, ok := instance.(adapter.MediaReader)
			if !ok {
				return debugapi.Result{}, precondition("media reader is unavailable")
			}
			data, next, eof, err := reader.ReadMedia(ctx, p.FileId, p.Offset, p.Length)
			response.Payload = debugPayload(&pb.AdapterMediaReadResponse{Data: data, NextOffset: next, Eof: eof})
			if err != nil {
				response.Err = &pb.ErrorInfo{Code: wire.CodeOf(err), SafeMessage: err.Error()}
			}
		}
		if message != nil {
			if s.opt.DebugCall == nil {
				return debugapi.Result{}, precondition("debug call service is unavailable")
			}
			response = s.opt.DebugCall(ctx, caller, &wire.Inbound{Target: service, Payload: debugPayload(message), TraceID: trace, ParentAttach: invocation})
		}
	} else {
		if operation.Scope == "ACCOUNT" && len(request.To) > 0 {
			return debugapi.Result{}, invalid("account operation does not accept to")
		}
		p := &pb.ActionRequest{Name: request.Name, Version: request.Version, ParamsJson: string(request.Params)}
		if operation.Scope == "CHAT" {
			p.To = target
		} else {
			p.AdapterInstanceId = request.Instance
		}
		if s.opt.DebugCall == nil {
			return debugapi.Result{}, precondition("debug call service is unavailable")
		}
		response = s.opt.DebugCall(ctx, caller, &wire.Inbound{Target: wire.ServiceAction, Payload: debugPayload(p), TraceID: trace, ParentAttach: invocation})
	}
	// 3. 分别编码结果和错误，保留平台部分成功的回执与同一 trace。
	result := debugapi.Result{TraceID: hex.EncodeToString(trace)}
	if response.Payload.SchemaID != "" {
		if err := proto.Unmarshal(response.Payload.Bytes, output); err != nil {
			return result, err
		}
		raw, err := protojson.Marshal(output)
		if err != nil {
			return result, err
		}
		result.Payload = raw
		if action, ok := output.(*pb.ActionResponse); ok && action.ResultJson != "" {
			result.Payload = json.RawMessage(action.ResultJson)
		}
	}
	if response.Err != nil {
		response.Err = proto.CloneOf(response.Err)
		response.Err.TraceId = trace
		raw, err := protojson.Marshal(response.Err)
		if err != nil {
			return result, err
		}
		result.Error = &debugapi.Error{Code: strings.TrimPrefix(response.Err.Code.String(), "ERROR_CODE_"), Message: response.Err.SafeMessage, Details: raw}
	}
	return result, nil
}

func debugPayload(message proto.Message) wire.Payload {
	encoded, _ := proto.Marshal(message)
	return wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: 1, Bytes: encoded}
}
