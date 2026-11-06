//go:build unix

package supervisor

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/memory"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	"github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const (
	incomingMediaPluginEnv = "ALONABOT_TEST_INCOMING_MEDIA_PLUGIN"
	incomingMediaFileID    = "platform-photo-id"
)

type incomingMediaAdapter struct {
	*memory.Adapter
	data  []byte
	reads atomic.Int64 // 记录跨插件共享对象的实际平台取件次数
}

// ReadMedia 返回内存中的短文件，确认 Core 用约定的首块请求取件。
func (a *incomingMediaAdapter) ReadMedia(_ context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	a.reads.Add(1)
	if fileID != incomingMediaFileID || offset != 0 || length != wire.MaxMediaChunk {
		return nil, 0, false, fmt.Errorf("unexpected platform read: file=%q offset=%d length=%d", fileID, offset, length)
	}
	return bytes.Clone(a.data), uint64(len(a.data)), true, nil
}

type incomingMediaReport struct {
	PluginID, Token string
	Data            []byte
	Chunks          int
	DeniedCode      pb.ErrorCode
	DeniedMessage   string
	GrantedToken    string
}

// incomingMediaPluginSource 独立编译 SDK 夹具，遵循 Core 测试包的依赖方向。
const incomingMediaPluginSource = `package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	sdkmessage "github.com/AlonaBot/AlonaBot/sdk/message"
)

type incomingMediaReport struct {
	PluginID, Token string
	Data            []byte
	Chunks          int
	DeniedCode      pb.ErrorCode
	DeniedMessage   string
	GrantedToken    string
}

type incomingMediaPlugin struct {
	sdk.Base
	id string
}

// Events 让 a/b 分块读取入站图片，c 保持激活但没有事件关联。
func (p *incomingMediaPlugin) Events(r *sdk.EventRegistry) {
	if p.id == "c" {
		return
	}
	var priority int32
	if p.id == "b" {
		// b 先返回，a 的授权必须依靠实例许可，避免被同事件在途关联放行。
		priority = -1
	}
	r.On("incoming-media", "message.created", func(c *sdk.Context) error {
		chain := c.Event().GetEventPayload().GetMessageCreated().GetChain()
		incoming := chain.GetSegments()[0].GetImage().GetSource().GetIncomingRef()
		if incoming.GetReadability() != pb.MediaReadability_MEDIA_READABILITY_READABLE {
			return fmt.Errorf("incoming media is not readable: %v", incoming)
		}
		report := incomingMediaReport{PluginID: p.id, Token: incoming.Token}
		var offset uint64
		for {
			data, next, eof, err := sdk.ReadMedia(c.Context(), incoming.Token, offset, 7)
			if err != nil {
				return err
			}
			if next != offset+uint64(len(data)) {
				return fmt.Errorf("media next offset=%d, want %d", next, offset+uint64(len(data)))
			}
			report.Data = append(report.Data, data...)
			report.Chunks++
			offset = next
			if eof {
				break
			}
		}
		if p.id == "a" {
			_, err := sdk.GrantMedia(c.Context(), incoming.Token, "c", c.Event().Target)
			if err != nil {
				var public *sdk.Error
				if !errors.As(err, &public) {
					return err
				}
				report.DeniedCode, report.DeniedMessage = public.Code, public.Message
			}
			report.GrantedToken, err = sdk.GrantMedia(c.Context(), incoming.Token, "b", c.Event().Target)
			if err != nil {
				return err
			}
			// SDK 的类型化入口保留父调用，将收到的 incoming_ref 原样交回消息服务。
			result, err := sdk.CallProto[*pb.SendResult](c.Context(), "` + wire.ServiceSend + `", &pb.SendRequest{Chain: chain})
			if err != nil {
				return err
			}
			if result.GetStatus() != pb.SendStatus_SEND_STATUS_ACCEPTED {
				return fmt.Errorf("media send rejected: %v", result)
			}
		}
		encoded, err := json.Marshal(report)
		if err != nil {
			return err
		}
		_, err = c.Send(sdkmessage.Chain{sdkmessage.Text(string(encoded))})
		return err
	}, sdk.Priority(priority))
}

func main() {
	if err := sdk.Run(&incomingMediaPlugin{id: os.Getenv("` + incomingMediaPluginEnv + `")}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`

