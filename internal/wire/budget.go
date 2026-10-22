// 本文件负责跨 Session 共享的 Core 全局 payload 字节预算（CONTRACTS §1.2）。
package wire

import "sync"

// Budget 是跨 Session 共享的 payload 字节预算（Core 合计 64 MiB）。
type Budget struct {
	mu          sync.Mutex
	used, limit int64
}

// NewBudget 创建上限为 limit 字节的预算。
func NewBudget(limit int64) *Budget { return &Budget{limit: limit} }

func (b *Budget) acquire(n int64) bool {
	if b == nil {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.used+n > b.limit {
		return false
	}
	b.used += n
	return true
}

func (b *Budget) release(n int64) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.used -= n
	b.mu.Unlock()
}

// Used 返回当前占用的字节数。
func (b *Budget) Used() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used
}
