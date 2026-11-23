package plugintest

// Background 按 Adapter 完成顺序消费尚未归属事件或任务的发送与编辑。
func (h *Harness) Background() []Sent {
	h.t.Helper()
	h.mu.Lock()
	defer h.mu.Unlock()
	var result []Sent
	remaining := h.sends[:0]
	for _, sent := range h.sends {
		if sent.eventID == "" {
			result = append(result, sent)
		} else {
			remaining = append(remaining, sent)
		}
	}
	h.sends = remaining
	return result
}
