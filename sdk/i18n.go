package sdk

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

type translationSession struct {
	messages      *i18n.Catalog
	defaultLocale string // Bootstrap 的快照；下一次启动重新下发。
}

// Bootstrap 到达后一次发布只读快照，后台任务与 Handler 可以并发格式化。
var translations atomic.Pointer[translationSession]

func setI18n(messages *i18n.Catalog, defaultLocale string) {
	translations.Store(&translationSession{messages: messages, defaultLocale: defaultLocale})
}

func collectI18n(root Root, reg *Registration) error {
	catalog, problems := i18n.Parse(generatedFor(root).I18n)
	var errs []error
	for _, problem := range problems {
		errs = append(errs, fmt.Errorf("%s:%s: %s", problem.File, problem.Key, problem.Message))
	}
	if len(problems) != 0 {
		return errors.Join(errs...)
	}
	reg.messages = catalog
	for _, command := range reg.Commands {
		if command.spec.GetUsageMode() == pb.UsageMode_USAGE_MODE_CUSTOM && (catalog == nil || !catalog.Has(command.spec.GetUsageKey())) {
			errs = append(errs, fmt.Errorf("route %q uses unknown usage key %q", commandRouteID(command), command.spec.GetUsageKey()))
		}
	}
	return errors.Join(errs...)
}

// T 按本次调用 Actor 的语言、Core 默认语言、en-US 格式化文案。
// args 为交替的字符串名字与值，值经 fmt.Sprint 格式化；取消的调用返回 context 错误。
func (c *Context) T(key string, args ...any) (string, error) {
	actor := c.actor
	if c.event != nil {
		actor = c.event.GetActor()
	}
	return translate(c.ctx, actor.GetLocale(), key, args)
}

// T 从 Core 默认语言开始选择文案，供没有 Actor 的后台调用使用。
// args 为交替的字符串名字与值；格式化在本地完成，无需业务 RPC 就绪。
func T(ctx context.Context, key string, args ...any) (string, error) {
	return translate(ctx, "", key, args)
}

func translate(ctx context.Context, actorLocale, key string, args []any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	session := translations.Load()
	if session == nil {
		return "", &Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "translation session has not received Bootstrap"}
	}
	if session.messages == nil {
		return "", &Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "plugin has no message catalog"}
	}
	if len(args)%2 != 0 {
		return "", fmt.Errorf("translation arguments must contain name/value pairs")
	}
	values := make(map[string]string, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		name, ok := args[i].(string)
		if !ok {
			return "", fmt.Errorf("translation argument name at index %d must be a string", i)
		}
		values[name] = fmt.Sprint(args[i+1])
	}
	return session.messages.Format([]string{actorLocale, session.defaultLocale}, key, values)
}
