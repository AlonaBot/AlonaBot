package plugintest_test

import (
	"fmt"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/kv"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func TestPluginHarness(t *testing.T) {
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	t.Run("commands and permissions", func(t *testing.T) {
		p := probe.New()
		p.Raw().Handle(func(c *sdk.Context) error {
			_, err := c.Send(message.Text(c.Text()))
			return err
		})
		h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Admin("root"), plugintest.Adapter(plugintest.AdapterID, adapter.PermissionDeclaration{Name: "plugintest.moderator", Implies: []string{"chat_admin"}, LabelKey: "moderator"}))
		result := h.Text(chat, user, "/ping")
		result.RequireReply(t, "pong")
		if result.Sends[0].To != message.SID(plugintest.AdapterID, "private:alice", "") {
			t.Fatalf("reply session = %q", result.Sends[0].To)
		}
		// 命令权限拒绝仍消费消息，后续 /ping 证明套件继续可用。
		if result := h.Text(chat, user, "/admin"); len(result.Sends) != 0 || len(result.Outcomes) != 0 {
			t.Fatalf("ordinary user reached admin Handler: %+v", result)
		}
		h.Text(chat, plugintest.User{ID: "root"}, "/admin").RequireReply(t, "admin")
		group := plugintest.Group("room")
		if result := h.Text(group, user, "/mod"); len(result.Sends) != 0 {
			t.Fatalf("user without chat_admin received reply: %+v", result)
		}
		moderator := user
		moderator.Permissions = []sdk.Permission{sdk.PermissionName("plugintest.moderator")}
		h.Text(group, moderator, "/mod").RequireReply(t, "mod")
		h.Text(chat, user, "/ping").RequireReply(t, "pong")
	})
	t.Run("buttons accepted", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()))
		result := h.Text(chat, user, "/buttons")
		result.RequireReply(t, "buttons")
		if result.Sends[0].Result.Status != message.SendAccepted || len(result.Sends[0].Chain) != 2 {
			t.Fatalf("button send = %+v", result.Sends[0])
		}
		if _, ok := result.Sends[0].Chain[1].(message.ButtonsSegment); !ok {
			t.Fatal("buttons segment missing")
		}
	})
	t.Run("buttons fallback", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.Capabilities("message.send", "message.markdown", "message.album", "message.reference", "message.forward", "message.edit"))
		result := h.Text(chat, user, "/buttons")
		result.RequireReply(t, "plain")
		if len(result.Sends[0].Chain) != 1 || len(result.Outcomes) != 1 {
			t.Fatalf("fallback = %+v", result)
		}
	})
	t.Run("wait and resume", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()))
		result := h.Text(chat, user, "/ask")
		result.RequireReply(t, "Your answer?")
		waiting := false
		for _, outcome := range result.Outcomes {
			waiting = waiting || outcome.Waiting && outcome.RouteID == "cmd.ask"
		}
		if !result.Waiting || !waiting {
			t.Fatalf("ask did not enter wait: %+v", result)
		}
		h.Text(chat, plugintest.User{ID: "bob"}, "/ping").RequireReply(t, "pong")
		h.Text(plugintest.Group("other"), user, "/ping").RequireReply(t, "pong")
		result = h.Text(chat, user, "answer")
		result.RequireReply(t, "answer")
		if result.Waiting {
			t.Fatal("completed continuation still waiting")
		}
	})
	t.Run("clicks and actor permissions", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.Admin("root"))
		group := plugintest.Group("room")
		menu := h.Text(group, user, "/menu")
		menu.RequireReply(t, "buttons")
		h.Click(user, menu.Sends[0], "first").RequireReply(t, "one:plugintest:alice")
		h.Click(user, menu.Sends[0], "second").RequireReply(t, "two:plugintest:alice")
		h.Click(plugintest.User{ID: "bob"}, menu.Sends[0], "first").RequireReply(t, "one:plugintest:bob")
		admin := h.Text(group, user, "/adminmenu").Sends[0]
		if denied := h.Click(user, admin, "admin"); len(denied.Sends) != 0 || len(denied.Outcomes) != 0 {
			t.Fatalf("ordinary user reached admin button: %+v", denied)
		}
		h.Click(plugintest.User{ID: "root"}, admin, "admin").RequireReply(t, "admin:plugintest:root")
	})
	t.Run("static and dynamic tasks", func(t *testing.T) {
		p := probe.New()
		h := plugintest.Start(t, plugintest.Plugin(p))
		static := h.Trigger(p, "tick")
		static.RequireReply(t, "tick")
		if len(static.Outcomes) != 1 || static.Outcomes[0].RouteID != "job.tick" {
			t.Fatalf("static outcomes = %+v", static.Outcomes)
		}
		created := h.Text(chat, user, "/remind lunch soup")
		if len(created.Sends) != 0 || len(created.Outcomes) != 1 || created.Outcomes[0].Err != nil {
			t.Fatalf("create task = %+v", created)
		}
		h.Text(chat, user, "/tasks").RequireReply(t, "1")
		dynamic := h.TriggerTask(p, "remind", "lunch")
		dynamic.RequireReply(t, "soup")
		if len(dynamic.Outcomes) != 1 || dynamic.Outcomes[0].RouteID != "task.remind" || dynamic.Sends[0].To != static.Sends[0].To {
			t.Fatalf("dynamic result = %+v", dynamic)
		}
		h.Text(chat, user, "/tasks").RequireReply(t, "0")
		if records := h.Background(); len(records) != 0 {
			t.Fatalf("consumed records = %+v", records)
		}
	})
	t.Run("scheduled send permission", func(t *testing.T) {
		p := probe.New()
		h := plugintest.Start(t, plugintest.Plugin(p), plugintest.AllowedInstances(p))
		result := h.Trigger(p, "tick")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || errcode.Of(result.Outcomes[0].Err) != errcode.PermissionDenied {
			t.Fatalf("denied scheduled send = %+v", result)
		}
		h.Text(chat, user, "/remind lunch soup")
		result = h.TriggerTask(p, "remind", "lunch")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || errcode.Of(result.Outcomes[0].Err) != errcode.PermissionDenied {
			t.Fatalf("denied task send = %+v", result)
		}
	})
	t.Run("background ordered consumption", func(t *testing.T) {
		p := probe.New()
		send, done := make(chan struct{}), make(chan error)
		background := &probe.Background{Send: send, Done: done, To: message.SID(plugintest.AdapterID, "private:alice", "")}
		p.OnReady(background.Ready)
		p.OnStop(background.Stop)
		h := plugintest.Start(t, plugintest.Plugin(p))
		send <- struct{}{}
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		h.Text(chat, user, "/ping").RequireReply(t, "pong")
		h.Trigger(p, "tick").RequireReply(t, "tick")
		records := h.Background()
		if len(records) != 1 || records[0].Text() != "background" {
			t.Fatalf("background = %+v", records)
		}
		if records := h.Background(); len(records) != 0 {
			t.Fatalf("duplicate background = %+v", records)
		}
	})
	t.Run("rule and media capability", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.Capabilities("message.send"))
		h.Text(chat, user, "/rule").RequireReply(t, "pong")
		result := h.Text(chat, user, "/album")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || errcode.Of(result.Outcomes[0].Err) != errcode.Unsupported {
			t.Fatalf("album capability rejection = %+v", result)
		}
	})
}

