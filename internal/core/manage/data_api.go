package manage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"modernc.org/sqlite"
	sqlitecodes "modernc.org/sqlite/lib"
)

const (
	dataPageDefault = 50 // ADMIN-API §21。
	dataPageMaximum = 500
	dataBusyTimeout = "5000" // CONTRACTS §6.1，单位毫秒。
	dataKVTable     = "_alonabot_kv"
)

// dataSelection 的路径与代次只在持有逐插件管理锁时使用。
type dataSelection struct {
	id               string
	database         string
	directory        string
	readonly         bool
	audit            []any  // 成功提交后记录定位信息，禁止放入取值或文件正文。
	beforeDataRename func() // 本地回归确定性复现检查与移动之间的文件变化。
}

type dataOperation func(*http.Request, *dataSelection) (any, error)

func (s *Service) dataRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/data/export", s.dataHandler(s.dataExport))
	mux.HandleFunc("POST /api/v1/plugins/{id}/data/clear", s.dataHandler(s.dataClear))

	mux.HandleFunc("GET /api/v1/plugins/{id}/data/files", s.dataHandler(s.dataFiles))
	for _, method := range []string{"GET", "PUT"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/data/files/content", s.dataHandler(s.dataContent))
	}
	for _, action := range []string{"mkdir", "rename", "delete"} {
		mux.HandleFunc("POST /api/v1/plugins/{id}/data/files/"+action, s.dataHandler(s.dataFileAction))
	}

	mux.HandleFunc("GET /api/v1/plugins/{id}/data/kv", s.dataHandler(s.dataKVList))
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/data/kv/{key...}", s.dataHandler(s.dataKV))
	}

	mux.HandleFunc("GET /api/v1/plugins/{id}/data/tables", s.dataHandler(s.dataTables))
	for _, method := range []string{"GET", "POST"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/data/tables/{table}/rows", s.dataHandler(s.dataRows))
	}
	for _, method := range []string{"PATCH", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/data/tables/{table}/rows/{rowid}", s.dataHandler(s.dataRows))
	}
}

func (s *Service) dataHandler(operation dataOperation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 选路径、访问文件和提交事务共用生命周期锁，避免代次被中途替换或回收。
		unlock := s.lock(r.PathValue("id"))
		defer unlock()
		release := s.opt.Store.LockPlugin(r.PathValue("id"))
		defer release()
		selected, err := s.selectData(r)
		var result any
		if err == nil {
			result, err = operation(r, selected)
		}
		if err == nil && len(selected.audit) != 0 {
			slog.InfoContext(r.Context(), "plugin data changed", selected.audit...)
		}
		if download, ok := result.(*dataDownload); ok {
			if err == nil {
				download.serve(w, r)
				return
			}
			err = errors.Join(err, download.file.Close())
			if download.cleanup != nil {
				err = errors.Join(err, download.cleanup())
			}
		}
		respond(w, result, dataFailure(err))
	}
}

func (s *Service) selectData(r *http.Request) (*dataSelection, error) {
	id := r.PathValue("id")
	in, err := s.opt.Store.GetInstall(r.Context(), id)
	if err != nil {
		return nil, err
	}
	if in.Uninstalled {
		return nil, state.ErrNotFound
	}
	active, err := s.opt.Store.Active(r.Context(), id)
	if err != nil {
		return nil, err
	}
	retained, err := s.opt.Store.Retained(r.Context(), id)
	if err != nil {
		return nil, err
	}
	generation := r.URL.Query().Get("generation")
	if generation == "" {
		generation = active
	}
	if generation != "" && generation != active && !slices.Contains(retained, generation) {
		return nil, state.ErrNotFound
	}
	selected := &dataSelection{id: id, readonly: generation != active, directory: filepath.Join(filepath.Dir(s.opt.Packages), "plugin_data", id)}
	if generation != "" {
		selected.database = s.generationDatabase(in, generation)
	}
	return selected, nil
}

func dataFailure(err error) error {
	if err == nil {
		return nil
	}
	var sqliteError *sqlite.Error
	if errors.As(err, &sqliteError) {
		switch sqliteError.Code() & 0xff {
		case sqlitecodes.SQLITE_BUSY, sqlitecodes.SQLITE_LOCKED:
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: err.Error()}
		case sqlitecodes.SQLITE_CONSTRAINT:
			return invalid(err.Error())
		}
	}
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, sql.ErrNoRows) {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: err.Error()}
	}
	if errors.Is(err, os.ErrExist) {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: err.Error()}
	}
	return err
}

func dataPage(query url.Values) (offset, limit int, err error) {
	limit = dataPageDefault
	for _, name := range []string{"offset", "limit"} {
		raw := query.Get(name)
		if raw == "" {
			continue
		}
		value, e := strconv.Atoi(raw)
		if e != nil || value < 0 || (name == "limit" && value == 0) {
			return 0, 0, invalid("invalid " + name)
		}
		if name == "offset" {
			offset = value
		} else {
			limit = min(value, dataPageMaximum)
		}
	}
	return offset, limit, nil
}

