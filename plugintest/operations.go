package plugintest

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"slices"

	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Call 保存测试 Adapter 的一次实际操作及返回值，TraceID 关联触发它的 Handler。
type Call struct {
	eventID                                                  string
	Operation, AdapterID, InstanceID, AccountID, To, TraceID string
	Chain                                                    message.Chain
	MessageIDs                                               []string
	Name                                                     string
	Version                                                  uint32
	Params                                                   json.RawMessage
	Menu                                                     adapter.CommandMenu
	SendResult                                               *message.SendResult
	DeleteResult                                             *message.DeleteResult
	Nodes                                                    []message.ForwardItem
	Payload                                                  json.RawMessage
	Data                                                     []byte
	Offset                                                   uint64
	Length                                                   uint32
	NextOffset                                               uint64
	EOF                                                      bool
	Err                                                      *errcode.Error
}

// Behavior 配置真实 SDK 回调的返回值；nil 回调沿用发送、编辑、删除、节点和菜单的成功实现。
// 声明的平台 Action 必须提供回调；缺失时返回 UNSUPPORTED。
// SendResult、DeleteResult 为缺省成功回执，回调可修改后返回；记录采用实际返回值。
// 回调可通过 context 等待测试信号，取消时应及时返回。
type Behavior struct {
	Send   func(context.Context, Call) (*message.SendResult, *errcode.Error)
	Edit   func(context.Context, Call) (*message.SendResult, *errcode.Error)
	Action func(context.Context, Call) (json.RawMessage, *errcode.Error)
	Delete func(context.Context, Call) (*message.DeleteResult, *errcode.Error)
	Nodes  func(context.Context, Call) ([]message.ForwardItem, *errcode.Error)
	Menu   func(context.Context, Call) ([]adapter.CommandMenuEntry, *errcode.Error)
}

// InstanceConfig 声明一个测试账号；额外平台以 Descriptor 提供声明。
type InstanceConfig struct {
	AdapterID, InstanceID, AccountID string
	Descriptor                       adapter.Descriptor // 平台声明全集；同平台实例共用。
	Capabilities                     []string           // 实际启用的能力名称；nil 沿用 Harness 的 Capabilities 选项。
	Behavior                         Behavior
	MenuConfig                       adapter.CommandMenuConfig
}

// Instance 添加一个测试实例；未提供该选项时沿用单实例默认设置。
func Instance(config InstanceConfig) Option {
	return func(o *options) { o.instances = append(o.instances, config) }
}

// DefaultInstance 指定该平台显式发送时使用的缺省实例。
func DefaultInstance(adapterID, instanceID string) Option {
	return func(o *options) {
		if o.defaults == nil {
			o.defaults = map[string]string{}
		}
		o.defaults[adapterID] = instanceID
	}
}

// Operations 配置缺省测试实例的操作回调。
func Operations(behavior Behavior) Option { return func(o *options) { o.behavior = behavior } }

// OnInstance 选择这条入站事件所属的测试账号，保持 Chat 的会话类别。
func (c Chat) OnInstance(instanceID string) Chat { c.instanceID = instanceID; return c }

// Media 保存入站平台文件 fixture；字节只由真实 Adapter 媒体 RPC 读取。
type Media struct {
	Data           []byte
	MIME, Filename string
}

// Fixture 登记实例平台文件并生成 SDK 入站媒体来源。
func (h *Harness) Fixture(instanceID, fileID string, fixture Media) message.MediaSource {
	h.t.Helper()
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.instances[instanceID]; !ok {
		h.t.Fatalf("unknown fixture instance %q", instanceID)
	}
	if h.fixtures[instanceID] == nil {
		h.fixtures[instanceID] = map[string]Media{}
	}
	fixture.Data = slices.Clone(fixture.Data)
	h.fixtures[instanceID][fileID] = fixture
	return message.MediaSource{Incoming: &message.IncomingMedia{Token: fileID, Size: uint64(len(fixture.Data)), MIME: fixture.MIME, Readability: message.Readable}}
}

// Image 解码已捕获的实际出站字节，返回原始解码错误。
func (m Media) Image() (image.Image, string, error) { return image.Decode(bytes.NewReader(m.Data)) }

// Calls 返回全部已完成操作的快照，保持 SDK 回调完成顺序。
func (h *Harness) Calls() []Call { h.mu.Lock(); defer h.mu.Unlock(); return slices.Clone(h.calls) }

// WaitCalls 等待至少 count 次完成操作或调用方取消，返回完整记录。
func (h *Harness) WaitCalls(ctx context.Context, count int) ([]Call, error) {
	for {
		h.mu.Lock()
		if len(h.calls) >= count {
			result := slices.Clone(h.calls)
			h.mu.Unlock()
			return result, nil
		}
		changed := h.changed
		h.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return h.Calls(), ctx.Err()
		case <-h.ctx.Done():
			return h.Calls(), h.ctx.Err()
		}
	}
}

func (h *Harness) recordCall(call Call) {
	h.mu.Lock()
	defer h.mu.Unlock()
	call.eventID = h.traceOwners[call.TraceID]
	h.calls = append(h.calls, call)
	close(h.changed)
	h.changed = make(chan struct{})
}
