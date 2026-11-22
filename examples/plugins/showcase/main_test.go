package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/task"
)

func TestShowcase(t *testing.T) {
	p := New()
	var ctx context.Context
	p.OnReady(func(ready context.Context) error { ctx = context.WithoutCancel(ready); return nil })
	h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Config(p, `{"greeting":"Welcome"}`))
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	check := func(result plugintest.Result, text string, count int) {
		t.Helper()
		if count == 1 {
			result.RequireReply(t, text)
			return
		}
		if len(result.Sends) != count {
			t.Fatalf("sends = %+v, want %d", result, count)
		}
		for _, sent := range result.Sends {
			if sent.Text() != text {
				t.Fatalf("text = %q, want %q", sent.Text(), text)
			}
		}
		for _, outcome := range result.Outcomes {
			if outcome.Err != nil {
				t.Fatal(outcome.Err)
			}
		}
	}
	check(h.Text(chat, user, "/demo hello"), "Welcome", 1)
	check(h.Text(chat, user, `/demo echo "hello world" 2`), "hello world", 2)
	invalid := h.Text(chat, user, "/demo echo hi invalid")
	invalid.RequireReply(t, "Usage: /demo echo <text> <times:int>")
	if len(invalid.Outcomes) != 0 {
		t.Fatalf("invalid argument usage = %+v", invalid)
	}
	check(h.Text(chat, user, "/demo remind lunch 3600 soup"), "lunch", 1)
	session := message.SID(plugintest.AdapterID, "private:alice", "")
	actor := message.UID(plugintest.AdapterID, user.ID)
	key := fmt.Sprintf("%d:%s/%d:%s/%s", len(actor), actor, len(session), session, "lunch")
	reminder := h.TriggerTask(p, "remind", key)
	check(reminder, "soup", 1)
	if reminder.Sends[0].To != session {
		t.Fatalf("reminder session = %q", reminder.Sends[0].To)
	}
	check(h.Text(chat, user, "/demo remind lunch 3600 soup"), "lunch", 1)
	check(h.Text(chat, user, "/demo cancel lunch"), "lunch", 1)
	instances, err := task.Define[ReminderArgs]("remind").List(ctx)
	if err != nil || len(instances) != 0 {
		t.Fatalf("cancelled instances = %+v, error = %v", instances, err)
	}
}
