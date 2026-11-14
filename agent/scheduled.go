package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/agent/provider"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/gormdb"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

func (a *component) scheduled(c *sdk.Context, task AgentTask) error {
	key := c.Event().Payload.(*message.ScheduledPayload).TaskKey
	err := a.runScheduled(c, "task/"+key, task)
	if err != nil {
		if errcode.Of(err) == errcode.PermissionDenied {
			c.Logger().WarnContext(c.Context(), "scheduled agent task denied", "task_key", key, "error", err)
		} else {
			c.Logger().ErrorContext(c.Context(), "scheduled agent task failed", "task_key", key, "error", err)
		}
	}
	return err
}

func (a *component) runScheduled(c *sdk.Context, session string, task AgentTask) error {
	// 1. 读取目标聊天的模型设置，历史仅使用独立的任务会话；不读取记忆服务。
	db, err := gormdb.Open(c.Context())
	if err != nil {
		return err
	}
	settings, err := a.settings(c.Context(), string(task.Target.session())+"@"+task.Target.Instance, db)
	if err != nil {
		return err
	}
	selected := a.config.provider(settings.Provider)
	if selected == nil {
		return fmt.Errorf("scheduled task has no configured provider")
	}
	history, err := recentConversation(c.Context(), session, a.config.historyEntries)
	if err != nil {
		return err
	}
	request := provider.Request{MaxTokens: selected.maxTokens}
	if persona := a.config.persona(settings.Persona); persona != nil && persona.prompt != "" {
		request.Messages = append(request.Messages, provider.Message{Role: roleSystem, Content: persona.prompt})
	}
	if len(a.skills) > 0 {
		request.Messages = append(request.Messages, provider.Message{Role: roleSystem, Content: a.skillPrompt(c.T("agent.skills.header"))})
	}
	for _, entry := range history {
		request.Messages = append(request.Messages, provider.Message{Role: entry.Role, Content: entry.Content})
	}
	request.Messages = append(request.Messages, provider.Message{Role: roleUser, Content: task.Prompt})
	catalog, err := a.listTools(c)
	if err != nil {
		return err
	}
	// 2. 工具循环和发送继承 Core 的 job scope、期限与发信许可。
	var assistant strings.Builder
	if err := runToolLoop(c.Context(), selected, request, catalog, func(text string) error { assistant.WriteString(text); return nil }); err != nil {
		return err
	}
	if _, err := sdk.Send(c.Context(), task.Target.session(), message.Text(assistant.String()), sdk.Via(task.Target.Instance)); err != nil {
		return err
	}
	userJSON, err := json.Marshal(conversationEntry{Role: roleUser, Content: task.Prompt})
	if err != nil {
		return err
	}
	assistantJSON, err := json.Marshal(conversationEntry{Role: roleAssistant, Content: assistant.String()})
	if err != nil {
		return err
	}
	_, err = sdk.AppendConversation(c.Context(), session, userJSON, assistantJSON)
	return err
}
