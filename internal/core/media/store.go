// Package media 实现 Session 所有的媒体上传、引用和暂存配额。
//
// [Store] 经 [Store.AddServices] 提供顺序上传，经 [Store.Lookup] 查询已提交文件。
// 调用方负责定期 Sweep 和在 Session 结束时 ReleaseSession（CONTRACTS §1.3）。
package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// MaxChunkBytes 固定协议分块上限为 1 MiB（CONTRACTS §1.3）。
const MaxChunkBytes = 1 << 20
const tokenBytes = 16
const schemaVersion = 1
const directoryMode = 0700
const fileMode = 0600

// Limits 指定字节配额、未完成上传数量和有效期。
type Limits struct {
	MaxFileBytes, MaxSessionBytes, MaxCoreBytes int64
	MaxUploads                                  int
	TTL                                         time.Duration
}

// NewLimits 读取媒体调优参数的生效值。
func NewLimits(v tune.Values) Limits {
	return Limits{v.MediaMaxObject, v.MediaSessionBudget, v.MediaDiskBudget, v.MediaMaxUploads, v.MediaTTL}
}

// File 描述已提交且尚未过期的媒体文件。
type File struct {
	Path      string
	Size      uint64
	MIME      string
	SHA256    []byte
	ExpiresAt time.Time
}

type entry struct {
	file   File
	owner  string
	offset uint64
}

// Store 串行管理文件变更与配额；上传提交后只转移索引，保留原有占用。
// 上传文件每块写入后关闭，Store 无需持有长期文件描述符。
type Store struct {
	mu      sync.Mutex // 同时保护文件、索引与配额，避免释放和写入交错
	log     *slog.Logger
	dir     string
	lim     Limits
	now     func() time.Time
	uploads map[string]*entry
	refs    map[string]*entry
	used    map[string]int64
	total   int64
}

// Open 清空独占媒体目录中的遗留文件；清理失败直接返回原错误。
func Open(log *slog.Logger, dir string, lim Limits, now func() time.Time) (*Store, error) {
	if err := os.MkdirAll(dir, directoryMode); err != nil {
		return nil, fmt.Errorf("create media directory: %w", err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list media directory: %w", err)
	}
	for _, f := range files {
		if err := os.RemoveAll(filepath.Join(dir, f.Name())); err != nil {
			return nil, fmt.Errorf("clean media directory: %w", err)
		}
	}
	return &Store{log: log, dir: dir, lim: lim, now: now, uploads: make(map[string]*entry), refs: make(map[string]*entry), used: make(map[string]int64)}, nil
}

// AddServices 注册三个 Core 固定上传服务。
func (s *Store) AddServices(m *router.Mux) {
	m.Handle(wire.ServiceMediaBeginUpload, s.begin)
	m.Handle(wire.ServiceMediaWriteChunk, s.write)
	m.Handle(wire.ServiceMediaCommitUpload, s.commit)
}

func decode(in *wire.Inbound, msg proto.Message) error {
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(msg) || in.Payload.SchemaVersion != schemaVersion {
		return fmt.Errorf("expected media protobuf schema version %d", schemaVersion)
	}
	return proto.Unmarshal(in.Payload.Bytes, msg)
}
func failure(code pb.ErrorCode, err error) wire.Response {
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}}
}
func response(msg proto.Message) wire.Response {
	b, err := proto.Marshal(msg)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, err)
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(msg), SchemaVersion: schemaVersion, Bytes: b}}
}
func token() string {
	b := make([]byte, tokenBytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) begin(_ context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	// 1. 校验上传声明，拒绝无效大小、MIME 和摘要。
	req := &pb.BeginUploadRequest{}
	if err := decode(in, req); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	if req.Size == 0 || req.Size > uint64(s.lim.MaxFileBytes) || req.Mime == "" || len(req.Sha256) != sha256.Size {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("invalid media size, MIME or SHA256"))
	}
	// 2. 准入、建文件和登记共用锁，避免回收或其他请求穿插占用配额。
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	// 3. 先核对上传槽与两级字节余量，再创建临时文件。
	count := 0
	for _, e := range s.uploads {
		if e.owner == c.SessionKey {
			count++
		}
	}
	size := int64(req.Size)
	if count >= s.lim.MaxUploads || size > s.lim.MaxSessionBytes-s.used[c.SessionKey] || size > s.lim.MaxCoreBytes-s.total {
		return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, errors.New("media quota exceeded"))
	}
	// 4. 创建独占文件；创建或关闭失败时保留原有配额。
	id := token()
	path := filepath.Join(s.dir, id)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		return s.storageFailure(path, err)
	}
	if err := f.Close(); err != nil {
		s.remove(path)
		return s.storageFailure(path, err)
	}
	// 5. 文件就绪后登记 Begin 期限与预留，之后才返回上传 ID。
	expires := s.now().Add(s.lim.TTL)
	s.uploads[id] = &entry{file: File{path, req.Size, req.Mime, bytes.Clone(req.Sha256), expires}, owner: c.SessionKey}
	s.used[c.SessionKey] += size
	s.total += size
	return response(&pb.BeginUploadResponse{UploadId: id, ExpiresAtUnixMs: expires.UnixMilli()})
}
func (s *Store) write(_ context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	req := &pb.WriteChunkRequest{}
	if err := decode(in, req); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	e, ok := s.uploads[req.UploadId]
	if !ok || e.owner != c.SessionKey {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("upload not found"))
	}
	if req.Offset != e.offset || len(req.Data) > MaxChunkBytes || uint64(len(req.Data)) > e.file.Size-e.offset {
		s.release(s.uploads, req.UploadId, e)
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("invalid media chunk offset or size"))
	}
	f, err := os.OpenFile(e.file.Path, os.O_WRONLY|os.O_APPEND, fileMode)
	if err == nil {
		_, writeErr := f.Write(req.Data)
		err = errors.Join(writeErr, f.Close())
	}
	if err != nil {
		s.release(s.uploads, req.UploadId, e)
		return s.storageFailure(e.file.Path, err)
	}
	e.offset += uint64(len(req.Data))
	return response(&pb.WriteChunkResponse{NextOffset: e.offset})
}
func (s *Store) commit(_ context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	// 1. 校验请求，在锁内清理过期上传并核对 Session 所有权。
	req := &pb.CommitUploadRequest{}
	if err := decode(in, req); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	e, ok := s.uploads[req.UploadId]
	if !ok || e.owner != c.SessionKey {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("upload not found"))
	}
	// 2. 核对真实文件大小与摘要；持锁使校验和签发之间不会发生回收。
	f, err := os.Open(e.file.Path)
	if err != nil {
		s.release(s.uploads, req.UploadId, e)
		return s.storageFailure(e.file.Path, err)
	}
	digest := sha256.New()
	n, readErr := io.Copy(digest, f)
	if err := errors.Join(readErr, f.Close()); err != nil {
		s.release(s.uploads, req.UploadId, e)
		return s.storageFailure(e.file.Path, err)
	}
	if uint64(n) != e.file.Size || !bytes.Equal(digest.Sum(nil), e.file.SHA256) {
		s.release(s.uploads, req.UploadId, e)
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("media size or SHA256 mismatch"))
	}

	// 3. 从 Commit 起计算引用期限，转移索引并保留原有配额占用。
	id := token()
	e.file.ExpiresAt = s.now().Add(s.lim.TTL)
	delete(s.uploads, req.UploadId)
	s.refs[id] = e
	return response(&pb.CommitUploadResponse{Ref: &pb.MediaRef{Token: id, Size: e.file.Size, Mime: e.file.MIME, Sha256: bytes.Clone(e.file.SHA256), ExpiresAtUnixMs: e.file.ExpiresAt.UnixMilli()}})
}

