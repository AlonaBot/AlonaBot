package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const (
	consoleAdapterID   = "console"
	consoleComponentID = "alonabot-adapter-console"
	consoleInstanceID  = "console"
	consoleUserID      = "local"
	consoleAccountID   = "bot"
)

type consoleAdapter struct {
	input         io.Reader
	output        io.Writer
	terminal      bool       // stdout 为终端时显示提示符。
	inputTerminal bool       // stdin 的终端回显自行换行；管道输入由读取方换行。
	mu            sync.Mutex // 串行化实例切换、输出与提示符状态。
	current       *consoleInstance
	waiting       bool // 唯一 stdin 读取器正在等待下一行，跨实例重启保留。
	prompted      bool
}

// Descriptor 声明内置终端平台身份。
func (*consoleAdapter) Descriptor() *adapter.Descriptor {
	return &adapter.Descriptor{AdapterID: consoleAdapterID}
}

// Start 绑定真实 Core；stdin 的读取由 CLI 在全部组件激活后启动。
func (a *consoleAdapter) Start(ctx context.Context, _ adapter.StartRequest, core adapter.Core) (adapter.Instance, error) {
	lifetime, cancel := context.WithCancel(context.WithoutCancel(ctx))
	instance := &consoleInstance{adapter: a, core: core, ctx: lifetime, cancel: cancel}
	a.mu.Lock()
	a.current = instance
	a.mu.Unlock()
	return instance, nil
}

type consoleInstance struct {
	adapter  *consoleAdapter
	core     adapter.Core
	ctx      context.Context
	cancel   context.CancelFunc
	sequence uint64
	buttons  []message.Button
	origin   string
	stopped  bool
}

// Account 返回终端机器人的固定账号。
func (*consoleInstance) Account() adapter.Account { return adapter.Account{ID: consoleAccountID} }

// Capabilities 声明终端支持的发送、编辑、按钮、Markdown 与引用能力。
func (*consoleInstance) Capabilities() []message.Capability {
	return []message.Capability{
		{Name: "message.send", Version: 1, Supported: true},
		{Name: "message.edit", Version: 1, Supported: true},
		{Name: "message.buttons", Version: 1, Supported: true},
		{Name: "message.markdown", Version: 1, Supported: true},
		{Name: "message.reference", Version: 1, Supported: true},
	}
}

// Send 将整条消息写入 stdout，并返回递增的平台消息 ID。
func (i *consoleInstance) Send(_ context.Context, to string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.adapter.mu.Lock()
	defer i.adapter.mu.Unlock()
	i.sequence++
	return i.writeMessage(to, strconv.FormatUint(i.sequence, 10), chain, "")
}

// Edit 将编辑结果写为带标记的新段，并沿用原消息 ID。
func (i *consoleInstance) Edit(_ context.Context, to, platformMessageID string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.adapter.mu.Lock()
	defer i.adapter.mu.Unlock()
	return i.writeMessage(to, platformMessageID, chain, "(edited) ")
}

// writeMessage 由持锁的发送与编辑入口调用，整段输出后才发布按钮快照。
func (i *consoleInstance) writeMessage(to, id string, chain message.Chain, prefix string) (*message.SendResult, *errcode.Error) {
	if i.stopped {
		return nil, errcode.FailedPrecondition.New("console instance has stopped")
	}
	var text strings.Builder
	text.WriteString(prefix)
	var buttons []message.Button
	hasButtons := false
	for _, segment := range chain {
		switch s := segment.(type) {
		case message.TextSegment:
			text.WriteString(s.Text)
		case message.ImageSegment:
			text.WriteString("[image]")
		case message.FileSegment:
			fmt.Fprintf(&text, "[file: %s]", s.Filename)
		case message.VideoSegment:
			text.WriteString("[video]")
		case message.VoiceSegment:
			text.WriteString("[voice]")
		case message.MentionSegment:
			text.WriteString("[mention]")
		case message.ReferenceSegment:
			text.WriteString("[reference]")
		case message.ForwardSegment:
			text.WriteString("[forward]")
		case message.ExtensionSegment:
			text.WriteString("[extension]")
		case message.ButtonsSegment:
			hasButtons = true
			for _, row := range s.Rows {
				buttons = append(buttons, row...)
			}
		}
	}
	if !strings.HasSuffix(text.String(), "\n") {
		text.WriteByte('\n')
	}
	var callbacks []message.Button
	for _, button := range buttons {
		if button.URL != "" {
			fmt.Fprintf(&text, "%s <%s>\n", button.Label, button.URL)
		} else {
			callbacks = append(callbacks, button)
			fmt.Fprintf(&text, "#%d %s\n", len(callbacks), button.Label)
		}
	}
	if err := i.adapter.write(text.String()); err != nil {
		return nil, errcode.Internal.Errorf("write console message: %v", err)
	}
	if hasButtons || prefix != "" && id == i.origin {
		i.buttons = callbacks
		i.origin = id
	}
	indexes := make([]uint32, len(chain))
	for n := range chain {
		indexes[n] = uint32(n)
	}
	return &message.SendResult{Status: message.SendAccepted, Session: to, AdapterInstanceID: consoleInstanceID,
		PrimaryPlatformMessageID: id, PlatformMessageIDs: []string{id},
		PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartAccepted, PlatformMessageIDs: []string{id}}}}, nil
}

