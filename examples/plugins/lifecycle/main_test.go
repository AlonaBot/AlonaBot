package main

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
)

func TestLifecycle(t *testing.T) {
	h := plugintest.Start(t, plugintest.Plugin(New()))
	user := plugintest.User{ID: "alice"}
	result := h.Text(plugintest.Private(user.ID), user, "/ping")
	if len(result.Sends) != 1 || result.Sends[0].Text() != "pong" || len(result.Outcomes) != 1 || result.Outcomes[0].Err != nil {
		t.Fatalf("ping = %+v", result)
	}
}
