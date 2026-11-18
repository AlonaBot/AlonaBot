//go:build unix

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

const (
	agentPackageID      = "alonabot-agent-official"
	agentPackageNote    = "package note"
	agentPackageProfile = "package profile"
	agentPackageSkill   = "Use the packaged agent and report the result.\n"
	agentPackageTask    = "package-task-fired"
)

type agentPackageCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type agentPackageMessage struct {
	Role       string             `json:"role"`
	Content    json.RawMessage    `json:"content"`
	ToolCallID string             `json:"tool_call_id"`
	ToolCalls  []agentPackageCall `json:"tool_calls"`
}

type agentPackageRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Stream    bool   `json:"stream"`
	Tools     []struct {
		Function struct{ Name, Description string }
	} `json:"tools"`
	Messages []agentPackageMessage `json:"messages"`
}

// TestAgentPackage 通过正式安装包、管理入口与真实进程验证官方 Agent 的协作主路径。
func TestAgentPackage(t *testing.T) {
	// 1. 正式打包；Agent 和探针始终从安装包运行，开发目录只提供 Telegram Adapter。
	agentArchive := packAgentComponent(t, "../../agent", agentPackageID)
	probeArchive := packAgentComponent(t, "../../plugintest/internal/probe", "plugintest_probe")
	resources := agentPackageFiles(t, agentArchive)
	parsed, err := manifest.Parse(resources["plugin.yaml"])
	if err != nil {
		t.Fatal(err)
	}
	m := parsed.Manifest
	if m.ID != agentPackageID || m.Role != "agent" || m.Source != "github.com/AlonaBot/AlonaBot" || m.Version != "0.1.0" || m.OS != runtime.GOOS || m.Arch != runtime.GOARCH || m.MigrationMode != manifest.MigrationVersioned || m.PrepareEffects != manifest.EffectsManagedDB {
		t.Fatalf("packaged manifest: %+v", m)
	}
	for _, name := range []string{m.Entrypoint, "build-info", "config_schema.json", "i18n/en-US.json", "i18n/zh-CN.json", m.Icon} {
		if len(resources[name]) == 0 {
			t.Fatalf("missing package resource %q", name)
		}
	}
	schema, problems := configschema.Parse(developmentRead(t, "../../agent/config_schema.json"))
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	if !bytes.Equal(resources["config_schema.json"], schema.Canonical()) {
		t.Fatal("package schema differs from canonical source")
	}
	for _, name := range []string{"i18n/en-US.json", "i18n/zh-CN.json", m.Icon} {
		if !bytes.Equal(resources[name], developmentRead(t, filepath.Join("../../agent", name))) {
			t.Fatalf("package resource differs from source: %s", name)
		}
	}
	mcp := filepath.Join(t.TempDir(), "mcptest")
	command := exec.Command("go", "build", "-o", mcp, "../../agent/mcptest")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build MCP: %v\n%s", err, output)
	}
	pidFile := filepath.Join(t.TempDir(), "mcp.pid")
	skills := t.TempDir()
	developmentWrite(t, filepath.Join(skills, "package-skill", "SKILL.md"), "---\nname: package-skill\ndescription: Package integration skill\n---\n"+agentPackageSkill)

	// Provider 记录用锁保护；测试只在 Core 停止后读取完整请求序列。
	var mu sync.Mutex
	var requests []agentPackageRequest
	var scheduledAt string
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request agentPackageRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode Provider request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		requests = append(requests, request)
		step := len(requests)
		mu.Unlock()
		t.Logf("Provider request %d: %+v", step, request)
		w.Header().Set("Content-Type", "text/event-stream")
		switch step {
		case 1:
			writeAgentPackageCall(t, w, "memory-call", agentPackageTool(t, request, "Update chat notes"), map[string]any{"operations": []any{
				map[string]any{"op": "add", "target": "memory", "text": agentPackageNote},
				map[string]any{"op": "add", "target": "user", "text": agentPackageProfile},
			}})
		case 2:
			writeAgentPackageCall(t, w, "skill-call", agentPackageTool(t, request, "Load the full body"), map[string]any{"name": "package-skill"})
		case 3:
			writeAgentPackageCall(t, w, "mcp-call", agentPackageTool(t, request, "MCP echo"), map[string]any{"text": "package MCP parameters"})
		case 4:
			writeAgentPackageText(t, w, "package ", "complete")
		case 5:
			writeAgentPackageText(t, w, "image accepted")
		case 6:
			writeAgentPackageText(t, w, "image replayed")
		case 7:
			writeAgentPackageText(t, w, "clean accepted")
		case 8:
			at := time.Now().Add(3 * time.Second).UTC().Format(time.RFC3339Nano)
			mu.Lock()
			scheduledAt = at
			mu.Unlock()
			writeAgentPackageCall(t, w, "schedule-call", agentPackageTool(t, request, "Schedule an agent prompt"), map[string]any{"prompt": agentPackageTask, "at": at})
		case 9:
			writeAgentPackageText(t, w, "task scheduled")
		case 10:
			writeAgentPackageText(t, w, agentPackageTask)
		case 11:
			writeAgentPackageText(t, w, "generated")
		default:
			t.Errorf("unexpected Provider request %d: %+v", step, request)
		}
	}))
	defer model.Close()
	server := telegramtest.NewServer(t, "agent-package-token")
	server.SetMe(101, "agent_package_bot")
	dir := t.TempDir()
	developmentWrite(t, filepath.Join(dir, "core_config.json"), fmt.Sprintf(`{
		"listen":"127.0.0.1:0","adapters":{"telegram":{
		"package":"alonabot-adapter-telegram","default_instance":"package_bot","instances":{"package_bot":{
		"token":"agent-package-token","api_base_url":%q,"proxy_mode":"none"}}}}}`, server.URL()))
	c := startEmptyCore(t, dir, "../../adapters/telegram")
	defer func() { t.Logf("Core log:\n%s\nBot API: %+v", c.log(), server.Requests()) }()
	developmentPID(t, c, "alonabot-adapter-telegram")
	telegramEventually(t, c, func() bool { return server.Calls("getUpdates") > 0 })
	c.management(t, dir, "", "admin", "add", "telegram:88")
	c.management(t, dir, "", "install", agentArchive, "--yes")
	installed := findPackage(t, c, dir, agentPackageID)
	if installed.State != "Installed" || installed.Enabled || installed.Active != "" || installed.Signed || installed.Role != "agent" || installed.Version != "0.1.0" || installed.Source != m.Source {
		t.Fatalf("installed Agent: %+v", installed)
	}
	developmentWrite(t, filepath.Join(dir, "plugins", agentPackageID, "config.json"), fmt.Sprintf(`{
		"providers":[{"custom":{"name":"synthetic","protocol":"openai-compatible","base_url":%q,
		"api_key":"synthetic-key"}}],
		"models":[{"name":"synthetic","provider":"synthetic","model":"test-model","type":"chat","max_tokens":128,"image_input":true}],
		"chat_model":"synthetic","personas":[{"name":"package","prompt":"Package persona"}],"default_persona":"package",
		"skills_dir":%q,"mcp_servers":[{"name":"package","transport":"stdio","command":%q,"args":["--pid-file",%q]}]
	}`, model.URL, skills, mcp, pidFile))
	c.management(t, dir, "", "enable", agentPackageID)
	active := findPackage(t, c, dir, agentPackageID)
	if !active.Enabled || active.Active == "" || active.State != "Activated" {
		t.Fatalf("enabled Agent: %+v", active)
	}
	agentPID := developmentPID(t, c, agentPackageID)
	developmentRead(t, filepath.Join(dir, "plugins", agentPackageID, "generations", active.Active, "db.sqlite"))
	c.management(t, dir, "", "install", probeArchive, "--yes")
	c.management(t, dir, "", "enable", "plugintest_probe")
	probe := findPackage(t, c, dir, "plugintest_probe")
	if !probe.Enabled || probe.Active == "" || probe.State != "Activated" {
		t.Fatalf("enabled probe: %+v", probe)
	}
	probePID := developmentPID(t, c, "plugintest_probe")

	sequence := 0
	send := func(text string) {
		sequence++
		telegramMessage(t, server, sequence, map[string]any{"text": text, "chat": map[string]any{"id": 9, "type": "private"}})
	}
	waitText := func(text string) {
		t.Helper()
		telegramEventually(t, c, func() bool {
			for _, request := range server.Requests() {
				if (request.Method == "sendMessage" || request.Method == "editMessageText") && request.Params["text"] == text && request.Params["chat_id"] == float64(9) {
					return true
				}
			}
			return false
		})
	}
	// 平台已显示回复时历史可能尚未提交；下一轮以管理入口可见的条目数作屏障。
	waitEntries := func(count int) {
		t.Helper()
		telegramEventually(t, c, func() bool {
			_, status, raw := t13Request(t, c, dir, http.MethodGet, "/api/v1/plugins/"+agentPackageID+"/conversations")
			var summaries []struct {
				Key     string `json:"key"`
				Entries int    `json:"entries"`
			}
			if status != http.StatusOK || json.Unmarshal(raw, &summaries) != nil {
				t.Fatalf("conversation summaries: %d %s", status, raw)
			}
			for _, summary := range summaries {
				if summary.Key == "telegram:9@package_bot" {
					return summary.Entries == count
				}
			}
			return count == 0
		})
	}
	confirm := func(operation string, arguments ...string) {
		t.Helper()
		var confirmation telegramtest.Request
		telegramEventually(t, c, func() bool {
			for _, request := range server.Requests() {
				text, _ := request.Params["text"].(string)
				if request.Method == "sendMessage" && strings.Contains(text, operation) {
					confirmation = request
					return true
				}
			}
			return false
		})
		for _, argument := range arguments {
			if !strings.Contains(confirmation.Params["text"].(string), argument) {
				t.Fatalf("confirmation omits %q: %+v", argument, confirmation)
			}
		}
		keyboard := confirmation.Params["reply_markup"].(map[string]any)["inline_keyboard"].([]any)
		allow := keyboard[0].([]any)[0].(map[string]any)
		if allow["text"] != "Allow" {
			t.Fatalf("confirmation allow button: %+v", allow)
		}
		// 尚未点击时 Provider 必须停在本次工具调用，确认之后才会收到结果。
		mu.Lock()
		before := len(requests)
		mu.Unlock()
		want := 3
		if operation == "agent/schedule_task" {
			want = 8
		}
		if before != want {
			t.Fatalf("tool continued before confirmation: requests=%d want %d", before, want)
		}
		sequence++
		telegramUpdate(t, server, map[string]any{"update_id": sequence, "callback_query": map[string]any{
			"id": fmt.Sprintf("package-confirm-%d", sequence), "from": map[string]any{"id": 88, "first_name": "User", "language_code": "en"},
			"data": allow["callback_data"], "message": map[string]any{"message_id": 1, "chat": map[string]any{"id": 9, "type": "private"}},
		}})
	}

	// 2. 开启自由对话，依次完成自身 memory、技能与需平台确认的 MCP 工具往返。
	send("/agent on")
	waitText("Free conversation is on.")
	send("package conversation")
	confirm("mcp/package/echo/tool", `"text":"package MCP parameters"`)
	waitText("package complete")
	waitEntries(3)
	pid, err := strconv.Atoi(strings.TrimSpace(string(developmentRead(t, pidFile))))
	if err != nil {
		t.Fatal(err)
	}

	// 3. 重置后读取长期笔记和真实入站图片；下一轮回放，再重置后图片清空。
	send("/reset")
	waitEntries(0)
	var picture bytes.Buffer
	if err := jpeg.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil); err != nil {
		t.Fatal(err)
	}
	server.AddFile("package-photo", "photos/package.jpg", picture.Bytes())
	sequence++
	telegramMessage(t, server, sequence, map[string]any{"caption": "image round", "chat": map[string]any{"id": 9, "type": "private"},
		"photo": []any{map[string]any{"file_id": "package-photo", "file_unique_id": "package-photo-unique", "width": 1, "height": 1, "file_size": picture.Len()}},
	})
	waitText("image accepted")
	waitEntries(3)
	send("replay round")
	waitText("image replayed")
	waitEntries(5)
	send("/reset")
	waitEntries(0)
	send("clean round")
	waitText("clean accepted")
	waitEntries(3)

	// 4. 管理员确认未来 at 任务，等待真实 Scheduler 派发到原聊天。
	send("schedule round")
	confirm("agent/schedule_task", "at:", "prompt: "+agentPackageTask)
	waitText("task scheduled")
	waitEntries(5)
	waitText(agentPackageTask)
	send("/llmask")
	waitText("generated")

	// 5. 管理停用、正常停止后再核对请求总数与进程回收，覆盖一次性任务的唯一发送。
	c.management(t, dir, "", "disable", agentPackageID)
	stopped := findPackage(t, c, dir, agentPackageID)
	if stopped.Enabled || stopped.State != "Stopped" || stopped.Active != active.Active {
		t.Fatalf("disabled Agent: %+v", stopped)
	}
	if !pluginExited(agentPID, 2*time.Second) || !pluginExited(pid, 2*time.Second) {
		t.Fatalf("Agent or MCP remains: Agent=%d MCP=%d", agentPID, pid)
	}
	stopTestCore(t, c)
	if !pluginExited(probePID, 2*time.Second) {
		t.Fatalf("probe remains: %d", probePID)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 11 {
		t.Fatalf("Provider requests=%d want 11", len(requests))
	}
	for i, request := range requests {
		if request.Model != "test-model" || request.MaxTokens != 128 || !request.Stream {
			t.Errorf("Provider request %d settings: %+v", i+1, request)
		}
	}
	checkAgentPackageResult(t, requests[1], "memory-call", "{}")
	checkAgentPackageResult(t, requests[2], "skill-call", agentPackageSkill)
	mcpResult := checkAgentPackageResult(t, requests[3], "mcp-call", "")
	var result struct {
		StructuredContent struct {
			Text  string `json:"text"`
			Calls int    `json:"calls"`
			PID   int    `json:"pid"`
		} `json:"structuredContent"`
	}
	if err := json.Unmarshal([]byte(mcpResult), &result); err != nil || result.StructuredContent.Text != "package MCP parameters" || result.StructuredContent.Calls != 1 || result.StructuredContent.PID != pid {
		t.Errorf("MCP actual parameters/count/PID: %s (%v)", mcpResult, err)
	}
	scheduled := checkAgentPackageResult(t, requests[8], "schedule-call", "")
	var task struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(scheduled), &task); err != nil || task.ID == "" {
		t.Errorf("schedule result: %s (%v)", scheduled, err)
	}
	var trigger struct {
		At string `json:"at"`
	}
	call := requests[8].Messages[len(requests[8].Messages)-2].ToolCalls[0]
	if err := json.Unmarshal([]byte(call.Function.Arguments), &trigger); err != nil || trigger.At != scheduledAt {
		t.Errorf("scheduled at: %+v, want %s (%v)", trigger, scheduledAt, err)
	}
	wantImage := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(picture.Bytes())
	for _, i := range []int{4, 5, 6} {
		texts, images := agentPackageContents(t, requests[i])
		if !strings.Contains(strings.Join(texts, "\n"), agentPackageNote) || !strings.Contains(strings.Join(texts, "\n"), agentPackageProfile) {
			t.Errorf("persistent memory missing from request %d: %q", i+1, texts)
		}
		want := []string{wantImage}
		if i == 6 {
			want = nil
		}
		if !reflect.DeepEqual(images, want) {
			t.Errorf("request %d images=%q want %q", i+1, images, want)
		}
	}
	background, images := agentPackageContents(t, requests[9])
	if len(images) != 0 || !reflect.DeepEqual(background, []string{"Package persona", "Available skills (use load_skill to read a skill):\npackage-skill: Package integration skill", agentPackageTask}) {
		t.Errorf("background history or memory: %q, images=%q", background, images)
	}
	llmText, llmImages := agentPackageContents(t, requests[10])
	if len(requests[10].Tools) != 0 || len(requests[10].Messages) != 1 || requests[10].Messages[0].Role != "user" || !reflect.DeepEqual(llmText, []string{"  original prompt\n"}) || len(llmImages) != 0 {
		t.Errorf("llm.Ask request: %+v", requests[10])
	}
	var fired, confirmations int
	for _, request := range server.Requests() {
		if request.Method == "sendMessage" && request.Params["text"] == agentPackageTask {
			fired++
		}
		if request.Method == "sendMessage" && request.Params["reply_markup"] != nil {
			confirmations++
		}
	}
	if fired != 1 || confirmations != 2 {
		t.Errorf("background sends=%d confirmations=%d, want 1/2", fired, confirmations)
	}
}