// TestIncomingMedia 验证真实插件进程的独立令牌、共享取件、授权与只读回发。
func TestIncomingMedia(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	if err := os.WriteFile(source, []byte(incomingMediaPluginSource), 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "incoming-media")
	if out, err := exec.Command("go", "build", "-race", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build incoming media plugin: %v\n%s", err, out)
	}
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) { testIncomingMedia(t, binary, stdio) })
	}
}

func testIncomingMedia(t *testing.T, binary string, stdio bool) {
	t.Helper()
	ctx := context.Background()
	store, err := state.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	logs := &echoLog{}
	output, err := logging.Open(logging.Options{Format: logging.FormatJSON, Level: slog.LevelDebug, Stderr: logs})
	if err != nil {
		t.Fatal(err)
	}
	setDefaultLogger(t, output.Logger())
	defer output.Close()
	defer func() {
		if t.Failed() {
			t.Logf("Core and media plugin logs:\n%s", logs.String())
		}
	}()
	// 1. 装配真实状态库、媒体目录、路由与消息服务，平台文件由内存 Adapter 提供。
	values := tune.Defaults()
	mediaStore, err := media.Open(t.TempDir(), media.NewLimits(values), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	registry := adapter.NewRegistry()
	content := []byte("\x89PNG\r\n\x1a\nincoming media 世界")
	platform := &incomingMediaAdapter{Adapter: memory.New("memory", "instance", "account"), data: content}
	if err := registry.Add(platform); err != nil {
		t.Fatal(err)
	}
	sender := message.New(registry, policy.New(store), mediaStore.LeaseForSend, values.RequestTimeout)
	events := router.New(router.Options{Concurrency: 1, QueueBytes: values.DispatchQueueBytes, QueueEvents: values.DispatchQueueEvents, Adapters: registry, Send: sender.Send, Prepare: mediaStore.PrepareIncoming})
	events.Start()
	defer func() {
		if err := events.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	if err := platform.Start(ctx, events); err != nil {
		t.Fatal(err)
	}
	defer platform.Stop(ctx)
	mux := router.NewMux()
	mux.Handle(wire.ServiceSend, sender.Send)
	mediaStore.AddReadServices(mux, registry, events, store.AdapterInstanceAllowed)

	// 2. 三个进程独立握手和激活，Mux 与 Router 绑定相同的可信 Caller。
	var running []*Running
	callers := make(map[string]router.Caller)
	defer func() {
		for _, instance := range running {
			id := instance.Instance.PluginID
			events.Detach(id)
			stopped := instance.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if stopped.Err != nil || stopped.Killed {
				t.Errorf("media plugin stop: %+v", stopped)
			}
			mediaStore.ReleaseSession(callers[id].SessionKey)
		}
	}()
	for _, id := range []string{"a", "b", "c"} {
		digest := []byte("incoming media integration fixture")
		if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: id, Source: "test", Entrypoint: binary, BuildDigest: digest}); err != nil {
			t.Fatal(err)
		}
		generation := "generation-" + id
		if err := store.AddGeneration(ctx, id, generation, "{}", 1); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(t.TempDir(), id)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		caller := router.NewCaller(id, generation)
		callers[id] = caller
		spec := Spec{PluginID: id, Role: state.RolePlugin, GenerationID: generation, PluginDir: dir, Executable: binary, BuildDigest: digest, NewCandidate: true, Stdio: stdio,
			ExtraEnv:  map[string]string{incomingMediaPluginEnv: id, "GORACE": "atexit_sleep_ms=0"},
			Bootstrap: &pb.Bootstrap{GenerationId: generation, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}},
			Limits:    host.NewLimits(values), Stop: host.NewStopTiming(values), Log: output.Plugin(id, generation), Handler: mux.For(caller),
		}
		instance, err := Launch(ctx, store, spec)
		if err != nil {
			t.Fatalf("launch %s: %v", id, err)
		}
		running = append(running, instance)
		events.Attach(router.FromInstance(instance.Instance), caller)
	}
	// a 自身的入站 token 与原目标回复在空许可集合下仍有效；b 沿用缺省全部实例。
	if err := store.SetAdapterInstanceAccess(ctx, "a", false, nil); err != nil {
		t.Fatal(err)
	}
	// c 被收紧到另一个实例，拒绝接收本实例入站对象的读授权。
	if err := store.SetAdapterInstanceAccess(ctx, "c", false, []string{"other-instance"}); err != nil {
		t.Fatal(err)
	}

	// 3. SendFunc 在实际发送租约内按 Adapter Session 回读文件，并保存原始发送快照。
	type sentMedia struct {
		target *pb.Target
		chain  *pb.MessageChain
		data   []byte
		err    error
	}
	sends := make(chan sentMedia, 3) // 两条插件读取报告和一条图片回发
	platform.SetSendFunc(func(target *pb.Target, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
		sent := sentMedia{target: proto.CloneOf(target), chain: proto.CloneOf(chain)}
		if image := chain.GetSegments()[0].GetImage(); image != nil {
			token := image.GetSource().GetMediaRef()
			file, err := mediaStore.Lookup(token, router.Caller{SessionKey: platform.MediaSessionKey()})
			if err == nil {
				sent.data, err = os.ReadFile(file.Path)
			}
			sent.err = err
			if _, err := mediaStore.Lookup(token, callers["a"]); err == nil {
				sent.err = errors.New("adapter media token is readable by plugin a")
			}
		}
		sends <- sent
		if sent.err != nil {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: sent.err.Error()}
		}
		return &pb.SendResult{Status: pb.SendStatus_SEND_STATUS_ACCEPTED, Target: target}, nil
	})
	objectID := mediaStore.RegisterIncoming(platform.InstanceID(), incomingMediaFileID, proto.Uint64(uint64(len(content))), proto.String("image/png"))
	if got := platform.reads.Load(); got != 0 {
		t.Fatalf("platform reads before dispatch=%d", got)
	}
	target := &pb.Target{AdapterId: "memory", AdapterInstanceId: "instance", AccountId: "account", ChatId: "chat", ThreadId: proto.String("topic")}
	original := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{Token: objectID}}}}}}}}
	event := &pb.Event{EventId: "incoming-media-event", EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, Target: target,
		EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: original}}},
	}
	if err := platform.Inject(ctx, event); err != nil {
		t.Fatal(err)
	}

	// 4. 按真实发送收集结果，核对独立读取、Grant 决策、只读租约及一次取件。
	reports := make(map[string]incomingMediaReport)
	var adapterToken string
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for range 3 {
		select {
		case sent := <-sends:
			if sent.err != nil || !proto.Equal(sent.target, target) {
				t.Fatalf("send target=%v error=%v", sent.target, sent.err)
			}
			if text := sent.chain.GetSegments()[0].GetText(); text != nil {
				var report incomingMediaReport
				if err := json.Unmarshal([]byte(text.Text), &report); err != nil {
					t.Fatal(err)
				}
				if _, duplicate := reports[report.PluginID]; duplicate {
					t.Fatalf("duplicate media report: %+v", report)
				}
				reports[report.PluginID] = report
			} else {
				adapterToken = sent.chain.GetSegments()[0].GetImage().GetSource().GetMediaRef()
				if adapterToken == "" || !bytes.Equal(sent.data, content) {
					t.Fatalf("adapter media token=%q data=%q", adapterToken, sent.data)
				}
			}
		case <-deadline.C:
			t.Fatalf("media replies missing: reports=%+v adapter_token=%q", reports, adapterToken)
		}
	}
	for _, id := range []string{"a", "b"} {
		report := reports[id]
		decoded, err := hex.DecodeString(report.Token)
		if err != nil || len(decoded) != 16 || report.Token == objectID || !bytes.Equal(report.Data, content) || report.Chunks != (len(content)+6)/7 {
			t.Fatalf("invalid %s media report: %+v, token error=%v", id, report, err)
		}
	}
	a, b := reports["a"], reports["b"]
	if a.Token == b.Token || a.DeniedCode != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || a.DeniedMessage == "" || a.GrantedToken == "" || a.GrantedToken == a.Token || a.GrantedToken == b.Token {
		t.Fatalf("independent tokens or grant decisions failed: a=%+v b=%+v", a, b)
	}
	if adapterToken == "" || adapterToken == a.Token || adapterToken == b.Token || adapterToken == a.GrantedToken {
		t.Fatalf("adapter token is not independent: %q", adapterToken)
	}
	if _, err := mediaStore.Lookup(adapterToken, router.Caller{SessionKey: platform.MediaSessionKey()}); err == nil {
		t.Fatal("adapter token survived send lease release")
	}
	if got := platform.reads.Load(); got != 1 {
		t.Fatalf("platform reads=%d, want 1", got)
	}
	if got := original.Segments[0].GetImage().GetSource().GetIncomingRef().Token; got != objectID {
		t.Fatalf("dispatch mutated original media token: %q", got)
	}
}
