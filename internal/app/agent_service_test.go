//go:build unix

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/AlonaBot/AlonaBot/agent"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	coreoptions "github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
	optionsapi "github.com/AlonaBot/AlonaBot/sdk/options"
	serviceapi "github.com/AlonaBot/AlonaBot/sdk/service"
	"github.com/google/jsonschema-go/jsonschema"
)

const (
	serviceAgentID  = "alonabot_agent_official"
	serviceCallerID = "llm_allowed"
	serviceDeniedID = "llm_denied"
	serviceModel    = "local-model"
	serviceAPIKey   = "synthetic-provider-key"
	serviceProvider = "Primary"
	serviceGenerate = "generate"
	serviceAsk      = "ask"
	servicePrivate  = "private"
	serviceSource   = "@agent/llm.models"
	serviceText     = "hello"
	serviceStream   = "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":5,\"completion_tokens\":7}}\n\n" +
		"data: [DONE]\n\n"
)

// TestAgentServices 通过真实 Core、SDK Session 与官方 Agent 验证生成及选项查询的主路径。
func TestAgentServices(t *testing.T) {
	for _, restricted := range []bool{false, true} {
		t.Run(fmt.Sprintf("restricted=%v", restricted), func(t *testing.T) {
			var mu sync.Mutex // Provider HTTP 请求与测试断言共享请求记录。
			var requests []map[string]any
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer "+serviceAPIKey {
					t.Error("models must share the provider API key")
				}
				var request map[string]any
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				mu.Lock()
				requests = append(requests, request)
				mu.Unlock()
				w.Header().Set("Content-Type", "text/event-stream")
				if _, err := fmt.Fprint(w, serviceStream); err != nil {
					t.Error(err)
				}
			}))
			defer upstream.Close()
			configuration := agentServiceConfig(upstream.URL)
			if restricted {
				configuration["llm_service_allowed_plugins"] = []string{serviceCallerID}
			}
			core := startAgentServices(t, configuration)
			request := llm.Request{Messages: []llm.Message{{Role: "user", Text: "question"}}, System: "system prompt", Model: "Backup", MaxOutputTokens: 37}
			response, failure := callAgentService(t, core, serviceCallerID, serviceGenerate, request)
			if failure != nil {
				t.Fatal(failure)
			}
			if response.Text != serviceText || response.Model != "Backup" || response.InputTokens == nil || *response.InputTokens != 5 || response.OutputTokens == nil || *response.OutputTokens != 7 {
				t.Fatalf("generation = %+v", response)
			}
			mu.Lock()
			before := len(requests)
			generated := requests[0]
			mu.Unlock()
			if generated["model"] != "backup-model" || generated["max_tokens"] != float64(37) {
				t.Fatalf("selected model upstream = %+v", generated)
			}
			_, failure = callAgentService(t, core, serviceDeniedID, serviceGenerate, request)
			if restricted {
				if failure.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
					t.Fatalf("denied caller = %v", failure)
				}
			} else if failure != nil {
				t.Fatal(failure)
			}
			mu.Lock()
			count := len(requests)
			mu.Unlock()
			if restricted && count != before || !restricted && count != before+1 {
				t.Fatalf("provider requests = %d after %d", count, before)
			}

			prompt := "  ask verbatim\n"
			response, failure = callAgentService(t, core, serviceCallerID, serviceAsk, llm.Request{Messages: []llm.Message{{Role: "user", Text: prompt}}})
			if failure != nil || response.Text != serviceText {
				t.Fatalf("Ask = %+v, %v", response, failure)
			}
			mu.Lock()
			ask := requests[len(requests)-1]
			mu.Unlock()
			if !reflect.DeepEqual(ask["messages"], []any{map[string]any{"role": "user", "content": prompt}}) || ask["model"] != serviceModel || ask["max_tokens"] != float64(23) || ask["tools"] != nil {
				t.Fatalf("Ask upstream = %+v", ask)
			}
			for _, id := range []string{serviceAgentID, serviceCallerID, serviceDeniedID} {
				entries, err := core.store.Conversations(t.Context(), id)
				if err != nil || len(entries) != 0 {
					t.Fatalf("conversations for %s = %v, %v", id, entries, err)
				}
			}

			for _, query := range []struct {
				component, field string
				input            coreoptions.Query
				want             []string
			}{
				{serviceAgentID, "chat_model", coreoptions.Query{}, []string{serviceProvider, "Backup"}},
				{serviceCallerID, "model", coreoptions.Query{Search: "pRiM"}, []string{serviceProvider}},
				{serviceCallerID, "model", coreoptions.Query{Values: []string{"Backup", "missing"}, Search: "absent"}, []string{"Backup"}},
			} {
				page, err := core.assembly.options.Query(t.Context(), query.component, coreoptions.ConfigSchema, query.field, query.input)
				if err != nil {
					t.Fatal(err)
				}
				var names []string
				for _, option := range page.Options {
					names = append(names, option.Value)
					model := serviceModel
					if option.Value == "Backup" {
						model = "backup-model"
					}
					if option.Label != option.Value || option.Description != serviceProvider+" · "+model {
						t.Fatalf("provider option = %+v", option)
					}
				}
				encoded, err := json.Marshal(page)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(names, query.want) || page.More || strings.Contains(string(encoded), upstream.URL) || strings.Contains(string(encoded), serviceAPIKey) {
					t.Fatalf("provider page = %s; want %v", encoded, query.want)
				}
			}
			_, failure = callAgentService(t, core, serviceCallerID, servicePrivate, request)
			if failure.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
				t.Fatalf("external options call = %v", failure)
			}

			const replacementID = "community_agent"
			generationCallers := make(chan string, 1) // 将 Handler 观察到的可信身份交回测试。
			replacement := sdk.NewPlugin()
			genresources.Bind(replacement, agentServiceResources(t, replacementID, "agent"))
			serviceapi.Serve(replacement, llm.GenerateService).Access(serviceapi.AccessPublic).Handle(func(c *sdk.Context, _ llm.Request) (llm.Response, error) {
				generationCallers <- c.Caller()
				return llm.Response{Text: replacementID, Model: replacementID}, nil
			})
			optionsapi.Serve(replacement, "llm.models", func(context.Context, optionsapi.Query) (optionsapi.Page, error) {
				return optionsapi.Page{Options: []optionsapi.Option{{Value: replacementID, Label: replacementID}}}, nil
			})
			if err := core.operation(t.Context(), func(ctx context.Context) error {
				if err := core.stopManaged(serviceAgentID, pb.StopReason_STOP_REASON_DISABLE); err != nil {
					return err
				}
				if err := core.store.Uninstall(ctx, serviceAgentID, false, false); err != nil {
					return err
				}
				installed, err := core.store.Installs(ctx)
				if err != nil {
					return err
				}
				entries := []Component{PluginComponent(replacement)}
				if err := validateComponents(entries, installed); err != nil {
					return err
				}
				return core.startDevelopment(ctx, core.dir, &entries[0])
			}); err != nil {
				t.Fatal(err)
			}
			response, failure = callAgentService(t, core, serviceCallerID, serviceGenerate, request)
			if failure != nil || response.Text != replacementID {
				t.Fatalf("replacement generation = %+v, %v", response, failure)
			}
			if caller := <-generationCallers; caller != serviceCallerID {
				t.Fatalf("replacement caller = %q", caller)
			}
			page, err := core.assembly.options.Query(t.Context(), serviceCallerID, coreoptions.ConfigSchema, "model", coreoptions.Query{})
			if err != nil || len(page.Options) != 1 || page.Options[0].Value != replacementID {
				t.Fatalf("replacement options = %+v, %v", page, err)
			}
		})
	}
}

