package media_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

type uploadHarness struct {
	dir    string
	logs   bytes.Buffer
	store  *media.Store
	mux    *router.Mux
	caller router.Caller
	now    time.Time
}

func newHarness(t *testing.T, lim media.Limits) *uploadHarness {
	t.Helper()
	h := &uploadHarness{mux: router.NewMux(), caller: router.NewCaller("plugin", "generation"), now: time.Unix(1000, 0), dir: t.TempDir()}
	var err error
	h.store, err = media.Open(slog.New(slog.NewTextHandler(&h.logs, nil)), h.dir, lim, func() time.Time { return h.now })
	if err != nil {
		t.Fatal(err)
	}
	h.store.AddServices(h.mux)
	t.Cleanup(func() { h.store.ReleaseSession(h.caller.SessionKey) })
	return h
}
func (h *uploadHarness) call(t *testing.T, c router.Caller, target string, req proto.Message, out proto.Message, code pb.ErrorCode) {
	t.Helper()
	b, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	r := h.mux.For(c)(context.Background(), &wire.Inbound{Target: target, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(req), SchemaVersion: 1, Bytes: b}})
	if code == pb.ErrorCode_ERROR_CODE_INTERNAL && strings.Contains(r.Err.GetSafeMessage(), h.dir) {
		t.Fatalf("internal response exposed media directory: %q", r.Err.GetSafeMessage())
	}
	if r.Err.GetCode() != code {
		t.Fatalf("%s: error = %v, want %v", target, r.Err, code)
	}
	if code != 0 {
		if len(r.Payload.Bytes) != 0 {
			t.Fatal("rejection returned payload")
		}
		return
	}
	if err := proto.Unmarshal(r.Payload.Bytes, out); err != nil {
		t.Fatal(err)
	}
}
func (h *uploadHarness) begin(t *testing.T, c router.Caller, b []byte, code pb.ErrorCode) string {
	t.Helper()
	sum := sha256.Sum256(b)
	out := &pb.BeginUploadResponse{}
	h.call(t, c, wire.ServiceMediaBeginUpload, &pb.BeginUploadRequest{Size: uint64(len(b)), Mime: "text/plain", Sha256: sum[:]}, out, code)
	return out.UploadId
}
func (h *uploadHarness) write(t *testing.T, c router.Caller, id string, offset uint64, b []byte, code pb.ErrorCode) {
	t.Helper()
	out := &pb.WriteChunkResponse{}
	h.call(t, c, wire.ServiceMediaWriteChunk, &pb.WriteChunkRequest{UploadId: id, Offset: offset, Data: b}, out, code)
	if code == 0 && out.NextOffset != offset+uint64(len(b)) {
		t.Fatal("incorrect next offset")
	}
}
func defaults() media.Limits {
	return media.Limits{MaxFileBytes: 32, MaxSessionBytes: 64, MaxCoreBytes: 128, MaxUploads: 4, TTL: time.Minute}
}
func TestUpload(t *testing.T) {
	for _, name := range []string{"commit", "offset", "length", "digest", "expiry", "release"} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, defaults())
			c := h.caller
			b := []byte("abcdef")
			id := h.begin(t, c, b, 0)
			other := router.NewCaller("plugin", "generation")
			h.write(t, other, id, 0, b, pb.ErrorCode_ERROR_CODE_NOT_FOUND)
			switch name {
			case "offset":
				h.write(t, c, id, 1, b, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)
			case "length":
				h.write(t, c, id, 0, b[:2], 0)
			case "digest":
				h.write(t, c, id, 0, []byte("xxxxxx"), 0)
			case "expiry":
				h.now = h.now.Add(time.Minute)
				h.write(t, c, id, 0, b, pb.ErrorCode_ERROR_CODE_NOT_FOUND)
			case "release":
				h.store.ReleaseSession(c.SessionKey)
				h.write(t, c, id, 0, b, pb.ErrorCode_ERROR_CODE_NOT_FOUND)
			case "commit":
				for i := 0; i < 3; i++ {
					h.write(t, c, id, uint64(i*2), b[i*2:i*2+2], 0)
				}
			}
			code := pb.ErrorCode_ERROR_CODE_NOT_FOUND
			if name == "length" || name == "digest" {
				code = pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			}
			if name == "commit" {
				code = 0
			}
			out := &pb.CommitUploadResponse{}
			h.call(t, c, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: id}, out, code)
			if name == "commit" {
				f, err := h.store.Lookup(out.Ref.Token, c)
				if err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(f.Path)
				if err != nil || !bytes.Equal(got, b) || f.Size != 6 || f.MIME != "text/plain" || len(f.SHA256) != 32 || !f.ExpiresAt.Equal(h.now.Add(time.Minute)) {
					t.Fatalf("file = %+v, bytes=%q, error=%v", f, got, err)
				}
				if _, err := h.store.Lookup(out.Ref.Token, other); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
					t.Fatalf("foreign lookup = %v", err)
				}
				h.store.ReleaseSession(c.SessionKey)
				if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
					t.Fatalf("released file: %v", err)
				}
				if _, err := h.store.Lookup(out.Ref.Token, c); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
					t.Fatalf("released lookup = %v", err)
				}
			}
			// 四个新上传可同时预留，证明失败和回收均释放原上传槽。
			for i := 0; i < 4; i++ {
				h.begin(t, c, b, 0)
			}
			h.begin(t, c, b, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
		})
	}
}
func TestQuota(t *testing.T) {
	for _, name := range []string{"session", "core", "concurrent"} {
		t.Run(name, func(t *testing.T) {
			lim := defaults()
			lim.MaxSessionBytes = 6
			lim.MaxCoreBytes = 6
			if name == "session" {
				lim.MaxCoreBytes = 12
			}
			h := newHarness(t, lim)
			b := []byte("abcdef")
			if name == "concurrent" {
				var wg sync.WaitGroup
				var accepted atomic.Int32
				for i := 0; i < 12; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						c := router.NewCaller("p", "g")
						sum := sha256.Sum256(b)
						req := &pb.BeginUploadRequest{Size: 6, Mime: "text/plain", Sha256: sum[:]}
						data, _ := proto.Marshal(req)
						r := h.mux.For(c)(context.Background(), &wire.Inbound{Target: wire.ServiceMediaBeginUpload, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(req), SchemaVersion: 1, Bytes: data}})
						if r.Err == nil {
							accepted.Add(1)
						} else if r.Err.Code != pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED {
							t.Errorf("unexpected error: %v", r.Err)
						}
					}()
				}
				wg.Wait()
				if accepted.Load() != 1 {
					t.Fatalf("accepted = %d", accepted.Load())
				}
				return
			}
			h.begin(t, h.caller, b, 0)
			c := h.caller
			if name == "core" {
				c = router.NewCaller("other", "g")
			}
			h.begin(t, c, b, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
			h.store.ReleaseSession(h.caller.SessionKey)
			h.begin(t, c, b, 0)
			h.store.ReleaseSession(c.SessionKey)
		})
	}
}

