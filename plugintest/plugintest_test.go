package plugintest_test

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func TestPluginHarness(t *testing.T) {
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	check := func(t *testing.T, result plugintest.Result, text string) {
		t.Helper()
		if len(result.Sends) != 1 || result.Sends[0].Text() != text {
			t.Fatalf("sends = %+v, want one %q", result.Sends, text)
		}
		for _, outcome := range result.Outcomes {
			if outcome.Err != nil {
				t.Fatalf("handler %s/%s: %v", outcome.PluginID, outcome.RouteID, outcome.Err)
			}
		}
	}
	t.Run("commands and permissions", func(t *testing.T) {
		p := probe.New()
		p.Raw("unconsumed").Handle(func(c *sdk.Context) error {
			_, err := c.Send(message.Text(c.Text()))
			return err
		})
		h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Admin("root"), plugintest.Adapter(plugintest.AdapterID, adapter.PermissionDeclaration{Name: "plugintest.moderator", Implies: []string{"chat_admin"}, LabelKey: "moderator"}))
		result := h.Text(chat, user, "/ping")
		check(t, result, "pong")
		if result.Sends[0].To != message.SID(plugintest.AdapterID, "private:alice", "") {
			t.Fatalf("reply session = %q", result.Sends[0].To)
		}
		// 命令权限拒绝仍消费消息，后续 /ping 证明套件继续可用。
		if result := h.Text(chat, user, "/admin"); len(result.Sends) != 0 || len(result.Outcomes) != 0 {
			t.Fatalf("ordinary user reached admin Handler: %+v", result)
		}
		check(t, h.Text(chat, plugintest.User{ID: "root"}, "/admin"), "admin")
		group := plugintest.Group("room")
		if result := h.Text(group, user, "/mod"); len(result.Sends) != 0 {
			t.Fatalf("user without chat_admin received reply: %+v", result)
		}
		moderator := user
		moderator.Permissions = []sdk.Permission{sdk.PermissionName("plugintest.moderator")}
		check(t, h.Text(group, moderator, "/mod"), "mod")
		check(t, h.Text(chat, user, "/ping"), "pong")
	})
	t.Run("buttons accepted", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()))
		result := h.Text(chat, user, "/buttons")
		check(t, result, "buttons")
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
		check(t, result, "plain")
		if len(result.Sends[0].Chain) != 1 || len(result.Outcomes) != 1 {
			t.Fatalf("fallback = %+v", result)
		}
	})
	t.Run("wait and resume", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()))
		result := h.Text(chat, user, "/ask")
		check(t, result, "Your answer?")
		waiting := false
		for _, outcome := range result.Outcomes {
			waiting = waiting || outcome.Waiting && outcome.RouteID == "cmd.ask"
		}
		if !result.Waiting || !waiting {
			t.Fatalf("ask did not enter wait: %+v", result)
		}
		check(t, h.Text(chat, plugintest.User{ID: "bob"}, "/ping"), "pong")
		check(t, h.Text(plugintest.Group("other"), user, "/ping"), "pong")
		result = h.Text(chat, user, "answer")
		check(t, result, "answer")
		if result.Waiting {
			t.Fatal("completed continuation still waiting")
		}
	})
	t.Run("clicks and actor permissions", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.Admin("root"))
		group := plugintest.Group("room")
		menu := h.Text(group, user, "/menu")
		check(t, menu, "buttons")
		check(t, h.Click(user, menu.Sends[0], "first"), "one:plugintest:alice")
		check(t, h.Click(user, menu.Sends[0], "second"), "two:plugintest:alice")
		check(t, h.Click(plugintest.User{ID: "bob"}, menu.Sends[0], "first"), "one:plugintest:bob")
		admin := h.Text(group, user, "/adminmenu").Sends[0]
		if denied := h.Click(user, admin, "admin"); len(denied.Sends) != 0 || len(denied.Outcomes) != 0 {
			t.Fatalf("ordinary user reached admin button: %+v", denied)
		}
		check(t, h.Click(plugintest.User{ID: "root"}, admin, "admin"), "admin:plugintest:root")
	})
	t.Run("static and dynamic tasks", func(t *testing.T) {
		p := probe.New()
		h := plugintest.Start(t, plugintest.Plugin(p))
		static := h.Trigger(p, "tick")
		check(t, static, "tick")
		if len(static.Outcomes) != 1 || static.Outcomes[0].RouteID != "job.tick" {
			t.Fatalf("static outcomes = %+v", static.Outcomes)
		}
		created := h.Text(chat, user, "/remind lunch soup")
		if len(created.Sends) != 0 || len(created.Outcomes) != 1 || created.Outcomes[0].Err != nil {
			t.Fatalf("create task = %+v", created)
		}
		check(t, h.Text(chat, user, "/tasks"), "1")
		dynamic := h.TriggerTask(p, "remind", "lunch")
		check(t, dynamic, "soup")
		if len(dynamic.Outcomes) != 1 || dynamic.Outcomes[0].RouteID != "task.remind" || dynamic.Sends[0].To != static.Sends[0].To {
			t.Fatalf("dynamic result = %+v", dynamic)
		}
		check(t, h.Text(chat, user, "/tasks"), "0")
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
		check(t, h.Text(chat, user, "/ping"), "pong")
		check(t, h.Trigger(p, "tick"), "tick")
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
		check(t, h.Text(chat, user, "/rule"), "pong")
		result := h.Text(chat, user, "/album")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || errcode.Of(result.Outcomes[0].Err) != errcode.Unsupported {
			t.Fatalf("album capability rejection = %+v", result)
		}
	})
}
