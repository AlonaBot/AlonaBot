package sdk

import (
	"fmt"
	"slices"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// Permission 保存不可变的权限名集合；零值不要求 Actor。
type Permission struct{ names []string }

var (
	// Anyone 允许没有 Actor 的事件。
	Anyone Permission
	// User 要求可信 Actor。
	User = PermissionName("user")
	// Admin 要求 Core 管理员权限。
	Admin = PermissionName("admin")
	// ChatAdmin 要求会话管理员权限。
	ChatAdmin = PermissionName("chat_admin")
)

// PermissionName 构造权限名；路由收集时检查名称格式。
func PermissionName(name string) Permission { return Permission{names: []string{name}} }

// Or 返回排序去重后的权限并集。
func (p Permission) Or(others ...Permission) Permission {
	names := p.Names()
	for _, other := range others {
		names = append(names, other.names...)
	}
	slices.Sort(names)
	return Permission{names: slices.Compact(names)}
}

// Names 返回权限名的副本。
func (p Permission) Names() []string { return slices.Clone(p.names) }

// Has 检查当前 Actor 是否具有集合中的任一权限；空集合始终满足。
func (c *Context) Has(p Permission) bool {
	if len(p.names) == 0 {
		return true
	}
	for _, name := range c.event.GetActor().GetPermissions() {
		if slices.Contains(p.names, name) {
			return true
		}
	}
	return false
}

func checkPermission(p Permission) error {
	for _, name := range p.names {
		if !validate.IsPermissionName(name) {
			return fmt.Errorf("invalid permission name %q", name)
		}
	}
	return nil
}