func dataDatabase(ctx context.Context, selected *dataSelection, write bool) (*sql.DB, error) {
	if selected.database == "" {
		return nil, state.ErrNotFound
	}
	mode := "rw"
	if selected.readonly {
		mode = "ro"
	}
	path := filepath.ToSlash(selected.database)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	query := url.Values{"mode": {mode}, "_pragma": {"synchronous(FULL)", "foreign_keys(ON)", "busy_timeout(" + dataBusyTimeout + ")"}}
	if write {
		query.Set("_txlock", "immediate")
	}
	if !selected.readonly {
		query.Add("_pragma", "journal_mode(WAL)")
	}
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: path, RawQuery: query.Encode()}).String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func dataTransaction(r *http.Request, selected *dataSelection, write bool, operation func(*sql.Tx) (any, error)) (result any, err error) {
	if write && selected.readonly {
		return nil, denied("retained generation is read-only")
	}
	// mode=rw/ro 阻止 SQLite 将缺失的受管库创建为空库。
	if selected.database == "" {
		return nil, state.ErrNotFound
	}
	if _, err := os.Stat(selected.database); err != nil {
		return nil, err
	}
	db, err := dataDatabase(r.Context(), selected, write)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	tx, err := db.BeginTx(r.Context(), &sql.TxOptions{ReadOnly: !write})
	if err != nil {
		return nil, err
	}
	result, err = operation(tx)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}
	return result, tx.Commit()
}

func (s *Service) dataExport(r *http.Request, selected *dataSelection) (result any, err error) {
	// 1. 所有收集和压缩先完成到临时文件，故障时仍可返回原始错误。
	directory, err := os.MkdirTemp("", "alonabot-data-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(directory))
		}
	}()
	snapshot := filepath.Join(directory, "db.sqlite")
	if err := dataBackup(r, selected, snapshot); err != nil {
		return nil, err
	}
	root, err := dataFilesRoot(selected)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	file, err := os.OpenFile(filepath.Join(directory, "data.zip"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := dataArchive(r.Context(), root, snapshot, file); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	// 2. 响应持有已完成的归档句柄，发送结束后释放临时目录。
	return &dataDownload{file: file, name: selected.id + "-data.zip", mediaType: "application/zip", cleanup: func() error { return os.RemoveAll(directory) }}, nil
}

func dataBackup(r *http.Request, selected *dataSelection, destination string) (err error) {
	if selected.database == "" {
		return state.ErrNotFound
	}
	if _, err := os.Stat(selected.database); err != nil {
		return err
	}
	db, err := dataDatabase(r.Context(), selected, false)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	connection, err := db.Conn(r.Context())
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, connection.Close()) }()
	tx, err := connection.BeginTx(r.Context(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer func() {
		if rollback := tx.Rollback(); !errors.Is(rollback, sql.ErrTxDone) {
			err = errors.Join(err, rollback)
		}
	}()
	// 固定本请求的读取快照；Backup 与事务共用这一个物理连接。
	var schemaCount int
	if err := tx.QueryRowContext(r.Context(), "SELECT count(*) FROM sqlite_schema").Scan(&schemaCount); err != nil {
		return err
	}
	temporary, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	err = connection.Raw(func(driverConnection any) error {
		destinationPath := filepath.ToSlash(destination)
		if !strings.HasPrefix(destinationPath, "/") {
			destinationPath = "/" + destinationPath
		}
		backup, err := driverConnection.(interface {
			NewBackup(string) (*sqlite.Backup, error)
		}).NewBackup((&url.URL{Scheme: "file", Path: destinationPath, RawQuery: "mode=rw"}).String())
		if err != nil {
			return err
		}
		_, stepErr := backup.Step(-1)
		return errors.Join(stepErr, backup.Finish())
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) dataClear(r *http.Request, selected *dataSelection) (any, error) {
	var request struct {
		Confirm bool `json:"confirm"`
	}
	if err := decode(r, &request); err != nil {
		return nil, err
	}
	if !request.Confirm {
		return nil, invalid("confirm=true is required")
	}
	// 1. 业务行和 KV 同一事务提交；延迟外键检查允许先删父表再删子表。
	_, err := dataTransaction(r, selected, true, func(tx *sql.Tx) (any, error) {
		if _, err := tx.ExecContext(r.Context(), "PRAGMA defer_foreign_keys=ON"); err != nil {
			return nil, err
		}
		tables, err := dataSchema(r, tx, false, "")
		if err != nil {
			return nil, err
		}
		var targets []string
		for _, table := range tables {
			if !table.Writable && table.Name != dataKVTable {
				continue
			}
			if _, err := tx.ExecContext(r.Context(), "DELETE FROM "+dataQuote(table.Name)); err != nil {
				return nil, err
			}
			targets = append(targets, table.Name)
		}
		// 删除触发器可能重新填充此前清空的表，提交前核对整个清空集合。
		for _, name := range targets {
			var remaining bool
			if err := tx.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM "+dataQuote(name)+")").Scan(&remaining); err != nil {
				return nil, err
			}
			if remaining {
				return nil, fmt.Errorf("clear left rows in table %q", name)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(r.Context(), "plugin data changed", "plugin_id", selected.id, "operation", "clear")
	// 2. 文件系统单独清理，错误原样返回；数据库已提交的结果保持可观察。
	root, err := dataFilesRoot(selected)
	if err != nil {
		return nil, err
	}
	err = dataClearFiles(root)
	return struct{}{}, errors.Join(err, root.Close())
}
