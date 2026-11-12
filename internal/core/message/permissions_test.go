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
	"github.com/AlonaBot/AlonaBot/internal/tune"
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
	if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: pluginID, Source: "dev-local", Entrypoint: "unused", BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	registry := adapter.NewRegistry()
	platforms := []*memory.Adapter{memory.New("memory", "a", "account-a"), memory.New("memory", "b", "account-b")}
	for _, platform := range platforms {
		if err := registry.Add(platform); err != nil {
			t.Fatal(err)
		}
	}
	service := message.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout)
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
	settings := []struct {
		name             string
		set, all         bool
		instances        []string
		allowsA, allowsB bool
	}{
		{name: "default", allowsA: true, allowsB: true},
		{name: "restricted_to_a", set: true, instances: []string{"a"}, allowsA: true},
		{name: "restored_all", set: true, all: true, allowsA: true, allowsB: true},
		{name: "restricted_empty", set: true},
	}
	for _, setting := range settings {
		t.Run(setting.name, func(t *testing.T) {
			if setting.set {
				if err := store.SetAdapterInstanceAccess(ctx, pluginID, setting.all, setting.instances); err != nil {
					t.Fatal(err)
				}
			}
			codeA, codeB := pb.ErrorCode_ERROR_CODE_UNSPECIFIED, pb.ErrorCode_ERROR_CODE_UNSPECIFIED
			accountCode, adapterCode := pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			missingCode := pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED
			if !setting.allowsA {
				codeA = pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED
				accountCode, adapterCode = codeA, codeA
			}
			if !setting.allowsB {
				codeB = pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED
			}
			if !setting.set || setting.all {
				missingCode = pb.ErrorCode_ERROR_CODE_UNAVAILABLE
			}
			tests := []struct {
				name string
				inv  *router.Invocation
				to   *pb.Target
				code pb.ErrorCode
			}{
				{name: "reply", inv: invocation},
				{name: "explicit_original_reply", inv: invocation, to: proto.CloneOf(original)},
				{name: "foreign_plugin_has_no_reply_exemption", inv: foreign, to: original, code: codeA},
				{name: "other_target", inv: invocation, to: otherChat, code: codeA},
				{name: "other_thread", inv: invocation, to: otherThread, code: codeA},
				{name: "background_a", to: original, code: codeA},
				{name: "background_b", to: otherInstance, code: codeB},
				{name: "same_instance_other_chat", to: otherChat, code: codeA},
				{name: "account_binding", to: wrongAccount, code: accountCode},
				{name: "adapter_binding", to: wrongAdapter, code: adapterCode},
				{name: "permission_before_instance_lookup", to: missing, code: missingCode},
				{name: "no_default_or_to", code: pb.ErrorCode_ERROR_CODE_NO_DEFAULT_TARGET},
			}
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
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
					platform, ok := registry.Get(target.GetAdapterInstanceId())
					if !ok {
						t.Fatal("accepted target has no adapter")
					}
					sent := platform.(*memory.Adapter).Sent()
					last := sent[len(sent)-1]
					if !proto.Equal(last.Target, target) || !proto.Equal(last.Chain, chain) {
						t.Fatalf("adapter received %v; want target=%v chain=%v", last, target, chain)
					}
				})
			}
		})
	}
}

