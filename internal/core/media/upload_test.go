package media_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

type uploadHarness struct {
	dir    string
	store  *media.Store
	mux    *router.Mux
	caller router.Caller
	now    time.Time
}

func newHarness(t *testing.T, lim media.Limits) *uploadHarness {
	t.Helper()
	h := &uploadHarness{mux: router.NewMux(), caller: router.NewCaller("plugin", "generation"), now: time.Unix(1000, 0), dir: t.TempDir()}
	var err error
	h.store, err = media.Open(h.dir, lim, func() time.Time { return h.now })
	if err != nil {
		t.Fatal(err)
	}
	h.store.AddServices(h.mux)
	t.Cleanup(func() { h.store.ReleaseSession(h.caller.SessionKey) })
	return h
}
func (h *uploadHarness) call(t *testing.T, c router.Caller, target string, req proto.Message, out proto.Message) {
	t.Helper()
	b, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	r := h.mux.For(c)(context.Background(), &wire.Inbound{Target: target, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(req), SchemaVersion: 1, Bytes: b}})
	if r.Err != nil {
		t.Fatalf("%s: %v", target, r.Err)
	}
	if err := proto.Unmarshal(r.Payload.Bytes, out); err != nil {
		t.Fatal(err)
	}
}
func (h *uploadHarness) begin(t *testing.T, c router.Caller, b []byte) string {
	t.Helper()
	sum := sha256.Sum256(b)
	out := &pb.BeginUploadResponse{}
	h.call(t, c, wire.ServiceMediaBeginUpload, &pb.BeginUploadRequest{Size: uint64(len(b)), Mime: "text/plain", Sha256: sum[:]}, out)
	return out.UploadId
}
func (h *uploadHarness) write(t *testing.T, c router.Caller, id string, offset uint64, b []byte) {
	t.Helper()
	out := &pb.WriteChunkResponse{}
	h.call(t, c, wire.ServiceMediaWriteChunk, &pb.WriteChunkRequest{UploadId: id, Offset: offset, Data: b}, out)
	if out.NextOffset != offset+uint64(len(b)) {
		t.Fatal("incorrect next offset")
	}
}
func defaults() media.Limits {
	return media.NewLimits(tune.Defaults())
}
func TestUpload(t *testing.T) {
	h := newHarness(t, defaults())
	c := h.caller
	b := []byte("abcdef")
	id := h.begin(t, c, b)
	for i := 0; i < 3; i++ {
		h.write(t, c, id, uint64(i*2), b[i*2:i*2+2])
	}
	out := &pb.CommitUploadResponse{}
	h.call(t, c, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: id}, out)
	f, err := h.store.Lookup(out.Ref.Token, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(f.Path)
	if err != nil || !bytes.Equal(got, b) || f.Size != 6 || f.MIME != "text/plain" || len(f.SHA256) != 32 || !f.ExpiresAt.Equal(h.now.Add(defaults().TTL)) {
		t.Fatalf("file = %+v, bytes=%q, error=%v", f, got, err)
	}
	h.store.ReleaseSession(c.SessionKey)
	if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
		t.Fatalf("released file: %v", err)
	}
}

func TestReferenceLifetime(t *testing.T) {
	for _, name := range []string{"lookup", "sweep", "release"} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, defaults())
			c := h.caller
			b := []byte("abcdef")
			id := h.begin(t, c, b)
			h.now = h.now.Add(defaults().TTL / 2)
			h.write(t, c, id, 0, b)
			out := &pb.CommitUploadResponse{}
			h.call(t, c, wire.ServiceMediaCommitUpload, &pb.CommitUploadRequest{UploadId: id}, out)
			if out.Ref.ExpiresAtUnixMs != h.now.Add(defaults().TTL).UnixMilli() {
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
			otherID := h.begin(t, other, b)
			defer h.store.ReleaseSession(other.SessionKey)
			h.now = h.now.Add(defaults().TTL / 2)
			if _, err := h.store.Lookup(out.Ref.Token, c); err != nil {
				t.Fatalf("reference expired at Begin TTL: %v", err)
			}
			switch name {
			case "release":
				h.store.ReleaseSession(c.SessionKey)
			case "sweep":
				h.now = h.now.Add(defaults().TTL / 2)
				h.store.Sweep()
			case "lookup":
				h.now = h.now.Add(defaults().TTL / 2)
			}
			if _, err := h.store.Lookup(out.Ref.Token, c); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
				t.Fatalf("reclaimed reference: %v", err)
			}
			if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
				t.Fatalf("reclaimed file: %v", err)
			}
			h.begin(t, c, b)
			if name == "release" {
				h.write(t, other, otherID, 0, b)
			}
		})
	}
}

func TestOpenCleanup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leftover")
	if err := os.WriteFile(path, []byte("leftover media"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := media.Open(dir, defaults(), time.Now); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("leftover files: %v, %v", entries, err)
	}
}
