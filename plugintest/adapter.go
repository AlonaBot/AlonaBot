package plugintest

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

type testAdapter struct {
	h          *Harness
	descriptor adapter.Descriptor
}

// Generated 按本次 Adapter ID 提供对象绑定的工程资源。
func (a *testAdapter) Generated() sdk.Generated {
	// %q 保留用户指定 ID 的边界；工程身份仍由正式 manifest 解析器校验。
	project := fmt.Sprintf("id: %s\nname: Plugin test adapter\nversion: 0.1.0\nauthor: AlonaBot\nrole: adapter\nadapter_id: %q\ndescription: Loopback test adapter\nsource: github.com/AlonaBot/AlonaBot\n", adapterComponentID, a.descriptor.AdapterID)
	return sdk.Generated{Manifest: []byte(project), AdapterConfigSchema: []byte(`{"$version":1}`)}
}

// Descriptor 返回本次测试的能力与平台声明。
func (a *testAdapter) Descriptor() *adapter.Descriptor { return &a.descriptor }

// Start 发布真实 Session 绑定的交付入口，返回固定账号实例。
func (a *testAdapter) Start(_ context.Context, _ adapter.StartRequest, core adapter.Core) (adapter.Instance, error) {
	a.h.mu.Lock()
	a.h.core = core
	a.h.mu.Unlock()
	return a, nil
}

// Account 返回固定的测试机器人账号。
func (a *testAdapter) Account() adapter.Account { return adapter.Account{ID: AccountID} }

// Capabilities 返回本次替换后的实例能力。
func (a *testAdapter) Capabilities() []message.Capability { return a.descriptor.Capabilities }

// Stop 结束测试实例；套件没有独立接收循环。
func (a *testAdapter) Stop(context.Context) error { return nil }

// Send 按实例能力预检整条链，通过后记录一个覆盖全部片段的操作。
func (a *testAdapter) Send(_ context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return a.record("send", to, "", chain)
}

// Edit 按实例能力预检编辑，并保留被编辑消息的句柄。
func (a *testAdapter) Edit(_ context.Context, to message.Session, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	return a.record("edit", to, message.NewMessageRef(to, id, InstanceID), chain)
}

func (a *testAdapter) record(op string, to message.Session, edited message.MessageRef, chain message.Chain) (*message.SendResult, *errcode.Error) {
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
		return &message.SendResult{Status: message.SendRejected, Session: to, AdapterInstanceID: InstanceID,
			PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartRejected, Error: safe}}}, safe
	}
	a.h.mu.Lock()
	defer a.h.mu.Unlock()
	a.h.messageID++
	id := strconv.FormatUint(a.h.messageID, 10)
	result := &message.SendResult{Status: message.SendAccepted, Session: to, AdapterInstanceID: InstanceID,
		PlatformMessageIDs: []string{id}, PrimaryPlatformMessageID: id, ProviderAcceptedAt: time.Now(),
		PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartAccepted, PlatformMessageIDs: []string{id}}}}
	// 链来自 SDK 解码的新副本，发送返回后仅由套件读取。
	a.h.sends = append(a.h.sends, Sent{Op: op, To: to, Chain: chain, Edited: edited, Result: result})
	return result, nil
}

func (a *testAdapter) checkCapabilities(chain message.Chain) *errcode.Error {
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
var _ adapter.Editor = (*testAdapter)(nil)
var _ sdk.GeneratedProvider = (*testAdapter)(nil)
