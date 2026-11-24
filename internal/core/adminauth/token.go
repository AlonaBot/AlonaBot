// Package adminauth 管理 Core 的凭据、首次领取与管理会话。
package adminauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const tokenBytes = 32

// FirstTokenFile 指定本机 CLI 共用的持久凭据文件。
const FirstTokenFile = "admin_token"

// ErrNotInitialized 表示管理凭据尚未初始化。
var ErrNotInitialized = errors.New("admin token is not initialized")

// InitToken 初始化并持久保存管理 token；已有凭据时保留原值。
func InitToken(ctx context.Context, st *state.Store, dataDir string) (string, error) {
	return persistToken(dataDir, func(token string) error {
		digest := sha256.Sum256([]byte(token))
		return st.InitAdminToken(ctx, digest[:], time.Now().UnixMilli())
	})
}

// FirstToken 在首次启动时保存凭据；已有摘要而文件缺失时保留摘要并提示显式 reset。
func FirstToken(ctx context.Context, st *state.Store, dataDir string) error {
	if _, err := st.AdminCredential(ctx); err == nil {
		if _, err := os.Stat(filepath.Join(dataDir, FirstTokenFile)); errors.Is(err, os.ErrNotExist) {
			slog.Warn(fmt.Sprintf("The admin token file is missing in %s; run admin-token reset explicitly", dataDir), "data_dir", dataDir)
		} else if err != nil {
			return err
		}
		return nil
	} else if !errors.Is(err, state.ErrNotFound) {
		return err
	}
	_, err := InitToken(ctx, st, dataDir)
	return err
}

// ResetToken 保存新凭据、递增代次并撤销旧会话；失败时恢复旧文件。
func ResetToken(ctx context.Context, st *state.Store, dataDir string) (string, error) {
	return persistToken(dataDir, func(token string) error {
		digest := sha256.Sum256([]byte(token))
		err := st.ResetAdminToken(ctx, digest[:], time.Now().UnixMilli())
		if errors.Is(err, state.ErrNotFound) {
			return ErrNotInitialized
		}
		return err
	})
}

// persistToken 先准备可恢复的文件，再提交数据库；失败时原子恢复旧文件。
// 调用方持有实例锁，在线更新还持有 credentials 锁。
func persistToken(dataDir string, save func(string) error) (result string, err error) {
	token := randomToken()
	name := filepath.Join(dataDir, FirstTokenFile)
	if info, err := os.Lstat(name); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("admin token path must be a regular file: %s", name)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect admin token file: %w", err)
	}
	file, err := os.CreateTemp(dataDir, ".admin-token-*")
	if err != nil {
		return "", fmt.Errorf("prepare admin token: %w", err)
	}
	temp := file.Name()
	defer func() { err = errors.Join(err, removeTemporary(temp)) }()
	_, writeErr := file.WriteString(token + "\n")
	syncErr := file.Sync()
	if err := errors.Join(writeErr, syncErr, file.Close()); err != nil {
		return "", fmt.Errorf("write admin token: %w", err)
	}
	// 保留旧 inode，以便数据库拒绝更新时恢复，文件缺失的旧实例也可显式 reset。
	backup := temp + ".previous"
	hadPrevious := true
	if err := os.Rename(name, backup); errors.Is(err, os.ErrNotExist) {
		hadPrevious = false
	} else if err != nil {
		return "", fmt.Errorf("preserve admin token: %w", err)
	}
	restore := func() error {
		if hadPrevious {
			return os.Rename(backup, name)
		}
		err := os.Remove(name)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.Rename(temp, name); err != nil {
		return "", errors.Join(err, restore())
	}
	if err := save(token); err != nil {
		return "", errors.Join(err, restore())
	}
	if hadPrevious {
		if err := os.Remove(backup); err != nil {
			return token, fmt.Errorf("remove previous admin token: %w", err)
		}
	}
	return token, nil
}

func randomToken() string {
	var raw [tokenBytes]byte
	// crypto/rand.Read 保证填满缓冲区；熵源不可用时由运行时终止进程。
	rand.Read(raw[:])
	return base64.RawURLEncoding.EncodeToString(raw[:])
}
