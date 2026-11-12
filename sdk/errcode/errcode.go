// Package errcode 提供 SDK 的错误码、可检查的错误与错误链查询。
//
// Code 的数值与协议保持一致；插件通过 Of 检查直接或包装后的错误。
package errcode

import (
	"errors"
	"fmt"
	"strconv"
)

// Code 标识调用失败的原因，零值表示没有 SDK 错误。
type Code int32

const (
	InvalidArgument      Code = iota + 1 // 参数不合法。
	Unauthenticated                      // 未通过身份认证。
	PermissionDenied                     // 权限不足。
	NotFound                             // 目标不存在。
	AlreadyExists                        // 目标已存在。
	FailedPrecondition                   // 调用前置条件不满足。
	Unsupported                          // 操作不受支持。
	ResourceExhausted                    // 资源配额耗尽。
	DeadlineExceeded                     // 调用期限已到。
	Cancelled                            // 调用已取消。
	Unavailable                          // 服务暂不可用。
	Internal                             // 内部错误。
	ProtocolError                        // 协议错误。
	CallCycle                            // 调用链成环。
	PluginNotReady                       // 插件尚未就绪。
	NoDefaultTarget                      // 缺少默认发送目标。
	NeedsManualMigration                 // 需要人工迁移。
)

var names = [...]string{
	"ERROR_CODE_UNSPECIFIED", "ERROR_CODE_INVALID_ARGUMENT", "ERROR_CODE_UNAUTHENTICATED",
	"ERROR_CODE_PERMISSION_DENIED", "ERROR_CODE_NOT_FOUND", "ERROR_CODE_ALREADY_EXISTS",
	"ERROR_CODE_FAILED_PRECONDITION", "ERROR_CODE_UNSUPPORTED", "ERROR_CODE_RESOURCE_EXHAUSTED",
	"ERROR_CODE_DEADLINE_EXCEEDED", "ERROR_CODE_CANCELLED", "ERROR_CODE_UNAVAILABLE",
	"ERROR_CODE_INTERNAL", "ERROR_CODE_PROTOCOL_ERROR", "ERROR_CODE_CALL_CYCLE",
	"ERROR_CODE_PLUGIN_NOT_READY", "ERROR_CODE_NO_DEFAULT_TARGET", "ERROR_CODE_NEEDS_MANUAL_MIGRATION",
}

// String 返回协议枚举名；未知码返回十进制数值。
func (c Code) String() string {
	if c >= 0 && int(c) < len(names) {
		return names[c]
	}
	return strconv.FormatInt(int64(c), 10)
}

// Error 携带供调用者检查的错误码与安全说明。
type Error struct {
	Code    Code
	Message string
}

// Error 返回「错误码: 说明」。
func (e *Error) Error() string { return e.Code.String() + ": " + e.Message }

// New 创建带此错误码的错误。
func (c Code) New(message string) *Error { return &Error{Code: c, Message: message} }

// Errorf 按格式创建带此错误码的错误。
func (c Code) Errorf(format string, args ...any) *Error {
	return c.New(fmt.Sprintf(format, args...))
}

// Of 返回错误链中首个 *Error 的错误码，没有时返回零。
func Of(err error) Code {
	var public *Error
	if errors.As(err, &public) {
		return public.Code
	}
	return 0
}
