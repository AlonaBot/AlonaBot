package plugintest

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

type testAdapter struct {
	h          *Harness
	descriptor adapter.Descriptor
}

func (a *testAdapter) bindResources() {
	// %q 保留用户指定 ID 的边界；工程身份仍由正式 manifest 解析器校验。
	project := fmt.Sprintf("id: %s\nname: Plugin test adapter\nversion: 0.1.0\nauthor: AlonaBot\nrole: adapter\nadapter_id: %q\ndescription: Loopback test adapter\nsource: github.com/AlonaBot/AlonaBot\n", adapterComponentID+"_"+a.descriptor.AdapterID, a.descriptor.AdapterID)
	generated.Bind(a, generated.Resources{Manifest: []byte(project), AdapterConfigSchema: []byte(`{}`)})
}

// Descriptor 返回本次测试的能力与平台声明。
func (a *testAdapter) Descriptor() *adapter.Descriptor { return &a.descriptor }

// Start 发布真实 Session 绑定的交付入口，返回固定账号实例。
func (a *testAdapter) Start(_ context.Context, req adapter.StartRequest, core adapter.Core) (adapter.Instance, error) {
	config := a.h.instances[req.InstanceID]
	a.h.mu.Lock()
	a.h.cores[req.InstanceID] = core
	a.h.mu.Unlock()
	var capabilities []message.Capability
	for _, capability := range config.Descriptor.Capabilities {
		if slices.Contains(config.Capabilities, capability.Name) {
			capabilities = append(capabilities, capability)
		}
	}
	return &testInstance{testAdapter: a, config: config, capabilities: capabilities, core: core, menus: map[menuLocation][]adapter.CommandMenuEntry{}}, nil
}

type testInstance struct {
	*testAdapter
	config       InstanceConfig
	capabilities []message.Capability // 实例选择的能力快照，独立于平台注册声明全集。
	core         adapter.Core
	menus        map[menuLocation][]adapter.CommandMenuEntry
}

// Account 返回固定的测试机器人账号。
func (a *testInstance) Account() adapter.Account { return adapter.Account{ID: a.config.AccountID} }

// Capabilities 返回本次替换后的实例能力。
func (a *testInstance) Capabilities() []message.Capability { return a.capabilities }

// Stop 结束测试实例；套件没有独立接收循环。
func (a *testInstance) Stop(context.Context) error { return nil }

// Send 按实例能力预检整条链，通过后记录一个覆盖全部片段的操作。
func (a *testInstance) Send(ctx context.Context, to string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return a.record(ctx, "send", to, "", chain)
}

// Edit 按实例能力预检编辑，并保留被编辑消息的句柄。
func (a *testInstance) Edit(ctx context.Context, to string, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return a.record(ctx, "edit", to, message.Ref(to, id, a.config.InstanceID), chain)
}

func (a *testInstance) record(ctx context.Context, op string, to string, edited string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	indexes := make([]uint32, len(chain))
	for i := range indexes {
		indexes[i] = uint32(i)
	}
	name := "message." + op
	var safe *errcode.Error
	if !slices.ContainsFunc(a.Capabilities(), func(c message.Capability) bool { return c.Name == name && c.Supported }) {
		safe = errcode.Unsupported.New("adapter capability is not declared: " + name)
	} else {
		safe = a.checkCapabilities(chain)
	}
	if safe != nil {
		result := &message.SendResult{Status: message.SendRejected, Session: to, AdapterInstanceID: a.config.InstanceID,
			PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartRejected, Error: safe}}}
		call := a.call(ctx, name, to)
		call.Chain = chain
		call.SendResult = result
		call.Err = safe
		a.h.recordCall(call)
		return result, safe
	}

	a.h.mu.Lock()
	var id string
	if op == "edit" {
		_, id, _, _ = message.ParseRef(edited)
	} else {
		a.h.messageID++
		id = strconv.FormatUint(a.h.messageID, 10)
	}
	a.h.mu.Unlock()
	result := &message.SendResult{Status: message.SendAccepted, Session: to, AdapterInstanceID: a.config.InstanceID, PlatformMessageIDs: []string{id}, PrimaryPlatformMessageID: id, ProviderAcceptedAt: time.Now(), PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartAccepted, PlatformMessageIDs: []string{id}}}}
	call := a.call(ctx, "message."+op, to)
	call.Chain = chain
	call.MessageIDs = []string{id}
	call.SendResult = result
	captured, mediaErr := a.capture(ctx, chain)
	if mediaErr != nil {
		safe = errcode.Unavailable.New(mediaErr.Error())
		result = nil
	} else if op == "send" && a.config.Behavior.Send != nil {
		result, safe = a.config.Behavior.Send(ctx, call)
	} else if op == "edit" && a.config.Behavior.Edit != nil {
		result, safe = a.config.Behavior.Edit(ctx, call)
	}
	call.SendResult, call.Err = result, safe
	a.h.mu.Lock()
	a.h.sends = append(a.h.sends, Sent{eventID: a.h.traceOwners[call.TraceID], AdapterID: a.config.AdapterID, InstanceID: a.config.InstanceID, AccountID: a.config.AccountID, TraceID: call.TraceID, Op: op, To: to, Chain: chain, Edited: edited, Result: result, Err: safe, Media: captured})
	a.h.mu.Unlock()
	a.h.recordCall(call)
	return result, safe
}

