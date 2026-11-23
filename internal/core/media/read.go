package media

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// AddReadServices 注册媒体读取与授权入口，绑定可信实例、派发关联与当前实例许可查询。
func (s *Store) AddReadServices(m *router.Mux, readers *adapter.Registry, dispatch *router.Router, allowed func(ctx context.Context, pluginID, instanceID string) (bool, error)) {
	s.mu.Lock()
	s.readers = readers
	s.mu.Unlock()
	m.Handle(wire.ServiceMediaRead, s.operation(s.read))
	m.Handle(wire.ServiceMediaGrant, s.operation(func(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
		return s.grant(ctx, c, in, dispatch, allowed)
	}))
}

func (s *Store) read(ctx context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	request := &pb.ReadRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	return s.readChunk(ctx, c, request)
}

// ReadDebugIncoming 为经管理认证的单次读取签发令牌，复用入站对象的取件和读取检查。
func (s *Store) ReadDebugIncoming(ctx context.Context, c router.Caller, instance adapter.Instance, request *pb.AdapterMediaReadRequest) wire.Response {
	s.mu.Lock()
	s.sweep()
	object := s.incoming[request.FileId]
	if object == nil || object.source.instance != instance {
		s.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("incoming media object not found for this adapter instance"))
	}
	current, ok := s.readers.Resolve(ctx, instance.InstanceID())
	if !ok || current != instance {
		s.mu.Unlock()
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, errors.New("media source adapter instance is no longer running"))
	}
	access := s.issueIncoming(object, c, nil)
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.incomingTokens, access.Token)
		if _, charged := object.charges[c.SessionKey]; charged {
			s.uncharge(object, c.SessionKey)
		}
		s.sweepIncoming()
	}()
	return s.readChunk(ctx, c, &pb.ReadRequest{Token: access.Token, Offset: request.Offset, Length: request.Length})
}

func (s *Store) readChunk(ctx context.Context, c router.Caller, request *pb.ReadRequest) wire.Response {
	if request.Length == 0 || request.Length > wire.MaxMediaChunk {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("invalid media read length"))
	}
	// 1. 入站对象先完成唯一取件；已提交的上传对象沿用已有索引和身份。
	s.mu.Lock()
	_, incoming := s.incomingTokens[request.Token]
	s.mu.Unlock()
	if incoming {
		if err := s.loadIncoming(ctx, c, request.Token); err != nil {
			var call *wire.CallError
			if errors.As(err, &call) {
				return wire.Response{Err: call.ErrorInfo}
			}
			return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, err)
		}
	}
	// 2. 持有同一回收锁读取最多一块，防止 Sweep 或 Session 释放抢先删除文件。
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	var file File
	if incoming {
		access, err := s.incomingAccess(request.Token, c.SessionKey)
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("media token not found"))
		}
		file = access.object.file
	} else {
		entry := s.refs[request.Token]
		if entry == nil || entry.owner != c.SessionKey {
			return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("media token not found"))
		}
		file = entry.file
	}
	if request.Offset > file.Size {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("media offset exceeds file size"))
	}
	if request.Offset == file.Size {
		return response(&pb.ReadResponse{NextOffset: request.Offset, Eof: true})
	}
	f, err := os.Open(file.Path)
	if err != nil {
		return s.storageFailure(file.Path, err)
	}
	data := make([]byte, min(uint64(request.Length), file.Size-request.Offset))
	_, readErr := f.ReadAt(data, int64(request.Offset))
	if err := errors.Join(readErr, f.Close()); err != nil {
		return s.storageFailure(file.Path, err)
	}
	next := request.Offset + uint64(len(data))
	return response(&pb.ReadResponse{Data: data, NextOffset: next, Eof: next == file.Size})
}

func (s *Store) fetchIncoming(ctx context.Context, id string, object *incomingObject, reader adapter.MediaReader, reserved int64) (result File, err error) {
	path := filepath.Join(s.dir, id)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		slog.Error("media storage failed", "path", path, "err", err)
		return File{}, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "media storage failed")
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Error("media storage failed", "path", path, "err", closeErr)
			err = errors.Join(err, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "media storage failed"))
		}
		// 最后一块返回或文件关闭期间也可能到期，成功结果发布前仍须遵守取件期限。
		if ctxErr := ctx.Err(); err == nil && ctxErr != nil {
			err = incomingFailure(ctxErr)
		}
		if err != nil {
			s.remove(path)
		}
	}()
	digest := sha256.New()
	writer := io.MultiWriter(f, digest)
	var offset uint64
	for {
		if err := ctx.Err(); err != nil {
			return File{}, incomingFailure(err)
		}
		data, next, eof, readErr := reader.ReadMedia(ctx, object.source.fileID, offset, wire.MaxMediaChunk)
		if readErr != nil {
			return File{}, incomingFailure(readErr)
		}
		if len(data) > wire.MaxMediaChunk || next < offset || next-offset != uint64(len(data)) || (len(data) == 0 && !eof) {
			return File{}, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "invalid adapter media read response")
		}
		if next > uint64(reserved) {
			return File{}, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "media content exceeds reserved size")
		}
		if _, err := writer.Write(data); err != nil {
			slog.Error("media storage failed", "path", path, "err", err)
			return File{}, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "media storage failed")
		}
		offset = next
		if eof {
			break
		}
	}
	if object.size != nil && offset != *object.size {
		return File{}, mediaCallError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, fmt.Sprintf("media size mismatch: received %d, expected %d", offset, *object.size))
	}
	result = File{Path: path, Size: offset, SHA256: digest.Sum(nil)}
	if object.mime != nil {
		result.MIME = *object.mime
	}
	return result, nil
}
