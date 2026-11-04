package policy

import (
	"cmp"
	"context"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Admin 绑定一个 Adapter 实例中的 Core 管理员身份。
type Admin struct {
	InstanceID, UserID string
}

// IsAdmin 按 Actor 的实例与用户身份读取当前 Core 管理员设置。
func (p *Policy) IsAdmin(ctx context.Context, actor *pb.Actor) (bool, error) {
	if actor == nil {
		return false, nil
	}
	return p.store.IsCoreAdmin(ctx, actor.GetAdapterInstanceId(), actor.GetExternalUserId())
}

// AddAdmin 幂等添加管理员，持久化提交后才返回成功。
func (p *Policy) AddAdmin(ctx context.Context, instanceID, userID string) error {
	if instanceID == "" || userID == "" {
		return invalidArgument("instance ID and user ID are required")
	}
	return p.store.AddCoreAdmin(ctx, state.CoreAdmin{InstanceID: instanceID, UserID: userID})
}

// RemoveAdmin 幂等移除管理员，下一次事件派发使用新的设置。
func (p *Policy) RemoveAdmin(ctx context.Context, instanceID, userID string) error {
	if instanceID == "" || userID == "" {
		return invalidArgument("instance ID and user ID are required")
	}
	return p.store.RemoveCoreAdmin(ctx, instanceID, userID)
}

// Admins 按实例与用户 ID 的字典序读取管理员列表。
func (p *Policy) Admins(ctx context.Context) ([]Admin, error) {
	rows, err := p.store.CoreAdmins(ctx)
	admins := make([]Admin, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, Admin{InstanceID: row.InstanceID, UserID: row.UserID})
	}
	slices.SortFunc(admins, func(a, b Admin) int {
		if order := cmp.Compare(a.InstanceID, b.InstanceID); order != 0 {
			return order
		}
		return cmp.Compare(a.UserID, b.UserID)
	})
	return admins, err
}

func invalidArgument(message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: message}}
}
