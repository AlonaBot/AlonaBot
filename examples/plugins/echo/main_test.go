package main

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func TestEcho(t *testing.T) {
	h := plugintest.Start(t, plugintest.Plugin(New()))
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	for _, content := range []message.Content{
		message.Text("hello"),
		message.Chain{message.Text("hel"), message.Mention("bob"), message.ImageURL("https://example.com/image.png"), message.Text("lo")},
	} {
		result := h.Message(chat, user, content)
		if len(result.Sends) != 1 || result.Sends[0].Text() != "hello" || len(result.Outcomes) != 2 {
			t.Fatalf("echo = %+v", result)
		}
		for _, outcome := range result.Outcomes {
			if outcome.Err != nil {
				t.Fatalf("echo outcome = %+v", outcome)
			}
		}
	}
}