func agentServiceConfig(url string) map[string]any {
	models := []any{}
	for _, name := range []string{serviceProvider, "Backup"} {
		model := serviceModel
		if name == "Backup" {
			model = "backup-model"
		}
		models = append(models, map[string]any{"name": name, "provider": serviceProvider, "model": model, "type": "chat", "max_tokens": 23})
	}
	return map[string]any{"providers": []any{map[string]any{"custom": map[string]any{"name": serviceProvider, "protocol": "openai-compatible", "base_url": url, "api_key": serviceAPIKey}}}, "models": models, "chat_model": serviceProvider, "personas": []any{map[string]any{"name": "unused", "prompt": "must not enter generation"}}, "default_persona": "unused"}
}

func startAgentServices(t *testing.T, configuration map[string]any) *Core {
	t.Helper()
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	dir := t.TempDir()
	path := filepath.Join(dir, "plugins", serviceAgentID)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	root := agent.New()
	genresources.Bind(root, agentServiceFixture)
	listen := "127.0.0.1:0"
	opts := Options{DataDir: dir, Listen: &listen, Args: []string{"STOP_DRAIN=50ms"}, Components: []Component{PluginComponent(root)}}
	for _, id := range []string{serviceCallerID, serviceDeniedID} {
		p := sdk.NewPlugin()
		genresources.Bind(p, agentServiceResources(t, id, "plugin"))
		serviceapi.Serve(p, serviceapi.JSON[llm.Request, llm.Response](id+"/"+serviceGenerate)).Handle(func(c *sdk.Context, req llm.Request) (llm.Response, error) {
			return llm.Generate(c.Context(), req)
		})
		serviceapi.Serve(p, serviceapi.JSON[llm.Request, llm.Response](id+"/"+serviceAsk)).Handle(func(c *sdk.Context, req llm.Request) (llm.Response, error) {
			text, err := llm.Ask(c.Context(), req.Messages[0].Text)
			return llm.Response{Text: text}, err
		})
		serviceapi.Serve(p, serviceapi.JSON[llm.Request, llm.Response](id+"/"+servicePrivate)).Handle(func(c *sdk.Context, _ llm.Request) (llm.Response, error) {
			_, err := serviceapi.JSON[optionsapi.Query, optionsapi.Page](serviceSource).Call(c.Context(), optionsapi.Query{Limit: 10})
			return llm.Response{}, err
		})
		opts.Components = append(opts.Components, PluginComponent(p))
	}
	ctx, cancel := context.WithCancel(t.Context())
	core, err := Start(ctx, opts)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		if err := core.Wait(); err != nil {
			t.Error(err)
		}
	})
	return core
}

