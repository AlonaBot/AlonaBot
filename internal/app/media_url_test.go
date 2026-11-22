//go:build unix

package app

import (
	"io"
	"net/http"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestMediaURLRouting 通过真实 Core 启动验证媒体签发服务与公开取件入口均已装配。
func TestMediaURLRouting(t *testing.T) {
	h := newWebUIHarness(t)
	h.start(nil)
	request := &pb.MediaURLRequest{Token: "unbound-media-probe"}
	raw, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	caller := router.NewCaller("media-route-probe", "test")
	response := h.core.assembly.mux.For(caller)(t.Context(), &wire.Inbound{
		Target: wire.ServiceMediaURL,
		Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF,
			SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: raw},
	})
	if response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("unbound media caller reached the wrong service: %v", response.Err)
	}
	result, err := h.client.Get("http://" + h.core.http.Addr + "/media/unbound-media-probe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(io.Discard, result.Body); err != nil {
		t.Error(err)
	}
	if err := result.Body.Close(); err != nil {
		t.Error(err)
	}
	if result.StatusCode != http.StatusNotFound || result.Header.Get("Content-Security-Policy") != "sandbox; default-src 'none'" || result.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("media endpoint reached the wrong handler: status=%d headers=%v", result.StatusCode, result.Header)
	}
}