// TestSendTargetCompletion 验证目标补齐后的原目标豁免、缺省实例、绑定与引用（V-101）。
func TestSendTargetCompletion(t *testing.T) {
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
	if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: pluginID, Source: "dev-local", Entrypoint: "unused", BuildDigest: []byte{1}}); err != nil {
		t.Fatal(err)
	}
	registry := adapter.NewRegistry()
	platforms := []*memory.Adapter{memory.New("memory", "a", "account-a"), memory.New("memory", "b", "account-b")}
	for _, platform := range platforms {
		if err := registry.Add(platform); err != nil {
			t.Fatal(err)
		}
	}
	service := message.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout)
	original := &pb.Target{AdapterId: "memory", AdapterInstanceId: "a", AccountId: "account-a", ChatId: "original", ThreadId: proto.String("topic")}
	invocation := &router.Invocation{PluginID: pluginID, DefaultTarget: original}
	tests := []struct {
		name           string
		to             *pb.Target
		defaults       map[string]string
		restricted     bool
		reference      bool
		instance, chat string
		code           pb.ErrorCode
	}{
		{name: "original_reply_without_permission", to: &pb.Target{AdapterId: "memory", ChatId: "original", ThreadId: proto.String("topic")}, restricted: true, instance: "a", chat: "original"},
		{name: "default_instance", to: &pb.Target{AdapterId: "memory", ChatId: "other"}, defaults: map[string]string{"memory": "b"}, instance: "b", chat: "other"},
		{name: "no_default_instance", to: &pb.Target{AdapterId: "memory", ChatId: "other"}, code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION},
		{name: "explicit_instance", to: &pb.Target{AdapterId: "memory", AdapterInstanceId: "b", ChatId: "other"}, instance: "b", chat: "other"},
		{name: "wrong_account", to: &pb.Target{AdapterId: "memory", AdapterInstanceId: "b", AccountId: "wrong", ChatId: "other"}, code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		{name: "reference", to: &pb.Target{AdapterId: "memory", ChatId: "original", ThreadId: proto.String("topic")}, reference: true, instance: "a", chat: "original"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry.SetDefaults(test.defaults)
			if err := store.SetAdapterInstanceAccess(ctx, pluginID, !test.restricted, nil); err != nil {
				t.Fatal(err)
			}
			chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "hello"}}}}}
			if test.reference {
				chain.Segments = append(chain.Segments, &pb.Segment{Data: &pb.Segment_Reference{Reference: &pb.ReferenceSegment{OriginalPlatformMessageId: "quoted", Target: &pb.Target{AdapterId: "memory", ChatId: "original"}}}})
			}
			beforeA, beforeB := len(platforms[0].Sent()), len(platforms[1].Sent())
			request := &pb.SendRequest{To: test.to, Chain: chain}
			encoded, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			in := &wire.Inbound{ParentAttach: invocation, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}}
			response := service.Send(ctx, router.NewCaller(pluginID, "generation"), in)
			if response.Err.GetCode() != test.code {
				t.Fatalf("error = %v; want %v", response.Err, test.code)
			}
			result := &pb.SendResult{}
			if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil {
				t.Fatal(err)
			}
			afterA, afterB := len(platforms[0].Sent()), len(platforms[1].Sent())
			if test.code != pb.ErrorCode_ERROR_CODE_UNSPECIFIED {
				if result.Status != pb.SendStatus_SEND_STATUS_REJECTED || beforeA != afterA || beforeB != afterB {
					t.Fatalf("rejected send: result=%v, sends=%d/%d -> %d/%d", result, beforeA, beforeB, afterA, afterB)
				}
				if test.code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION && !proto.Equal(result.Target, test.to) {
					t.Fatalf("completion failure target = %v; want %v", result.Target, test.to)
				}
				return
			}
			want := &pb.Target{AdapterId: "memory", AdapterInstanceId: test.instance, AccountId: "account-" + test.instance, ChatId: test.chat, ThreadId: test.to.ThreadId}
			if result.Status != pb.SendStatus_SEND_STATUS_ACCEPTED || !proto.Equal(result.Target, want) {
				t.Fatalf("result = %v; want accepted target %v", result, want)
			}
			index := 0
			if test.instance == "b" {
				index = 1
			}
			if afterA-beforeA != 1-index || afterB-beforeB != index {
				t.Fatalf("sends = %d/%d -> %d/%d; want only %s", beforeA, beforeB, afterA, afterB, test.instance)
			}
			sent := platforms[index].Sent()
			last := sent[len(sent)-1]
			if !proto.Equal(last.Target, want) {
				t.Fatalf("adapter target = %v; want %v", last.Target, want)
			}
			if test.reference {
				quoted := last.Chain.Segments[1].GetReference().Target
				wantReference := &pb.Target{AdapterId: "memory", AdapterInstanceId: "a", AccountId: "account-a", ChatId: "original"}
				if !proto.Equal(quoted, wantReference) {
					t.Fatalf("reference target = %v; want %v", quoted, wantReference)
				}
			}
		})
	}
}
