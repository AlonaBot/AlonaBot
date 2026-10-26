package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, p Protocol, h http.HandlerFunc) Client {
	t.Helper()
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	c, err := New(Config{Protocol: p, BaseURL: s.URL + "/v1", APIKey: "test-secret", Model: "test-model"}, s.Client())
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestOpenAIStream(t *testing.T) {
	c := newTestClient(t, OpenAICompatible, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Error("missing auth")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"c1","function":{"name":"lookup","arguments":"{\"q\":\"x\""}}]}}]}

`)
		fmt.Fprint(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"}"}}]}}]}

`)
		fmt.Fprint(w, "data: [DONE]\n\n")
	})
	var got strings.Builder
	calls := []ToolCall{}
	err := c.Stream(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}, MaxTokens: 20}, func(e Event) error {
		if e.Kind == TextDelta {
			got.WriteString(e.Text)
		}
		if e.ToolCall != nil {
			calls = append(calls, *e.ToolCall)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "hello" {
		t.Fatalf("text=%q", got.String())
	}
	if len(calls) != 1 || calls[0].Name != "lookup" {
		t.Fatalf("calls=%#v", calls)
	}
	var args map[string]any
	if json.Unmarshal(calls[0].Arguments, &args) != nil || args["q"] != "x" {
		t.Fatalf("args=%s", calls[0].Arguments)
	}
}
func TestAnthropicStreamAndSystem(t *testing.T) {
	c := newTestClient(t, AnthropicMessages, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-secret" || r.Header.Get("anthropic-version") == "" {
			t.Error("missing auth headers")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["system"] != "one\n\ntwo" {
			t.Errorf("system=%#v", body["system"])
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\"}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\n")
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	})
	var got string
	err := c.Stream(context.Background(), Request{Messages: []Message{{Role: "system", Content: "one"}, {Role: "system", Content: "two"}, {Role: "user", Content: "hi"}}, MaxTokens: 20}, func(e Event) error { got += e.Text; return nil })
	if err != nil {
		t.Fatal(err)
	}
	if got != "ok" {
		t.Fatalf("got %q", got)
	}
}
func TestStreamErrors(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		c := newTestClient(t, OpenAICompatible, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401); fmt.Fprint(w, "test-secret") })
		err := c.Stream(context.Background(), Request{Messages: []Message{{Role: "user"}}, MaxTokens: 1}, func(Event) error { return nil })
		if err == nil || strings.Contains(err.Error(), "test-secret") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("callback", func(t *testing.T) {
		want := errors.New("stop")
		c := newTestClient(t, OpenAICompatible, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n\n")
		})
		err := c.Stream(context.Background(), Request{Messages: []Message{{Role: "user"}}, MaxTokens: 1}, func(Event) error { return want })
		if !errors.Is(err, want) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("missing terminal", func(t *testing.T) {
		c := newTestClient(t, OpenAICompatible, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n\n")
		})
		if err := c.Stream(context.Background(), Request{Messages: []Message{{Role: "user"}}, MaxTokens: 1}, func(Event) error { return nil }); err == nil {
			t.Fatal("expected error")
		}
	})
}
