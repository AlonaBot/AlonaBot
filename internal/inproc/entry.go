package inproc

// Plugin 由 SDK 初始化时登记，使用显式启动参数运行插件。
var Plugin func(any, *Launch) error

// Adapter 由 SDK 初始化时登记，使用显式启动参数运行 Adapter。
var Adapter func(any, *Launch) error
