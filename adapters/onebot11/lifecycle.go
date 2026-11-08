package onebot11

import (
	"context"
	"errors"
	"log/slog"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/coder/websocket"
)

const (
	initialReconnectBackoff = time.Second
	maxReconnectBackoff     = 60 * time.Second
)

// startRead 为当前连接启动唯一读循环；调用者必须等待返回的 channel 后再建立下一条连接。
func (i *instance) startRead(conn *websocket.Conn) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		i.readConnection(conn)
	}()
	return done
}

func (i *instance) maintainConnection(readDone <-chan struct{}) {
	defer close(i.done)
	defer i.cancel()
	<-readDone // 初始连接由 Start 建立，先等其读取与在途调用收尾。
	backoff := initialReconnectBackoff
	for waitReconnect(i.ctx, backoff) {
		// 1. 只有这个循环创建新连接；Stop 取消同一个 ctx，覆盖退避和握手。
		conn, err := i.config.dial(i.ctx)
		if i.ctx.Err() != nil {
			if conn != nil {
				conn.CloseNow()
			}
			return
		}
		if err != nil {
			var public *sdk.Error
			if errors.As(err, &public) && public.Code == pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
				i.reportFailure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "onebot implementation rejected the access token")
				return
			}
			slog.Warn("onebot websocket reconnect failed", "adapter_instance_id", i.id, "error", i.config.safeError(err))
			backoff = nextReconnectBackoff(backoff)
			continue
		}

		// 2. 发布连接后立即恢复读取，让登录 API 也通过同一条 echo 通道配对。
		i.mu.Lock()
		if i.ctx.Err() != nil {
			i.mu.Unlock()
			conn.CloseNow()
			return
		}
		i.conn = conn
		i.mu.Unlock()
		readDone = i.startRead(conn)
		account, err := i.login(i.ctx)
		if err != nil || i.ctx.Err() != nil || account.ID != i.account.ID {
			conn.CloseNow()
			<-readDone
			if i.ctx.Err() != nil {
				return
			}
			if err != nil {
				slog.Warn("onebot reconnect login failed", "adapter_instance_id", i.id, "error", i.config.safeError(err))
				backoff = nextReconnectBackoff(backoff)
				continue
			}
			i.reportFailure(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "onebot account changed")
			return
		}

		// 3. 身份复核成功后恢复初始退避；不重放旧连接的任何请求。
		slog.Debug("reconnected onebot websocket", "adapter_instance_id", i.id, "url", i.config.safeURL)
		backoff = initialReconnectBackoff
		<-readDone
	}
}

func (i *instance) reportFailure(code pb.ErrorCode, message string) {
	if err := i.core.ReportFailure(i.ctx, &pb.ErrorInfo{Code: code, SafeMessage: message}); err != nil {
		slog.Error("report onebot failure", "adapter_instance_id", i.id, "error", i.config.safeError(err))
	}
}

func waitReconnect(ctx context.Context, backoff time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select { // Stop 必须能立即结束退避，无需等待计时器。
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil
	}
}

func nextReconnectBackoff(current time.Duration) time.Duration {
	return min(current*2, maxReconnectBackoff)
}
