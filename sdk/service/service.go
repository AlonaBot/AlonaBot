// Package service 声明、注册与调用带类型的 JSON 和 Protobuf 服务。
//
// [Spec] 绑定服务目标与类型，[Serve] 构建访问规则和 Handler；调用继承有效父请求。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"github.com/AlonaBot/AlonaBot/sdk/tool"
	"google.golang.org/protobuf/proto"
)

// Access 指定服务允许的直接调用方。
type Access int

const (
	// AccessPrivate 只允许 Core 和本插件调用，作为默认访问规则。
	AccessPrivate Access = Access(registry.AccessPrivate)
	// AccessAllowlist 允许 Core、本插件和明确列出的插件调用。
	AccessAllowlist Access = Access(registry.AccessAllowlist)
	// AccessPublic 允许已启用插件调用。
	AccessPublic Access = Access(registry.AccessPublic)
)

// Spec 绑定服务目标与输入输出类型；零值无效。
type Spec[In, Out any] struct {
	target   string
	encoding pb.PayloadEncoding
}

// JSON 声明 JSON 服务；调用使用完整目标，服务端可用 /name 绑定本插件身份，Schema 由生成器提供。
func JSON[In, Out any](target string) Spec[In, Out] {
	return Spec[In, Out]{target: target, encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON}
}

// Proto 声明完整目标的 Protobuf 服务，Schema 版本保持为 1。
func Proto[In, Out proto.Message](target string) Spec[In, Out] {
	return Spec[In, Out]{target: target, encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF}
}

// Builder 构建声明的访问规则与 Handler；Handle 结束注册。
type Builder[In, Out any] struct {
	builder          *registry.ServiceBuilder[*sdk.Context, sdk.Permission]
	handler          func(*sdk.Context, In) (Out, error)
	toolDescription  *string // nil 表示未声明为工具，空说明留给注册校验拒绝。
	toolDefault      tool.Default
	toolEffect       tool.Effect
	effectBeforeTool bool // 保留调用顺序错误，后续 Tool 调用也不会消除。
}

// Serve 开始在插件或子插件上注册声明，运行前统一检查构建结果。
func Serve[In, Out any](owner sdk.Registrar, declaration Spec[In, Out]) *Builder[In, Out] {
	scope := registry.For[*sdk.Context, sdk.Permission](owner)
	scope.Check()
	_, name, _ := strings.Cut(declaration.target, "/")
	b := &registry.ServiceBuilder[*sdk.Context, sdk.Permission]{Scope: scope, Target: declaration.target, Name: name, Encoding: declaration.encoding}
	if strings.HasPrefix(declaration.target, "/") {
		b.Target = ""
	}
	s := &Builder[In, Out]{builder: b}
	b.Prepare = func() (registry.Service[*sdk.Context], error) {
		if s.effectBeforeTool {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("Effect must be called after Tool")
		}
		if s.toolEffect < tool.Unspecified || s.toolEffect > tool.Destructive {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("unknown tool effect %d", s.toolEffect)
		}
		if b.Encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
			if s.toolDescription != nil {
				return registry.Service[*sdk.Context]{}, fmt.Errorf("Protobuf service cannot be a tool")
			}
			return s.prepareProto()
		}
		return s.prepareJSON()
	}
	scope.Services.Builders = append(scope.Services.Builders, b)
	return s
}

