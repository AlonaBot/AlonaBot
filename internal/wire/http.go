package wire

// HeaderProxyAuth 指定 Core 注入、SDK 校验的内部凭证头，值为本 Session 代理 token 的十六进制。
const HeaderProxyAuth = "X-Alonabot-Proxy-Auth"

// CapabilityHTTPListener 指定 HTTP Listener 协议能力名，版本为 1（WIRE-FORMAT §3.1）。
const CapabilityHTTPListener = "http_listener"
