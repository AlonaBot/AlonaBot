package policy

import (
	"cmp"
	"context"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Admin 绑定一个 Adapter 中的 Core 管理员身份。
type Admin struct{ AdapterID, UserID string }

// IsAdmin 从装配处设置的内存列表判定 Actor 的管理员身份。
func (p *Policy) IsAdmin(ctx context.Context, adapterID string, actor *pb.Actor) (bool, error) {
	if actor == nil {
		return false, nil
	}
	p.adminsMu.RLock()
	defer p.adminsMu.RUnlock()
	return slices.Contains(p.admins, Admin{AdapterID: adapterID, UserID: actor.GetExternalUserId()}), nil
}

// SetAdmins 整体替换管理员列表并复制输入，可与 IsAdmin、Admins 并发调用。
// 装配处负责从 core_config.json 读取并校验身份；首次调用前列表为空。
func (p *Policy) SetAdmins(admins []Admin) {
	copied := slices.Clone(admins)
	slices.SortFunc(copied, func(a, b Admin) int {
		if order := cmp.Compare(a.AdapterID, b.AdapterID); order != 0 {
			return order
		}
		return cmp.Compare(a.UserID, b.UserID)
	})
	// 只在发布完整列表时持写锁，读取始终观察到一次完整替换。
	p.adminsMu.Lock()
	p.admins = copied
	p.adminsMu.Unlock()
}

// Admins 返回按 AdapterID、UserID 排序的管理员列表副本。
func (p *Policy) Admins() []Admin {
	p.adminsMu.RLock()
	defer p.adminsMu.RUnlock()
	return slices.Clone(p.admins)
}
