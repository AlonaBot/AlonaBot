package dbhook

import (
	"context"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Migrator 承载数据库入口识别到的一项迁移声明及其执行函数。
type Migrator struct {
	Mode pb.MigrationMode
	Run  func(context.Context) (*pb.MigrationState, error)
}

var registry struct {
	mu        sync.Mutex
	recognize []func(any) []Migrator
}

// Register 追加数据库入口的迁移识别函数；仅保存函数值，不执行 I/O。
func Register(recognize func(any) []Migrator) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.recognize = append(registry.recognize, recognize)
}

// Recognize 收集已导入的数据库入口从 root 识别出的全部迁移声明。
func Recognize(root any) []Migrator {
	registry.mu.Lock()
	recognizers := append([]func(any) []Migrator(nil), registry.recognize...)
	registry.mu.Unlock()
	var migrators []Migrator
	for _, recognize := range recognizers {
		migrators = append(migrators, recognize(root)...)
	}
	return migrators
}
