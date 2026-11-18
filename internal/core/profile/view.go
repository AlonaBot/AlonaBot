package profile

import (
	"context"
	"encoding/json"
	"sync"
)

// View 保留一次插件准入及消息过滤共用的配置文件选择与有效配置。
// Config 只存活到本次匹配结束，不写入路由快照。
type View struct {
	ProfileID string
	Enabled   bool
	Config    json.RawMessage
}

type viewKey struct{}
type viewResult struct {
	once sync.Once
	view View
	err  error
}

// WithView 为一次插件派发建立视图作用域，供准入和过滤复用同一快照。
func WithView(ctx context.Context) context.Context {
	return context.WithValue(ctx, viewKey{}, &viewResult{})
}

// ReadView 在作用域内只读取一次；独立读取在每次操作时建立新视图。
func ReadView(ctx context.Context, load func() (View, error)) (View, error) {
	result, ok := ctx.Value(viewKey{}).(*viewResult)
	if !ok {
		return load()
	}
	result.once.Do(func() { result.view, result.err = load() })
	return result.view, result.err
}
