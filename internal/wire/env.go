package wire

// Core 启动插件进程时传递的环境变量（CONTRACTS §1.1、§12.2）。子进程只收到这些变量与少量
// 平台必需变量；Core 管理凭证与其他插件的凭证从不进入插件环境。
const (
	EnvPluginID     = "ALONABOT_PLUGIN_ID"
	EnvSecret       = "ALONABOT_SESSION_SECRET" // 32 字节，十六进制编码
	EnvTransport    = "ALONABOT_TRANSPORT"      // TransportTCP 或 TransportStdio
	EnvCoreAddr     = "ALONABOT_CORE_ADDR"      // TCP：Core 持有的 loopback 地址
	EnvBuildDigest  = "ALONABOT_BUILD_DIGEST"   // 安装时已验证的归档 SHA256，十六进制；SDK 原样回显
	EnvRuntimeLock  = "ALONABOT_RUNTIME_LOCK"   // <plugins>/<id>/runtime.lock 的绝对路径
	TransportTCP    = "tcp"
	TransportStdio  = "stdio"
	RuntimeLockFile = "runtime.lock"
)
