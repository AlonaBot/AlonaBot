package plugintest_test

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// TestLLMHarness 经公开入口验证真实 Agent 定位、Service 调用及回复与错误。
func TestLLMHarness(t *testing.T) {
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	for _, usage := range []string{"input", "output"} {
		t.Run("generate with zero "+usage+" tokens", func(t *testing.T) {
			var calls atomic.Int32 // Service goroutine 与测试线程共享调用次数。
			requests := make(chan llm.Request, 1)
			h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.LLM(func(ctx context.Context, req llm.Request) (llm.Response, error) {
				calls.Add(1)
				requests <- req
				zero := 0
				response := llm.Response{Text: "generated", Model: "synthetic", OutputTokens: &zero}
				if usage == "input" {
					response.InputTokens, response.OutputTokens = &zero, nil
				}
				return response, nil
			}))
			result := h.Text(chat, user, "/llmgenerate "+usage)
			result.RequireReply(t, "generated")
			if len(result.Outcomes) != 1 || result.Waiting {
				t.Fatalf("generation = %+v", result)
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want 1", calls.Load())
			}
			want := llm.Request{Messages: []llm.Message{{Role: "user", Text: "first"}, {Role: "assistant", Text: "second"}, {Role: "user", Text: "third"}}, System: "system prompt", Model: "synthetic", MaxOutputTokens: 37}
			if got := <-requests; !reflect.DeepEqual(got, want) {
				t.Fatalf("request = %+v, want %+v", got, want)
			}
		})
	}
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
	t.Run("unavailable", func(t *testing.T) {
		var calls atomic.Int32
		h := plugintest.Start(t, plugintest.Plugin(probe.New()), plugintest.LLM(func(context.Context, llm.Request) (llm.Response, error) {
			calls.Add(1)
			return llm.Response{}, errcode.Unavailable.New("synthetic generation failure")
		}))
		result := h.Text(chat, user, "/llmask")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || result.Waiting {
			t.Fatalf("failure = %+v", result)
		}
		failure := result.Outcomes[0].Err
		if errcode.Of(failure) != errcode.Unavailable || failure.Message != "synthetic generation failure" || calls.Load() != 1 {
			t.Fatalf("failure = %v, calls = %d", failure, calls.Load())
		}
	})
	t.Run("no substitute", func(t *testing.T) {
		h := plugintest.Start(t, plugintest.Plugin(probe.New()))
		result := h.Text(chat, user, "/llmask")
		if len(result.Sends) != 0 || len(result.Outcomes) != 1 || errcode.Of(result.Outcomes[0].Err) != errcode.Unavailable || result.Waiting {
			t.Fatalf("missing Agent = %+v", result)
		}
	})
}
