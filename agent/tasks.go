package agent

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/agent/provider"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const minimumTaskInterval = 60 * time.Second

var agentTask = sdk.Task[AgentTask]("agent")

type taskCreator struct {
	Instance string       `json:"instance"`
	User     message.User `json:"user"`
}

type taskTarget struct {
	Instance string          `json:"instance"`
	Chat     message.Session `json:"chat"` // 不含话题的会话地址。
	Thread   string          `json:"thread"`
}

// AgentTask 保存创建者、显式发送目标和后台一轮的提示词。
type AgentTask struct {
	Creator   taskCreator `json:"creator"`
	Target    taskTarget  `json:"target"`
	Prompt    string      `json:"prompt"`
	CreatedAt string      `json:"created_at"` // RFC 3339 纳秒时间。
}

func currentTaskTarget(c *sdk.Context) taskTarget {
	adapter, chat, thread, _ := c.Session().Parts()
	return taskTarget{Instance: c.Instance(), Chat: message.NewSession(adapter, chat, ""), Thread: thread}
}

func (t taskTarget) session() message.Session {
	adapter, chat, _, _ := t.Chat.Parts()
	return message.NewSession(adapter, chat, t.Thread)
}

func (a *component) addTaskTools(c *sdk.Context, catalog *toolCatalog) {
	for _, tool := range []roundTool{
		{local: func(args json.RawMessage) string { return scheduleTask(c, args) }, definition: provider.Tool{Name: "schedule_task", Description: "Schedule an agent prompt for this chat (Core Admin only)", InputSchema: json.RawMessage(`{"type":"object","properties":{"prompt":{"type":"string"},"at":{"type":"string"},"interval_seconds":{"type":"integer","minimum":60},"cron":{"type":"string"},"timezone":{"type":"string"}},"required":["prompt"],"additionalProperties":false}`)}},
		{local: func(args json.RawMessage) string { return listTasks(c) }, definition: provider.Tool{Name: "list_tasks", Description: "List scheduled agent tasks in this chat", InputSchema: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}},
		{local: func(args json.RawMessage) string { return cancelTask(c, args) }, definition: provider.Tool{Name: "cancel_task", Description: "Cancel a scheduled agent task in this chat (Core Admin only)", InputSchema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`)}},
	} {
		catalog.byName[tool.definition.Name] = tool
		catalog.definitions = append(catalog.definitions, tool.definition)
	}
}

func scheduleTask(c *sdk.Context, arguments json.RawMessage) string {
	if !c.Has(sdk.Admin) {
		return "error: PERMISSION_DENIED"
	}
	// 1. 在调用 Core 前验证输入组合与 Agent 的间隔下限。
	var input struct {
		Prompt   string  `json:"prompt"`
		At       *string `json:"at"`
		Interval *int64  `json:"interval_seconds"`
		Cron     *string `json:"cron"`
		Timezone *string `json:"timezone"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return toolResult("", errcode.InvalidArgument.New(err.Error()))
	}
	triggers := 0
	for _, present := range []bool{input.At != nil, input.Interval != nil, input.Cron != nil} {
		if present {
			triggers++
		}
	}
	if strings.TrimSpace(input.Prompt) == "" || triggers != 1 || input.Timezone != nil && input.Cron == nil {
		return toolResult("", errcode.InvalidArgument.New("prompt and exactly one trigger are required; timezone is only valid with cron"))
	}
	var at time.Time
	var interval time.Duration
	var err error
	if input.At != nil {
		at, err = time.Parse(time.RFC3339, *input.At)
	}
	if input.Interval != nil {
		interval, err = time.ParseDuration(strconv.FormatInt(*input.Interval, 10) + "s")
		if err == nil && interval < minimumTaskInterval {
			err = errcode.InvalidArgument.New("interval_seconds must be at least 60")
		}
	}
	if err != nil {
		return toolResult("", errcode.InvalidArgument.New(err.Error()))
	}
	// 2. 确认实际参数后，持久化与触发规则校验交给 Core。
	if err := sdk.ConfirmOperation(c.Context(), "agent/schedule_task", string(arguments)); err != nil {
		return toolResult("", err)
	}
	args := AgentTask{Creator: taskCreator{Instance: c.Instance(), User: c.User()}, Target: currentTaskTarget(c), Prompt: input.Prompt, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	var key string
	switch {
	case input.At != nil:
		key, err = agentTask.At(c.Context(), at, args)
	case input.Interval != nil:
		key, err = agentTask.Every(c.Context(), interval, args)
	case input.Cron != nil:
		var options []sdk.TaskOption
		if input.Timezone != nil {
			options = append(options, sdk.Timezone(*input.Timezone))
		}
		key, err = agentTask.Cron(c.Context(), *input.Cron, args, options...)
	}
	if err != nil {
		return toolResult("", err)
	}
	return taskJSON(struct {
		ID string `json:"id"`
	}{key})
}

func listTasks(c *sdk.Context) string {
	instances, err := agentTask.List(c.Context())
	if err != nil {
		return toolResult("", err)
	}
	type listedTask struct {
		ID       string     `json:"id"`
		Prompt   string     `json:"prompt"`
		At       *time.Time `json:"at,omitempty"`
		Interval int64      `json:"interval_seconds,omitempty"`
		Cron     string     `json:"cron,omitempty"`
		Timezone string     `json:"timezone,omitempty"`
		NextAt   *time.Time `json:"next_at"`
	}
	tasks := []listedTask{}
	target := currentTaskTarget(c)
	for _, instance := range instances {
		if instance.Args.Target != target {
			continue
		}
		task := listedTask{ID: instance.Key, Prompt: instance.Args.Prompt, Interval: int64(instance.Interval / time.Second), Cron: instance.Cron, Timezone: instance.Timezone, NextAt: instance.NextAt}
		if !instance.At.IsZero() {
			task.At = &instance.At
		}
		tasks = append(tasks, task)
	}
	return taskJSON(tasks)
}

func cancelTask(c *sdk.Context, arguments json.RawMessage) string {
	if !c.Has(sdk.Admin) {
		return "error: PERMISSION_DENIED"
	}
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return toolResult("", errcode.InvalidArgument.New(err.Error()))
	}
	if input.ID == "" {
		return toolResult("", errcode.InvalidArgument.New("id is required"))
	}
	if err := sdk.ConfirmOperation(c.Context(), "agent/cancel_task", string(arguments)); err != nil {
		return toolResult("", err)
	}
	instances, err := agentTask.List(c.Context())
	if err != nil {
		return toolResult("", err)
	}
	for _, instance := range instances {
		if instance.Key == input.ID && instance.Args.Target == currentTaskTarget(c) {
			err := agentTask.Cancel(c.Context(), input.ID)
			if errcode.Of(err) == errcode.NotFound {
				return "error: NOT_FOUND"
			}
			if err != nil {
				return toolResult("", err)
			}
			return "{}"
		}
	}
	return "error: NOT_FOUND"
}

func taskJSON(value any) string {
	encoded, err := json.Marshal(value)
	return toolResult(string(encoded), err)
}
