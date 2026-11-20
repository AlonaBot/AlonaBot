// Package metrics 按插件、业务种类和原因累计 Core 的内存计数。
//
// Counters 的零值可用；Add 与 Snapshot 可并发调用，快照独立于后续写入。
// 计数随 Core 生命周期保留，遵循 CONTRACTS §3。
package metrics

import "sync"

// Reason 标识 denied、cancelled、offline、dropped_queue、dropped_budget、cycle 或 depth。
// 唤醒使用 started、ready、failed、cold_start、dropped_queue、dropped_bytes、
// dropped_budget、cancelled、deadline、dropped_failed 和 dropped_other。
type Reason string

// Kind 标识 service、bus、event 或 wake。
type Kind string

// Counters 保存进程内累计值；使用后应通过指针共享。
type Counters struct {
	mu     sync.Mutex // 保护整个计数树，使快照与更新互斥。
	counts map[string]map[Kind]map[Reason]uint64
}

// Add 累加一次原因计数；nil 接收者关闭计数。
func (c *Counters) Add(pluginID string, kind Kind, reason Reason) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[string]map[Kind]map[Reason]uint64)
	}
	if c.counts[pluginID] == nil {
		c.counts[pluginID] = make(map[Kind]map[Reason]uint64)
	}
	if c.counts[pluginID][kind] == nil {
		c.counts[pluginID][kind] = make(map[Reason]uint64)
	}
	c.counts[pluginID][kind][reason]++
}

// Snapshot 返回 plugin → kind → reason → count 的完整副本。
func (c *Counters) Snapshot() map[string]map[Kind]map[Reason]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make(map[string]map[Kind]map[Reason]uint64, len(c.counts))
	for pluginID, kinds := range c.counts {
		result[pluginID] = make(map[Kind]map[Reason]uint64, len(kinds))
		for kind, reasons := range kinds {
			copy := make(map[Reason]uint64, len(reasons))
			for reason, count := range reasons {
				copy[reason] = count
			}
			result[pluginID][kind] = copy
		}
	}
	return result
}