// Lookup 校验引用的 Session 所有权与期限，返回独立的摘要副本。
func (s *Store) Lookup(token string, c router.Caller) (File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	e, ok := s.refs[token]
	if !ok || e.owner != c.SessionKey {
		return File{}, &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, SafeMessage: "media reference not found"}}
	}
	f := e.file
	f.SHA256 = bytes.Clone(f.SHA256)
	return f, nil
}

// ReleaseSession 回收指定 Session 的上传与引用，其他 Session 保持原状。
func (s *Store) ReleaseSession(sessionKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entries := range []map[string]*entry{s.uploads, s.refs} {
		for id, e := range entries {
			if e.owner == sessionKey {
				s.release(entries, id, e)
			}
		}
	}
}

// Sweep 回收全部到期上传与引用。
func (s *Store) Sweep() { s.mu.Lock(); defer s.mu.Unlock(); s.sweep() }
func (s *Store) sweep() {
	now := s.now()
	for _, entries := range []map[string]*entry{s.uploads, s.refs} {
		for id, e := range entries {
			if !now.Before(e.file.ExpiresAt) {
				s.release(entries, id, e)
			}
		}
	}
}
func (s *Store) release(entries map[string]*entry, id string, e *entry) {
	s.remove(e.file.Path)
	delete(entries, id)
	s.total -= int64(e.file.Size)
	s.used[e.owner] -= int64(e.file.Size)
	if s.used[e.owner] == 0 {
		delete(s.used, e.owner)
	}
}
func (s *Store) remove(path string) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.log.Warn("remove media file failed", "path", path, "err", err)
	}
}

func (s *Store) storageFailure(path string, err error) wire.Response {
	s.log.Error("media storage failed", "path", path, "err", err)
	return failure(pb.ErrorCode_ERROR_CODE_INTERNAL, errors.New("media storage failed"))
}
