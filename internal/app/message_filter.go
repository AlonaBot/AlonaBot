package app

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// activeMessageConfig 沿用状态库的代次提交点；已安装和开发组件都从 Active 读取。
func activeMessageConfig(store *state.Store) func(context.Context, string, *pb.Event) (json.RawMessage, error) {
	return func(ctx context.Context, pluginID string, _ *pb.Event) (json.RawMessage, error) {
		generation, err := store.Active(ctx, pluginID)
		if err != nil {
			return nil, fmt.Errorf("read active generation: %w", err)
		}
		if generation == "" {
			return nil, fmt.Errorf("plugin %s has no active generation", pluginID)
		}
		snapshot, err := store.GenerationConfig(ctx, generation)
		if err != nil {
			return nil, fmt.Errorf("read active generation config: %w", err)
		}
		return json.RawMessage(snapshot), nil
	}
}

// profileMessageConfig 按同一会话解析有效覆盖，仅以 Active 基础配置构造本次匹配视图。
func (a *assembly) profileMessageConfig(ctx context.Context, pluginID string, event *pb.Event) (json.RawMessage, error) {
	view, err := a.profileView(ctx, pluginID, event.GetTarget())
	return view.Config, err
}
