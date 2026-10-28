package message_test

import (
	"context"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestSendPermissions 验证原目标回复、显式发送、后台发送和动态实例许可（V-07）。
func TestSendPermissions(t *testing.T) {
	ctx := context.Background()
	store, err := state.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	const pluginID = "sender"
	if err := store.AddInstall(ctx, state.Install{PluginID: pluginID, Source: "dev-local", Entrypoint: "unused", BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	registry := adapter.NewRegistry()
	platforms := []*memory.Adapter{memory.New("memory", "a", "account-a"), memory.New("memory", "b", "account-b")}
	for _, platform := range platforms {
		if err := registry.Add(platform); err != nil {
			t.Fatal(err)
		}
	}
	service := message.New(registry, policy.New(store))
	original := &pb.Target{AdapterId: "memory", AdapterInstanceId: "a", AccountId: "account-a", ChatId: "original", ThreadId: proto.String("topic")}
	otherChat := proto.CloneOf(original)
	otherChat.ChatId = "other"
	otherThread := proto.CloneOf(original)
	otherThread.ThreadId = proto.String("other-topic")
	otherInstance := &pb.Target{AdapterId: "memory", AdapterInstanceId: "b", AccountId: "account-b", ChatId: "other"}
	missing := proto.CloneOf(original)
	missing.AdapterInstanceId = "missing"
	wrongAccount := proto.CloneOf(original)
	wrongAccount.AccountId = "wrong"
	wrongAdapter := proto.CloneOf(original)
	wrongAdapter.AdapterId = "wrong"
	invocation := &router.Invocation{PluginID: pluginID, DefaultTarget: original}
	foreign := &router.Invocation{PluginID: "other-plugin", DefaultTarget: original}
	chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "hello"}}}}}
	tests := []struct {
		name          string
		inv           *router.Invocation
		to            *pb.Target
		grant, revoke string
		code          pb.ErrorCode
	}{
		{name: "reply_without_permission", inv: invocation},
		{name: "explicit_original_without_permission", inv: invocation, to: proto.CloneOf(original)},
		{name: "foreign_plugin_has_no_reply_exemption", inv: foreign, to: original, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "other_target_denied", inv: invocation, to: otherChat, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "thread_change_needs_permission", inv: invocation, to: otherThread, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "background_denied", to: original, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "other_target_granted", inv: invocation, to: otherChat, grant: "a"},
		{name: "revoke_affects_next_send", inv: invocation, to: otherChat, revoke: "a", code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "background_granted", to: original, grant: "a"},
		{name: "same_instance_other_chat", to: otherChat},
		{name: "same_instance_other_thread", inv: invocation, to: otherThread},
		{name: "instance_a_does_not_grant_b", to: otherInstance, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "account_binding_still_checked", to: wrongAccount, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{name: "adapter_binding_still_checked", to: wrongAdapter, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{name: "permission_before_instance_lookup", to: missing, code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED},
		{name: "granted_missing_instance", to: missing, grant: "missing", code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE},
		{name: "no_default_or_to", code: pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.grant != "" {
				if err := store.GrantAdapterInstance(ctx, pluginID, test.grant); err != nil {
					t.Fatal(err)
				}
			}
			if test.revoke != "" {
				if err := store.RevokeAdapterInstance(ctx, pluginID, test.revoke); err != nil {
					t.Fatal(err)
				}
			}
			before := len(platforms[0].Sent()) + len(platforms[1].Sent())
			request := &pb.SendRequest{To: test.to, Chain: chain}
			encoded, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			in := &wire.Inbound{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}}
			if test.inv != nil {
				in.ParentAttach = test.inv
			}
			response := service.Send(ctx, router.NewCaller(pluginID, "generation"), in)
			if response.Err.GetCode() != test.code {
				t.Fatalf("error = %v; want %v", response.Err, test.code)
			}
			result := &pb.SendResult{}
			if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil {
				t.Fatal(err)
			}
			target := test.to
			if target == nil && test.inv != nil {
				target = test.inv.DefaultTarget
			}
			if !proto.Equal(result.Target, target) {
				t.Fatalf("result target = %v; want %v", result.Target, target)
			}
			after := len(platforms[0].Sent()) + len(platforms[1].Sent())
			if test.code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
				if result.Status != pb.SendStatus_SEND_STATUS_REJECTED || after != before {
					t.Fatalf("denied send: result=%v, sends=%d->%d", result, before, after)
				}
				return
			}
			if result.Status != pb.SendStatus_SEND_STATUS_ACCEPTED || after != before+1 {
				t.Fatalf("accepted send: result=%v, sends=%d->%d", result, before, after)
			}
			sent := platforms[0].Sent()
			last := sent[len(sent)-1]
			if !proto.Equal(last.Target, target) || !proto.Equal(last.Chain, chain) {
				t.Fatalf("adapter received %v; want target=%v chain=%v", last, target, chain)
			}
		})
	}
}
