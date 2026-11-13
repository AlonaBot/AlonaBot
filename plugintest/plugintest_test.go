package plugintest_test

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
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
		if result.Sends[0].To != message.NewSession(plugintest.AdapterID, "private:alice", "") {
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
		if !result.Waiting {
			t.Fatalf("ask did not enter wait: %+v", result)
		}
		result = h.Text(chat, user, "answer")
		check(t, result, "answer")
		if result.Waiting {
			t.Fatal("completed continuation still waiting")
		}
	})
}