func TestReferenceLifetime(t *testing.T) {
	for _, name := range []string{"lookup", "sweep", "release"} {
		t.Run(name, func(t *testing.T) {
			lim := defaults()
			lim.MaxSessionBytes = 6
			lim.MaxCoreBytes = 12
			h := newHarness(t, lim)
			c := h.caller
			b := []byte("abcdef")
			id := h.begin(t, c, b, 0)
			h.now = h.now.Add(30 * time.Second)
			h.write(t, c, id, 0, b, 0)
			out := &pb.CommitUploadResponse{}
			h.call(t, c, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: id}, out, 0)
			if out.Ref.ExpiresAtUnixMs != h.now.Add(time.Minute).UnixMilli() {
				t.Fatal("commit did not start a new TTL")
			}
			sum := sha256.Sum256(b)
			if !bytes.Equal(out.Ref.Sha256, sum[:]) {
				t.Fatal("incorrect reference digest")
			}
			f, err := h.store.Lookup(out.Ref.Token, c)
			if err != nil {
				t.Fatal(err)
			}
			other := router.NewCaller("other", "generation")
			otherID := h.begin(t, other, b, 0)
			defer h.store.ReleaseSession(other.SessionKey)
			h.begin(t, c, b, pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
			h.write(t, c, id, 0, nil, pb.ErrorCode_ERROR_CODE_NOT_FOUND)
			h.now = h.now.Add(30 * time.Second)
			if _, err := h.store.Lookup(out.Ref.Token, c); err != nil {
				t.Fatalf("reference expired at Begin TTL: %v", err)
			}
			switch name {
			case "release":
				h.store.ReleaseSession(c.SessionKey)
				h.store.ReleaseSession(c.SessionKey)
			case "sweep":
				h.now = h.now.Add(30 * time.Second)
				h.store.Sweep()
				h.store.Sweep()
			case "lookup":
				h.now = h.now.Add(30 * time.Second)
			}
			if _, err := h.store.Lookup(out.Ref.Token, c); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
				t.Fatalf("reclaimed reference: %v", err)
			}
			if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
				t.Fatalf("reclaimed file: %v", err)
			}
			// 每次回收后可再次恰好占满 Session，且超出一个字节仍被拒绝。
			h.begin(t, c, b, 0)
			h.begin(t, c, []byte("x"), pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED)
			if name == "release" {
				h.write(t, other, otherID, 0, b, 0)
			}
		})
	}
}

func TestUploadFileFailures(t *testing.T) {
	for _, name := range []string{"write", "commit", "remove"} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, defaults())
			c := h.caller
			b := []byte("abcdef")
			id := h.begin(t, c, b, 0)
			// 文件真实变为非空目录：写入、读取和删除均产生可复现的文件系统错误。
			entries, err := os.ReadDir(h.dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("entries=%v err=%v", entries, err)
			}
			path := filepath.Join(h.dir, entries[0].Name())
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(path, "child"), b, 0600); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "write":
				h.write(t, c, id, 0, b, pb.ErrorCode_ERROR_CODE_INTERNAL)
			case "commit":
				h.call(t, c, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: id}, &pb.CommitUploadResponse{}, pb.ErrorCode_ERROR_CODE_INTERNAL)
			case "remove":
				h.store.ReleaseSession(c.SessionKey)
			}
			h.write(t, c, id, 0, nil, pb.ErrorCode_ERROR_CODE_NOT_FOUND)
			if logs := h.logs.String(); !strings.Contains(logs, "level=WARN") || !strings.Contains(logs, "path=") || !strings.Contains(logs, "err=") {
				t.Fatalf("missing cleanup warning: %s", logs)
			}
			for i := 0; i < 4; i++ {
				h.begin(t, c, b, 0)
			}
		})
	}
}

func TestOpenCleanup(t *testing.T) {
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	dir := t.TempDir()
	path := filepath.Join(dir, "leftover")
	if err := os.WriteFile(path, []byte("leftover media"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := media.Open(log, dir, defaults(), time.Now); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("leftover files: %v, %v", entries, err)
	}
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := media.Open(log, filepath.Join(path, "media"), defaults(), time.Now); err == nil {
		t.Fatal("Open accepted an unusable directory")
	}
}
