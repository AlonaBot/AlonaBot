//go:build unix

package remote_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	coremessage "github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

type deleteHarness struct {
	handler  wire.Handler
	process  *remote.Process
	running  *supervisor.Running
	registry *adapter.Registry
	audit    string
}

func startDeleteHarness(t *testing.T, stdio bool) *deleteHarness {
	t.Helper()
	ctx := context.Background()
	store, spec, _ := setup(t)
	spec.Stdio = stdio
	registry := adapter.NewRegistry()
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: &eventSink{}, RegisterIncoming: openMediaStore(t).RegisterIncoming})
	service := coremessage.New(registry, policy.New(store), nil, nil, time.Second)
	mux := router.NewMux()
	p.Services(mux)
	mux.Handle(wire.ServiceDelete, service.Delete)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Unbind()
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	})
	if err := p.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	audit := filepath.Join(t.TempDir(), "delete.jsonl")
	if err := os.WriteFile(audit, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, instance := range []struct {
		id        string
		supported bool
	}{{"delete-a", true}, {"delete-b", true}, {"plain", false}} {
		config, err := json.Marshal(map[string]any{"account": instance.id, "delete": instance.supported, "delete_audit": audit})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.StartInstance(ctx, instance.id, config); err != nil {
			t.Fatal(err)
		}
	}
	registry.SetDefaults(map[string]string{"loop": "delete-a"})
	return &deleteHarness{handler: mux.For(caller), process: p, running: running, registry: registry, audit: audit}
}

func (h *deleteHarness) call(t *testing.T, ctx context.Context, instance string, ids []string) (*pb.DeleteResult, *pb.ErrorInfo) {
	t.Helper()
	request := &pb.DeleteRequest{To: &pb.Target{AdapterId: "loop", AdapterInstanceId: instance, ChatId: "chat", ThreadId: proto.String("topic")}, PlatformMessageIds: ids}
	data, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	out := h.handler(ctx, &wire.Inbound{Target: wire.ServiceDelete, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: data}})
	if out.Payload.SchemaID == "" {
		return nil, out.Err
	}
	result := &pb.DeleteResult{}
	if err := proto.Unmarshal(out.Payload.Bytes, result); err != nil {
		t.Fatal(err)
	}
	return result, out.Err
}

func (h *deleteHarness) calls(t *testing.T) map[string]int {
	t.Helper()
	data, err := os.ReadFile(h.audit)
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var call struct{ Instance string }
		if err := json.Unmarshal([]byte(line), &call); err != nil {
			t.Fatal(err)
		}
		counts[call.Instance]++
	}
	return counts
}

// TestDeleteMessages 经真实回环 Adapter 进程验证有序部分回执与能力拒绝。
func TestDeleteMessages(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		name := "tcp"
		if stdio {
			name = "stdio"
		}
		t.Run(name, func(t *testing.T) {
			h := startDeleteHarness(t, stdio)
			ids := []string{"accepted", "reject", "not-attempted"}
			result, safe := h.call(t, context.Background(), "delete-a", ids)
			if safe.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || safe.GetSafeMessage() != "message deletion rejected" {
				t.Fatalf("safe error = %v", safe)
			}
			if result == nil || len(result.Items) != len(ids) || result.Target.AdapterInstanceId != "delete-a" || result.Target.AccountId != "delete-a" || result.Target.GetThreadId() != "topic" {
				t.Fatalf("delete result = %v", result)
			}
			statuses := []pb.PartStatus{pb.PartStatus_PART_STATUS_ACCEPTED, pb.PartStatus_PART_STATUS_REJECTED, pb.PartStatus_PART_STATUS_NOT_ATTEMPTED}
			for i, item := range result.Items {
				if item.PlatformMessageId != ids[i] || item.Status != statuses[i] {
					t.Errorf("item %d = %v", i, item)
				}
			}
			if !proto.Equal(result.Items[1].Error, safe) {
				t.Errorf("item error = %v, top-level error = %v", result.Items[1].Error, safe)
			}
			if result, safe := h.call(t, context.Background(), "plain", []string{"m"}); result != nil || safe.GetCode() != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
				t.Fatalf("undeclared capability: result=%v error=%v", result, safe)
			}
			counts := h.calls(t)
			if counts["delete-a"] != 1 || counts["plain"] != 0 {
				t.Fatalf("adapter.delete calls = %v", counts)
			}
		})
	}
}