func packAgentComponent(t *testing.T, project, id string) string {
	t.Helper()
	generated := filepath.Join(project, "alonabot_gen.go")
	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Fatalf("unexpected generated source: %v", err)
	}
	out := t.TempDir()
	command := exec.Command(filepath.Join(binDir, "alonabot"), "pack", "--dir", project, "--asset-url", "https://example.invalid/"+id+".tar.gz", "--out", out)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("pack %s: %v\n%s", project, err, output)
	} else {
		t.Logf("pack %s:\n%s", project, output)
	}
	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Fatalf("pack wrote generated source: %v", err)
	}
	return filepath.Join(out, id+"-0.1.0-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz")
}

func agentPackageFiles(t *testing.T, archive string) map[string][]byte {
	t.Helper()
	file, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	files := make(map[string][]byte)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files
		}
		if err != nil {
			t.Fatal(err)
		}
		files[header.Name], err = io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func agentPackageTool(t *testing.T, request agentPackageRequest, description string) string {
	t.Helper()
	for _, tool := range request.Tools {
		if strings.HasPrefix(tool.Function.Description, description) {
			return tool.Function.Name
		}
	}
	t.Errorf("missing tool %q: %+v", description, request.Tools)
	return ""
}

func writeAgentPackageCall(t *testing.T, w http.ResponseWriter, id, name string, args any) {
	t.Helper()
	arguments, err := json.Marshal(args)
	if err != nil {
		t.Error(err)
		return
	}
	writeAgentPackageEvent(t, w, map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": id, "type": "function", "function": map[string]any{"name": name, "arguments": string(arguments)}}}}, "tool_calls")
}

func writeAgentPackageText(t *testing.T, w http.ResponseWriter, texts ...string) {
	t.Helper()
	for i, text := range texts {
		finish := ""
		if i == len(texts)-1 {
			finish = "stop"
		}
		writeAgentPackageEvent(t, w, map[string]any{"content": text}, finish)
	}
}

func writeAgentPackageEvent(t *testing.T, w http.ResponseWriter, delta map[string]any, finish string) {
	t.Helper()
	event := map[string]any{"choices": []any{map[string]any{"index": 0, "delta": delta}}}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", raw); err != nil {
		t.Error(err)
		return
	}
	if finish != "" {
		if _, err := fmt.Fprintf(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":%q}]}\n\ndata: [DONE]\n\n", finish); err != nil {
			t.Error(err)
		}
	}
	w.(http.Flusher).Flush()
}

