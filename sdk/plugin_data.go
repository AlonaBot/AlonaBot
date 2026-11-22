package sdk

import "context"

// PluginDataDir 返回 ctx 所属插件 Session 的数据目录绝对路径。
// 传入 Handler 或生命周期回调的 ctx 及其派生 context；独立进程可使用默认 Session。
// Session 就绪前调用会 panic；目录中的文件跨 restart 保留，随卸载的数据清理选项删除。
func PluginDataDir(ctx context.Context) string {
	session := componentFor(ctx)
	if session == nil {
		panic("plugin session is not ready")
	}
	active := session.active.Load()
	if active == nil || !active.ready.Load() {
		panic("plugin session is not ready")
	}
	return active.ack.GetBootstrap().GetPluginDataDir()
}