func agentServiceResources(t *testing.T, id, role string) genresources.Resources {
	t.Helper()
	schemas := maps.Clone(agentServiceFixture.ServiceSchemas)
	input, err := jsonschema.For[optionsapi.Query](nil)
	if err != nil {
		t.Fatal(err)
	}
	output, err := jsonschema.For[optionsapi.Page](nil)
	if err != nil {
		t.Fatal(err)
	}
	for id, schema := range map[string]*jsonschema.Schema{validate.OptionsQuerySchemaID: input, validate.OptionsPageSchemaID: output} {
		data, err := json.Marshal(schema)
		if err != nil {
			t.Fatal(err)
		}
		schemas[id] = data
	}
	return genresources.Resources{Manifest: []byte(fmt.Sprintf("id: %s\nname: Service test\nversion: 1.0.0\nsource: test/services\nrole: %s\n", id, role)), ServiceSchemas: schemas, ConfigSchema: []byte(`{"model":{"type":"string","options_from":{"source":"@agent/llm.models","filter":{"type":"chat"}}}}`)}
}

func callAgentService(t *testing.T, core *Core, id, service string, req llm.Request) (llm.Response, *pb.ErrorInfo) {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	response := core.assembly.broker.Call(t.Context(), router.Caller{}, &wire.Inbound{Target: id + "/" + service, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: llm.SchemaIn, SchemaVersion: llm.SchemaVersion, Bytes: data}})
	if response.Err != nil {
		return llm.Response{}, response.Err
	}
	if response.Payload.SchemaID != llm.SchemaOut || response.Payload.SchemaVersion != llm.SchemaVersion {
		t.Fatalf("response schema = %+v", response.Payload)
	}
	var result llm.Response
	if err := json.Unmarshal(response.Payload.Bytes, &result); err != nil {
		t.Fatal(err)
	}
	return result, nil
}

