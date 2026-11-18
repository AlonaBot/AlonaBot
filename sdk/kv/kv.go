// Package kv 在当前 Session 的受管 SQLite 中保存键值数据。
//
// 数据保存在保留表 _alonabot_kv，随受管数据库的候选副本、回退与清理处理（CONTRACTS §6.1）。
// 调用须传入 Handler 或生命周期回调的 context；连接由 SDK 管理。
// [Get]、[Set]、[Delete] 操作单个键，[List] 按字节前缀与排他游标分页返回 [Page]。
// 键须为 1～512 字节 UTF-8，且须避开 HTTP 路径中的独立点段 . 与 ..。
package kv

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

const (
	maxKeyBytes   = 512
	maxValueBytes = 1 << 20
)

// Page 返回按键的字节序排列的一页结果。
type Page struct {
	Keys []string
	Next string // 下一页的排他游标；空串表示没有更多键。
}

// Get 读取 key；键不存在时返回 ok=false。
func Get(ctx context.Context, key string) (value []byte, ok bool, err error) {
	if err := checkKey(key); err != nil {
		return nil, false, err
	}
	db, err := connection(ctx)
	if err != nil {
		return nil, false, err
	}
	err = db.QueryRowContext(ctx, "SELECT value FROM _alonabot_kv WHERE key = ?", key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return value, err == nil, err
}

// Set 覆盖 key 的值；键须为 1～512 字节的 UTF-8，值最多 1 MiB。
func Set(ctx context.Context, key string, value []byte) error {
	if err := checkKey(key); err != nil {
		return err
	}
	if len(value) > maxValueBytes {
		return errcode.InvalidArgument.New("KV value exceeds 1 MiB")
	}
	db, err := connection(ctx)
	if err != nil {
		return err
	}
	if value == nil {
		value = []byte{} // nil 在 database/sql 中编码为 NULL；空值须保存为 BLOB。
	}
	_, err = db.ExecContext(ctx, "INSERT INTO _alonabot_kv (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at", key, value, time.Now().UnixMilli())
	return err
}

// Delete 删除 key；键不存在时同样成功。
func Delete(ctx context.Context, key string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	db, err := connection(ctx)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "DELETE FROM _alonabot_kv WHERE key = ?", key)
	return err
}

// List 按字节序返回以 prefix 开头、严格大于 cursor 的键，limit 须大于零。
// 空 prefix 匹配全部键，空 cursor 从头开始，后续请求传入上一页的 Next。
func List(ctx context.Context, prefix, cursor string, limit int) (page Page, err error) {
	if !utf8.ValidString(prefix) {
		return Page{}, errcode.InvalidArgument.New("KV prefix must be UTF-8")
	}
	if cursor != "" {
		if err := checkKey(cursor); err != nil {
			return Page{}, err
		}
	}
	if limit <= 0 {
		return Page{}, errcode.InvalidArgument.New("KV list limit must be positive")
	}
	db, err := connection(ctx)
	if err != nil {
		return Page{}, err
	}
	// BLOB 前缀比较保留 NUL 与通配符的字面含义；逐行读取最多 limit+1 个键。
	rows, err := db.QueryContext(ctx, "SELECT key FROM _alonabot_kv WHERE substr(CAST(key AS BLOB), 1, ?) = ? AND key > ? ORDER BY key", len(prefix), []byte(prefix), cursor)
	if err != nil {
		return Page{}, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return page, err
		}
		if len(page.Keys) == limit {
			page.Next = page.Keys[len(page.Keys)-1]
			break
		}
		page.Keys = append(page.Keys, key)
	}
	return page, rows.Err()
}

func checkKey(key string) error {
	if key == "." || key == ".." {
		return errcode.InvalidArgument.New("KV key must not be . or ..")
	}
	if len(key) == 0 || len(key) > maxKeyBytes || !utf8.ValidString(key) {
		return errcode.InvalidArgument.New("KV key must be 1 to 512 bytes of UTF-8")
	}
	return nil
}

func connection(ctx context.Context) (*sql.DB, error) {
	db, err := sqlitedb.Open(ctx)
	if errors.Is(err, sqlitedb.ErrNoDatabase) {
		return nil, errors.Join(errcode.FailedPrecondition.New("plugin has no managed SQLite database"), err)
	}
	return db, err
}
