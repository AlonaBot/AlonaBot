package gormdb

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func up(context.Context, *gorm.DB) error { return nil }

type none struct{}
type self struct{}
type versioned struct{ ms []Migration }
type both struct{ versioned }

func (self) Migrate(context.Context, *gorm.DB) error { return nil }
func (v versioned) Migrations() []Migration          { return v.ms }
func (both) Migrate(context.Context, *gorm.DB) error { return nil }

func TestCheckMigrators(t *testing.T) {
	ok := map[string]struct {
		root any
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
