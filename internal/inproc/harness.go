package inproc

import (
	"context"
	"sync"
)

// Outcome 保留一次 Handler 调用的插件、路由与安全错误。
type Outcome struct {
	PluginID, RouteID string
	ErrorCode         int32 // 0 表示 Handler 成功。
	SafeMessage       string
	Waiting           bool
}

// Dispatch 汇总一个来源事件的分发结果。
type Dispatch struct {
	InstanceID, SourceEventID string
	RunID                     string // 显式调度触发的运行 ID；事件注入时为空。
	Waiting                   bool   // 某个 Handler 已进入多轮等待。
	Outcomes                  []Outcome
}

// Harness 连接同一数据目录的 Core 与测试套件；字段在组件启动前装配，运行期间只读。
type Harness struct {
	AllowedInstances map[string][]string // 未列出的插件保持全部实例许可；空列表收紧为全部拒绝。
	Dispatched       func(Dispatch)      // 整条链及自动回复结束，或 Handler 进入等待后调用。
	Trigger          func(ctx context.Context, pluginID, jobID string, completed func(Dispatch)) error
	TriggerTask      func(ctx context.Context, pluginID, taskName, key string, completed func(Dispatch)) error
}

var harnesses sync.Map // 以绝对数据目录隔离并行运行的 Core。

// RegisterHarness 登记绝对数据目录对应的套件；同目录重复登记时 panic。
// 返回的注销函数可重复调用，注销后该目录可再次登记。
func RegisterHarness(dataDir string, h *Harness) (unregister func()) {
	if _, loaded := harnesses.LoadOrStore(dataDir, h); loaded {
		panic("harness already registered for data directory: " + dataDir)
	}
	// 旧注销函数只执行一次，避免误删同目录随后登记的套件。
	return sync.OnceFunc(func() { harnesses.Delete(dataDir) })
}

// HarnessFor 读取绝对数据目录对应的套件；未登记时返回 nil。
func HarnessFor(dataDir string) *Harness {
	h, ok := harnesses.Load(dataDir)
	if !ok {
		return nil
	}
	return h.(*Harness)
}
