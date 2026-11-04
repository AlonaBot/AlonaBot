// Package candidate 准备独立的受管 SQLite 候选资源。
//
// 调用方持有插件操作锁，并在复制前确认旧进程树退出；PrepareDatabase 暂持
// runtime.lock，返回前释放，随后由 SDK 取锁打开候选。规则见 CONTRACTS §6.1、§6.3、§12.2。
//
// 包的入口为 [PrepareDatabase] 与 [Discard]，分别返回 [Prepared] 和清理未受保护的候选。
// 管理层按以下顺序使用本包：
//
//  1. 持插件操作锁，冻结并停止旧进程，确认进程树退出。
//  2. 调用 PrepareDatabase。
//  3. 调用 state.Store.AddGeneration 保存未提交代次。
//  4. 调用 supervisor.Launch：NewCandidate=true，ExpectedOld=FromGeneration，
//     Bootstrap.database_descriptor.activation_reason=Reason，sqlite_path=Prepared.DatabasePath。
//  5. PREPARED 后由管理层的唯一提交点在同一事务中提交 Active 与 COMMITTED。
//  6. 提交前失败时关闭候选并确认进程树退出，再调用 Discard，随后用旧代次启动新 Session
//     （NewCandidate=false）。
//
// Active 与 COMMITTED 的写入及恢复编排由管理层负责；本包只准备和清理资源。
package candidate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"modernc.org/sqlite"
)

const (
	generationsDirectory = "generations"
	databaseFile         = "db.sqlite"
	temporaryPattern     = ".db-*.sqlite"
	integrityOK          = "ok"
)

// Request 指定插件身份、启动原因及候选的源代次和目标代次。
type Request struct {
	PluginID       string
	PluginDir      string              // <plugins>/<id> 的绝对路径，目录由调用方准备
	Reason         pb.ActivationReason // 支持 FIRST_INSTALL、RELOAD、UPDATE、RESUME
	FromGeneration string              // Active 代次；FIRST_INSTALL 时为空
	NewGeneration  string              // RESUME 时等于 FromGeneration
}

// Prepared 交接候选身份与受管数据库路径，供 Bootstrap 及失败清理使用。
type Prepared struct {
	PluginID     string
	Generation   string
	DatabasePath string // 绝对路径，写入 Bootstrap 的 sqlite_path
	Created      bool   // 本次是否创建了数据库文件；首装文件由 SDK 创建
}

// PrepareDatabase 在 runtime.lock 内准备候选；返回前释放锁，失败清理本次新建的目录。
// FIRST_INSTALL 只建目录，RELOAD/UPDATE 复制 Active，RESUME 复用 Active 路径。
func PrepareDatabase(ctx context.Context, req Request) (prepared *Prepared, err error) {
	switch req.Reason {
	case pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL,
		pb.ActivationReason_ACTIVATION_REASON_RELOAD,
		pb.ActivationReason_ACTIVATION_REASON_UPDATE,
		pb.ActivationReason_ACTIVATION_REASON_RESUME:
	default:
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "unsupported activation reason"}
	}
	if !filepath.IsAbs(req.PluginDir) || !generationName(req.NewGeneration) {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "absolute plugin directory and generation name are required"}
	}
	if req.Reason == pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL {
		if req.FromGeneration != "" {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "first install must not have a source generation"}
		}
	} else if !generationName(req.FromGeneration) || (req.Reason == pb.ActivationReason_ACTIVATION_REASON_RESUME && req.NewGeneration != req.FromGeneration) {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid source generation or resume generation mismatch"}
	}
	// 1. 先取跨代次进程锁，确保目录和数据库操作不会与旧 SDK 重叠。
	lock, err := oslock.TryLock(filepath.Join(req.PluginDir, wire.RuntimeLockFile))
	if errors.Is(err, oslock.ErrLocked) {
		return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "plugin process still holds runtime.lock"}
	}
	if err != nil {
		return nil, err
	}
	ownedDirectory := ""
	defer func() {
		if releaseErr := lock.Unlock(); releaseErr != nil {
			err = errors.Join(err, releaseErr)
		}
		if err != nil {
			prepared = nil
			if ownedDirectory != "" {
				if cleanupErr := os.RemoveAll(ownedDirectory); cleanupErr != nil {
					err = errors.Join(err, cleanupErr)
				}
			}
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	directory := filepath.Join(req.PluginDir, generationsDirectory, req.NewGeneration)
	prepared = &Prepared{PluginID: req.PluginID, Generation: req.NewGeneration, DatabasePath: filepath.Join(directory, databaseFile)}
	if req.Reason == pb.ActivationReason_ACTIVATION_REASON_RESUME {
		return prepared, nil
	}
	// 2. 独占创建目标目录，既有候选保持原样。
	if err := os.MkdirAll(filepath.Dir(directory), 0o700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "candidate directory already exists"}
		}
		return nil, err
	}
	ownedDirectory = directory
	if req.Reason == pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL {
		return prepared, nil
	}
	// 3. 先备份并校验临时库，完整落盘后才发布 db.sqlite。
	temporary, err := os.CreateTemp(directory, temporaryPattern)
	if err != nil {
		return nil, err
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return nil, err
	}
	source := filepath.Join(req.PluginDir, generationsDirectory, req.FromGeneration, databaseFile)
	if err := backupDatabase(ctx, source, temporaryPath); err != nil {
		return nil, err
	}
	if err := checkIntegrity(ctx, temporaryPath); err != nil {
		return nil, err
	}
	if err := syncPath(temporaryPath); err != nil {
		return nil, err
	}
	if err := os.Rename(temporaryPath, prepared.DatabasePath); err != nil {
		return nil, err
	}
	if err := syncPath(directory); err != nil {
		return nil, err
	}
	prepared.Created = true
	return prepared, nil
}

// generationName 限定文件系统路径段，防止候选清理越过代次目录。
func generationName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name
}

func backupDatabase(ctx context.Context, source, destination string) (err error) {
	uri := url.URL{Scheme: "file", Path: source, RawQuery: "mode=rw"}
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	connection, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := connection.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	return connection.Raw(func(driverConnection any) error {
		// 驱动固定为 modernc；Raw 保持物理连接独占，直到备份结束。
		backup, err := driverConnection.(interface {
			NewBackup(string) (*sqlite.Backup, error)
		}).NewBackup(destination)
		if err != nil {
			return err
		}
		_, stepErr := backup.Step(-1)
		finishErr := backup.Finish()
		if stepErr != nil {
			if finishErr != nil {
				return errors.Join(stepErr, finishErr)
			}
			return stepErr
		}
		return finishErr
	})
}

func checkIntegrity(ctx context.Context, path string) (err error) {
	uri := url.URL{Scheme: "file", Path: path, RawQuery: "mode=rw"}
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	rows, err := db.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	var results []string
	for rows.Next() {
		var result string
		if err := rows.Scan(&result); err != nil {
			return err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(results) != 1 || results[0] != integrityOK {
		return fmt.Errorf("candidate integrity_check: %s", strings.Join(results, "\n"))
	}
	return nil
}

func syncPath(path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	return file.Sync()
}
