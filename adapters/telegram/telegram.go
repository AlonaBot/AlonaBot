// Package telegram 实现官方 Telegram Adapter 的账号检查与 long polling 接收。
//
// [New] 创建供 sdk.RunAdapter 使用的 Adapter。每个实例独占 HTTP 客户端和接收循环；
// Core 持久化逐条游标，实例仅在收到应答后推进内存状态（MESSAGE-MODEL §6.1）。
package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

// AdapterID 标识 Telegram 平台。
const AdapterID = "telegram"
const (
	pollTimeoutSeconds    = 30
	receiveRoundDuration  = 7 * 24 * time.Hour
	initialReceiveBackoff = time.Second
	maxReceiveBackoff     = 30 * time.Second
)

var allowedUpdates = []string{"message", "edited_message", "channel_post", "edited_channel_post", "callback_query", "my_chat_member", "chat_member", "message_reaction"}

type telegramAdapter struct {
	mu        sync.Mutex // 保护 starting/active 实例槽；网络请求期间不持锁。
	instances map[string]*instance
	now       func() time.Time
}

// New 创建独立的 Telegram Adapter；账号身份在实例 Start 时通过 getMe 验证。
func New() adapter.Adapter {
	return &telegramAdapter{instances: make(map[string]*instance), now: time.Now}
}

func receiveCapabilities() []*pb.Capability {
	return []*pb.Capability{{Name: "message.created", Version: 1, Supported: true}, {Name: "message.edited", Version: 1, Supported: true}, {Name: "member.joined", Version: 1, Supported: true}, {Name: "member.left", Version: 1, Supported: true}, {Name: "reaction.added", Version: 1, Supported: true}}
}

// Descriptor 声明接收能力，不声明发送或平台 extension。
func (*telegramAdapter) Descriptor() *pb.AdapterDescriptor {
	return &pb.AdapterDescriptor{AdapterId: AdapterID, Capabilities: receiveCapabilities()}
}

// Start 验证账号和 webhook 后启动该实例唯一的接收循环。
func (a *telegramAdapter) Start(ctx context.Context, request adapter.StartRequest, core adapter.Core) (adapter.Instance, error) {
	// 1. 在配置边界验证参数，错误只返回字段名。
	var config struct {
		Token      string `json:"token"`
		APIBaseURL string `json:"api_base_url"`
		Proxy      string `json:"proxy"`
	}
	if err := json.Unmarshal(request.ConfigJSON, &config); err != nil || config.Token == "" {
		return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid telegram configuration"}
	}
	if config.APIBaseURL == "" {
		config.APIBaseURL = defaultAPIBaseURL
	}
	client, err := newBotClient(config.Token, config.APIBaseURL, config.Proxy)
	if err != nil {
		return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: err.Error()}
	}
	saved := cursor{RoundStartUnixMS: a.now().UnixMilli()}
	if len(request.Cursor) > 0 {
		if err := json.Unmarshal(request.Cursor, &saved); err != nil {
			client.http.CloseIdleConnections()
			return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid telegram cursor"}
		}
	}
	i := &instance{client: client, core: core, id: request.InstanceID, now: a.now, done: make(chan struct{})}
	a.mu.Lock()
	_, exists := a.instances[i.id]
	if !exists {
		a.instances[i.id] = i
	}
	a.mu.Unlock()
	if exists {
		client.http.CloseIdleConnections()
		return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: "telegram instance already started"}
	}
	started := false
	defer func() {
		if !started {
			client.http.CloseIdleConnections()
			a.mu.Lock()
			delete(a.instances, i.id)
			a.mu.Unlock()
		}
	}()
	// 2. 验证账号与 webhook；启动失败时释放实例槽。
	var me struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	if err := client.call(ctx, "getMe", struct{}{}, &me); err != nil {
		return nil, startError(err)
	}
	if me.ID == 0 {
		return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, Message: "telegram getMe returned no account ID"}
	}
	i.account = adapter.Account{ID: strconv.FormatInt(me.ID, 10), DisplayName: me.Username}
	i.botID = me.ID
	var webhook struct {
		URL string `json:"url"`
	}
	if err := client.call(ctx, "getWebhookInfo", struct{}{}, &webhook); err != nil {
		return nil, startError(err)
	}
	if webhook.URL != "" {
		return nil, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "telegram webhook is already configured"}
	}
	// 3. Start 的 RPC 结束不取消实例；循环与 Stop 单独拥有其 context。
	receiveContext, cancel := context.WithCancel(context.WithoutCancel(ctx))
	i.cancel = cancel
	started = true
	go func() {
		defer cancel()
		defer close(i.done) // Stop 等待全部回调完成后才返回。
		defer client.http.CloseIdleConnections()
		defer func() { a.mu.Lock(); delete(a.instances, i.id); a.mu.Unlock() }()
		i.receive(receiveContext, saved)
	}()
	return i, nil
}
func startError(err error) error {
	code := pb.ErrorCode_ERROR_CODE_UNAVAILABLE
	var api *apiError
	if errors.As(err, &api) && api.status == http.StatusUnauthorized {
		code = pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED
	}
	return &sdk.Error{Code: code, Message: err.Error()}
}

