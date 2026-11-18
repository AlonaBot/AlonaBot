package sdk

import (
	"context"
	"time"
)

var _ context.Context = (*Context)(nil)

// Deadline 返回本次调用的截止时刻。
func (c *Context) Deadline() (time.Time, bool) { return c.Context().Deadline() }

// Done 返回本次调用的取消通知。
func (c *Context) Done() <-chan struct{} { return c.Context().Done() }

// Err 返回本次调用的取消或超期错误。
func (c *Context) Err() error { return c.Context().Err() }

// Value 返回本次调用携带的值，保留 Session 与父调用关联。
func (c *Context) Value(key any) any { return c.Context().Value(key) }