func (s *Builder[In, Out]) prepareJSON() (registry.Service[*sdk.Context], error) {
	b := s.builder
	r := b.Scope.Services

	// 1. 解析输入输出 Schema 并冻结 Handler，调用时复用校验器。
	inputID, inputVersion, err := registry.JSONType[In]()
	if err != nil {
		return registry.Service[*sdk.Context]{}, err
	}
	outputID, outputVersion, err := registry.JSONType[Out]()
	if err != nil {
		return registry.Service[*sdk.Context]{}, err
	}
	inputID = registry.SchemaID(r.Schemas, b.Scope.Entry(), inputID)
	outputID = registry.SchemaID(r.Schemas, b.Scope.Entry(), outputID)
	inputSchema, err := registry.ResolveSchema(r.Schemas, inputID)
	if err != nil {
		return registry.Service[*sdk.Context]{}, err
	}
	outputSchema, err := registry.ResolveSchema(r.Schemas, outputID)
	if err != nil {
		return registry.Service[*sdk.Context]{}, err
	}
	handler := s.handler // 收集后冻结 Handler，保持 Session 注册固定。
	if handler == nil {
		return registry.Service[*sdk.Context]{}, fmt.Errorf("handler is nil")
	}
	route := b.Route(inputID, outputID, inputVersion, outputVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, func(c *sdk.Context, data []byte) ([]byte, error) {
		if err := registry.ValidateJSON(inputSchema, data); err != nil {
			return nil, &sdk.Error{Code: errcode.InvalidArgument, Message: "invalid service input"}
		}
		var input In
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, &sdk.Error{Code: errcode.InvalidArgument, Message: "decode service input: " + err.Error()}
		}
		out, handlerErr := handler(c, input)
		encoded, err := json.Marshal(out)
		if err != nil {
			return nil, errors.Join(handlerErr, fmt.Errorf("encode service output: %w", err))
		}
		if err := registry.ValidateJSON(outputSchema, encoded); err != nil {
			c.Logger().Error("service output schema validation failed", "schema_id", outputID)
			return nil, errors.Join(handlerErr, &sdk.Error{Code: errcode.Internal, Message: "invalid service output"})
		}
		return encoded, handlerErr
	})
	// 2. 工具描述携带同一输入 Schema 原文，完整注册校验交给注册收集阶段。
	if s.toolDescription != nil {
		var policy pb.ToolPolicy
		switch s.toolDefault {
		case tool.DefaultAllow:
			policy = pb.ToolPolicy_TOOL_POLICY_ALLOW
		case tool.DefaultConfirm:
			policy = pb.ToolPolicy_TOOL_POLICY_CONFIRM
		default:
			return registry.Service[*sdk.Context]{}, fmt.Errorf("unknown tool default policy %d", s.toolDefault)
		}
		route.Descriptor.GetService().Tool = &pb.ToolSpec{Description: *s.toolDescription, DefaultPolicy: policy, InputSchemaJson: string(r.Schemas[inputID]), Effect: pb.ToolEffect(s.toolEffect)}
	}
	return route, nil
}

func (s *Builder[In, Out]) prepareProto() (registry.Service[*sdk.Context], error) {
	b := s.builder

	handler := s.handler // 收集后冻结 Handler，保持 Session 注册固定。
	if handler == nil {
		return registry.Service[*sdk.Context]{}, fmt.Errorf("handler is nil")
	}
	input, output := registry.NewMessage[In](), registry.NewMessage[Out]()
	return b.Route(validate.SchemaID(input), validate.SchemaID(output), registry.SchemaVersion, registry.SchemaVersion, pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, func(c *sdk.Context, data []byte) ([]byte, error) {
		input := registry.NewMessage[In]()
		if err := proto.Unmarshal(data, input); err != nil {
			return nil, &sdk.Error{Code: errcode.InvalidArgument, Message: "decode service input: " + err.Error()}
		}
		output, handlerErr := handler(c, input.(In))
		encoded, err := proto.Marshal(any(output).(proto.Message))
		if err != nil {
			return nil, errors.Join(handlerErr, fmt.Errorf("encode service output: %w", err))
		}
		return encoded, handlerErr
	}), nil
}

// Access 设置直接调用方规则，重复调用采用最后一次。
func (s *Builder[In, Out]) Access(a Access, allowlist ...string) *Builder[In, Out] {
	setAccess(s.builder, a, allowlist)
	return s
}

// Permission 覆盖服务要求的权限，满足其中任一权限名即可；缺省不要求权限。
func (s *Builder[In, Out]) Permission(p sdk.Permission) *Builder[In, Out] {
	s.builder.Scope.Check()
	if s.builder.Handles != 0 {
		s.builder.AfterHandle = true
	}
	s.builder.Permission = p
	return s
}

// Tool 设置工具说明与缺省策略；Proto 声明或 Handle 后调用在 Run 收集注册时报错。
func (s *Builder[In, Out]) Tool(description string, defaultPolicy tool.Default) *Builder[In, Out] {
	s.builder.Scope.Check()
	if s.builder.Handles != 0 {
		s.builder.AfterHandle = true
	}
	s.toolDescription, s.toolDefault = &description, defaultPolicy
	return s
}

// Effect 设置工具的副作用提示；Tool 之前、Handle 之后调用或类别非法时在收集注册时报错。
func (s *Builder[In, Out]) Effect(e tool.Effect) *Builder[In, Out] {
	s.builder.Scope.Check()
	if s.builder.Handles != 0 {
		s.builder.AfterHandle = true
	}
	s.effectBeforeTool = s.effectBeforeTool || s.toolDescription == nil
	s.toolEffect = e
	return s
}

// Handle 结束服务注册，冻结的 Handler 在 Session 内按声明类型调用。
func (s *Builder[In, Out]) Handle(h func(*sdk.Context, In) (Out, error)) {
	s.builder.Scope.Check()
	s.builder.Handles++
	s.handler = h
}