func (a *testInstance) call(ctx context.Context, operation, to string) Call {
	return Call{Operation: operation, AdapterID: a.config.AdapterID, InstanceID: a.config.InstanceID, AccountID: a.config.AccountID, To: to, TraceID: wire.TraceID(ctx)}
}

func (a *testInstance) capture(ctx context.Context, chain message.Chain) ([]Media, error) {
	var result []Media
	for _, segment := range chain {
		var source message.MediaSource
		var name string
		switch s := segment.(type) {
		case message.ImageSegment:
			source = s.Source
		case message.VideoSegment:
			source = s.Source
		case message.VoiceSegment:
			source = s.Source
		case message.FileSegment:
			source = s.Source
			name = s.Filename
		case message.ForwardSegment:
			for _, item := range s.Items {
				nested, err := a.capture(ctx, item.Chain)
				if err != nil {
					return result, err
				}
				result = append(result, nested...)
			}
		default:
			continue
		}
		if source.MediaRef == "" && source.PlatformFile == "" && source.InlineBase64 == "" {
			continue
		}
		var data []byte
		if source.MediaRef != "" {
			var offset uint64
			for {
				chunk, next, eof, err := a.core.ReadMedia(ctx, source.MediaRef, offset, wire.MaxMediaChunk)
				data = append(data, chunk...)
				if err != nil {
					return result, err
				}
				if eof {
					break
				}
				offset = next
			}
		} else if source.PlatformFile != "" {
			var offset uint64
			for {
				chunk, next, eof, err := a.ReadPlatformFile(ctx, source.PlatformFile, offset, wire.MaxMediaChunk)
				data = append(data, chunk...)
				if err != nil {
					return result, err
				}
				if eof {
					break
				}
				offset = next
			}
		} else {
			decoded, err := base64.StdEncoding.DecodeString(source.InlineBase64)
			if err != nil {
				return result, err
			}
			data = decoded
		}
		result = append(result, Media{Data: data, MIME: source.MIME, Filename: name})
	}
	return result, nil
}

func (a *testInstance) checkCapabilities(chain message.Chain) *errcode.Error {
	var required []string
	media := 0
	for _, segment := range chain {
		switch data := segment.(type) {
		case message.ImageSegment, message.VideoSegment:
			media++
			if media == 2 {
				required = append(required, "message.album")
			}
			continue
		case message.ReferenceSegment:
			required = append(required, "message.reference")
			continue
		case message.ButtonsSegment:
			required = append(required, "message.buttons")
			continue
		case message.TextSegment:
			if data.Format == message.FormatMarkdown {
				required = append(required, "message.markdown")
			}
		case message.ForwardSegment:
			required = append(required, "message.forward")
			for _, item := range data.Items {
				if safe := a.checkCapabilities(item.Chain); safe != nil {
					return safe
				}
			}
		}
		// 引用和按钮属于元数据；其他内容会分隔相邻图片、视频组成的相册。
		media = 0
	}
	for _, name := range required {
		if !slices.ContainsFunc(a.Capabilities(), func(c message.Capability) bool { return c.Name == name && c.Supported }) {
			return errcode.Unsupported.New("adapter capability is not declared: " + name)
		}
	}
	return nil
}

// 保持编译期接口检查，防止测试 Adapter 绕过标准编辑入口。
var _ adapter.Editor = (*testInstance)(nil)
