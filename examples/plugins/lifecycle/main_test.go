package main

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
)

func TestLifecycle(t *testing.T) {
	h := plugintest.Start(t, plugintest.Plugin(New()))
	user := plugintest.User{ID: "alice"}
	result := h.Text(plugintest.Private(user.ID), user, "/ping")
	result.RequireReply(t, "pong")
	if len(result.Outcomes) != 1 {
		t.Fatalf("ping = %+v", result)
	}
}
