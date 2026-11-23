package plugintest

import (
	"context"

	"github.com/AlonaBot/AlonaBot/internal/inproc"
	"github.com/AlonaBot/AlonaBot/sdk"
)

const scheduledSourcePrefix = "@job:"

// Trigger 立即运行已激活插件的静态任务，返回本次 RPC 完成通知前的发送与结果。
func (h *Harness) Trigger(p *sdk.Plugin, jobID string) Result {
	h.t.Helper()
	return h.trigger(p, func(ctx context.Context, pluginID string, completed func(inproc.Dispatch)) error {
		return h.bridge.Trigger(ctx, pluginID, jobID, completed)
	})
}

// TriggerTask 立即触发已创建的动态实例，参数与运行准入由 Core 持有。
func (h *Harness) TriggerTask(p *sdk.Plugin, taskName, key string) Result {
	h.t.Helper()
	return h.trigger(p, func(ctx context.Context, pluginID string, completed func(inproc.Dispatch)) error {
		return h.bridge.TriggerTask(ctx, pluginID, taskName, key, completed)
	})
}

func (h *Harness) trigger(p *sdk.Plugin, run func(context.Context, string, func(inproc.Dispatch)) error) Result {
	h.t.Helper()
	if !h.busy.CompareAndSwap(false, true) {
		h.t.Fatal("concurrent event injection on one Harness")
	}
	defer h.busy.Store(false)
	pluginID, ok := h.plugins[p]
	if !ok {
		h.t.Fatal("trigger requires a plugin registered with this Harness")
	}
	var notification dispatchResult
	completed := func(d inproc.Dispatch) {
		h.mu.Lock()
		defer h.mu.Unlock()
		notification = dispatchResult{dispatch: d}
	}
	if err := run(h.ctx, pluginID, completed); err != nil {
		h.t.Fatalf("trigger scheduled handler: %v", err)
	}
	return h.result(notification)
}