var agentServiceFixture = genresources.Resources{
	Manifest:     []byte("# 官方 Agent 组件的工程身份与静态图标。\nid: alonabot_agent_official\nname: AlonaBot Agent\nversion: 0.1.0\nauthor: AlonaBot\nrole: agent\ndescription: Official conversational agent\nsource: github.com/AlonaBot/AlonaBot\nicon: assets/icons/agent.svg\n"),
	ConfigSchema: []byte("{\"providers\":{\"type\":\"template_list\",\"description\":\"Providers\",\"templates\":{\"openai\":{\"name\":\"OpenAI\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"openai\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"default\":\"https://api.openai.com/v1\"},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}},\"anthropic\":{\"name\":\"Anthropic\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"anthropic-messages\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"default\":\"https://api.anthropic.com\"},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}},\"gemini\":{\"name\":\"Gemini\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"gemini-generate-content\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"default\":\"https://generativelanguage.googleapis.com/v1beta\"},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}},\"deepseek\":{\"name\":\"DeepSeek\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"openai-compatible\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"default\":\"https://api.deepseek.com/v1\"},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}},\"ollama\":{\"name\":\"Ollama\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"ollama-chat\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"default\":\"http://localhost:11434\"},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}},\"custom\":{\"name\":\"Custom\",\"display_item\":\"name\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"protocol\":{\"type\":\"string\",\"description\":\"Protocol\",\"options\":[\"openai\",\"openai-compatible\",\"anthropic-messages\",\"openai-responses\",\"gemini-generate-content\",\"ollama-chat\"],\"default\":\"openai-compatible\"},\"base_url\":{\"type\":\"string\",\"description\":\"Base URL\",\"required\":true},\"api_key\":{\"type\":\"string\",\"description\":\"API key\",\"secret\":true}}}}},\"personas\":{\"type\":\"list\",\"description\":\"Personas\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"prompt\":{\"type\":\"text\",\"description\":\"System prompt\"}}},\"default_persona\":{\"type\":\"string\",\"description\":\"Default persona\"},\"history_entries\":{\"type\":\"int\",\"description\":\"History entries\",\"default\":20,\"slider\":{\"min\":0,\"max\":500,\"step\":1}},\"mcp_servers\":{\"type\":\"list\",\"description\":\"MCP servers\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"transport\":{\"type\":\"string\",\"description\":\"Transport\",\"options\":[\"stdio\",\"streamable-http\"],\"default\":\"stdio\"},\"command\":{\"type\":\"string\",\"description\":\"Command\",\"condition\":{\"transport\":\"stdio\"}},\"args\":{\"type\":\"list\",\"description\":\"Arguments\",\"condition\":{\"transport\":\"stdio\"}},\"env\":{\"type\":\"list\",\"description\":\"Environment variables\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"value\":{\"type\":\"string\",\"description\":\"Secret value\",\"secret\":true}},\"condition\":{\"transport\":\"stdio\"}},\"url\":{\"type\":\"string\",\"description\":\"URL\",\"condition\":{\"transport\":\"streamable-http\"}},\"headers\":{\"type\":\"list\",\"description\":\"HTTP headers\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"value\":{\"type\":\"string\",\"description\":\"Secret value\",\"secret\":true}},\"condition\":{\"transport\":\"streamable-http\"}}}},\"skills_dir\":{\"type\":\"string\",\"description\":\"Skills directory\",\"hint\":\"Empty disables skills. Relative paths use the plugin working directory. Restart the plugin after changing skills.\",\"default\":\"\"},\"llm_service_allowed_plugins\":{\"type\":\"list\",\"description\":\"Plugins allowed to generate text\",\"default\":[\"*\"]},\"chat_model\":{\"type\":\"string\",\"description\":\"Chat model\",\"options_from\":{\"source\":\"@agent/llm.models\",\"filter\":{\"type\":\"chat\"}}},\"models\":{\"type\":\"list\",\"description\":\"Models\",\"items\":{\"name\":{\"type\":\"string\",\"description\":\"Name\",\"required\":true},\"provider\":{\"type\":\"string\",\"description\":\"Provider\",\"required\":true},\"model\":{\"type\":\"string\",\"description\":\"Model ID\",\"required\":true},\"type\":{\"type\":\"string\",\"description\":\"Model type\",\"options\":[\"chat\",\"speech_to_text\",\"text_to_speech\",\"embedding\",\"rerank\",\"decision\"],\"default\":\"chat\"},\"max_tokens\":{\"type\":\"int\",\"description\":\"Maximum output tokens\",\"hint\":\"Must be at least 1.\",\"default\":2048,\"condition\":{\"type\":\"chat\"}},\"image_input\":{\"type\":\"bool\",\"description\":\"Image input\",\"default\":false,\"condition\":{\"type\":\"chat\"}}}}}"),
	I18n: map[string][]byte{
		"en-US.json": []byte("{\n  \"metadata\": {\n    \"name\": \"AlonaBot Agent\",\n    \"description\": \"Official conversational agent\"\n  },\n  \"messages\": {\n    \"agent.skills.header\": \"Available skills (use load_skill to read a skill):\",\n    \"agent.memory_error\": \"Memory could not be read ({code}).\",\n    \"agent.error\": \"The provider request failed ({code}).\",\n    \"agent.enabled\": \"Free conversation is on.\",\n    \"agent.disabled\": \"Free conversation is off.\",\n    \"agent.personas\": \"Personas: {names}\",\n    \"agent.persona_selected\": \"Persona selected: {name}\",\n    \"agent.unknown_persona\": \"Unknown persona: {name}\",\n    \"agent.reset\": \"Conversation cleared.\",\n    \"agent.tool_limit\": \"The tool call limit has been reached. This conversation turn has ended.\",\n    \"agent.tools_error\": \"The tool directory request failed ({code}).\",\n    \"agent.no_model\": \"Ask an administrator to configure and select a chat model.\",\n    \"agent.models\": \"Models: {names}\",\n    \"agent.model_selected\": \"Model selected: {name}\",\n    \"agent.unknown_model\": \"Unknown chat model: {name}\"\n  },\n  \"config\": {\n    \"skills_dir\": {\n      \"description\": \"Skills directory\",\n      \"hint\": \"Empty disables skills. Relative paths use the plugin working directory. Restart the plugin after changing skills.\"\n    },\n    \"mcp_servers\": {\n      \"description\": \"MCP servers\",\n      \"items\": {\n        \"name\": {\n          \"description\": \"Name\"\n        },\n        \"transport\": {\n          \"description\": \"Transport\"\n        },\n        \"command\": {\n          \"description\": \"Command\"\n        },\n        \"args\": {\n          \"description\": \"Arguments\"\n        },\n        \"env\": {\n          \"description\": \"Environment variables\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"Name\"\n            },\n            \"value\": {\n              \"description\": \"Secret value\"\n            }\n          }\n        },\n        \"url\": {\n          \"description\": \"URL\"\n        },\n        \"headers\": {\n          \"description\": \"HTTP headers\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"Name\"\n            },\n            \"value\": {\n              \"description\": \"Secret value\"\n            }\n          }\n        }\n      }\n    },\n    \"llm_service_allowed_plugins\": {\n      \"description\": \"Plugins allowed to generate text\"\n    },\n    \"providers\": {\n      \"templates\": {\n        \"openai\": {\n          \"items\": {}\n        },\n        \"anthropic\": {\n          \"items\": {}\n        },\n        \"gemini\": {\n          \"items\": {}\n        },\n        \"deepseek\": {\n          \"items\": {}\n        },\n        \"ollama\": {\n          \"items\": {}\n        },\n        \"custom\": {\n          \"items\": {}\n        }\n      }\n    },\n    \"chat_model\": {\n      \"description\": \"Chat model\"\n    },\n    \"models\": {\n      \"description\": \"Models\",\n      \"items\": {\n        \"name\": {\n          \"description\": \"Name\"\n        },\n        \"provider\": {\n          \"description\": \"Provider\"\n        },\n        \"model\": {\n          \"description\": \"Model ID\"\n        },\n        \"type\": {\n          \"description\": \"Model type\"\n        },\n        \"max_tokens\": {\n          \"description\": \"Maximum output tokens\",\n          \"hint\": \"Must be at least 1.\"\n        },\n        \"image_input\": {\n          \"description\": \"Image input\"\n        }\n      }\n    }\n  }\n}\n"),
		"zh-CN.json": []byte("{\n  \"metadata\": {\n    \"name\": \"AlonaBot Agent\",\n    \"description\": \"官方对话 Agent\"\n  },\n  \"messages\": {\n    \"agent.skills.header\": \"可用技能（使用 load_skill 读取技能）：\",\n    \"agent.memory_error\": \"记忆读取失败（{code}）。\",\n    \"agent.error\": \"Provider 请求失败（{code}）。\",\n    \"agent.enabled\": \"已开启自由对话。\",\n    \"agent.disabled\": \"已关闭自由对话。\",\n    \"agent.personas\": \"人格：{names}\",\n    \"agent.persona_selected\": \"已选择人格：{name}\",\n    \"agent.unknown_persona\": \"找不到人格：{name}\",\n    \"agent.reset\": \"已清除会话历史。\",\n    \"agent.tool_limit\": \"已达到本轮工具调用上限，本轮对话已结束。\",\n    \"agent.tools_error\": \"工具目录请求失败（{code}）。\",\n    \"agent.no_model\": \"请管理员配置并选择对话模型。\",\n    \"agent.models\": \"模型：{names}\",\n    \"agent.model_selected\": \"已选择模型：{name}\",\n    \"agent.unknown_model\": \"找不到对话模型：{name}\"\n  },\n  \"commands\": {\n    \"ask\": {\n      \"description\": \"向 Agent 提问\"\n    },\n    \"agent on\": {\n      \"description\": \"开启本聊天的自由对话\"\n    },\n    \"agent off\": {\n      \"description\": \"关闭本聊天的自由对话\"\n    },\n    \"persona\": {\n      \"description\": \"列出或选择人格\"\n    },\n    \"provider\": {\n      \"description\": \"列出或选择对话模型\"\n    },\n    \"reset\": {\n      \"description\": \"清除本聊天的会话历史\"\n    }\n  },\n  \"config\": {\n    \"skills_dir\": {\n      \"description\": \"技能目录\",\n      \"hint\": \"留空关闭技能。相对路径以插件工作目录为基准，修改技能后需重启插件。\"\n    },\n    \"providers\": {\n      \"description\": \"Provider\",\n      \"templates\": {\n        \"openai\": {\n          \"name\": \"OpenAI\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        },\n        \"anthropic\": {\n          \"name\": \"Anthropic\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        },\n        \"gemini\": {\n          \"name\": \"Gemini\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        },\n        \"deepseek\": {\n          \"name\": \"DeepSeek\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        },\n        \"ollama\": {\n          \"name\": \"Ollama\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        },\n        \"custom\": {\n          \"name\": \"自定义\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"protocol\": {\n              \"description\": \"协议\"\n            },\n            \"base_url\": {\n              \"description\": \"服务地址\"\n            },\n            \"api_key\": {\n              \"description\": \"API 密钥\"\n            }\n          }\n        }\n      }\n    },\n    \"personas\": {\n      \"description\": \"人格\",\n      \"items\": {\n        \"name\": {\n          \"description\": \"名称\"\n        },\n        \"prompt\": {\n          \"description\": \"系统提示词\"\n        }\n      }\n    },\n    \"default_persona\": {\n      \"description\": \"默认人格\"\n    },\n    \"history_entries\": {\n      \"description\": \"历史条数\"\n    },\n    \"mcp_servers\": {\n      \"description\": \"MCP 服务器\",\n      \"items\": {\n        \"name\": {\n          \"description\": \"名称\"\n        },\n        \"transport\": {\n          \"description\": \"传输方式\"\n        },\n        \"command\": {\n          \"description\": \"命令\"\n        },\n        \"args\": {\n          \"description\": \"参数\"\n        },\n        \"env\": {\n          \"description\": \"环境变量\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"value\": {\n              \"description\": \"秘密值\"\n            }\n          }\n        },\n        \"url\": {\n          \"description\": \"服务地址\"\n        },\n        \"headers\": {\n          \"description\": \"HTTP 请求头\",\n          \"items\": {\n            \"name\": {\n              \"description\": \"名称\"\n            },\n            \"value\": {\n              \"description\": \"秘密值\"\n            }\n          }\n        }\n      }\n    },\n    \"llm_service_allowed_plugins\": {\n      \"description\": \"允许生成文本的插件\"\n    },\n    \"chat_model\": {\n      \"description\": \"对话模型\"\n    },\n    \"models\": {\n      \"description\": \"模型\",\n      \"items\": {\n        \"name\": {\n          \"description\": \"名称\"\n        },\n        \"provider\": {\n          \"description\": \"提供商\"\n        },\n        \"model\": {\n          \"description\": \"模型 ID\"\n        },\n        \"type\": {\n          \"description\": \"模型类型\"\n        },\n        \"max_tokens\": {\n          \"description\": \"最大输出 token 数\",\n          \"hint\": \"至少为 1。\"\n        },\n        \"image_input\": {\n          \"description\": \"图片输入\"\n        }\n      }\n    }\n  }\n}\n"),
	},
	ServiceSchemas: map[string][]byte{
		"github.com/AlonaBot/AlonaBot/agent.memoryRequest":  []byte("{\"$schema\":\"https://json-schema.org/draft/2020-12/schema\",\"properties\":{\"operations\":{\"items\":{\"properties\":{\"old_text\":{\"type\":\"string\"},\"op\":{\"type\":\"string\"},\"target\":{\"type\":\"string\"},\"text\":{\"type\":\"string\"}},\"required\":[\"op\",\"target\"],\"type\":\"object\"},\"type\":[\"array\",\"null\"]}},\"required\":[\"operations\"],\"type\":\"object\"}"),
		"github.com/AlonaBot/AlonaBot/agent.memoryResponse": []byte("{\"$schema\":\"https://json-schema.org/draft/2020-12/schema\",\"properties\":{},\"type\":\"object\"}"),
		"github.com/AlonaBot/AlonaBot/sdk/llm.Request":      []byte("{\"$schema\":\"https://json-schema.org/draft/2020-12/schema\",\"properties\":{\"max_output_tokens\":{\"maximum\":9223372036854775807,\"minimum\":-9223372036854775808,\"type\":\"integer\"},\"messages\":{\"items\":{\"properties\":{\"role\":{\"type\":\"string\"},\"text\":{\"type\":\"string\"}},\"required\":[\"role\",\"text\"],\"type\":\"object\"},\"type\":[\"array\",\"null\"]},\"system\":{\"type\":\"string\"},\"model\":{\"type\":\"string\"}},\"required\":[\"messages\"],\"type\":\"object\"}"),
		"github.com/AlonaBot/AlonaBot/sdk/llm.Response":     []byte("{\"$schema\":\"https://json-schema.org/draft/2020-12/schema\",\"properties\":{\"input_tokens\":{\"maximum\":9223372036854775807,\"minimum\":-9223372036854775808,\"type\":[\"integer\",\"null\"]},\"model\":{\"type\":\"string\"},\"output_tokens\":{\"maximum\":9223372036854775807,\"minimum\":-9223372036854775808,\"type\":[\"integer\",\"null\"]},\"text\":{\"type\":\"string\"}},\"required\":[\"text\",\"model\"],\"type\":\"object\"}"),
	},
	ParamNames: map[string][]string{
		"github.com/AlonaBot/AlonaBot/agent.(*component).addTaskTools":       {"catalog"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).appendImages":       {"db", "generation", "user", "assistant", "images"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).conversationMemory": {"db", "key"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).converse":           {"text"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).generate":           {"req"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).listTools":          {},
		"github.com/AlonaBot/AlonaBot/agent.(*component).resetConversation":  {},
		"github.com/AlonaBot/AlonaBot/agent.(*component).runScheduled":       {"session", "task"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).scheduled":          {"task"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).selectPersona":      {"name"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).selectModel":        {"name"},
		"github.com/AlonaBot/AlonaBot/agent.(*component).updateSettings":     {"update"},
		"github.com/AlonaBot/AlonaBot/agent.cancelTask":                      {"arguments"},
		"github.com/AlonaBot/AlonaBot/agent.currentTaskTarget":               {},
		"github.com/AlonaBot/AlonaBot/agent.listTasks":                       {},
		"github.com/AlonaBot/AlonaBot/agent.memoryError":                     {"err"},
		"github.com/AlonaBot/AlonaBot/agent.memoryScope":                     {"target"},
		"github.com/AlonaBot/AlonaBot/agent.readImages":                      {"text"},
		"github.com/AlonaBot/AlonaBot/agent.scheduleTask":                    {"arguments"},
		"github.com/AlonaBot/AlonaBot/agent.sessionKey":                      {},
		"github.com/AlonaBot/AlonaBot/agent:conversation.go:35":              {"chain"},
		"github.com/AlonaBot/AlonaBot/agent:conversation.go:37":              {},
		"github.com/AlonaBot/AlonaBot/agent:conversation.go:39":              {},
		"github.com/AlonaBot/AlonaBot/agent:conversation.go:48":              {},
		"github.com/AlonaBot/AlonaBot/agent:memory.go:79":                    {"in"},
		"github.com/AlonaBot/AlonaBot/agent:settings.go:61":                  {},
		"github.com/AlonaBot/AlonaBot/agent:settings.go:68":                  {},
		"github.com/AlonaBot/AlonaBot/agent:settings.go:77":                  {},
	},
}
