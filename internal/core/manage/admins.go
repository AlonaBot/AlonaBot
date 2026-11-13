package manage

import (
	"context"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// AddAdmin 保存用户字符串并立即发布新的 Core Admin 列表。
func (s *Service) AddAdmin(ctx context.Context, user string) error {
	return s.editAdmin(ctx, user, true)
}

// RemoveAdmin 移除用户字符串并立即发布新的 Core Admin 列表。
func (s *Service) RemoveAdmin(ctx context.Context, user string) error {
	return s.editAdmin(ctx, user, false)
}

// Admins 从当前策略读取排序后的 Core Admin 用户字符串。
func (s *Service) Admins(ctx context.Context) []string {
	s.admins.Lock()
	defer s.admins.Unlock()
	result := make([]string, 0)
	for _, admin := range s.opt.Policy.Admins() {
		result = append(result, admin.AdapterID+":"+admin.UserID)
	}
	return result
}

func (s *Service) editAdmin(ctx context.Context, user string, add bool) error {
	if err := ctx.Err(); err != nil {
		return failure(err)
	}
	adapterID, userID, found := strings.Cut(user, ":")
	if !found || adapterID == "" || userID == "" {
		return invalid("admin must be <adapter_id>:<external_user_id>")
	}
	if _, err := validate.PluginID(adapterID); err != nil || strings.HasPrefix(adapterID, validate.MarketPrefix) {
		return invalid("invalid admin adapter ID")
	}
	// 写回设置和更新内存列表共用一把锁，后来的管理请求只会读到成对状态。
	s.admins.Lock()
	defer s.admins.Unlock()
	users := make([]string, 0)
	for _, admin := range s.opt.Settings.Settings().Admins {
		users = append(users, admin.String())
	}
	if add && !slices.Contains(users, user) {
		users = append(users, user)
	} else if !add {
		users = slices.DeleteFunc(users, func(value string) bool { return value == user })
	}
	err := s.opt.Settings.Update(func(doc *settings.Document) error { return doc.Set([]string{"admins"}, users) })
	// Rename 后的目录同步错误仍会发布设置快照，因此总是按当前有效设置更新策略。
	admins := make([]policy.Admin, 0)
	for _, admin := range s.opt.Settings.Settings().Admins {
		admins = append(admins, policy.Admin{AdapterID: admin.AdapterID, UserID: admin.UserID})
	}
	s.opt.Policy.SetAdmins(admins)
	if err != nil {
		return failure(err)
	}
	return nil
}
