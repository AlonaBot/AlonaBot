package sdk

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"google.golang.org/protobuf/proto"
)

// OpenMedia 立即读取来源令牌的第一块，成功后返回按顺序读取的 reader。
// Read 沿用 ctx，Close 停止读取而保留令牌授权。
func OpenMedia(ctx context.Context, src message.MediaSource) (io.ReadCloser, error) {
	if src.InlineBase64 != "" || src.HTTPSURL != "" || src.PlatformFile != "" || (src.Incoming != nil && src.MediaRef != "") {
		return nil, errcode.InvalidArgument.New("media source must contain only an incoming or media_ref token")
	}
	token := src.MediaRef
	if src.Incoming != nil {
		token = src.Incoming.Token
	}
	if token == "" {
		return nil, errcode.InvalidArgument.New("media source requires a token")
	}
	r := &mediaReader{ctx: ctx, token: token}
	if err := r.nextChunk(); err != nil {
		return nil, err
	}
	return r, nil
}

type mediaReader struct {
	ctx         context.Context
	token       string
	data        []byte
	next        uint64
	eof, closed bool
}

func (r *mediaReader) nextChunk() error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	result, err := Proto[*pb.ReadRequest, *pb.ReadResponse](wire.ServiceMediaRead).Call(r.ctx, &pb.ReadRequest{Token: r.token, Offset: r.next, Length: wire.MaxMediaChunk})
	if err != nil {
		if cancelled := r.ctx.Err(); cancelled != nil {
			return cancelled
		}
		return err
	}
	r.data, r.next, r.eof = result.Data, result.NextOffset, result.Eof
	return nil
}
func (r *mediaReader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	for len(r.data) == 0 {
		if r.eof {
			return 0, io.EOF
		}
		if err := r.nextChunk(); err != nil {
			return 0, err
		}
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}
func (r *mediaReader) Close() error { r.closed = true; r.data = nil; return nil }

// GrantMedia 为接收插件申请会话范围受限的新令牌。
func GrantMedia(ctx context.Context, token, granteePluginID string, sessions ...string) (string, error) {
	var targets []*pb.Target
	for _, s := range sessions {
		t, err := convert.TargetToProto(s, "")
		if err != nil {
			return "", err
		}
		targets = append(targets, t)
	}
	result, err := Proto[*pb.GrantRequest, *pb.GrantResponse](wire.ServiceMediaGrant).Call(ctx, &pb.GrantRequest{Token: token, GranteePluginId: granteePluginID, Targets: targets})
	if err != nil {
		return "", err
	}
	return result.Token, nil
}

// Upload 计算内容大小与摘要，再分块上传；mime 为空时按开头内容判断。
// 可 Seek 的来源回到当前起始位置；其他来源通过 0600 临时文件完成第二遍读取。
func Upload(ctx context.Context, r io.Reader, mime string) (ref *message.MediaRef, err error) {
	// 1. 先通过就绪门禁，再准备两遍读取的来源。
	session, err := currentSession(ctx)
	if err != nil {
		return nil, err
	}
	var input io.Reader = r
	var spool *os.File
	var seeker io.Seeker
	var start int64
	if s, ok := r.(io.Seeker); ok {
		seeker = s
		start, err = s.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("seek media upload: %w", err)
		}
	} else {
		spool, err = os.CreateTemp("", "alonabot-upload-*")
		if err != nil {
			return nil, fmt.Errorf("create media upload: %w", err)
		}
		defer func() {
			if cleanup := errors.Join(spool.Close(), os.Remove(spool.Name())); cleanup != nil {
				err = errors.Join(err, cleanup)
			}
		}()
		seeker = spool
		input = spool
	}
	// 2. 一遍计算摘要与大小，只保留 MIME 检测需要的开头。
	hash := sha256.New()
	var writer io.Writer = hash
	if spool != nil {
		writer = io.MultiWriter(hash, spool)
	}
	buffer := make([]byte, wire.MaxMediaChunk)
	const sniffBytes = 512 // http.DetectContentType 最多使用开头 512 字节。
	var prefix []byte
	var size uint64
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, readErr := r.Read(buffer)
		if n > 0 {
			if len(prefix) < sniffBytes {
				take := min(n, sniffBytes-len(prefix))
				prefix = append(prefix, buffer[:take]...)
			}
			if _, err := writer.Write(buffer[:n]); err != nil {
				return nil, fmt.Errorf("prepare media upload: %w", err)
			}
			size += uint64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read media upload: %w", readErr)
		}
	}
	if mime == "" {
		mime = http.DetectContentType(prefix)
	}
	if _, err := seeker.Seek(start, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind media upload: %w", err)
	}
	call := func(target string, request, result proto.Message) error {
		encoded, err := proto.Marshal(request)
		if err != nil {
			return errcode.InvalidArgument.New("encode media request: " + err.Error())
		}
		response, err := session.Call(ctx, wire.Request{Target: target, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
		if err != nil {
			info := err.(*wire.CallError).ErrorInfo
			return &Error{Code: errcode.Code(validate.PublicCode(info.GetCode())), Message: info.GetSafeMessage()}
		}
		if err := decodeInvocation(response.Payload, result); err != nil {
			return errcode.Internal.New(err.Error())
		}
		return nil
	}
	begin := &pb.BeginUploadResponse{}
	if err := call(wire.ServiceMediaBeginUpload, &pb.BeginUploadRequest{Size: size, Mime: mime, Sha256: hash.Sum(nil)}, begin); err != nil {
		return nil, err
	}
	// 3. 第二遍按 Core 确认的偏移上传，内容改变由 Core 校验拒绝。
	var offset uint64
	for {
		n, readErr := io.ReadFull(input, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("read media upload: %w", readErr)
		}
		if n > 0 {
			chunk := &pb.WriteChunkResponse{}
			if err := call(wire.ServiceMediaWriteChunk, &pb.WriteChunkRequest{UploadId: begin.UploadId, Offset: offset, Data: buffer[:n]}, chunk); err != nil {
				return nil, err
			}
			offset = chunk.NextOffset
		}
		if readErr != nil {
			break
		}
	}
	commit := &pb.CommitUploadResponse{}
	if err := call(wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: begin.UploadId}, commit); err != nil {
		return nil, err
	}
	return convert.MediaRefFromProto(commit.Ref), nil
}
