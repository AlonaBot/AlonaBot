// 本文件测试迁移模式的识别与迁移列表校验。
package gormdb

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

// up 是什么也不做的迁移步骤。
func up(context.Context, *gorm.DB) error { return nil }

// none 不声明迁移。
type none struct{}

// self 自行管理迁移。
type self struct{}

// versioned 声明一组按版本执行的迁移。
type versioned struct {
	// ms 是声明的迁移列表。
	ms []Migration
}

// both 同时实现两种迁移接口，应被拒绝。
type both struct {
	// versioned 提供 Migrations。
	versioned
}

// Migrate 是空的自管迁移。
func (self) Migrate(context.Context, *gorm.DB) error { return nil }

// Migrations 返回声明的迁移列表。
func (v versioned) Migrations() []Migration { return v.ms }

// Migrate 使 both 同时满足 SelfMigrator。
func (both) Migrate(context.Context, *gorm.DB) error { return nil }

// 三种迁移模式被正确识别；两种接口同时实现、ID 为 0、重复、递减或缺少 Up 时拒绝。
func TestCheckMigrators(t *testing.T) {
	ok := map[string]struct {
		// root 是被检查的根插件。
		root any
		// want 是期望的迁移模式。
		want string
	}{
		"无迁移":       {none{}, ModeNone},
		"自管迁移":      {self{}, ModeSelf},
		"辅助迁移":      {versioned{[]Migration{{1, up}, {2, up}, {10, up}}}, ModeVersioned},
		"辅助迁移（空列表）": {versioned{}, ModeVersioned},
	}
	for name, c := range ok {
		if got, err := CheckMigrators(c.root); err != nil || got != c.want {
			t.Errorf("%s: got %q, %v，期望 %q", name, got, err, c.want)
		}
	}
	bad := map[string]any{
		"Migrate 与 Migrations 同时实现": both{versioned{[]Migration{{1, up}}}},
		"ID 为 0":                    versioned{[]Migration{{0, up}}},
		"ID 重复":                     versioned{[]Migration{{1, up}, {1, up}}},
		"ID 递减":                     versioned{[]Migration{{2, up}, {1, up}}},
		"缺少 Up":                     versioned{[]Migration{{1, nil}}},
	}
	for name, root := range bad {
		if _, err := CheckMigrators(root); err == nil {
			t.Errorf("%s 应被拒绝", name)
		}
	}
}