func checkAgentPackageResult(t *testing.T, request agentPackageRequest, id, expected string) string {
	t.Helper()
	if len(request.Messages) < 2 {
		t.Fatalf("missing tool exchange %s: %+v", id, request)
	}
	last, previous := request.Messages[len(request.Messages)-1], request.Messages[len(request.Messages)-2]
	var result string
	if err := json.Unmarshal(last.Content, &result); err != nil {
		t.Fatal(err)
	}
	if last.Role != "tool" || last.ToolCallID != id || previous.Role != "assistant" || len(previous.ToolCalls) != 1 || previous.ToolCalls[0].ID != id {
		t.Fatalf("tool exchange %s: previous=%+v result=%+v", id, previous, last)
	}
	if expected != "" && result != expected {
		t.Errorf("tool %s result=%q want %q", id, result, expected)
	}
	return result
}

func agentPackageContents(t *testing.T, request agentPackageRequest) (texts, images []string) {
	t.Helper()
	for _, message := range request.Messages {
		var text string
		if err := json.Unmarshal(message.Content, &text); err == nil {
			texts = append(texts, text)
			continue
		}
		var parts []struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			ImageURL struct {
				URL string `json:"url"`
			} `json:"image_url"`
		}
		if err := json.Unmarshal(message.Content, &parts); err != nil {
			t.Fatal(err)
		}
		for _, part := range parts {
			switch part.Type {
			case "text":
				text += part.Text
			case "image_url":
				images = append(images, part.ImageURL.URL)
			}
		}
		texts = append(texts, text)
	}
	return texts, images
}
