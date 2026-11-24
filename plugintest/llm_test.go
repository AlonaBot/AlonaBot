package plugintest_test

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// TestLLMHarness 经公开入口验证真实 Agent 定位、Service 调用及回复与错误。
func TestLLMHarness(t *testing.T) {
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	t.Run("ask", func(t *testing.T) {
		var calls atomic.Int32
		requests := make(chan llm.Request, 1)
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error) {
			calls.Add(1)
			requests <- req
			return llm.Response{Text: "answer", Model: "synthetic"}, nil
		}))
		result := h.Text(chat, user, "/llmask")
		result.RequireReply(t, "answer")
		if len(result.Outcomes) != 1 || calls.Load() != 1 {
			t.Fatalf("ask = %+v, calls = %d", result, calls.Load())
		}
		want := llm.Request{Messages: []llm.Message{{Role: "user", Text: "  original prompt\n"}}}
		if got := <-requests; !reflect.DeepEqual(got, want) {
			t.Fatalf("ask request = %+v, want %+v", got, want)
		}
	})
}
