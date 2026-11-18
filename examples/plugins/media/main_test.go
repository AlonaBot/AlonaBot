package main

import (
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func TestMedia(t *testing.T) {
	for _, album := range []bool{true, false} {
		t.Run(map[bool]string{true: "album", false: "unsupported"}[album], func(t *testing.T) {
			options := []plugintest.Option{plugintest.Plugin(New())}
			if !album {
				options = append(options, plugintest.Capabilities("message.send", "message.buttons", "message.edit"))
			}
			h := plugintest.Start(t, options...)
			user := plugintest.User{ID: "alice"}
			result := h.Text(plugintest.Private(user.ID), user, "/media")
			if len(result.Outcomes) != 1 {
				t.Fatalf("outcomes = %+v", result)
			}
			if !album {
				if len(result.Sends) != 0 || errcode.Of(result.Outcomes[0].Err) != errcode.Unsupported {
					t.Fatalf("album rejection = %+v", result)
				}
				return
			}
			result.RequireReply(t, "Click to edit this text.")
			if len(result.Sends[0].Chain) != 4 {
				t.Fatalf("album = %+v", result)
			}
			chain := result.Sends[0].Chain
			for _, segment := range chain[:2] {
				image, ok := segment.(message.ImageSegment)
				if !ok || image.Source.MediaRef == "" {
					t.Fatalf("uploaded image = %+v", segment)
				}
			}
			if _, ok := chain[2].(message.TextSegment); !ok {
				t.Fatalf("text after album = %+v", chain[2])
			}
			if _, ok := chain[3].(message.ButtonsSegment); !ok {
				t.Fatalf("buttons after text = %+v", chain[3])
			}
		})
	}
}
