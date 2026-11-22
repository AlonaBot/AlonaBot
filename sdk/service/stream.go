package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
	"google.golang.org/protobuf/proto"
	"sync"
)

// StreamSpec 绑定流式输入、增量与最终结果 Schema。
type StreamSpec[In, Delta, Out any] struct{ target string }

// StreamJSON 声明 JSON 服务的流式模式，完整模式可使用同目标的 JSON 声明。
func StreamJSON[In, Delta, Out any](target string) StreamSpec[In, Delta, Out] {
	return StreamSpec[In, Delta, Out]{target: target}
}

type streamCallKey struct{}
type streamCall struct{ id string }
type streamReceiverKey struct {
	session *wire.Session
	id      string
}
type streamReceiver struct {
	mu       sync.Mutex
	sequence uint64
	closed   bool
	deliver  func(*pb.ServiceStreamDeltaRequest) error
}

var streamReceivers sync.Map // 以 Session 与随机流 ID 隔离不同组件的调用。

func init() { registry.StreamReceiver = receiveStream }

// Call 等待最终响应并同步确认每段增量；回调错误协作取消父调用。
func (s StreamSpec[In, Delta, Out]) Call(ctx context.Context, in In, onDelta func(Delta) error) (Out, error) {
	var zero Out
	if onDelta == nil {
		return zero, errcode.InvalidArgument.New("stream callback is nil")
	}
	if err := registry.RequireCapability(ctx, CapabilityStream); err != nil {
		return zero, err
	}
	session, err := registry.Session(ctx)
	if err != nil {
		return zero, err
	}
	id := fmt.Sprintf("%x", randomStreamID())
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	schemaID, version, err := registry.JSONType[Delta]()
	if err != nil {
		return zero, err
	}
	_, entry, schemas := registry.Runtime(ctx)
	schemaID = registry.SchemaID(schemas, entry, schemaID)
	schema, err := registry.ResolveSchema(schemas, schemaID)
	if err != nil {
		return zero, err
	}
	var callbackErr error
	receiver := &streamReceiver{}
	receiver.deliver = func(delta *pb.ServiceStreamDeltaRequest) error {
		if receiver.closed {
			return errcode.NotFound.New("stream call is no longer active")
		}
		if delta.Sequence != receiver.sequence {
			return errcode.InvalidArgument.New("stream sequence mismatch")
		}
		if delta.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON || delta.SchemaId != schemaID || delta.SchemaVersion != version {
			return errcode.InvalidArgument.New("stream delta schema mismatch")
		}
		if err := registry.ValidateJSON(schema, delta.Payload); err != nil {
			return errcode.InvalidArgument.Errorf("invalid stream delta: %v", err)
		}
		var value Delta
		if err := json.Unmarshal(delta.Payload, &value); err != nil {
			return err
		}
		if callbackErr != nil {
			return callbackErr
		}
		callbackErr = onDelta(value)
		if callbackErr != nil {
			cancel()
			return callbackErr
		}
		receiver.sequence++
		return nil
	}
	key := streamReceiverKey{session, id}
	streamReceivers.Store(key, receiver)
	defer streamReceivers.Delete(key)
	out, callErr := JSON[In, Out](s.target).Call(context.WithValue(child, streamCallKey{}, streamCall{id}), in)
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.closed = true
	return out, errors.Join(callErr, callbackErr)
}

func randomStreamID() []byte {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return value
}

func receiveStream(ctx context.Context, in *wire.Inbound) wire.Response {
	fail := func(err error) wire.Response {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: err.Error()}}
	}
	if err := registry.RequireCapability(ctx, CapabilityStream); err != nil {
		return fail(err)
	}
	session, err := registry.Session(ctx)
	if err != nil {
		return fail(err)
	}
	request := &pb.ServiceStreamDeltaRequest{}
	if err := registry.Decode(in.Payload, request); err != nil {
		return fail(err)
	}
	value, ok := streamReceivers.Load(streamReceiverKey{session, request.StreamId})
	if !ok {
		return wire.Response{Err: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, SafeMessage: "stream call is no longer active"}}
	}
	receiver := value.(*streamReceiver)
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	if err := receiver.deliver(request); err != nil {
		return fail(err)
	}
	out := &pb.ServiceStreamDeltaResponse{}
	encoded, err := proto.Marshal(out)
	if err != nil {
		return fail(err)
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(out), SchemaVersion: 1, Bytes: encoded}}
}

