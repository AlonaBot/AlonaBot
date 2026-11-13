// Package adapter 提供开发示例的回环平台，通过私有 HTTP 交付标准消息并观察发送。
//
// New 返回的 Adapter 在独立进程和同进程运行中共用生成资源与实例配置。
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

//go:generate go run ../../../cmd/alonabot plugin generate

const adapterID = "devloop"

type loopback struct {
	mu       sync.Mutex
	clients  map[string]adapter.Core // mu 保护实例启停与 HTTP 查找。
	sequence atomic.Uint64           // HTTP 注入和发送共享单调递增的消息 ID。
}

// New 创建无需外部平台的回环 Adapter。
func New() adapter.Adapter { return &loopback{clients: make(map[string]adapter.Core)} }

func (*loopback) Descriptor() *adapter.Descriptor { return &adapter.Descriptor{AdapterID: adapterID} }

func (a *loopback) HTTP(route func(string, string) *sdk.RouteBuilder) {
	route(http.MethodPost, "/inject").HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("instance")
		text := r.URL.Query().Get("text")
		if id == "" || text == "" {
			http.Error(w, "instance and text are required", http.StatusBadRequest)
			return
		}
		a.mu.Lock()
		core := a.clients[id]
		a.mu.Unlock()
		if core == nil {
			http.Error(w, "instance unavailable", http.StatusServiceUnavailable)
			return
		}
		number := fmt.Sprint(a.sequence.Add(1))
		chain := message.Chain{message.Text(text)}
		_, err := core.Deliver(r.Context(), &message.Event{
			ID: number, Type: "message.created", Source: message.SourceAdapter, OccurredAt: time.Now(),
			AdapterID: adapterID, AdapterInstanceID: id, AccountID: id,
			Session: message.NewSession(adapterID, "chat", ""), ChatType: message.ChatPrivate,
			Actor:   &message.Actor{ExternalUserID: "user", AdapterInstanceID: id},
			Message: &message.PlatformMessage{PlatformMessageID: number, Chain: chain},
			Payload: &message.CreatedPayload{PlatformMessageID: number, Chain: chain, ToMe: true},
		}, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func (a *loopback) Start(_ context.Context, req adapter.StartRequest, core adapter.Core) (adapter.Instance, error) {
	var config struct{ Output string }
	if err := json.Unmarshal(req.ConfigJSON, &config); err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.clients[req.InstanceID] = core
	a.mu.Unlock()
	slog.Info("loopback instance started", "instance_id", req.InstanceID)
	return &instance{owner: a, account: req.InstanceID, output: config.Output}, nil
}

type instance struct {
	owner   *loopback
	account string
	output  string      // 配置非空时覆盖写入最后一条发送的文本。
	stopped atomic.Bool // Stop 与 Send 可并发调用。
}

func (i *instance) Account() adapter.Account         { return adapter.Account{ID: i.account} }
func (*instance) Capabilities() []message.Capability { return nil }
func (i *instance) Send(_ context.Context, to message.Session, chain message.Chain) (*message.SendResult, *errcode.Error) {
	if i.stopped.Load() {
		return nil, errcode.FailedPrecondition.New("loopback instance has stopped")
	}
	// 原有库模式调用者可以只使用回执，文件观察由实例配置显式启用。
	if i.output != "" {
		if err := os.WriteFile(i.output, []byte(chain.Text()), 0o600); err != nil {
			return nil, errcode.Internal.Errorf("write loopback output: %v", err)
		}
	}
	number := fmt.Sprint(i.owner.sequence.Add(1))
	indexes := make([]uint32, len(chain))
	for n := range chain {
		indexes[n] = uint32(n)
	}
	return &message.SendResult{
		Status: message.SendAccepted, Session: to, AdapterInstanceID: i.account,
		PrimaryPlatformMessageID: number, PlatformMessageIDs: []string{number},
		PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: indexes, Status: message.PartAccepted, PlatformMessageIDs: []string{number}}},
	}, nil
}
func (i *instance) Stop(context.Context) error {
	i.stopped.Store(true)
	i.owner.mu.Lock()
	delete(i.owner.clients, i.account)
	i.owner.mu.Unlock()
	return nil
}
