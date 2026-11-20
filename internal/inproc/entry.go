package inproc

// Plugin 由 SDK 初始化时登记，使用显式启动参数运行插件。
var Plugin func(any, *Launch) error

// Adapter 由 SDK 初始化时登记，使用显式启动参数运行 Adapter。
// 通过接口调用，使仅导入 SDK 的插件可裁剪同进程 Adapter 实现。
var Adapter interface {
	// Run 使用本次显式启动参数运行 Adapter，返回其完整运行结果。
	Run(any, *Launch) error
}
