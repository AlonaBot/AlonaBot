// Package adminauth 管理 Core 的管理凭证与认证。
//
// FirstToken 在首次启动时向终端或 admin_token 文件交付管理凭证。
// InitToken 和 ResetToken 通过已持有 core.lock 的 state.Store 离线更新凭证；
// 返回的随机 token 由命令层只向本机终端显示一次，状态库只保存摘要和认证代次。
// [Auth] 提供登录、会话校验、CSRF 中间件与插件路由授权。
// 规则见 CONTRACTS §7.3、§7.4。
package adminauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// tokenBytes 采用 CONTRACTS §7.4 规定的 32 字节随机凭证长度。
const tokenBytes = 32

// FirstTokenFile 指定首次非终端启动的凭证文件（CONTRACTS §7.4）。
const FirstTokenFile = "admin_token"

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

// FirstToken 在尚无凭证时初始化管理 token；调用方应已取得 data_dir 的 core.lock。
// interactive 为 true 时返回 token，由调用方只向终端显示一次；否则写入
// dataDir 下权限为 0600 的 admin_token 并返回空串。已有凭证时返回空串和 nil。
func FirstToken(ctx context.Context, st *state.Store, dataDir string, interactive bool) (token string, err error) {
	// 1. 已初始化的凭证与尚未领取的文件保持不变。
	if _, err := st.AdminCredential(ctx); err == nil {
		return "", nil
	} else if !errors.Is(err, state.ErrNotFound) {
		return "", err
	}
	if interactive {
		return InitToken(ctx, st)
	}

	// 2. 先完成文件交付再提交摘要，文件写入失败时仍可在下次启动初始化。
	token = randomToken()
	name := filepath.Join(dataDir, FirstTokenFile)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("create first admin token file: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(name))
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return "", fmt.Errorf("protect first admin token file: %w", errors.Join(err, file.Close()))
	}
	_, writeErr := file.WriteString(token + "\n")
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return "", fmt.Errorf("write first admin token file: %w", err)
	}

	// 3. 数据库拒绝初始化时删除本次文件，保留原始错误与清理错误。
	digest := sha256.Sum256([]byte(token))
	if err := st.InitAdminToken(ctx, digest[:], time.Now().UnixMilli()); err != nil {
		return "", err
	}
	return "", nil
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
