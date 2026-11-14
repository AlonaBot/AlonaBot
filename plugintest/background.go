package plugintest

// Background 按 Adapter 接受顺序取走尚未交付给注入结果的发送与编辑。
// 与事件和任务注入一样，调用方按顺序读取同一套件。
func (h *Harness) Background() []Sent {
	h.t.Helper()
	if !h.busy.CompareAndSwap(false, true) {
		h.t.Fatal("concurrent event injection on one Harness")
	}
	defer h.busy.Store(false)
	h.mu.Lock()
	defer h.mu.Unlock()
	sends := h.sends
	h.sends = nil
	return sends
}
