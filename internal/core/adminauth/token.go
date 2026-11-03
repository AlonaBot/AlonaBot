// Package adminauth 管理 Core 的管理凭证与认证。
//
// InitToken 和 ResetToken 通过已持有 core.lock 的 state.Store 离线更新凭证；
// 返回的随机 token 由命令层只向本机终端显示一次，状态库只保存摘要和认证代次。
// [Auth] 提供登录、会话校验与 CSRF 中间件，并实现 webhost 的管理会话检查。
// 规则见 CONTRACTS §7.3、§7.4。
package adminauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// tokenBytes 采用 CONTRACTS §7.4 规定的 32 字节随机凭证长度。
const tokenBytes = 32

// ErrNotInitialized 表示管理凭证尚未初始化。
var ErrNotInitialized = errors.New("admin token is not initialized")

// InitToken 生成无 padding 的 base64url 管理 token，并在事务内保存摘要。
// 已有凭证时失败；调用方负责只向本机终端显示一次返回值。
func InitToken(ctx context.Context, st *state.Store) (token string, err error) {
	token = randomToken()
	digest := sha256.Sum256([]byte(token))
	if err := st.InitAdminToken(ctx, digest[:], time.Now().UnixMilli()); err != nil {
		return "", err
	}
	return token, nil
}

// ResetToken 替换管理 token、递增认证代次，并撤销全部管理会话。
// 尚未初始化时返回 ErrNotInitialized；调用方负责只向本机终端显示一次返回值。
func ResetToken(ctx context.Context, st *state.Store) (token string, err error) {
	token = randomToken()
	digest := sha256.Sum256([]byte(token))
	if err := st.ResetAdminToken(ctx, digest[:], time.Now().UnixMilli()); err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return "", ErrNotInitialized
		}
		return "", err
	}
	return token, nil
}

func randomToken() string {
	var raw [tokenBytes]byte
	// crypto/rand.Read 保证填满缓冲区；熵源不可用时由运行时终止进程。
	rand.Read(raw[:])
	return base64.RawURLEncoding.EncodeToString(raw[:])
}
