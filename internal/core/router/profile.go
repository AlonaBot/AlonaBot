package router

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
)

// Profile 从装配的快照入口读取配置文件；无默认目标时始终使用 default。
func (r *Router) Profile(ctx context.Context, pluginID string, target *pb.Target) (string, bool, error) {
	if r.profile == nil {
		return profile.Default, true, nil
	}
	return r.profile(ctx, pluginID, target)
}

func encodedProfile(id string) string {
	if id == profile.Default {
		return ""
	}
	return id
}
