package sdk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

type translationSession struct {
	messages      *i18n.Catalog
	defaultLocale string   // Bootstrap 的快照；下一次启动重新下发。
	warned        sync.Map // [2]string{key, problem}；同一 Session 内并发去重。
}

// Bootstrap 到达后一次发布文案快照，后台任务与 Handler 可以并发格式化。
var translations atomic.Pointer[translationSession]

func setI18n(messages *i18n.Catalog, defaultLocale string) {
	translations.Store(&translationSession{messages: messages, defaultLocale: defaultLocale})
}

func collectI18n(root *Plugin, reg *Registration) error {
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
// args 为交替的字符串名字与值，值经 fmt.Sprint 格式化；问题在 Session 内去重后记 WARN。
func (c *Context) T(key string, args ...any) string {
	actor := c.actor
	if c.event != nil {
		actor = c.event.GetActor()
	}
	return translate(c.ctx, actor.GetLocale(), key, args)
}

// T 从 Core 默认语言开始选择文案，供没有 Actor 的后台调用使用。
// args 为交替的字符串名字与值；格式化在本地完成，无需业务 RPC 就绪。
func T(ctx context.Context, key string, args ...any) string {
	return translate(ctx, "", key, args)
}

func translate(ctx context.Context, actorLocale, key string, args []any) string {
	// 1. Bootstrap 前的调用也共用去重记录，收到 Bootstrap 后换成新 Session。
	session := translations.Load()
	if session == nil {
		translations.CompareAndSwap(nil, &translationSession{})
		session = translations.Load()
	}
	// 2. 只格式化成对参数；诊断中仅保留名字或位置。
	var problems []string
	if len(args)%2 != 0 {
		problems = append(problems, "translation arguments must contain name/value pairs")
	}
	values := make(map[string]string, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		name, ok := args[i].(string)
		if !ok {
			problems = append(problems, fmt.Sprintf("translation argument name at index %d must be a string", i))
			continue
		}
		values[name] = fmt.Sprint(args[i+1])
	}
	text := key
	if session.messages == nil {
		problems = append(problems, "plugin has no message catalog")
	} else {
		var catalogProblems []string
		text, catalogProblems = session.messages.Text([]string{actorLocale, session.defaultLocale}, key, values)
		problems = append(problems, catalogProblems...)
	}
	// 3. LoadOrStore 保证并发调用同一问题时只产生一条日志。
	for _, problem := range problems {
		if _, loaded := session.warned.LoadOrStore([2]string{key, problem}, struct{}{}); !loaded {
			slog.WarnContext(ctx, "translation problem", "key", key, "problem", problem)
		}
	}
	return text
}
