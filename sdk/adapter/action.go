package adapter

import (
	"context"
	"encoding/json"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
)

// ActionScope 指定动作需要会话还是只需要账号实例。
type ActionScope int

const (
	// ActionScopeUnspecified 保留未声明的零值，注册时拒绝。
	ActionScopeUnspecified ActionScope = iota
	// ActionScopeChat 要求调用方提供会话。
	ActionScopeChat
	// ActionScopeAccount 只接收账号实例，执行时会话为空。
	ActionScopeAccount
)

// ActionDeclaration 保存 Action 从参数与结果的 Go 类型推导的声明。
type ActionDeclaration = convert.ActionDeclaration

// Action 声明平台动作；P 须为结构体，类型推导错误在 REGISTER 前报告。
func Action[P, R any](name string, version uint32, scope ActionScope) ActionDeclaration {
	return convert.DeclareAction[P, R](name, version, int(scope))
}

// ActionExecutor 提供实例可选的平台动作入口；账号动作的 to 为空。
// 返回的 JSON 原文与安全错误可以同时存在。
type ActionExecutor interface {
	Action(ctx context.Context, to string, name string, version uint32, params json.RawMessage) (json.RawMessage, *errcode.Error)
}
