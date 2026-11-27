package media_test

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestPrepareCheckSource 经真实上传、文件与受管来源验证只读 MIME 和完整目标授权（A2）。
func TestPrepareCheckSource(t *testing.T) {
	h := newHarness(t, defaults())
	ctx := t.Context()
	target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "one", AccountId: "bot", ChatId: "chat", ThreadId: proto.String("topic")}
	for _, mime := range []string{"image/jpeg", "image/gif"} {
		content := []byte(mime)
		digest := sha256.Sum256(content) // 上传协议要求的完整性摘要。
		started := &pb.BeginUploadResponse{}
		h.call(t, h.caller, wire.ServiceMediaBeginUpload, &pb.BeginUploadRequest{Size: uint64(len(content)), Mime: mime, Sha256: digest[:]}, started)
		h.write(t, h.caller, started.UploadId, 0, content)
		committed := &pb.CommitUploadResponse{}
		h.call(t, h.caller, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: started.UploadId}, committed)
		source := &pb.MediaSource{Source: &pb.MediaSource_MediaRef{MediaRef: committed.Ref.Token}}
		before := proto.CloneOf(source)
		prepared, err := h.store.PrepareCheckSource(ctx, h.caller, target, source)
		if err != nil || prepared.GetMime() != mime || prepared.GetMediaRef() != committed.Ref.Token || !proto.Equal(source, before) {
			t.Fatalf("prepared=%v source=%v error=%v", prepared, source, err)
		}
		if _, err := h.store.PrepareCheckSource(ctx, router.NewCaller("other", "generation"), target, source); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
			t.Fatalf("foreign session: %v", err)
		}
		// 原上传引用仍可查询，预检保留其所有权及元数据。
		file, err := h.store.Lookup(committed.Ref.Token, h.caller)
		if err != nil {
			t.Fatal(err)
		}
		if file.MIME != mime {
			t.Fatalf("upload metadata changed: %+v", file)
		}
	}
	instance := memory.New("memory", "one", "bot")
	id := h.store.RegisterIncoming(instance, &pb.IncomingMedia{Token: "platform-file", Mime: proto.String("image/gif")})
	incoming := h.store.IssueIncoming(id, h.caller, []*pb.Target{target})
	source := &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: proto.CloneOf(incoming)}}
	source.GetIncomingRef().Mime = proto.String("image/jpeg")
	prepared, err := h.store.PrepareCheckSource(ctx, h.caller, target, source)
	if err != nil || prepared.GetMime() != "image/gif" || prepared.GetIncomingRef().Token != incoming.Token {
		t.Fatalf("trusted incoming metadata=%v error=%v", prepared, err)
	}
	other := proto.CloneOf(target)
	other.ThreadId = proto.String("other")
	if _, err := h.store.PrepareCheckSource(ctx, h.caller, other, source); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
		t.Fatalf("target grant: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := h.store.PrepareCheckSource(canceled, h.caller, target, source); err != context.Canceled {
		t.Fatalf("canceled check: %v", err)
	}
	h.now = h.now.Add(defaults().TTL + time.Millisecond)
	if _, err := h.store.PrepareCheckSource(ctx, h.caller, target, source); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
		t.Fatalf("expired incoming: %v", err)
	}
}
