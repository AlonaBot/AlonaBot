// Package tool 列出和调用 Core 工具，并在有效 Handler 内请求用户确认。
package tool

// Default 指定工具未被管理策略覆盖时的调用方式。
type Default int

const (
	// DefaultAllow 允许工具调用直接执行。
	DefaultAllow Default = iota
	// DefaultConfirm 要求工具调用先取得用户确认。
	DefaultConfirm
)
