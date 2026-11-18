package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlonaBot/AlonaBot/agent/provider"
)

const providerRetryInterval = time.Second

type conversationStream func(provider.Request, func(provider.Event) error) error

// retryStream 为工具循环的每次请求重新尝试模型链；工具结果由循环持有。
func (s *configuration) retryStream(ctx context.Context, logger *slog.Logger, selected *configuredModel, prepare func(*configuredModel, provider.Request) provider.Request) conversationStream {
	models := []*configuredModel{selected}
	seen := map[string]bool{selected.name: true}
	for _, name := range s.fallbackChatModels {
		if !seen[name] {
			models = append(models, s.model(name))
			seen[name] = true
		}
	}
	return func(request provider.Request, emit func(provider.Event) error) error {
		var lastError error
		for _, model := range models {
			for attempt := int64(1); ; attempt++ {
				if lastError != nil {
					timer := time.NewTimer(providerRetryInterval)
					select {
					case <-ctx.Done():
						timer.Stop()
						return ctx.Err()
					case <-timer.C:
					}
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				current := request
				if prepare != nil {
					current = prepare(model, request)
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				current.MaxTokens = model.maxTokens
				delivered := false
				_, err := model.client.Stream(ctx, current, func(event provider.Event) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					delivered = delivered || event.Text != "" || event.ToolCall != nil
					return emit(event)
				})
				if err == nil {
					return nil
				}
				logger.WarnContext(ctx, "provider conversation request failed", "model", model.name, "attempt", attempt, "error", err)
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}
				if delivered || !provider.Retryable(err) {
					return err
				}
				lastError = err
				if attempt == s.maxAttempts {
					break
				}
			}
		}
		return lastError
	}
}