// write 将异步输出移到提示符后的新行，再恢复仍在等待输入的提示符；调用方持有 mu。
func (a *consoleAdapter) write(text string) error {
	if a.prompted {
		text = "\n" + text
		a.prompted = false
	}
	if a.waiting && a.terminal {
		text += "> "
		a.prompted = true
	}
	_, err := io.WriteString(a.output, text)
	return err
}

// readInput 在首次激活后启动一次，每行交给 Start 最近发布的实例。
func (a *consoleAdapter) readInput() error {
	reader := bufio.NewReader(a.input)
	for {
		a.mu.Lock()
		a.waiting = true
		if a.terminal && !a.prompted {
			if _, err := io.WriteString(a.output, "> "); err != nil {
				a.mu.Unlock()
				return err
			}
			a.prompted = true
		}
		a.mu.Unlock()
		// stdin 由 CLI 进程持有；信号终止时进程释放仍在等输入的读取方。
		line, readErr := reader.ReadString('\n')
		a.mu.Lock()
		a.waiting = false
		var promptErr error
		if a.prompted && (!a.inputTerminal || errors.Is(readErr, io.EOF)) {
			_, promptErr = io.WriteString(a.output, "\n")
		}
		a.prompted = false
		instance := a.current
		a.mu.Unlock()
		if promptErr != nil {
			return promptErr
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line != "" {
			// 重启取消旧实例的在途交付时，读取器继续接收新实例的输入。
			if err := instance.deliver(line); err != nil && instance.ctx.Err() == nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read console input: %w", readErr)
		}
	}
}

func (i *consoleInstance) deliver(line string) error {
	i.adapter.mu.Lock()
	if i.stopped {
		i.adapter.mu.Unlock()
		return i.ctx.Err()
	}
	i.sequence++
	id := "input-" + strconv.FormatUint(i.sequence, 10)
	event := &message.Event{ID: id, Type: "message.created", Source: message.SourceAdapter, OccurredAt: time.Now(),
		AdapterID: consoleAdapterID, AdapterInstanceID: consoleInstanceID, AccountID: consoleAccountID,
		Session: message.SID(consoleAdapterID, consoleUserID, ""), ChatType: message.ChatPrivate,
		Actor: &message.Actor{ExternalUserID: consoleUserID, AdapterInstanceID: consoleInstanceID}}
	if strings.HasPrefix(line, "#") && len(line) > 1 && strings.Trim(line[1:], "0123456789") == "" {
		n, err := strconv.Atoi(line[1:])
		if err != nil || n < 1 || n > len(i.buttons) {
			err = i.adapter.write("Invalid callback button number.\n")
			i.adapter.mu.Unlock()
			return err
		}
		event.Type = "button.clicked"
		event.Payload = &message.ButtonClickedPayload{Value: i.buttons[n-1].CoreToken, CallbackQueryID: id, OriginPlatformMessageID: i.origin}
	} else {
		chain := message.Chain{message.Text(line)}
		event.Message = &message.PlatformMessage{PlatformMessageID: id, Chain: chain}
		event.Payload = &message.CreatedPayload{PlatformMessageID: id, Chain: chain, ToMe: true}
	}
	i.adapter.mu.Unlock()
	accepted, err := i.core.Deliver(i.ctx, event, nil)
	if err != nil {
		return fmt.Errorf("deliver console event: %w", err)
	}
	if !accepted {
		return errors.New("console event not accepted: dispatch queue full")
	}
	return nil
}

// Stop 取消事件交付；进程持有的 stdin 随 CLI 退出释放。
func (i *consoleInstance) Stop(context.Context) error {
	i.adapter.mu.Lock()
	defer i.adapter.mu.Unlock()
	i.stopped = true
	i.cancel()
	return nil
}
