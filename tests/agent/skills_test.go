//go:build unix

package agent_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/AlonaBot/alonabot-agent/provider"
)

// TestAgentSkills 经真实 Core、SDK Session、SQLite 和技能文件验证技能加载与工具循环。
func TestAgentSkills(t *testing.T) {
	directory := t.TempDir()
	files := map[string]string{
		"a-zeta":  "---\nname: zeta\ndescription: Zeta description\n---\nZeta body\n",
		"b-alpha": "---\nname: alpha\ndescription: Alpha description\nextra: ignored\n---\nAlpha body\n",
	}
	for name, content := range files {
		path := filepath.Join(directory, name)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var calls atomic.Int32 // Provider 回调与测试 goroutine 共享请求计数。
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request completionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var found bool
		for _, tool := range request.Tools {
			if tool.Function.Name == "load_skill" {
				found = true
			}
		}
		if !found {
			t.Error("load_skill missing from provider tools")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		callSkill := func(id, name string) {
			arguments, err := json.Marshal(map[string]string{"name": name})
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Fprintf(w, `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":%q,"type":"function","function":{"name":"load_skill","arguments":%q}}]}}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`, id, string(arguments))
		}
		switch calls.Add(1) {
		case 1:
			want := "Available skills (use load_skill to read a skill):\nalpha: Alpha description\nzeta: Zeta description"
			if len(request.Messages) != 4 || request.Messages[0].Role != "system" || request.Messages[0].Content != "Persona prompt" || request.Messages[1].Role != "system" || request.Messages[1].Content != want || request.Messages[2].Role != "system" || request.Messages[2].Content != "" || request.Messages[3].Role != "user" {
				t.Errorf("unexpected initial prompt: %+v", request.Messages)
			}
			callSkill("load", "alpha")
		case 2:
			checkToolResult(t, request, "load", "Alpha body\n")
			writeDelta(w, provider.OpenAICompatible, "skills finished")
			writeEnd(w, provider.OpenAICompatible)
		default:
			t.Error("unexpected provider request")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	t.Cleanup(model.Close)
	configuration := agentConfig(model.URL, provider.OpenAICompatible)
	configuration["skills_dir"] = directory
	// Adapter 未声明按钮能力，成功返回也证明加载技能无需额外确认。
	h, admin := startModelAdmin(t, configuration)
	admin.request(t, "POST", pageRoutePrefix+"personas", `{"name":"test","prompt":"Persona prompt"}`, 201)
	profile := admin.createProfile(t, "Skills", map[string]any{"default_persona": "test"})
	admin.assignProfile(t, "main", loopSession, profile)
	h.deliver(t, "/ask skills", "member", "en-US", false)
	h.waitText(t, "skills finished")
	h.waitEntries(t, 2)
	if calls.Load() != 2 {
		t.Errorf("provider requests = %d, want 2", calls.Load())
	}
	if err := h.stop(); err != nil {
		t.Fatal(err)
	}
}
