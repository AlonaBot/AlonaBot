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
	admin := Admin{AdapterID: adapterID, UserID: actor.GetExternalUserId()}
	return slices.Contains(p.admins, admin) || slices.Contains(p.runtimeAdmins, admin), nil
}

// SetAdmins 整体替换配置管理员列表并复制输入，运行期管理员继续保留。
// 装配处负责从 core_config.json 读取并校验身份；首次调用前列表为空。
func (p *Policy) SetAdmins(admins []Admin) {
	p.adminsMu.Lock()
	p.admins = slices.Clone(admins)
	p.adminsMu.Unlock()
}

// SetRuntimeAdmins 登记仅在本次运行生效的管理员，不随配置管理员的重新发布移除。
func (p *Policy) SetRuntimeAdmins(admins []Admin) {
	p.adminsMu.Lock()
	p.runtimeAdmins = slices.Clone(admins)
	p.adminsMu.Unlock()
}

// Admins 返回合并、去重并按 AdapterID、UserID 排序的有效管理员列表副本。
func (p *Policy) Admins() []Admin {
	p.adminsMu.RLock()
	copied := append(slices.Clone(p.admins), p.runtimeAdmins...)
	p.adminsMu.RUnlock()
	slices.SortFunc(copied, func(a, b Admin) int {
		if order := cmp.Compare(a.AdapterID, b.AdapterID); order != 0 {
			return order
		}
		return cmp.Compare(a.UserID, b.UserID)
	})
	return slices.Compact(copied)
}
