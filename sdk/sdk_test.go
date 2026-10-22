// 本文件测试 SDK 公共类型：Handler 签名检查、注册收集与拒绝、分组路径、StopPropagation、Send 与 Chain 转换。
package sdk

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/plugin"
)

// root 是注册入口可注入的根插件。
type root struct {
	// Base 提供根插件的默认实现。
	plugin.Base
	// commands 是 Commands 的行为；为 nil 时不注册命令。
	commands func(*CommandRegistry)
	// http 是 HTTP 的行为；为 nil 时不注册路由。
	http func(*HTTPRegistry)
	// schedules 是 Schedules 的行为；为 nil 时不注册任务。
	schedules func(*ScheduleRegistry)
}

// Commands 调用注入的 commands。
func (r *root) Commands(c *CommandRegistry) {
	if r.commands != nil {
		r.commands(c)
	}
}

// HTTP 调用注入的 http。
func (r *root) HTTP(h *HTTPRegistry) {
	if r.http != nil {
		r.http(h)
	}
}

// Schedules 调用注入的 schedules。
func (r *root) Schedules(s *ScheduleRegistry) {
	if r.schedules != nil {
		r.schedules(s)
	}
}

// 两种合法签名通过；nil、非函数、缺 Context、可变参数、不支持的参数或返回类型被拒绝。
func TestCheckHandler(t *testing.T) {
	ok := []any{
		func(*Context) error { return nil },
		func(*Context) (message.Chain, error) { return nil, nil },
		func(*Context, string, int64, float64, bool) error { return nil },
	}
	for _, h := range ok {
		if _, _, err := checkHandler(h); err != nil {
			t.Errorf("%T 应通过: %v", h, err)
		}
	}
	bad := map[string]any{
		"nil":         nil,
		"nil 函数":      (func(*Context) error)(nil),
		"非函数":         42,
		"缺少 Context":  func(string) error { return nil },
		"Context 非指针": func(Context) error { return nil },
		"不支持的参数 int":  func(*Context, int) error { return nil },
		"可变参数":        func(*Context, ...string) error { return nil },
		"无返回值":        func(*Context) {},
		"返回值顺序错误":     func(*Context) (error, message.Chain) { return nil, nil },
		"返回 string":   func(*Context) (string, error) { return "", nil },
		"三个返回值":       func(*Context) (message.Chain, error, error) { return nil, nil, nil },
	}
	for name, h := range bad {
		if _, _, err := checkHandler(h); err == nil {
			t.Errorf("%s 应被拒绝", name)
		}
	}
}

// 重复命令、非法路径、重复或非规范的 HTTP 路由、公开路由缺验证方式、非法定时任务都使 Collect 失败。
func TestCollectRejects(t *testing.T) {
	h := func(*Context) error { return nil }
	cases := map[string]*root{
		"重复命令": {commands: func(r *CommandRegistry) {
			r.Command("a", h)
			r.Group("x", func(g *CommandGroup) { g.Command("b", h) })
			r.Command("a", h)
		}},
		"组内与组外重复路径": {commands: func(r *CommandRegistry) {
			r.Group("x", func(g *CommandGroup) { g.Command("b", h) })
			r.Group("x", func(g *CommandGroup) { g.Command("b", h) })
		}},
		"命令 token 含空白": {commands: func(r *CommandRegistry) { r.Command("a b", h) }},
		"非法 Handler":   {commands: func(r *CommandRegistry) { r.Command("a", 1) }},
		"重复 HTTP 路由": {http: func(r *HTTPRegistry) {
			r.Handle("get", "/s", http.NotFoundHandler())
			r.Handle("GET", "/s", http.NotFoundHandler())
		}},
		"HTTP 路径不以 / 开头": {http: func(r *HTTPRegistry) { r.Handle("GET", "s", http.NotFoundHandler()) }},
		"HTTP 路径含 ..":    {http: func(r *HTTPRegistry) { r.Handle("GET", "/a/../b", http.NotFoundHandler()) }},
		"HTTP 路径含 //":    {http: func(r *HTTPRegistry) { r.Handle("GET", "//x", http.NotFoundHandler()) }},
		"HTTP 路径尾部 /":    {http: func(r *HTTPRegistry) { r.Handle("GET", "/a/", http.NotFoundHandler()) }},
		"公开路由未声明验证": {http: func(r *HTTPRegistry) {
			r.Handle("POST", "/cb", http.NotFoundHandler(), Public(pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION))
		}},
		"重复任务": {schedules: func(r *ScheduleRegistry) {
			r.Every("t", time.Second, func(*Context) error { return nil })
			r.Cron("t", "0 * * * *", "", func(*Context) error { return nil })
		}},
		"间隔小于 1s":   {schedules: func(r *ScheduleRegistry) { r.Every("t", time.Millisecond, func(*Context) error { return nil }) }},
		"cron 非五字段": {schedules: func(r *ScheduleRegistry) { r.Cron("t", "* * *", "", func(*Context) error { return nil }) }},
	}
	for name, r := range cases {
		if _, err := Collect(r); err == nil {
			t.Errorf("%s 应被拒绝", name)
		}
	}
	// 多个错误一起报告，便于一次修完。
	_, err := Collect(&root{commands: func(r *CommandRegistry) { r.Command("a", 1); r.Command("b", 2) }})
	if err == nil || !strings.Contains(err.Error(), "/a") || !strings.Contains(err.Error(), "/b") {
		t.Fatalf("应同时报告两个错误，实际 %v", err)
	}
}

// HTTP 根路径 / 是规范路径，允许注册。
func TestHTTPRootPathAllowed(t *testing.T) {
	if _, err := Collect(&root{http: func(r *HTTPRegistry) { r.Handle("GET", "/", http.NotFoundHandler()) }}); err != nil {
		t.Fatalf("根路径 / 应允许: %v", err)
	}
}

// 嵌套分组的前缀逐层并入命令路径，顺序与注册顺序一致。
func TestCollectGroupPath(t *testing.T) {
	reg, err := Collect(&root{commands: func(r *CommandRegistry) {
		r.Group("rss", func(g *CommandGroup) {
			g.Command("list", func(*Context) (message.Chain, error) { return nil, nil })
			g.Group("admin", func(a *CommandGroup) { a.Command("purge", func(*Context) error { return nil }) })
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range reg.Commands {
		got = append(got, strings.Join(c.Path, " "))
	}
	if strings.Join(got, ",") != "rss list,rss admin purge" {
		t.Fatalf("命令路径 = %v", got)
	}
}

// StopPropagation 一经设置保持锁存。
func TestStopPropagationLatches(t *testing.T) {
	c := &Context{}
	c.StopPropagation()
	c.StopPropagation()
	if !c.stopped {
		t.Fatal("StopPropagation 应锁存")
	}
}

// Send 尚未实现时返回 UNSUPPORTED。
func TestSendNotImplemented(t *testing.T) {
	var e *Error
	if _, err := (&Context{}).Send(message.Chain{message.Text("x")}); !errors.As(err, &e) || e.Code != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("Send 应返回 UNSUPPORTED，实际 %v", err)
	}
}

// message.Chain 按顺序转换为 MessageChain。
func TestChainProto(t *testing.T) {
	c := message.Chain{message.Text("a"), message.Text("b")}.Proto()
	if len(c.Segments) != 2 || c.Segments[1].GetText().GetText() != "b" {
		t.Fatalf("Proto() = %v", c)
	}
}
