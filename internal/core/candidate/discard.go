package candidate

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Discard 按状态记录、候选目录、孤立 secret 的顺序清理；任一步失败立即返回。
// 调用方持插件操作锁且已确认候选进程退出；p 须来自 PrepareDatabase。
// Active 或已提交的代次返回 FAILED_PRECONDITION，保留其目录与引用。
func Discard(ctx context.Context, st *state.Store, secrets *secret.Store, p *Prepared) error {
	// 1. 先核对 Active，再由状态库事务保护提交标记与代次归属。
	active, err := st.Active(ctx, p.PluginID)
	if err != nil {
		return err
	}
	if p.Generation == active {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: "cannot discard active generation"}
	}
	if err := st.DeleteCandidateGeneration(ctx, p.PluginID, p.Generation); err != nil {
		if errors.Is(err, state.ErrCommittedGeneration) {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: err.Error()}
		}
		return err
	}
	// 2. 状态已解除引用，才移除该候选目录；失败时保留 secret 供后续清理。
	if err := os.RemoveAll(filepath.Dir(p.DatabasePath)); err != nil {
		return err
	}
	// 3. 其余全部代次（含未提交候选）与 Pending 共同保护不可变 secret。
	snapshots, err := st.PluginGenerationConfigs(ctx, p.PluginID)
	if err != nil {
		return err
	}
	protected := make([][]byte, 0, len(snapshots)+1)
	for _, snapshot := range snapshots {
		protected = append(protected, []byte(snapshot))
	}
	pending, err := st.GetPendingConfig(ctx, p.PluginID)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return err
	}
	if err == nil {
		protected = append(protected, []byte(pending.ConfigJSON))
	}
	pluginDir := filepath.Dir(filepath.Dir(filepath.Dir(p.DatabasePath)))
	_, err = secrets.Collect(ctx, p.PluginID, protected, filepath.Join(pluginDir, "config.json"))
	return err
}
