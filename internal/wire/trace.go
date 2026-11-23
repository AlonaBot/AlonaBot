package wire

import (
	"context"
	"encoding/hex"
)

type traceContextKey struct{}

// WithTraceID 绑定可信入口生成的追踪标识，后续 Session.Call 沿用该标识。
func WithTraceID(ctx context.Context, trace []byte) context.Context {
	return context.WithValue(ctx, traceContextKey{}, append([]byte(nil), trace...))
}

// TraceID 返回当前可信 RPC 的十六进制追踪标识；后台调用没有业务 trace 时返回空串。
func TraceID(ctx context.Context) string {
	trace, _ := ctx.Value(traceContextKey{}).([]byte)
	return hex.EncodeToString(trace)
}
