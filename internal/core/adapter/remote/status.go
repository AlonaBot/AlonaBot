package remote

// InstanceStatus 返回实例的当前运行状态与最近失败说明，配置由装配处提供。
func (p *Process) InstanceStatus(instanceID string) (state, safeError string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry := p.instances[instanceID]; entry != nil && entry.running {
		return "running", ""
	}
	if message, failed := p.failures[instanceID]; failed {
		return "failed", message
	}
	return "stopped", ""
}

// RecordFailure 保存装配阶段发现的实例配置失败，原配置与启用选择保持不变。
func (p *Process) RecordFailure(instanceID, safeError string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failures[instanceID] = safeError
}

// ForgetFailure 在配置创建或删除后清除旧失败说明，避免实例名复用携带旧诊断。
func (p *Process) ForgetFailure(instanceID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.failures, instanceID)
}
