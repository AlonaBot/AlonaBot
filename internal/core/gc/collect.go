// Package gc 按状态库引用回收代次目录、程序包与不可变 secret。
//
// Collect 保护 Active、最近三个已提交代次、Pending 和未完成 Journal。
// 调用方在 Core 持锁且管理操作串行化时调用；恢复操作完成前仍有 Journal 引用的资源持续保留。
// 备份与恢复步骤见 docs/deployment.md。
package gc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const packageDigestPrefix = 12 // packages.Candidate.MoveTo 的不可变目录摘要前缀。

// Collect 删除无人引用的代次、程序包和 secret；返回已删除的文件系统路径及完整错误。
// 调用方须串行化管理操作；本函数另持逐插件状态锁，保护引用读取与删除的顺序。
func Collect(ctx context.Context, st *state.Store, secrets *secret.Store, packagesDir string) (removed []string, err error) {
	packagesDir, err = filepath.Abs(packagesDir)
	if err != nil {
		return nil, err
	}
	installs, err := st.Installs(ctx)
	if err != nil {
		return nil, err
	}
	installed := make(map[string]bool, len(installs))
	for _, in := range installs {
		installed[in.PluginID] = true
		unlock := st.LockPlugin(in.PluginID)
		paths, collectErr := collectPlugin(ctx, st, secrets, packagesDir, in)
		unlock()
		removed = append(removed, paths...)
		if collectErr != nil {
			return removed, fmt.Errorf("collect plugin %s: %w", in.PluginID, collectErr)
		}
	}
	entries, err := os.ReadDir(packagesDir)
	if errors.Is(err, os.ErrNotExist) {
		return removed, nil
	}
	if err != nil {
		return removed, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || installed[entry.Name()] {
			continue
		}
		// purge 同事务移除安装及其引用；外键保证孤立目录没有 Pending 或 Journal 引用。
		unlock := st.LockPlugin(entry.Name())
		paths, collectErr := removeChildren(ctx, filepath.Join(packagesDir, entry.Name()), func(string) bool { return false })
		unlock()
		removed = append(removed, paths...)
		if collectErr != nil {
			return removed, fmt.Errorf("collect uninstalled plugin %s: %w", entry.Name(), collectErr)
		}
	}
	return removed, nil
}

func collectPlugin(ctx context.Context, st *state.Store, secrets *secret.Store, packagesDir string, in state.Install) (removed []string, err error) {
	// 1. 在任何删除之前收集保护快照与程序引用，Pending 同时保护其基线和独立 secret。
	ids, err := st.ReferencedGenerations(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	snapshots, err := st.PluginGenerationConfigs(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	protected := make([][]byte, 0, len(ids)+1)
	for _, id := range ids {
		if snapshot, found := snapshots[id]; found {
			protected = append(protected, []byte(snapshot))
		}
	}
	pending, err := st.GetPendingConfig(ctx, in.PluginID)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		protected = append(protected, []byte(pending.ConfigJSON))
	}
	programs := make(map[string]bool)
	if in.Source != "dev" {
		digest := hex.EncodeToString(in.BuildDigest)
		if len(digest) < packageDigestPrefix {
			return nil, fmt.Errorf("installed build digest is invalid")
		}
		programs[in.Version+"-"+digest[:packageDigestPrefix]] = true
	}
	operations, err := st.Journals(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		unfinished := operation.Phase != "committed" && operation.Phase != "cleaned"
		if operation.PackageDir != "" && (unfinished || operation.Phase == "committed" && slices.Contains(ids, operation.NewGeneration)) {
			// 包目录名是持久身份，data_dir 迁移后仍在当前 packages 下保护。
			programs[filepath.Base(operation.PackageDir)] = true
		}
	}

	// 2. 事务先解除无人引用的元数据，再删除目录；文件删除失败留下可再次回收的残留。
	for id := range snapshots {
		if slices.Contains(ids, id) {
			continue
		}
		if err := st.DeleteUnreferencedGeneration(ctx, in.PluginID, id); err != nil {
			return removed, err
		}
	}
	generationDir := filepath.Join(filepath.Dir(packagesDir), "plugins", in.PluginID, "generations")
	paths, err := removeChildren(ctx, generationDir, func(path string) bool { return slices.Contains(ids, filepath.Base(path)) })
	removed = append(removed, paths...)
	if err != nil {
		return removed, err
	}
	paths, err = removeChildren(ctx, filepath.Join(packagesDir, in.PluginID), func(path string) bool { return programs[filepath.Base(path)] })
	removed = append(removed, paths...)
	if err != nil {
		return removed, err
	}

	// 3. 文件清理成功后回收 secret，日志只记录数量，避免暴露引用或秘密值。
	snapshotPath := filepath.Join(filepath.Dir(packagesDir), "plugins", in.PluginID, "config.json")
	count, err := secrets.Collect(ctx, in.PluginID, protected, snapshotPath)
	if count != 0 {
		slog.Info(fmt.Sprintf("Removed %d unreferenced secrets for component %s", count, in.PluginID), "plugin_id", in.PluginID, "count", count)
	}
	return removed, err
}

func removeChildren(ctx context.Context, parent string, protected func(string) bool) (removed []string, err error) {
	entries, err := os.ReadDir(parent)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return removed, err
		}
		path := filepath.Join(parent, entry.Name())
		if protected(path) {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			return removed, err
		}
		removed = append(removed, path)
		slog.Info(fmt.Sprintf("Removed unreferenced resource %s", path), "path", path)
	}
	return removed, nil
}
