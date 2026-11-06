package wire

import "sync"

// Budget 限制 Core 的共享载荷预算，覆盖 Session 在途载荷与事件等待队列。
// Acquire/Release 接受 nil *Budget，表示不限额（插件侧没有全局预算）；Used 需要非 nil。
type Budget struct {
	// mu 保护 used。
	mu    sync.Mutex
	used  int64 // 当前已占用的字节数
	limit int64 // 上限字节数，创建后不变
}

// NewBudget 创建上限为 limit 字节的预算。
func NewBudget(limit int64) *Budget { return &Budget{limit: limit} }

// Acquire 尝试占用 n 字节，超过上限时不占用并返回 false；不等待。
func (b *Budget) Acquire(n int64) bool {
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

// Release 归还 n 字节；调用方保证每次占用只归还一次。
func (b *Budget) Release(n int64) {
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