// TestSimplifiedSDKUsage 经真实 Core、Session、SQLite 与 Adapter 验证简化注册和片段回复。
func TestSimplifiedSDKUsage(t *testing.T) {
	p := probe.New()
	p.FullMatch("hello").Handle(func(c *sdk.Context) error {
		if err := kv.Set(c, "greeting", []byte("saved")); err != nil {
			return err
		}
		return c.Reply("hi", message.ImageURL("https://example.com/image.png"))
	})
	p.On(sdk.MessageCreated).Handle(func(c *sdk.Context) error {
		if c.Text() != "hello" {
			return nil
		}
		value, ok, err := kv.Get(c, "greeting")
		if err != nil {
			return err
		}
		if !ok || string(value) != "saved" {
			return fmt.Errorf("greeting KV value = %q, found = %v", value, ok)
		}
		return nil
	})
	p.Command("plain").Handle(func(c *sdk.Context) error { return c.Reply("plain reply") })
	h := plugintest.Start(t, plugintest.Plugin(p))
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	result := h.Text(chat, user, "hello")
	result.RequireReply(t, "hi")
	if len(result.Sends[0].Chain) != 2 {
		t.Fatalf("reply chain = %+v", result.Sends[0].Chain)
	}
	if text, ok := result.Sends[0].Chain[0].(message.TextSegment); !ok || text.Text != "hi" {
		t.Fatalf("text segment = %+v", result.Sends[0].Chain[0])
	}
	if image, ok := result.Sends[0].Chain[1].(message.ImageSegment); !ok || image.Source.HTTPSURL != "https://example.com/image.png" {
		t.Fatalf("image segment = %+v", result.Sends[0].Chain[1])
	}
	if len(result.Outcomes) != 2 {
		t.Fatalf("outcomes = %+v", result.Outcomes)
	}
	h.Text(chat, user, "/plain").RequireReply(t, "plain reply")
}