func setAccess(b *registry.ServiceBuilder[*sdk.Context, sdk.Permission], a Access, allowlist []string) {
	b.Scope.Check()
	if b.Handles != 0 {
		b.AfterHandle = true
	}
	b.Access, b.Allowlist = registry.Access(a), slices.Clone(allowlist)
}

// Call 调用声明的服务，JSON 输入输出按生成的 Schema 校验。
// Handler 中传入 Context.Context() 以关联父请求；后台工作使用独立 context。
func (s Spec[In, Out]) Call(ctx context.Context, in In) (Out, error) {
	if s.target == "" || s.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
		var out Out
		return out, errcode.InvalidArgument.Errorf("invalid service declaration %q", s.target)
	}
	if s.encoding == pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
		return s.callProto(ctx, in)
	}
	return s.callJSON(ctx, in)
}

func (s Spec[In, Out]) callJSON(ctx context.Context, in In) (Out, error) {
	var out Out
	inputID, inputVersion, err := registry.JSONType[In]()
	if err != nil {
		return out, err
	}
	outputID, outputVersion, err := registry.JSONType[Out]()
	if err != nil {
		return out, err
	}
	session, err := registry.Session(ctx)
	if err != nil {
		return out, err
	}
	_, entry, schemas := registry.Runtime(ctx)
	inputID = registry.SchemaID(schemas, entry, inputID)
	outputID = registry.SchemaID(schemas, entry, outputID)
	inputSchema, err := registry.ResolveSchema(schemas, inputID)
	if err != nil {
		return out, err
	}
	outputSchema, err := registry.ResolveSchema(schemas, outputID)
	if err != nil {
		return out, err
	}
	encoded, err := json.Marshal(in)
	if err != nil {
		return out, &sdk.Error{Code: errcode.InvalidArgument, Message: "encode service input: " + err.Error()}
	}
	if err := registry.ValidateJSON(inputSchema, encoded); err != nil {
		// 校验器错误包含载荷原值，公开错误与日志只保留声明信息。
		registry.Logger(ctx).DebugContext(ctx, fmt.Sprintf("Input for service %s does not match schema %s", s.target, inputID), "target", s.target, "schema_id", inputID)
		return out, errcode.InvalidArgument.Errorf("invalid service input for %q", s.target)
	}
	target := s.target
	payload := wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: inputID, SchemaVersion: inputVersion, Bytes: encoded}
	if stream, ok := ctx.Value(streamCallKey{}).(streamCall); ok {
		request := &pb.ServiceStreamCallRequest{Target: target, StreamId: stream.id, Encoding: payload.Encoding, SchemaId: inputID, SchemaVersion: inputVersion, Payload: encoded}
		data, err := proto.Marshal(request)
		if err != nil {
			return out, err
		}
		payload = wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: data}
		target = "core.service.stream.call"
	}
	response, callErr := registry.Call(ctx, session, target, payload)
	if response.SchemaID == "" && callErr != nil {
		return out, callErr
	}
	if response.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || response.SchemaID != outputID || response.SchemaVersion != outputVersion {
		return out, errors.Join(callErr, errcode.Internal.New("expected JSON service output"))
	}
	if err := registry.ValidateJSON(outputSchema, response.Bytes); err != nil {
		registry.Logger(ctx).DebugContext(ctx, fmt.Sprintf("Output from service %s does not match schema %s", s.target, outputID), "target", s.target, "schema_id", outputID)
		return out, errors.Join(callErr, errcode.Internal.Errorf("invalid service output for %q", s.target))
	}
	if err := json.Unmarshal(response.Bytes, &out); err != nil {
		return out, errors.Join(callErr, &sdk.Error{Code: errcode.Internal, Message: "decode service output: " + err.Error()})
	}
	return out, callErr
}

func (s Spec[In, Out]) callProto(ctx context.Context, in In) (Out, error) {
	return registry.Proto[In, Out](s.target).Call(ctx, in)
}

// Reentrant 允许通过显式声明的服务回调当前插件；调用仍受授权与深度检查。
func (s *Builder[In, Out]) Reentrant() *Builder[In, Out] {
	s.builder.Scope.Check()
	if s.builder.Handles != 0 {
		s.builder.AfterHandle = true
	}
	s.builder.Reentrant = true
	return s
}

// Priority 设置发现与执行排序，数值越小越先。
func (s *Builder[In, Out]) Priority(priority int32) *Builder[In, Out] {
	s.builder.Scope.Check()
	if s.builder.Handles != 0 {
		s.builder.AfterHandle = true
	}
	s.builder.Priority = priority
	return s
}
