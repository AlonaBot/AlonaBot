// Package adapter 提供库模式示例的内存回环平台。
package adapter

import (
	"context"
	"sync/atomic"

	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

//go:generate go run ../../../cmd/alonabot plugin generate

type loopback struct{}

// New 创建无需外部平台的回环 Adapter。
func New() adapter.Adapter { return &loopback{} }

func (*loopback) Descriptor() *adapter.Descriptor { return &adapter.Descriptor{AdapterID: "devloop"} }
func (*loopback) Start(_ context.Context, req adapter.StartRequest, _ adapter.Core) (adapter.Instance, error) {
	return &instance{account: req.InstanceID}, nil
}

type instance struct {
	account string
	stopped atomic.Bool
}

func (i *instance) Account() adapter.Account         { return adapter.Account{ID: i.account} }
func (*instance) Capabilities() []message.Capability { return nil }
func (i *instance) Send(_ context.Context, to message.Session, _ message.Chain) (*message.SendResult, *errcode.Error) {
	if i.stopped.Load() {
		return nil, errcode.FailedPrecondition.New("loopback instance has stopped")
	}
	return &message.SendResult{Status: message.SendAccepted, Session: to}, nil
}
func (i *instance) Stop(context.Context) error { i.stopped.Store(true); return nil }
