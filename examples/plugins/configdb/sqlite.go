//go:build !configdb_gorm

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

// registerMigrations 登记两步迁移；每步 DDL 与账本由 SDK 在同一事务内提交。
func (p *plugin) registerMigrations(root *sdk.Plugin) {
	sqlitedb.Migrations(root, []sqlitedb.Migration{
		{ID: createNotesMigration, Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, notesDDL)
			return err
		}},
		{ID: indexNotesMigration, Up: func(ctx context.Context, tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, noteIndexDDL); err != nil {
				return err
			}
			if p.plugin.Config().Bool("fail_migration") {
				return errors.New("requested failure in second migration")
			}
			return nil
		}},
	})
}

// OnStart 在当前代次的数据库写入启动记录，然后执行可控失败点。
func (p *plugin) OnStart(ctx context.Context) error {
	if err := saveStartup(ctx, p.plugin.Config().String("greeting")); err != nil {
		return err
	}
	return p.started()
}

// saveStartup 覆盖写入启动记录；它只写启动表，不占用用户的笔记键。
func saveStartup(ctx context.Context, value string) error {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO "+startupTable+" (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", startupKey, value)
	return err
}

func saveNote(ctx context.Context, key, value string) error {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO "+notesTable+" (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
	return err
}

func readNote(ctx context.Context, key string) (string, error) {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return "", err
	}
	var value string
	err = db.QueryRowContext(ctx, "SELECT value FROM "+notesTable+" WHERE key = ?", key).Scan(&value)
	return value, err
}

func deleteNote(ctx context.Context, key string) error {
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "DELETE FROM "+notesTable+" WHERE key = ?", key)
	return err
}
