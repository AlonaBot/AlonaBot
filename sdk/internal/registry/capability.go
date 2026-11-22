package registry

import "context"

// RequireCapability 检查当前 Session 的协商结果，能力缺失时返回 Unsupported。
var RequireCapability func(context.Context, string) error
