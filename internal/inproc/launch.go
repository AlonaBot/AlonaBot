// Package inproc 在根包与 SDK 之间传递同进程组件的显式启动参数。
//
// 参数仅描述一次经 loopback TCP 建立的 Session；本包只依赖标准库类型。
package inproc

import "log/slog"

// Launch 绑定一个组件本次启动的身份、认证信息与锁位置。
type Launch struct {
	PluginID        string
	Secret          []byte
	Addr            string
	BuildDigest     []byte
	RuntimeLockPath string       // <data_dir>/plugins/<id>/runtime.lock 的绝对路径。
	Log             slog.Handler // 已带组件来源属性、不过滤级别的日志目标。
}
