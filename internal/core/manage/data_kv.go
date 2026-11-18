package manage

import (
	"database/sql"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	dataKeyMaximum   = 512 // CONTRACTS §6.1，单位字节。
	dataValueMaximum = 1 << 20
)

// dataKVRequest 在 ServeMux 规范化路径前接收不透明的键；管理认证与 CSRF 已完成。
func (s *Service) dataKVRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "PUT" && r.Method != "DELETE" {
		return false
	}
	parts := strings.SplitN(strings.TrimPrefix(r.URL.EscapedPath(), "/"), "/", 7)
	if len(parts) != 7 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "plugins" || parts[4] != "data" || parts[5] != "kv" {
		return false
	}
	id, err := url.PathUnescape(parts[3])
	if err != nil {
		respond(w, nil, invalid(err.Error()))
		return true
	}
	key, err := url.PathUnescape(parts[6])
	if err != nil {
		respond(w, nil, invalid(err.Error()))
		return true
	}
	r.SetPathValue("id", id)
	r.SetPathValue("key", key)
	s.dataHandler(s.dataKV)(w, r)
	return true
}

func (s *Service) dataKVList(r *http.Request, selected *dataSelection) (any, error) {
	offset, limit, err := dataPage(r.URL.Query())
	if err != nil {
		return nil, err
	}
	prefix := r.URL.Query().Get("prefix")
	if !utf8.ValidString(prefix) {
		return nil, invalid("KV prefix must be UTF-8")
	}
	return dataTransaction(r, selected, false, func(tx *sql.Tx) (any, error) {
		const condition = " WHERE substr(CAST(key AS BLOB),1,?) = ?"
		var total int64
		if err := tx.QueryRowContext(r.Context(), "SELECT count(*) FROM "+dataKVTable+condition, len(prefix), []byte(prefix)).Scan(&total); err != nil {
			return nil, err
		}
		rows, err := tx.QueryContext(r.Context(), "SELECT key,length(value),updated_at FROM "+dataKVTable+condition+" ORDER BY key LIMIT ? OFFSET ?", len(prefix), []byte(prefix), limit, offset)
		if err != nil {
			return nil, err
		}
		type entry struct {
			Key       string    `json:"key"`
			Size      int64     `json:"size"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		items := make([]entry, 0)
		for rows.Next() {
			var item entry
			var millis int64
			if err := rows.Scan(&item.Key, &item.Size, &millis); err != nil {
				return nil, closeDataRows(rows, err)
			}
			item.UpdatedAt = time.UnixMilli(millis).UTC()
			items = append(items, item)
		}
		return struct {
			Items []entry `json:"items"`
			Total int64   `json:"total"`
		}{items, total}, closeDataRows(rows, rows.Err())
	})
}

func (s *Service) dataKV(r *http.Request, selected *dataSelection) (any, error) {
	key := r.PathValue("key")
	if key == "." || key == ".." {
		return nil, invalid("KV key must not be . or ..")
	}
	if key == "" || len(key) > dataKeyMaximum || !utf8.ValidString(key) {
		return nil, invalid("KV key must be 1 to 512 bytes of UTF-8")
	}
	write := r.Method == "PUT" || r.Method == "DELETE"
	return dataTransaction(r, selected, write, func(tx *sql.Tx) (any, error) {
		switch r.Method {
		case "GET", "HEAD":
			var value []byte
			if err := tx.QueryRowContext(r.Context(), "SELECT value FROM "+dataKVTable+" WHERE key = ?", key).Scan(&value); err != nil {
				return nil, err
			}
			response := struct {
				ValueBase64 string  `json:"value_base64"`
				Text        *string `json:"text,omitempty"`
			}{ValueBase64: base64.StdEncoding.EncodeToString(value)}
			if utf8.Valid(value) {
				text := string(value)
				response.Text = &text
			}
			return response, nil
		case "PUT":
			var request struct {
				ValueBase64 *string `json:"value_base64"`
				Text        *string `json:"text"`
			}
			if err := decode(r, &request); err != nil {
				return nil, err
			}
			if (request.ValueBase64 == nil) == (request.Text == nil) {
				return nil, invalid("provide exactly one of value_base64 or text")
			}
			var value []byte
			if request.Text != nil {
				value = []byte(*request.Text)
			} else {
				var err error
				value, err = base64.StdEncoding.DecodeString(*request.ValueBase64)
				if err != nil {
					return nil, invalid(err.Error())
				}
			}
			if len(value) > dataValueMaximum {
				return nil, invalid("KV value exceeds 1 MiB")
			}
			if _, err := tx.ExecContext(r.Context(), "INSERT INTO "+dataKVTable+" (key,value,updated_at) VALUES(?,?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_at=excluded.updated_at", key, value, time.Now().UnixMilli()); err != nil {
				return nil, err
			}
		case "DELETE":
			result, err := tx.ExecContext(r.Context(), "DELETE FROM "+dataKVTable+" WHERE key = ?", key)
			if err := dataAffected(result, err); err != nil {
				return nil, err
			}
		}
		selected.audit = []any{"plugin_id", selected.id, "key", key}
		return struct{}{}, nil
	})
}
