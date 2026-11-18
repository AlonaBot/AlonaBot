package tool

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type policySnapshot struct {
	profiles *profile.Service
	id       string
}

// profileSnapshot 只接受 Session 恢复的父附件；无父 list 使用 default。
func (c *Catalog) profileSnapshot(ctx context.Context, in *wire.Inbound) (policySnapshot, error) {
	if c.resolve == nil {
		return policySnapshot{}, nil
	}
	var target *pb.Target
	if in.ParentCallID != 0 {
		target = in.ParentAttach.(*router.Invocation).DefaultTarget
	}
	profiles, id, err := c.resolve(ctx, target)
	return policySnapshot{profiles: profiles, id: id}, err
}

func (v policySnapshot) policy(pluginID, name string, operation bool) (pb.ToolPolicy, error) {
	if v.profiles == nil {
		return pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED, nil
	}
	policy, enabled, err := v.profiles.ToolPolicy(v.id, pluginID, name, operation)
	if err != nil {
		return pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED, err
	}
	if !enabled {
		return pb.ToolPolicy_TOOL_POLICY_DENY, nil
	}
	return policy, nil
}