type cursor struct {
	Offset             int64 `json:"offset"`
	RoundStartUnixMS   int64 `json:"round_start_unix_ms"`
	LastAcceptedUnixMS int64 `json:"last_accepted_unix_ms"`
}

// instance 的接收状态仅由 receive goroutine 修改；Stop 通过 cancel/done 交接生命周期。
type instance struct {
	client  *botClient
	core    adapter.Core
	id      string
	account adapter.Account
	botID   int64
	cancel  context.CancelFunc
	done    chan struct{}
	now     func() time.Time
}

// Account 返回 getMe 验证过的机器人身份。
func (i *instance) Account() adapter.Account { return i.account }

// Capabilities 返回本实例的接收能力。
func (*instance) Capabilities() []*pb.Capability { return receiveCapabilities() }

// Send 返回 UNSUPPORTED，且不会发起平台请求。
func (*instance) Send(context.Context, *pb.Target, *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "telegram sending is unsupported"}
}

// Stop 取消在途请求，等待全部接收回调退出后返回。
func (i *instance) Stop(_ context.Context) error {
	i.cancel()
	<-i.done // 接收循环拥有全部交付，退出后再确认停止。
	return nil
}

func (i *instance) receive(ctx context.Context, saved cursor) {
	backoff := initialReceiveBackoff
	for ctx.Err() == nil {
		var batch []json.RawMessage
		err := i.client.call(ctx, "getUpdates", struct {
			Offset  int64    `json:"offset,omitempty"`
			Timeout int      `json:"timeout"`
			Allowed []string `json:"allowed_updates"`
		}{saved.Offset, pollTimeoutSeconds, allowedUpdates}, &batch)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			var api *apiError
			if errors.As(err, &api) && (api.status == http.StatusUnauthorized || api.status == http.StatusConflict) {
				message := "telegram rejected the token"
				if api.status == http.StatusConflict {
					message = "another getUpdates or webhook owns this bot"
				}
				code := pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED
				if api.status == http.StatusConflict {
					code = pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION
				}
				if reportErr := i.core.ReportFailure(ctx, &pb.ErrorInfo{Code: code, SafeMessage: message}); reportErr != nil {
					slog.Error("report telegram failure", "error", i.client.safeError(reportErr))
				}
				return
			}
		} else {
			// 每条交付独立等待 Core 确认，批内任何失败都从上次确认的 offset 重拉。
			for _, raw := range batch {
				if ctx.Err() != nil {
					return
				}
				var update struct {
					ID *int64 `json:"update_id"`
				}
				if parseErr := json.Unmarshal(raw, &update); parseErr != nil {
					err = parseErr
					break
				}
				if update.ID == nil {
					err = errors.New("telegram update_id is missing")
					break
				}
				var events []*pb.Event
				events, err = i.normalize(raw)
				if err != nil {
					break
				}
				next := saved
				next.Offset = *update.ID + 1
				if len(events) > 0 {
					next.LastAcceptedUnixMS = i.now().UnixMilli()
					for n, event := range events {
						if ctx.Err() != nil {
							return
						}
						// 一个 update 的全部事件完成前，Core 仍保存旧游标以便重拉。
						deliveryCursor := saved
						if n == len(events)-1 {
							deliveryCursor = next
						}
						err = i.save(ctx, event, deliveryCursor)
						if err != nil {
							break
						}
					}
				} else {
					err = i.save(ctx, nil, next)
				}
				if err != nil {
					break
				}
				saved = next
			}
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			var public *sdk.Error
			if errors.As(err, &public) && public.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
				return
			}
			slog.Warn("telegram receive failed", "adapter_instance_id", i.id, "error", i.client.safeError(err))
			delay := backoff
			var api *apiError
			if errors.As(err, &api) && api.status == http.StatusTooManyRequests && api.retryAfter > 0 {
				delay = api.retryAfter
			}
			if !waitReceive(ctx, delay) {
				return
			}
			backoff = nextReceiveBackoff(backoff)
			continue
		}
		backoff = initialReceiveBackoff
		// 当前 long poll 已结束；先持久化轮次，再发起新的请求。
		now := i.now()
		if now.UnixMilli()-max(saved.RoundStartUnixMS, saved.LastAcceptedUnixMS) >= receiveRoundDuration.Milliseconds() {
			next := cursor{RoundStartUnixMS: now.UnixMilli()}
			if err := i.save(ctx, nil, next); err != nil {
				var public *sdk.Error
				if errors.As(err, &public) && public.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
					return
				}
				slog.Warn("save telegram receive round failed", "error", i.client.safeError(err))
				if !waitReceive(ctx, backoff) {
					return
				}
				backoff = nextReceiveBackoff(backoff)
				continue
			}
			saved = next
		}
	}
}
func (i *instance) save(ctx context.Context, event *pb.Event, next cursor) error {
	encoded, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if event == nil {
		return i.core.SaveCursor(ctx, encoded)
	}
	_, err = i.core.Deliver(ctx, event, encoded)
	return err
}
func waitReceive(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func nextReceiveBackoff(current time.Duration) time.Duration {
	return min(current*2, maxReceiveBackoff)
}