// ServeStream 注册流式处理器，提供方的 emit 等待对端处理确认才返回。
func ServeStream[In, Delta, Out any](owner sdk.Registrar, declaration StreamSpec[In, Delta, Out]) *StreamBuilder[In, Delta, Out] {
	return &StreamBuilder[In, Delta, Out]{builder: Serve(owner, JSON[In, Out](declaration.target))}
}

// StreamBuilder 复用普通服务访问、权限、优先级及重入声明。
type StreamBuilder[In, Delta, Out any] struct{ builder *Builder[In, Out] }

// Access 设置直接调用者。
func (b *StreamBuilder[In, Delta, Out]) Access(a Access, ids ...string) *StreamBuilder[In, Delta, Out] {
	b.builder.Access(a, ids...)
	return b
}

// Permission 设置要求的调用权限。
func (b *StreamBuilder[In, Delta, Out]) Permission(p sdk.Permission) *StreamBuilder[In, Delta, Out] {
	b.builder.Permission(p)
	return b
}

// Priority 设置服务顺序。
func (b *StreamBuilder[In, Delta, Out]) Priority(p int32) *StreamBuilder[In, Delta, Out] {
	b.builder.Priority(p)
	return b
}

// Reentrant 声明允许受控回调重入。
func (b *StreamBuilder[In, Delta, Out]) Reentrant() *StreamBuilder[In, Delta, Out] {
	b.builder.Reentrant()
	return b
}

// Handle 冻结流式处理器；每次调用的序号与状态单独保存。
func (b *StreamBuilder[In, Delta, Out]) Handle(handler func(*sdk.Context, In, func(Delta) error) (Out, error)) {
	id, version, err := registry.JSONType[Delta]()
	original := b.builder.builder.Prepare
	b.builder.builder.Prepare = func() (registry.Service[*sdk.Context], error) {
		if err != nil {
			return registry.Service[*sdk.Context]{}, err
		}
		if handler == nil {
			return registry.Service[*sdk.Context]{}, fmt.Errorf("stream handler is nil")
		}
		id = registry.SchemaID(b.builder.builder.Scope.Services.Schemas, b.builder.builder.Scope.Entry(), id)
		if _, err := registry.ResolveSchema(b.builder.builder.Scope.Services.Schemas, id); err != nil {
			return registry.Service[*sdk.Context]{}, err
		}
		b.builder.builder.DeltaSchemaID = id
		b.builder.builder.DeltaSchemaVersion = version
		return original()
	}
	b.builder.Handle(func(c *sdk.Context, in In) (Out, error) {
		var zero Out
		if err := registry.RequireCapability(c, CapabilityStream); err != nil {
			return zero, err
		}
		streamID := registry.StreamID(c)
		if streamID == "" {
			return zero, errcode.InvalidArgument.New("stream invocation is missing stream_id")
		}
		var sequence uint64
		var emitMu sync.Mutex // 同一次调用的 emit 顺序串行化，ACK 后才推进序号。
		emit := func(delta Delta) error {
			emitMu.Lock()
			defer emitMu.Unlock()
			data, err := json.Marshal(delta)
			if err != nil {
				return err
			}
			_, _, schemas := registry.Runtime(c)
			schema, err := registry.ResolveSchema(schemas, id)
			if err != nil {
				return err
			}
			if err := registry.ValidateJSON(schema, data); err != nil {
				return err
			}
			_, err = registry.Proto[*pb.ServiceStreamDeltaRequest, *pb.ServiceStreamDeltaResponse]("core.service.stream.emit").Call(c, &pb.ServiceStreamDeltaRequest{StreamId: streamID, Sequence: sequence, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaId: id, SchemaVersion: version, Payload: data})
			if err == nil {
				sequence++
			}
			return err
		}
		return handler(c, in, emit)
	})
}
