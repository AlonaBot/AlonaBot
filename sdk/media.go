package sdk

import (
	"context"
	"fmt"
	"io"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// Upload 顺序上传 r 的全部内容，由 Core 校验声明的大小、MIME 和摘要后签发引用。
// Session 未就绪或已停止时返回 FAILED_PRECONDITION；读取失败保留原始错误。
// 中断的上传由 Core 在期限到达或 Session 结束时回收。
func Upload(ctx context.Context, r io.Reader, size uint64, mime string, sum []byte) (*pb.MediaRef, error) {
	// 1. 经既有 ready gate 取得同一个 Session，三种服务共用它的身份。
	session, err := currentSession()
	if err != nil {
		return nil, err
	}
	call := func(target string, request, result proto.Message) error {
		encoded, err := proto.Marshal(request)
		if err != nil {
			return &Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "encode media request: " + err.Error()}
		}
		response, err := session.Call(ctx, wire.Request{Target: target, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: invocationSchemaVersion, Bytes: encoded}})
		if err != nil {
			info := err.(*wire.CallError).ErrorInfo
			return &Error{Code: validate.PublicCode(info.GetCode()), Message: info.GetSafeMessage()}
		}
		if err := decodeInvocation(response.Payload, result); err != nil {
			return &Error{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, Message: err.Error()}
		}
		return nil
	}
	begin := &pb.BeginUploadResponse{}
	if err := call(wire.ServiceMediaBeginUpload, &pb.BeginUploadRequest{Size: size, Mime: mime, Sha256: sum}, begin); err != nil {
		return nil, err
	}
	// 2. 固定大小读取直到 EOF，大小不符由 Core 的写入和提交检查拒绝。
	buffer := make([]byte, wire.MaxMediaChunk)
	var offset uint64
	for {
		n, readErr := io.ReadFull(r, buffer)
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
	// 3. 只有成功提交的响应才向调用者提供引用。
	commit := &pb.CommitUploadResponse{}
	if err := call(wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: begin.UploadId}, commit); err != nil {
		return nil, err
	}
	return commit.Ref, nil
}
