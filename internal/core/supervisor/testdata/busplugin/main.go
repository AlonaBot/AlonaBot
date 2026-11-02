// busplugin 命令运行 Event Bus 集成测试中的发布者与订阅者。
package main

import (
	"errors"
	"log/slog"
	"os"
	"slices"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
)

const (
	pluginIDEnv    = "ALONABOT_TEST_BUS_ID"
	controlSchema  = "test.bus.control.v1"
	replySchema    = "test.bus.reply.v1"
	eventSchema    = "test.bus.number.v1"
	publishService = "core.bus.publish"
	schemaVersion  = 1
	exitFailure    = 1
)

type control struct {
	Operation string
	Event     string
	Number    int
	Version   uint32
}

type reply struct {
	EventID   string
	Seen      []int
	Versioned []int
	Entered   bool
	Code      pb.ErrorCode
	Error     string
}

type plugin struct {
	sdk.Base
	id        string
	release   chan struct{} // stuck 的回调等待测试显式放行，停止时响应 Context 取消
	mu        sync.Mutex    // 保护回调记录与 Entered；控制 Service 可与回调并发执行
	seen      []int
	versioned []int
	entered   bool
}

// Subscriptions 注册顺序收件回调，以及 fast/slow 的不同版本订阅。
func (p *plugin) Subscriptions(r *sdk.SubscriptionRegistry) {
	if p.id == "pub" {
		return
	}
	sdk.Subscribe(r, "number", "pub/x", eventSchema, schemaVersion, func(c *sdk.Context, n int) error {
		if p.id == "stuck" {
			p.mu.Lock()
			p.entered = true
			p.mu.Unlock()
			// 等待业务放行时，SDK 的读循环和其他插件仍应继续。
			select {
			case <-p.release:
			case <-c.Context().Done():
				return c.Context().Err()
			}
		}
		p.mu.Lock()
		p.seen = append(p.seen, n)
		p.mu.Unlock()
		return nil
	})
	if p.id == "fast" || p.id == "slow" {
		version := uint32(schemaVersion)
		if p.id == "fast" {
			version++
		}
		sdk.Subscribe(r, "versioned", "pub/version", eventSchema, version, func(_ *sdk.Context, n int) error {
			p.mu.Lock()
			p.versioned = append(p.versioned, n)
			p.mu.Unlock()
			return nil
		})
	}
}

// Services 暴露测试控制入口；发布仍经过插件到 Core 的真实 Publish RPC。
func (p *plugin) Services(r *sdk.ServiceRegistry) {
	sdk.JSONService(r, "control", controlSchema, replySchema, schemaVersion, func(c *sdk.Context, in control) (reply, error) {
		out := reply{}
		var err error
		switch in.Operation {
		case "publish":
			out.EventID, err = sdk.Publish(c.Context(), in.Event, eventSchema, in.Version, in.Number)
		case "foreign":
			// 绕过 Publish 的自动前缀，实际请求 Core 检查越权命名空间。
			var result *pb.PublishResponse
			result, err = sdk.CallProto[*pb.PublishResponse](c.Context(), publishService, &pb.PublishRequest{
				EventName: "fast/x", SchemaId: eventSchema, SchemaVersion: schemaVersion,
				Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, Payload: []byte(`1`),
			})
			if err == nil {
				out.EventID = result.GetEventId()
			}
		case "snapshot":
			p.mu.Lock()
			out.Seen, out.Versioned, out.Entered = slices.Clone(p.seen), slices.Clone(p.versioned), p.entered
			p.mu.Unlock()
		case "release":
			close(p.release)
		default:
			return out, errors.New("unknown bus test operation")
		}
		if err != nil {
			var public *sdk.Error
			if !errors.As(err, &public) {
				return out, err
			}
			out.Code, out.Error = public.Code, public.Message
		}
		return out, nil
	})
}

func main() {
	if err := sdk.Run(&plugin{id: os.Getenv(pluginIDEnv), release: make(chan struct{})}); err != nil {
		slog.Error("bus plugin failed", "error", err)
		os.Exit(exitFailure)
	}
}
