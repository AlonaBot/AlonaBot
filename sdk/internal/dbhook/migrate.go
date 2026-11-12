package dbhook

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// LedgerTable 指定两个数据库入口共用的迁移账本名。
const LedgerTable = "_alonabot_migrations"

// LedgerDDL 固定迁移账本的持久化格式。
const LedgerDDL = "CREATE TABLE IF NOT EXISTS " + LedgerTable + " (id INTEGER PRIMARY KEY, applied_at INTEGER NOT NULL)"

// Migrator 保存对象上登记的一种迁移方式；Check 只核对声明，不执行 I/O。
type Migrator struct {
	Mode  pb.MigrationMode
	Check func() error
	Run   func(context.Context) (*pb.MigrationState, error)
}

// State 由 Plugin 以私有字段持有，数据库入口通过封闭接口登记。
type State struct {
	migrators []Migrator
	frozen    atomic.Bool
}

func (s *State) migrationState() *State { return s }

type owner interface{ migrationState() *State }

// Register 为同一个插件保存迁移；重复登记由 Validate 报错。
func Register(p owner, m Migrator) {
	s := p.migrationState()
	if s.frozen.Load() {
		panic("plugin registration is closed after Run starts")
	}
	s.migrators = append(s.migrators, m)
}

// Freeze 关闭插件迁移登记入口。
func Freeze(p owner) { p.migrationState().frozen.Store(true) }

// Migrators 返回插件对象持有的迁移声明。
func Migrators(p owner) []Migrator { return p.migrationState().migrators }

// Validate 核对四种登记入口合计一次及每种方式的声明。
func Validate(p owner) error {
	list := Migrators(p)
	var errs []error
	if len(list) > 1 {
		errs = append(errs, fmt.Errorf("plugin has multiple migration registrations"))
	}
	for _, m := range list {
		errs = append(errs, m.Check())
	}
	return errors.Join(errs...)
}
