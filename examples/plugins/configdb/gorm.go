//go:build configdb_gorm

package main

import (
	"context"
	"errors"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/gormdb"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type note struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// startupRecord 是启动表的行，与用户笔记分开存放。
type startupRecord struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// registerMigrations 登记 GORM 版本的两步迁移，回调始终使用传入的 tx。
func (p *plugin) registerMigrations(root *sdk.Plugin) {
	gormdb.Migrations(root, []gormdb.Migration{
		{ID: createNotesMigration, Up: func(ctx context.Context, tx *gorm.DB) error {
			return tx.WithContext(ctx).Exec(notesDDL).Error
		}},
		{ID: indexNotesMigration, Up: func(ctx context.Context, tx *gorm.DB) error {
			if err := tx.WithContext(ctx).Exec(noteIndexDDL).Error; err != nil {
				return err
			}
			if p.plugin.Config().Bool("fail_migration") {
				return errors.New("requested failure in second migration")
			}
			return nil
		}},
	})
}

// OnStart 经 GORM 写入，再经 sqlitedb 读取，核对两个入口共用连接与数据。
func (p *plugin) OnStart(ctx context.Context) error {
	if err := saveStartup(ctx, p.plugin.Config().String("greeting")); err != nil {
		return err
	}
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return err
	}
	wrapped, err := gormdb.Open(ctx)
	if err != nil {
		return err
	}
	connection, err := wrapped.DB()
	if err != nil {
		return err
	}
	if connection != db {
		return errors.New("database entry points use different connections")
	}
	var value string
	if err := db.QueryRowContext(ctx, "SELECT value FROM "+startupTable+" WHERE key = ?", startupKey).Scan(&value); err != nil {
		return err
	}
	if value != p.plugin.Config().String("greeting") {
		return errors.New("database entry points read different startup values")
	}
	return p.started()
}

// saveStartup 经 GORM 覆盖写入启动记录；它只写启动表，不占用用户的笔记键。
func saveStartup(ctx context.Context, value string) error {
	db, err := gormdb.Open(ctx)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Table(startupTable).Clauses(clause.OnConflict{UpdateAll: true}).Create(&startupRecord{Key: startupKey, Value: value}).Error
	})
}

func saveNote(ctx context.Context, key, value string) error {
	db, err := gormdb.Open(ctx)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Table(notesTable).Clauses(clause.OnConflict{UpdateAll: true}).Create(&note{Key: key, Value: value}).Error
	})
}

func readNote(ctx context.Context, key string) (string, error) {
	db, err := gormdb.Open(ctx)
	if err != nil {
		return "", err
	}
	var stored note
	err = db.WithContext(ctx).Table(notesTable).Where(clause.Eq{Column: clause.PrimaryColumn, Value: key}).Take(&stored).Error
	return stored.Value, err
}

func deleteNote(ctx context.Context, key string) error {
	db, err := gormdb.Open(ctx)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(notesTable).Where(clause.Eq{Column: clause.PrimaryColumn, Value: key}).Delete(&note{}).Error
}
