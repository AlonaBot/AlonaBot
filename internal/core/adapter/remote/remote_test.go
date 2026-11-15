//go:build unix

package remote_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter/remote"
	"github.com/AlonaBot/AlonaBot/internal/core/broker"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/httpproxy"
	"github.com/AlonaBot/AlonaBot/internal/core/media"
	coremessage "github.com/AlonaBot/AlonaBot/internal/core/message"
	"github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/core/webhost"
	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk"
	sdkadapter "github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	optionsapi "github.com/AlonaBot/AlonaBot/sdk/options"
	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/protobuf/proto"
)

const helperEnv = "ALONABOT_LOOP_ADAPTER"
const observerAuditEnv = "ALONABOT_LOOP_OBSERVER_AUDIT"
const streamAuditEnv = "ALONABOT_LOOP_STREAM_AUDIT"

// TestMain 将测试二进制切换为经真实 SDK 运行的回环 Adapter 或插件。
func TestMain(m *testing.M) {
	mode := os.Getenv(helperEnv)
	if mode == "" {
		os.Exit(m.Run())
	}
	var err error
	if mode == "plugin" || mode == "command-plugin" || mode == "stream-plugin" || mode == "names-plugin" {
		err = sdk.Run(newLoopPlugin())
	} else if mode == "plugin-object" {
		err = sdk.Run(newObjectPlugin())
	} else if mode == "options-plugin" {
		err = sdk.Run(newOptionsPlugin())
	} else if mode == "schema-parse" || mode == "schema-resolve" {
		err = invalidSchemaAdapter(mode)
	} else if mode == "http-upgrade" {
		a := &httpUpgradeAdapter{}
		genresources.Bind(a, genresources.Resources{})
		err = sdk.RunAdapter(a)
	} else {
		a := &loopAdapter{}
		genresources.Bind(a, genresources.Resources{})
		err = sdk.RunAdapter(a)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func newLoopPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{})
	if os.Getenv(helperEnv) == "command-plugin" {
		genresources.Bind(p, genresources.Resources{I18n: commandCatalog()})
	}
	registerLoopEvents(p)
	registerLoopCommands(p)
	return p
}

type observedEvent struct {
	ChatType     message.ChatType
	Segments     []string
	VoiceURL     string
	ButtonURLs   []string
	ButtonTokens []string

	Route       string
	Permissions []string
	Version     uint32
	Payload     string
}

type poke struct {
	User string `json:"user"`
}

type loopActionParams struct {
	Value string `json:"value"`
}

type loopActionResult struct {
	Value string `json:"value"`
}

func observe(route string, c *sdk.Context, value *poke) error {
	report := observedEvent{Route: route, ChatType: c.ChatType()}
	if payload, ok := c.Event().Payload.(*message.CreatedPayload); ok {
		for _, segment := range payload.Chain {
			switch value := segment.(type) {
			case message.TextSegment:
				report.Segments = append(report.Segments, "text")
			case message.VoiceSegment:
				report.Segments = append(report.Segments, "voice")
				report.VoiceURL = value.Source.HTTPSURL
			case message.ButtonsSegment:
				report.Segments = append(report.Segments, "buttons")
				for _, row := range value.Rows {
					for _, button := range row {
						report.ButtonURLs = append(report.ButtonURLs, button.URL)
						report.ButtonTokens = append(report.ButtonTokens, button.CoreToken)
					}
				}
			}
		}
	}
	if actor := c.Event().Actor; actor != nil {
		report.Permissions = actor.Permissions
	}
	if value != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		report.Version = c.Event().Payload.(*message.AdapterEventPayload).Version
		report.Payload = string(data)
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(os.Getenv(observerAuditEnv), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return errors.Join(err, file.Close())
}

// registerLoopEvents 记录真实插件收到的权限和扩展载荷，barrier 用于确定完整派发已结束。
func registerLoopEvents(r *sdk.Plugin) {
	if os.Getenv(helperEnv) == "stream-plugin" {
		r.On("stream", sdk.MessageCreated).Handle(func(c *sdk.Context) error {
			parts := []string{"a😀", "b", "c", "d😀"}
			index := 0
			results, err := c.Stream(func(context.Context) (string, error) {
				if index == len(parts) {
					return "", io.EOF
				}
				text := parts[index]
				index++
				return text, nil
			})
			report := streamReport{Results: results, Capabilities: c.Event().Capabilities}
			if err != nil {
				report.Error = err.Error()
			}
			data, encodeError := json.Marshal(report)
			if encodeError != nil {
				return encodeError
			}
			return os.WriteFile(os.Getenv(observerAuditEnv), data, 0600)
		})
		return
	}
	r.On("loop", sdk.MessageCreated).Handle(func(c *sdk.Context) error { return observe("loop", c, nil) })
	r.On("extension", sdk.AdapterEvent[poke]("loop.poke", 1)).Handle(func(c *sdk.Context, value *poke) error { return observe("extension", c, value) })
	r.On("uninstalled", sdk.AdapterEvent[poke]("other.poke", 1)).Handle(func(c *sdk.Context, value *poke) error { return observe("uninstalled", c, value) })
	r.On("barrier", sdk.Scheduled).Handle(func(c *sdk.Context) error { return observe("barrier", c, nil) })
}

// registerLoopCommands 让扩展事件误入命令层时留下可观察记录。
func registerLoopCommands(r *sdk.Plugin) {
	if os.Getenv(helperEnv) == "names-plugin" {
		r.Command("echo").Alias("e").Args("value").Handle(func(c *sdk.Context, value string) error {
			return observe("cmd.echo:"+value, c, nil)
		})
		return
	}
	if os.Getenv(helperEnv) == "command-plugin" {
		handler := func(*sdk.Context) error { return nil }
		r.Command("ping").Description("Ping the bot").Alias("p").Handle(handler)
		r.Command("bare").Handle(handler)
		r.Command("admin").Description("Admin command").Permission(sdk.Admin).Handle(handler)
		r.Group("rss").Command("add").Description("Add a feed").Handle(handler)
		return
	}
	if os.Getenv(helperEnv) == "stream-plugin" {
		return
	}
	r.Raw("raw").Handle(func(c *sdk.Context) error { return observe("raw", c, nil) })
}

type loopAdapter struct{}

const httpPlatformToken = "loop-platform-test-token"

type httpUpgradeAdapter struct{ loopAdapter }

// Descriptor 声明公开平台凭证入口；最小 WebSocket echo 只收发四字节文本帧。
func (a *httpUpgradeAdapter) Descriptor() *sdkadapter.Descriptor {
	d := a.loopAdapter.Descriptor()
	d.HTTP = []sdkadapter.HTTPRoute{{Method: "GET", Path: "/ws", Auth: sdkadapter.AuthPlatformToken, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if r.Header.Get("Authorization") == "" {
			token = r.URL.Query().Get("access_token")
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(httpPlatformToken)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		for name := range r.Header {
			lower := strings.ToLower(name)
			if strings.EqualFold(name, wire.HeaderProxyAuth) || lower == "cookie" || lower == "x-alonabot-csrf" || lower == "forwarded" || strings.HasPrefix(lower, "x-forwarded-") {
				http.Error(w, "identity leaked", http.StatusInternalServerError)
				return
			}
		}
		conn, rw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		// RFC 6455 的握手摘要固定使用 SHA-1，与平台凭证的校验分开。
		accept := sha1.Sum([]byte(r.Header.Get("Sec-WebSocket-Key") + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		if _, err := fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Accept: %s\r\n\r\n", base64.StdEncoding.EncodeToString(accept[:])); err != nil {
			return
		}
		if err := rw.Flush(); err != nil {
			return
		}
		for {
			var frame [10]byte
			if _, err := io.ReadFull(rw, frame[:]); err != nil {
				return
			}
			if frame[0] != 0x81 || frame[1] != 0x84 {
				return
			}
			for i := range frame[6:] {
				frame[6+i] ^= frame[2+i]
			}
			if _, err := rw.Write(append([]byte{0x81, 4}, frame[6:]...)); err != nil {
				return
			}
			if err := rw.Flush(); err != nil {
				return
			}
		}
	})}}
	return d
}

func remoteHTTPUpgrade(t *testing.T, address, target, authorization string) (net.Conn, *bufio.Reader) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nAuthorization: %s\r\nCookie: custom=leak\r\nX-Alonabot-CSRF: forged\r\nForwarded: forged\r\nX-Forwarded-For: forged\r\nX-Forwarded-Test: forged\r\n%s: forged\r\n\r\n", target, authorization, wire.HeaderProxyAuth)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	response, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusSwitchingProtocols || response.Header.Get("Upgrade") != "websocket" || response.Header.Get("Sec-WebSocket-Accept") != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Fatalf("WebSocket handshake failed: status=%s", response.Status)
	}
	return conn, reader
}

// TestRemoteHTTPUpgrade 通过统一 Listener 验证真实 SDK Adapter 的平台认证、WebSocket 往返与停止回收。
func TestRemoteHTTPUpgrade(t *testing.T) {
	store, spec, logs := setup(t)
	spec.ExtraEnv[helperEnv] = "http-upgrade"
	spec.Stop.Drain = 10 * time.Millisecond
	spec.HTTP = httpproxy.New(nil)
	base, err := url.Parse("http://127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(webhost.New(webhost.Options{BaseURL: base, Proxy: spec.HTTP}))
	defer server.Close()
	running, err := supervisor.Launch(t.Context(), store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE) })
	if running.Instance.HTTPListener == nil {
		t.Fatal("SDK did not register HTTP listener")
	}
	conn, reader := remoteHTTPUpgrade(t, server.Listener.Addr().String(), "/plugins/loop-package/ws", "Bearer "+httpPlatformToken)
	if _, err := conn.Write([]byte{0x81, 0x84, 1, 2, 3, 4, 'p' ^ 1, 'i' ^ 2, 'n' ^ 3, 'g' ^ 4}); err != nil {
		t.Fatal(err)
	}
	var frame [6]byte
	if _, err := io.ReadFull(reader, frame[:]); err != nil || string(frame[:]) != "\x81\x04ping" {
		t.Fatalf("WebSocket echo=%q, error=%v", frame, err)
	}
	if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil || result.Killed {
		t.Fatalf("adapter stop=%+v", result)
	}
	if _, err := reader.ReadByte(); err == nil {
		t.Fatal("WebSocket survived adapter stop")
	} else if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
		t.Fatal("WebSocket left open after adapter stop")
	}
	logs.mu.Lock()
	leaked := bytes.Contains(logs.b.Bytes(), []byte(httpPlatformToken))
	logs.mu.Unlock()
	if leaked {
		t.Fatal("platform credential entered logs")
	}
}

func (*loopAdapter) Descriptor() *sdkadapter.Descriptor {
	descriptor := &sdkadapter.Descriptor{AdapterID: "loop", Permissions: []sdkadapter.PermissionDeclaration{
		{Name: "loop.owner", Implies: []string{"loop.admin"}, LabelKey: "permissions.owner"},
		{Name: "loop.admin", Implies: []string{"chat_admin"}, LabelKey: "permissions.admin"},
	}, Events: []sdkadapter.EventDeclaration{sdkadapter.Event[poke]("loop.poke", 1)}}
	if os.Getenv(helperEnv) == "stream-adapter" {
		descriptor.Capabilities = streamCapabilities()
	}
	if os.Getenv(helperEnv) == "actions" {
		descriptor.Actions = []sdkadapter.ActionDeclaration{
			sdkadapter.Action[loopActionParams, loopActionResult]("loop.echo", 1, sdkadapter.ActionScopeChat),
			sdkadapter.Action[loopActionParams, loopActionResult]("loop.account", 1, sdkadapter.ActionScopeAccount),
		}
	}
	switch os.Getenv(helperEnv) {
	case "permission-prefix":
		descriptor.Permissions[0].Name = "other.owner"
	case "implies-user":
		descriptor.Permissions[0].Implies = []string{"user"}
	case "implies-admin":
		descriptor.Permissions[0].Implies = []string{"admin"}
	case "permission-cycle":
		descriptor.Permissions[1].Implies = []string{"loop.owner"}
	case "event-prefix":
		descriptor.Events[0] = sdkadapter.Event[poke]("other.poke", 1)
	}
	return descriptor
}

// invalidSchemaAdapter 直接发送非法 REGISTER，保留 Core 对恶意 wire Schema 的原有验收。
func invalidSchemaAdapter(mode string) error {
	var conn *wire.Conn
	if os.Getenv(wire.EnvTransport) == wire.TransportStdio {
		conn = wire.NewConn(os.Stdin, os.Stdout, os.Stdin)
	} else {
		stream, err := net.Dial("tcp", os.Getenv(wire.EnvCoreAddr))
		if err != nil {
			return err
		}
		conn = wire.NewStreamConn(stream)
	}
	defer conn.Close()
	secret, err := hex.DecodeString(os.Getenv(wire.EnvSecret))
	if err != nil {
		return err
	}
	digest, err := hex.DecodeString(os.Getenv(wire.EnvBuildDigest))
	if err != nil {
		return err
	}
	id := os.Getenv(wire.EnvPluginID)
	hello, err := wire.ControlEnvelope(nil, 0, id, pb.FrameKind_FRAME_KIND_HELLO, &pb.Hello{PluginId: id, Secret: secret, BuildDigest: digest, SupportedMajor: validate.ProtocolMajor, SupportedMinor: validate.ProtocolMinor})
	if err != nil {
		return err
	}
	if err := conn.WriteFrame(hello); err != nil {
		return err
	}
	frame, err := conn.ReadFrame()
	if err != nil {
		return err
	}
	control, err := validate.DecodeControl(frame)
	if err != nil {
		return err
	}
	ack := control.(*pb.HelloAck)
	schema := `{"type":1}`
	if mode == "schema-resolve" {
		schema = `{"$ref":"#/$defs/missing"}`
	}
	registration := &pb.Register{Adapter: &pb.AdapterDescriptor{AdapterId: "loop", Events: []*pb.EventDeclaration{{Name: "loop.poke", Version: 1, SchemaJson: schema}}}, CandidateConfigJson: ack.Bootstrap.CandidateConfigJson, CandidateConfigVersion: ack.Bootstrap.TargetConfigVersion, Migration: &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	frame, err = wire.ControlEnvelope(ack.SessionId, ack.ChosenMinor, id, pb.FrameKind_FRAME_KIND_REGISTER, registration)
	if err != nil {
		return err
	}
	if err := conn.WriteFrame(frame); err != nil {
		return err
	}
	_, err = conn.ReadFrame()
	return err
}

func expectedDescriptor(t *testing.T) *pb.AdapterDescriptor {
	t.Helper()
	schema, err := jsonschema.For[poke](nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	return &pb.AdapterDescriptor{AdapterId: "loop", Permissions: []*pb.PermissionDeclaration{
		{Name: "loop.owner", Implies: []string{"loop.admin"}, LabelKey: "permissions.owner"},
		{Name: "loop.admin", Implies: []string{"chat_admin"}, LabelKey: "permissions.admin"},
	}, Events: []*pb.EventDeclaration{{Name: "loop.poke", Version: 1, SchemaJson: string(raw)}}}
}
func (*loopAdapter) Start(_ context.Context, req sdkadapter.StartRequest, core sdkadapter.Core) (sdkadapter.Instance, error) {
	var config struct {
		Account       string `json:"account"`
		Token         string `json:"token"`
		DisplayName   string `json:"display_name"`
		Username      string `json:"username"`
		Media         bool   `json:"media"`
		CommandsAudit string `json:"commands_audit"`
		Stream        bool   `json:"stream"`
		Delete        bool   `json:"delete"`
		DeleteAudit   string `json:"delete_audit"`
		Segments      bool   `json:"segments"`
	}
	if err := json.Unmarshal(req.ConfigJSON, &config); err != nil {
		return nil, err
	}
	if config.Account == "start-error" {
		return nil, fmt.Errorf("invalid token: %s", config.Token)
	}
	n := 0
	if len(req.Cursor) > 0 {
		var err error
		n, err = strconv.Atoi(string(req.Cursor))
		if err != nil {
			return nil, err
		}
	}
	instance := &loopInstance{id: req.InstanceID, account: config.Account, displayName: config.DisplayName, username: config.Username, core: core, sequence: n, initial: string(req.Cursor), token: config.Token, commandsAudit: config.CommandsAudit, stream: config.Stream, segments: config.Segments}
	instance.deleteSupported, instance.deleteAudit = config.Delete, config.DeleteAudit
	if config.Media {
		return &mediaLoopInstance{loopInstance: instance}, nil
	}
	return instance, nil
}

type mediaLoopInstance struct{ *loopInstance }

func (i *mediaLoopInstance) Capabilities() []message.Capability {
	return append(i.loopInstance.Capabilities(), message.Capability{Name: "media.read", Version: 2})
}

func (i *mediaLoopInstance) ReadPlatformFile(_ context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	i.mu.Lock()
	i.mediaReads++
	i.mu.Unlock()
	if fileID != "platform-file" {
		return nil, 0, false, &sdk.Error{Code: errcode.NotFound, Message: "platform file not found"}
	}
	const content = "image bytes"
	if offset > uint64(len(content)) {
		return nil, 0, false, &sdk.Error{Code: errcode.InvalidArgument, Message: "media offset exceeds file size"}
	}
	end := min(offset+uint64(length), uint64(len(content)))
	return []byte(content[offset:end]), end, end == uint64(len(content)), nil
}

type loopInstance struct {
	commandsAudit         string
	stream                bool
	token                 string
	stopFailed            bool
	id, account, initial  string
	displayName, username string
	core                  sdkadapter.Core
	mu                    sync.Mutex
	sequence              int
	mediaReads            int
	deleteSupported       bool
	deleteAudit           string
	actionCalls           int
	segments              bool
}

func (i *loopInstance) Account() sdkadapter.Account {
	return sdkadapter.Account{ID: i.account, DisplayName: i.displayName, Username: i.username}
}
func (i *loopInstance) Capabilities() []message.Capability {
	if i.stream {
		return streamCapabilities()
	}
	capabilities := []message.Capability{{Name: "message.text", Version: 1, Supported: true}}
	if i.commandsAudit != "" {
		capabilities = append(capabilities, message.Capability{Name: "commands.sync", Version: 1, Supported: true})
	}
	if i.deleteSupported {
		capabilities = append(capabilities, message.Capability{Name: "message.delete", Version: 1, Supported: true})
	}
	if i.segments {
		capabilities = append(capabilities, message.Capability{Name: "message.voice", Version: 1, Supported: true})
	}
	return capabilities
}

// Delete 记录真实 RPC 到达并模拟顺序平台操作，失败或未知后保留未尝试项。
func (i *loopInstance) Delete(ctx context.Context, target string, ids []string) (*message.DeleteResult, *errcode.Error) {
	data, err := json.Marshal(struct {
		Instance string
		IDs      []string
	}{Instance: i.id, IDs: ids})
	if err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	file, err := os.OpenFile(i.deleteAudit, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	_, writeError := file.Write(append(data, '\n'))
	if err := errors.Join(writeError, file.Close()); err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	result := &message.DeleteResult{Session: target, AdapterInstanceID: i.id}
	var safe *errcode.Error
	for _, id := range ids {
		item := message.DeleteItemResult{PlatformMessageID: id, Status: message.PartAccepted}
		if safe != nil {
			item.Status = message.PartNotAttempted
		} else {
			switch id {
			case "reject":
				item.Status, safe = message.PartRejected, errcode.PermissionDenied.New("message deletion rejected")
			case "unknown":
				item.Status, safe = message.PartUnknown, errcode.Unavailable.New("message deletion outcome unknown")
			case "wait":
				<-ctx.Done()
				item.Status, safe = message.PartUnknown, errcode.DeadlineExceeded.New("message deletion outcome unknown")
			}
			item.Error = safe
		}
		result.Items = append(result.Items, item)
	}
	return result, safe
}
func (i *loopInstance) Stop(context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.account == "stop-error" && !i.stopFailed {
		i.stopFailed = true
		return errors.New("platform stop failed: " + i.token)
	}
	return nil
}
func (i *loopInstance) Send(ctx context.Context, target string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	text := chain.Text()
	if i.stream && text != "stream-trigger" {
		i.sequence++
		return i.streamReceipt("send", target, strconv.Itoa(i.sequence), text)
	}
	if text == "media-source" {
		return i.inspectMediaSource(ctx, chain)
	}
	if text == "fail" {
		if err := i.core.ReportFailure(ctx, errcode.Unauthenticated.New("account rejected")); err != nil {
			return nil, errcode.Internal.New(err.Error())
		}
		if _, err := i.core.Deliver(ctx, &message.Event{}, nil); err == nil {
			return nil, errcode.Internal.New("failed client accepted event")
		}
		return &message.SendResult{Status: message.SendAccepted}, nil
	}
	if text == "cursor" {
		return &message.SendResult{Status: message.SendAccepted, PlatformMessageIDs: []string{i.initial}}, nil
	}
	if text == "send-count" {
		return &message.SendResult{Status: message.SendAccepted, PlatformMessageIDs: []string{strconv.Itoa(i.sequence)}}, nil
	}
	if text == "media-read-count" {
		return &message.SendResult{Status: message.SendAccepted, PlatformMessageIDs: []string{strconv.Itoa(i.mediaReads)}}, nil
	}
	if text == "action-count" {
		return &message.SendResult{Status: message.SendAccepted, PlatformMessageIDs: []string{strconv.Itoa(i.actionCalls)}}, nil
	}
	i.sequence++
	id := i.id
	if text == "forge" {
		id = "other-instance"
	}
	ev := &message.Event{ID: "adapter-forged-id", Type: "message.created", AdapterID: "loop", AdapterInstanceID: id, AccountID: i.account, Session: target, Payload: &message.CreatedPayload{Chain: chain, PlatformMessageID: strconv.Itoa(i.sequence)}}
	switch text {
	case "chat-directory-group":
		ev.ChatType, ev.ChatTitle = message.ChatGroup, "Garden Group"
	case "chat-directory-private":
		ev.ChatType, ev.ChatTitle = message.ChatPrivate, "Garden Person"
	case "segments":
		ev.ChatType = message.ChatGroup
	case "bot.online", "bot.offline", "scheduled":
		ev.Type = text
	case "mismatched-payload":
		ev.Type = "message.edited"
	case "media-created":
		ev.Message = &message.PlatformMessage{Chain: chain}
	case "media-edited":
		ev.Message = &message.PlatformMessage{Chain: chain}
		ev.Type = "message.edited"
		ev.Payload = &message.EditedPayload{Chain: chain, OldChain: chain}
	}
	if text == "permissions" || strings.HasPrefix(text, "extension") || text == "actor-forge" || text == "target-forge" {
		ev.Actor = &message.Actor{ExternalUserID: "external-7", AdapterInstanceID: i.id, Permissions: []string{"loop.owner", "loop.unknown", "user", "admin", "chat_admin", "loop.owner"}}
	}
	if strings.HasPrefix(text, "extension") {
		ev.Type = "loop.poke"
		payload := &message.AdapterEventPayload{Version: 1, PayloadJSON: `{"user":"external-7"}`}
		ev.Payload = payload
		switch text {
		case "extension-unknown":
			ev.Type = "loop.unknown"
		case "extension-other":
			ev.Type = "other.poke"
		case "extension-version":
			payload.Version = 2
		case "extension-schema":
			payload.PayloadJSON = `{"user":7}`
		case "extension-array":
			payload.PayloadJSON = `[]`
		case "extension-json":
			payload.PayloadJSON = `{`
		case "extension-extra":
			payload.PayloadJSON = `{"user":"external-7","extra":true}`
		case "extension-branch":
			ev.Payload = &message.CreatedPayload{Chain: chain}
		}
	}
	if text == "actor-forge" {
		ev.Actor.AdapterInstanceID = "other"
	}
	if text == "target-forge" {
		ev.AccountID = "other"
	}
	accepted, err := i.core.Deliver(ctx, ev, []byte(strconv.Itoa(i.sequence)))
	if err != nil {
		return nil, errcode.InvalidArgument.New(err.Error())
	}
	if !accepted {
		return nil, errcode.ResourceExhausted.New("event queue full")
	}
	return &message.SendResult{Status: message.SendAccepted}, nil
}

type eventSink struct {
	mu     sync.Mutex
	events []*pb.Event
}

func (s *eventSink) Deliver(_ context.Context, ev *pb.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, proto.Clone(ev).(*pb.Event))
	return nil
}
func (s *eventSink) snapshot() []*pb.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*pb.Event(nil), s.events...)
}

type processLog struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *processLog) Line(_, _ string, b []byte) { l.mu.Lock(); defer l.mu.Unlock(); l.b.Write(b) }
func (*processLog) SetSession([]byte)            {}

func (l *processLog) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(b)
}

// SetDefault 会同时修改 log 包的输出与标志，清理时一并恢复。
func setDefaultLogger(t *testing.T, logger *slog.Logger) {
	t.Helper()
	previous, writer, flags := slog.Default(), log.Writer(), log.Flags()
	slog.SetDefault(logger)
	t.Cleanup(func() {
		slog.SetDefault(previous)
		log.SetOutput(writer)
		log.SetFlags(flags)
	})
}

func setup(t *testing.T) (*state.Store, supervisor.Spec, *processLog) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	store, err := state.Open(ctx, filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	digest := bytes.Repeat([]byte{1}, 32)
	if err := store.AddInstall(ctx, state.Install{PluginID: "loop-package", Role: state.RoleAdapter, AdapterID: "loop", BuildDigest: digest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, "loop-package", "gen", "{}", 1); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugin")
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		t.Fatal(err)
	}
	values := tune.Defaults()
	limits := host.NewLimits(values)
	limits.Handshake = 3 * time.Second
	limits.OnReady = 3 * time.Second
	logs := &processLog{}
	setDefaultLogger(t, slog.New(slog.NewTextHandler(logs, nil)))
	return store, supervisor.Spec{PluginID: "loop-package", Role: state.RoleAdapter, AdapterID: "loop", PluginDir: pluginDir, Executable: os.Args[0], BuildDigest: digest, GenerationID: "gen", Bootstrap: &pb.Bootstrap{GenerationId: "gen", CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}, NewCandidate: true, Limits: limits, Stop: host.NewStopTiming(values), Log: logs, ExtraEnv: map[string]string{helperEnv: "adapter", "GORACE": "atexit_sleep_ms=0"}}, logs
}

func openMediaStore(t *testing.T) *media.Store {
	t.Helper()
	store, err := media.Open(t.TempDir(), media.NewLimits(tune.Defaults()), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

// TestAdapterRegistrationDeclarations 验证真实 Adapter 进程的非法权限和事件声明在注册阶段被拒绝。
func TestAdapterRegistrationDeclarations(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		for _, tc := range []struct {
			mode string
			want string
		}{
			{"permission-prefix", "invalid adapter permission name"},
			{"implies-user", "implies invalid permission"},
			{"implies-admin", "implies invalid permission"},
			{"permission-cycle", "implication cycle"},
			{"event-prefix", "invalid adapter event name"},
			{"schema-parse", "parse adapter event"},
			{"schema-resolve", "resolve adapter event"},
		} {
			t.Run(fmt.Sprintf("%s/stdio=%v", tc.mode, stdio), func(t *testing.T) {
				store, spec, _ := setup(t)
				spec.Stdio = stdio
				spec.ExtraEnv[helperEnv] = tc.mode
				running, err := supervisor.Launch(context.Background(), store, spec)
				if err == nil {
					stopped := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
					t.Fatalf("invalid adapter registration activated: stop=%+v", stopped)
				}
				var phase *host.PhaseError
				if !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_REGISTER || phase.Code != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("registration error = %v, want REGISTER/INVALID_ARGUMENT containing %q", err, tc.want)
				}
			})
		}
	}
}

// TestRemoteLifecycle 覆盖真实进程的多实例收发、账号唯一性、伪造拒绝、故障与游标恢复。
func TestRemoteLifecycle(t *testing.T) {
	ctx := context.Background()
	store, spec, logs := setup(t)
	registry := adapter.NewRegistry()
	sink := &eventSink{}
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: sink, RegisterIncoming: openMediaStore(t).RegisterIncoming})
	launch := func() *supervisor.Running {
		t.Helper()
		mux := router.NewMux()
		p.Services(mux)
		caller := router.NewCaller(spec.PluginID, spec.GenerationID)
		spec.Handler = mux.For(caller)
		running, err := supervisor.Launch(ctx, store, spec)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if res := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); res.Err != nil {
				t.Error(res.Err)
			}
		})
		if err := p.Bind(ctx, running.Instance, caller); err != nil {
			t.Fatal(err)
		}
		return running
	}
	running := launch()
	pid := running.Process.Pid()
	start := func(id, account string) {
		t.Helper()
		if err := p.StartInstance(ctx, id, []byte(`{"account":"`+account+`","display_name":"Loop display","token":"credential-canary"}`)); err != nil {
			t.Fatal(err)
		}
	}
	send := func(id, text string) (*pb.SendResult, *pb.ErrorInfo) {
		t.Helper()
		in, ok := registry.Get(id)
		if !ok {
			t.Fatalf("missing instance %s", id)
		}
		return in.Send(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: id, AccountId: in.AccountID(), ChatId: "chat"}, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}})
	}
	mustSend := func(id, text string) *pb.SendResult {
		t.Helper()
		out, safe := send(id, text)
		if safe != nil {
			t.Fatal(safe)
		}
		return out
	}
	count := func(eventType, id string, reason pb.BotOfflineReason) int {
		n := 0
		for _, ev := range sink.snapshot() {
			if ev.EventType == eventType && ev.GetAdapterInstanceId() == id && (eventType != "bot.offline" || ev.GetEventPayload().GetBotOffline().GetReason() == reason) {
				n++
			}
		}
		return n
	}
	if err := p.StartInstance(ctx, "one", []byte(`{"account":"a","display_name":"Loop display","username":"loop_bot","token":"credential-canary"}`)); err != nil {
		t.Fatal(err)
	}
	start("two", "b")
	for _, c := range []struct{ id, username string }{{"one", "loop_bot"}, {"two", ""}} {
		in, ok := registry.Get(c.id)
		if !ok {
			t.Fatalf("missing instance %s", c.id)
		}
		named, ok := in.(adapter.Named)
		if !ok || named.AccountName() != c.username {
			t.Fatalf("instance %s username = %v, want %q", c.id, in, c.username)
		}
	}
	for _, ev := range sink.snapshot() {
		if got := ev.GetEventPayload().GetBotOnline().GetAccountDisplayName(); got != "Loop display" {
			t.Fatalf("online display name = %q, want Loop display", got)
		}
	}
	if count("bot.online", "one", 0) != 1 || count("bot.online", "two", 0) != 1 {
		t.Fatal("missing online events")
	}
	mustSend("one", "hello")
	mustSend("two", "world")
	ev := sink.snapshot()[2]
	if ev.GetEventPayload().GetMessageCreated().GetChain().GetSegments()[0].GetText().GetText() != "hello" || ev.EventId == "adapter-forged-id" || len(ev.EventId) != 32 || ev.SourceKind != pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER || len(ev.Capabilities) != 1 {
		t.Fatalf("event was not normalized: %v", ev)
	}
	before, err := store.AdapterCursor(ctx, "two")
	if !bytes.Equal(before, []byte("1")) {
		t.Fatalf("delivered cursor = %q, want 1", before)
	}
	if err != nil {
		t.Fatal(err)
	}
	eventsBefore := len(sink.snapshot())
	if _, safe := send("two", "forge"); safe == nil {
		t.Fatal("forged instance accepted")
	}
	after, err := store.AdapterCursor(ctx, "two")
	if err != nil || !bytes.Equal(before, after) || len(sink.snapshot()) != eventsBefore {
		t.Fatal("forgery advanced cursor or delivered event")
	}
	for _, text := range []string{"bot.online", "bot.offline", "scheduled", "mismatched-payload"} {
		if _, safe := send("two", text); safe == nil {
			t.Fatalf("invalid event %s accepted", text)
		}
		current, err := store.AdapterCursor(ctx, "two")
		if err != nil || !bytes.Equal(before, current) || len(sink.snapshot()) != eventsBefore {
			t.Fatalf("invalid event %s changed cursor or Sink", text)
		}
	}
	if err := p.StartInstance(ctx, "duplicate", []byte(`{"account":"a"}`)); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
		t.Fatalf("duplicate account: %v", err)
	}
	if _, ok := registry.Get("duplicate"); ok {
		t.Fatal("duplicate account registered")
	}
	if err := p.StopInstance(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	if count("bot.offline", "one", pb.BotOfflineReason_BOT_OFFLINE_REASON_STOPPED) != 1 || running.Process.Pid() != pid {
		t.Fatal("instance stop changed process or missed offline")
	}
	mustSend("two", "still alive")
	start("failed", "c")
	mustSend("failed", "fail")
	if _, ok := registry.Get("failed"); ok {
		t.Fatal("failed instance remained registered")
	}
	if count("bot.offline", "failed", pb.BotOfflineReason_BOT_OFFLINE_REASON_INSTANCE_FAILED) != 1 {
		t.Fatal("missing failure event")
	}
	for _, ev := range sink.snapshot() {
		if ev.EventType == "bot.offline" && ev.GetAdapterInstanceId() == "failed" && ev.EventPayload.GetBotOffline().GetSafeError().GetSafeMessage() != "account rejected" {
			t.Fatal("safe error lost")
		}
	}
	// 故障实例可在同一个 Adapter 进程内重新启用。
	start("failed", "c")
	mustSend("failed", "recovered")
	if err := p.StopInstance(ctx, "failed"); err != nil {
		t.Fatal(err)
	}
	// 停止 RPC 报错仍关闭 Core 运行登记，且保留原始错误。
	start("stop-error", "stop-error")
	if err := p.StopInstance(ctx, "stop-error"); err == nil || !strings.Contains(err.Error(), "platform stop failed") {
		t.Fatalf("stop error lost: %v", err)
	}
	if _, ok := registry.Get("stop-error"); ok {
		t.Fatal("failed stop remained registered")
	}
	if count("bot.offline", "stop-error", pb.BotOfflineReason_BOT_OFFLINE_REASON_STOPPED) != 1 {
		t.Fatal("failed stop missed offline event")
	}
	// 未清理的 SDK 实例在显式重新启用前先完成清理。
	start("stop-error", "stop-error")
	if err := p.StopInstance(ctx, "stop-error"); err == nil {
		t.Fatal("expected new instance stop failure")
	}
	err = p.StartInstance(ctx, "start-error", []byte(`{"account":"start-error","token":"credential-canary"}`))
	if err == nil || strings.Contains(err.Error(), "credential-canary") {
		t.Fatalf("startup credential leaked: %v", err)
	}
	start("three", "d")
	mustSend("three", "before crash")
	saved := map[string][]byte{}
	for _, id := range []string{"two", "three"} {
		saved[id], err = store.AdapterCursor(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	oldInstance, _ := registry.Get("two")
	oldSessionKey := oldInstance.(adapter.MediaSession).MediaSessionKey()
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	if _, _, err := running.WaitExit(); err != nil {
		t.Fatal(err)
	}
	p.Unbind()
	p.Unbind()
	for _, id := range []string{"two", "three"} {
		if count("bot.offline", id, pb.BotOfflineReason_BOT_OFFLINE_REASON_ADAPTER_EXITED) != 1 {
			t.Fatalf("wrong exit event count for %s", id)
		}
	}
	spec.NewCandidate = false
	running = launch()
	restored, _ := registry.Get("two")
	if restored.(adapter.MediaSession).MediaSessionKey() == oldSessionKey || oldInstance.(adapter.MediaSession).MediaSessionKey() != oldSessionKey {
		t.Fatal("restored instance reused its old media session or changed the old instance identity")
	}
	for _, id := range []string{"one", "failed", "duplicate"} {
		if _, ok := registry.Get(id); ok {
			t.Fatalf("disabled instance %s restored", id)
		}
	}
	for _, id := range []string{"two", "three"} {
		out := mustSend(id, "cursor")
		if len(out.PlatformMessageIds) != 1 || out.PlatformMessageIds[0] != string(saved[id]) {
			t.Fatalf("cursor not restored for %s: %v", id, out)
		}
	}
	for _, ev := range sink.snapshot() {
		if ev.EventType == "bot.online" || ev.EventType == "bot.offline" {
			if ev.Actor != nil || ev.Target != nil || ev.Message != nil {
				t.Fatal("lifecycle event contains platform context")
			}
		}
	}
	logs.mu.Lock()
	defer logs.mu.Unlock()
	if bytes.Contains(logs.b.Bytes(), []byte("credential-canary")) {
		t.Fatal("configuration leaked into logs")
	}
}

// TestRemoteRoleMismatch 在进程启动边界确认两种角色伪装都在 REGISTER 阶段失败。
func TestRemoteRoleMismatch(t *testing.T) {
	for _, mode := range []string{"adapter", "plugin"} {
		t.Run(mode, func(t *testing.T) {
			store, spec, _ := setup(t)
			spec.ExtraEnv[helperEnv] = mode
			if mode == "adapter" {
				spec.Role = state.RolePlugin
			}
			running, err := supervisor.Launch(context.Background(), store, spec)
			if running != nil {
				running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
				t.Fatal("role mismatch activated")
			}
			var phase *host.PhaseError
			if !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_REGISTER {
				t.Fatalf("role mismatch failed outside REGISTER: %v", err)
			}
		})
	}
}

// TestRemoteMediaRead 经真实 Adapter 进程登记并共享入站媒体，保持 Session 身份与平台错误。
func TestRemoteMediaRead(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	registry := adapter.NewRegistry()
	mediaStore := openMediaStore(t)
	sink := &eventSink{}
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: sink, RegisterIncoming: mediaStore.RegisterIncoming})
	mux := router.NewMux()
	p.Services(mux)
	dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: 1, Adapters: registry, Prepare: mediaStore.PrepareIncoming})
	t.Cleanup(func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	mediaStore.AddReadServices(mux, registry, dispatch, store.AdapterInstanceAllowed)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	})
	if err := p.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ id, config string }{
		{"readable", `{"account":"media-reader-account","media":true}`},
		{"unsupported", `{"account":"unsupported-account"}`},
	} {
		if err := p.StartInstance(ctx, c.id, []byte(c.config)); err != nil {
			t.Fatal(err)
		}
	}
	in, _ := registry.Get("readable")
	unsupported, _ := registry.Get("unsupported")
	for _, check := range []struct {
		in       adapter.Instance
		readable bool
	}{
		{in, true},
		{unsupported, false},
	} {
		_, implements := check.in.(adapter.MediaReader)
		if implements != check.readable {
			t.Fatalf("instance %s MediaReader = %v, want %v", check.in.InstanceID(), implements, check.readable)
		}
		var declared bool
		declarations := 0
		for _, capability := range check.in.Capabilities() {
			if capability.Name == "media.read" {
				declarations++
			}
			if capability.Name == "media.read" && capability.Version == 1 && capability.Supported {
				declared = true
			}
		}
		if declared != check.readable {
			t.Fatalf("instance %s media.read declaration = %v", check.in.InstanceID(), declared)
		}
		if declarations > 1 {
			t.Fatalf("instance %s has duplicate media.read declarations", check.in.InstanceID())
		}
	}
	if got := in.(adapter.MediaSession).MediaSessionKey(); got != caller.SessionKey {
		t.Fatalf("media session = %q, want %q", got, caller.SessionKey)
	}
	// 1. 真实交付先登记对象，所有消息链和转发条目均改写，接纳期间不下载。
	target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "readable", AccountId: in.AccountID(), ChatId: "chat"}
	chain := func(command, token string) *pb.MessageChain {
		source := func() *pb.MediaSource {
			return &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{Token: token, Size: proto.Uint64(11), Mime: proto.String("application/octet-stream")}}}
		}
		return &pb.MessageChain{Segments: []*pb.Segment{
			{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: command}}},
			{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: source(), Caption: proto.String("image")}}},
			{Data: &pb.Segment_Video{Video: &pb.VideoSegment{Source: source(), Caption: proto.String("video")}}},
			{Data: &pb.Segment_File{File: &pb.FileSegment{Source: source(), Filename: proto.String("file")}}},
			{Data: &pb.Segment_Forward{Forward: &pb.ForwardSegment{Items: []*pb.ForwardItem{{Chain: &pb.MessageChain{Segments: []*pb.Segment{
				{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: source()}}},
			}}}}}}},
			{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: &pb.MediaSource{Source: &pb.MediaSource_HttpsUrl{HttpsUrl: "https://example.com/image"}}}}},
		}}
	}
	var objectID string
	for _, command := range []string{"media-created", "media-edited"} {
		if _, safe := in.Send(ctx, target, chain(command, "platform-file")); safe != nil {
			t.Fatal(safe)
		}
		events := sink.snapshot()
		ev := events[len(events)-1]
		id := ev.GetMessage().GetChain().GetSegments()[1].GetImage().GetSource().GetIncomingRef().GetToken()
		if id == "" || id == "platform-file" || objectID != "" && id != objectID {
			t.Fatalf("incoming object ID = %q, previous %q", id, objectID)
		}
		objectID = id
		chains := []*pb.MessageChain{ev.GetMessage().GetChain()}
		if command == "media-created" {
			chains = append(chains, ev.GetEventPayload().GetMessageCreated().GetChain())
		} else {
			chains = append(chains, ev.GetEventPayload().GetMessageEdited().GetChain(), ev.GetEventPayload().GetMessageEdited().GetOldChain())
		}
		for index, got := range chains {
			if want := chain(command, objectID); !proto.Equal(got, want) {
				t.Fatalf("%s chain %d = %v, want %v", command, index, got, want)
			}
		}
	}
	readCount := func(want string) {
		t.Helper()
		out, safe := in.Send(ctx, target, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "media-read-count"}}}}})
		if safe != nil {
			t.Fatal(safe)
		}
		if len(out.PlatformMessageIds) != 1 || out.PlatformMessageIds[0] != want {
			t.Fatalf("platform read count = %v, want %s", out.PlatformMessageIds, want)
		}
	}
	readCount("0")
	if id := mediaStore.RegisterIncoming("readable", "platform-file", nil, nil); id != objectID {
		t.Fatalf("registered object ID = %q, want %q", id, objectID)
	}
	unsupportedID := mediaStore.RegisterIncoming("unsupported", "platform-file", nil, nil)
	if unsupportedID == objectID {
		t.Fatal("different adapter instances shared an incoming object")
	}
	consumer := router.NewCaller("unsupported-consumer", "generation")
	descriptor := mediaStore.IssueIncoming(unsupportedID, consumer, nil)
	if descriptor.Readability != pb.MediaReadability_MEDIA_READABILITY_UNSUPPORTED {
		t.Fatalf("unsupported readability = %v", descriptor.Readability)
	}
	readRequest := &pb.ReadRequest{Token: descriptor.Token, Length: 4}
	encoded, err := proto.Marshal(readRequest)
	if err != nil {
		t.Fatal(err)
	}
	if out := mux.For(consumer)(ctx, &wire.Inbound{Target: wire.ServiceMediaRead, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(readRequest), SchemaVersion: 1, Bytes: encoded}}); out.Err.GetCode() != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("unsupported Core read = %v", out.Err)
	}

	// 2. 两个接收 Session 使用独立令牌，经 Core 读取同一份惰性缓存。
	var firstToken string
	for _, pluginID := range []string{"a", "b"} {
		consumer := router.NewCaller(pluginID, "generation")
		incoming := mediaStore.IssueIncoming(objectID, consumer, []*pb.Target{target})
		if incoming == nil {
			t.Fatal("incoming object disappeared before issue")
		}
		if incoming.Token == firstToken || incoming.Token == objectID || len(incoming.Token) != 32 || incoming.GetSize() != 11 || incoming.GetMime() != "application/octet-stream" || incoming.Readability != pb.MediaReadability_MEDIA_READABILITY_READABLE {
			t.Fatalf("invalid incoming descriptor: %v", incoming)
		}
		firstToken = incoming.Token
		var content []byte
		for offset := uint64(0); ; {
			request := &pb.ReadRequest{Token: incoming.Token, Offset: offset, Length: 4}
			encoded, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			out := mux.For(consumer)(ctx, &wire.Inbound{Target: wire.ServiceMediaRead, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
			if out.Err != nil {
				t.Fatal(out.Err)
			}
			response := &pb.ReadResponse{}
			if err := proto.Unmarshal(out.Payload.Bytes, response); err != nil {
				t.Fatal(err)
			}
			content = append(content, response.Data...)
			if response.NextOffset != offset+uint64(len(response.Data)) {
				t.Fatalf("Core read offset = %d after %d bytes", response.NextOffset, len(response.Data))
			}
			offset = response.NextOffset
			if response.Eof {
				break
			}
		}
		if string(content) != "image bytes" {
			t.Fatalf("consumer %s content = %q", pluginID, content)
		}
		readCount("1")
	}

	// 3. 底层读取继续保留分块、EOF、平台错误和实例生命周期语义。
	reader := in.(adapter.MediaReader)
	var content []byte
	var offset uint64
	for {
		data, next, eof, err := reader.ReadMedia(ctx, "platform-file", offset, 4)
		if err != nil {
			t.Fatal(err)
		}
		content = append(content, data...)
		if next != offset+uint64(len(data)) {
			t.Fatalf("next offset = %d after %d bytes", next, len(data))
		}
		offset = next
		if eof {
			break
		}
	}
	if string(content) != "image bytes" {
		t.Fatalf("content = %q", content)
	}
	if data, next, eof, err := reader.ReadMedia(ctx, "platform-file", offset, 4); err != nil || len(data) != 0 || next != offset || !eof {
		t.Fatalf("read at EOF = %q %d %v %v", data, next, eof, err)
	}
	if _, _, _, err := reader.ReadMedia(ctx, "missing", 0, 4); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_NOT_FOUND {
		t.Fatalf("missing file error = %v", err)
	} else {
		var call *wire.CallError
		if !errors.As(err, &call) || call.ErrorInfo.SafeMessage != "platform file not found" {
			t.Fatalf("platform safe message changed: %v", err)
		}
	}
	for _, length := range []uint32{0, wire.MaxMediaChunk + 1} {
		if _, _, _, err := reader.ReadMedia(ctx, "platform-file", 0, length); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
			t.Fatalf("length %d error = %v", length, err)
		}
	}
	mediaRequest := &pb.AdapterMediaReadRequest{InstanceId: "unsupported", FileId: "platform-file", Length: 4}
	encoded, err = proto.Marshal(mediaRequest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := running.Instance.Session.Call(ctx, wire.Request{Target: wire.ServiceAdapterMediaRead, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(mediaRequest), SchemaVersion: 1, Bytes: encoded}}); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("unsupported instance error = %v", err)
	}
	if err := p.StopInstance(ctx, "readable"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := reader.ReadMedia(ctx, "platform-file", 0, 4); wire.CodeOf(err) != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("stopped instance error = %v", err)
	}
}

// TestRemotePermissionsAndEvents 验证真实 Adapter/插件进程之间的声明、权限展开和扩展事件准入。
func TestRemotePermissionsAndEvents(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprintf("stdio=%v", stdio), func(t *testing.T) {
			ctx := t.Context()
			store, spec, logs := setup(t)
			spec.Stdio = stdio
			registry := adapter.NewRegistry()
			dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: tune.Defaults().DispatchQueueEvents, Adapters: registry})
			dispatch.Start()
			defer func() {
				if err := dispatch.Close(context.Background()); err != nil {
					t.Error(err)
				}
			}()
			process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: dispatch, RegisterIncoming: openMediaStore(t).RegisterIncoming})
			mux := router.NewMux()
			process.Services(mux)
			caller := router.NewCaller(spec.PluginID, spec.GenerationID)
			spec.Handler = mux.For(caller)
			running, err := supervisor.Launch(ctx, store, spec)
			if err != nil {
				t.Fatal(err)
			}
			defer running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			if !proto.Equal(running.Instance.Adapter, expectedDescriptor(t)) {
				t.Fatalf("activated descriptor=%v", running.Instance.Adapter)
			}
			if err := process.Bind(ctx, running.Instance, caller); err != nil {
				t.Fatal(err)
			}
			if err := process.StartInstance(ctx, "one", []byte(`{"account":"a"}`)); err != nil {
				t.Fatal(err)
			}
			audit := filepath.Join(t.TempDir(), "observer.jsonl")
			if err := os.WriteFile(audit, nil, 0600); err != nil {
				t.Fatal(err)
			}
			observer := spec
			observer.PluginID = "observer"
			observer.Role = state.RolePlugin
			observer.AdapterID = ""
			observer.GenerationID = "observer-gen"
			observer.PluginDir = t.TempDir()
			observer.Handler = nil
			observer.Bootstrap = &pb.Bootstrap{GenerationId: observer.GenerationID, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
			observer.ExtraEnv = map[string]string{helperEnv: "plugin", observerAuditEnv: audit, "GORACE": "atexit_sleep_ms=0"}
			if err := store.AddInstall(ctx, state.Install{PluginID: observer.PluginID, Role: state.RolePlugin, BuildDigest: observer.BuildDigest}); err != nil {
				t.Fatal(err)
			}
			if err := store.AddGeneration(ctx, observer.PluginID, observer.GenerationID, "{}", 1); err != nil {
				t.Fatal(err)
			}
			plugin, err := supervisor.Launch(ctx, store, observer)
			if err != nil {
				t.Fatal(err)
			}
			defer plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
			dispatch.Attach(router.FromInstance(plugin.Instance), router.NewCaller(observer.PluginID, observer.GenerationID))
			reports := func() []observedEvent {
				t.Helper()
				file, err := os.Open(audit)
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				decoder := json.NewDecoder(file)
				var all []observedEvent
				for {
					var report observedEvent
					err := decoder.Decode(&report)
					if err == io.EOF {
						return all
					}
					if err != nil {
						t.Fatal(err)
					}
					all = append(all, report)
				}
			}
			instance, _ := registry.Get("one")
			target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}
			for _, tc := range []struct {
				name     string
				rejected bool
				routes   []string
			}{
				{"permissions", false, []string{"raw", "loop"}},
				{"extension", false, []string{"extension"}},
				{"extension-unknown", true, nil}, {"extension-other", true, nil}, {"extension-version", true, nil}, {"extension-schema", true, nil}, {"extension-array", true, nil}, {"extension-json", true, nil}, {"extension-branch", true, nil}, {"actor-forge", true, nil}, {"target-forge", true, nil},
			} {
				t.Run(tc.name, func(t *testing.T) {
					before := len(reports())
					cursor, err := store.AdapterCursor(ctx, "one")
					if err != nil {
						t.Fatal(err)
					}
					chain := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: tc.name}}}}}
					_, safe := instance.Send(ctx, target, chain)
					if (safe != nil) != tc.rejected {
						t.Fatalf("send error=%v rejected=%v", safe, tc.rejected)
					}
					if tc.rejected {
						after, err := store.AdapterCursor(ctx, "one")
						if err != nil || !bytes.Equal(cursor, after) {
							t.Fatalf("rejected event advanced cursor: %q -> %q error=%v", cursor, after, err)
						}
					}
					if err := dispatch.Deliver(ctx, &pb.Event{EventType: "scheduled", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_SCHEDULER, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_Scheduled{Scheduled: &pb.ScheduledPayload{}}}}); err != nil {
						t.Fatal(err)
					}
					deadline := time.Now().Add(5 * time.Second)
					var observed []observedEvent
					for {
						observed = reports()[before:]
						if len(observed) > 0 && observed[len(observed)-1].Route == "barrier" {
							break
						}
						if time.Now().After(deadline) {
							t.Fatalf("observer barrier missing: %v", observed)
						}
						time.Sleep(10 * time.Millisecond)
					}
					observed = observed[:len(observed)-1]
					var routes []string
					for _, report := range observed {
						routes = append(routes, report.Route)
						if !reflect.DeepEqual(report.Permissions, []string{"chat_admin", "loop.admin", "loop.owner", "user"}) {
							t.Fatalf("permissions=%v", report.Permissions)
						}
						if report.Route == "extension" && (report.Version != 1 || report.Payload != `{"user":"external-7"}`) {
							t.Fatalf("extension report=%+v", report)
						}
					}
					if !reflect.DeepEqual(routes, tc.routes) {
						t.Fatalf("routes=%v want %v", routes, tc.routes)
					}
				})
			}
			// 同一安装包的旧 Caller 也无法推进当前 Session 的游标。
			cursor, err := store.AdapterCursor(ctx, "one")
			if err != nil {
				t.Fatal(err)
			}
			request := &pb.AdapterDeliverRequest{InstanceId: "one", Cursor: []byte("stale"), Event: &pb.Event{EventType: "message.created", AdapterId: proto.String("loop"), AdapterInstanceId: proto.String("one"), AccountId: proto.String("a"), Target: target, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "stale"}}}}}}}}}}
			encoded, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			response := mux.For(router.NewCaller(spec.PluginID, "old-generation"))(ctx, &wire.Inbound{Target: wire.ServiceCoreAdapterDeliver, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
			after, err := store.AdapterCursor(ctx, "one")
			if response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT || err != nil || !bytes.Equal(cursor, after) {
				t.Fatalf("stale Caller result=%v cursor=%q -> %q error=%v", response.Err, cursor, after, err)
			}
			logs.mu.Lock()
			text := logs.b.String()
			logs.mu.Unlock()
			if strings.Count(text, "discard undeclared adapter permission") != 8 {
				t.Fatalf("discard diagnostics=%s", text)
			}
			for _, name := range []string{"loop.unknown", "user", "admin", "chat_admin"} {
				if !strings.Contains(text, "name="+name) {
					t.Errorf("missing discarded permission diagnostic %q", name)
				}
			}
			if !strings.Contains(text, "level=WARN") || !strings.Contains(text, "plugin_id=loop-package instance_id=one") || !strings.Contains(text, "invalid adapter instance event") {
				t.Fatalf("missing rejection identity or level: %s", text)
			}
		})
	}
}

func commandCatalog() map[string][]byte {
	return map[string][]byte{"zh-CN.json": []byte(`{"commands":{"ping":{"description":"测试机器人"},"rss add":{"description":"添加订阅"}}}`)}
}

func (i *loopInstance) SyncCommands(_ context.Context, commands []sdkadapter.CommandMenuItem) *errcode.Error {
	encoded, err := json.Marshal(commands)
	if err != nil {
		return errcode.Internal.New(err.Error())
	}
	file, err := os.OpenFile(i.commandsAudit, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return errcode.Internal.New(err.Error())
	}
	_, writeErr := file.Write(append(encoded, '\n'))
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return errcode.Internal.New(err.Error())
	}
	return nil
}

// TestRemoteCommandSync 验证真实 SDK 进程的命令说明、菜单筛选和持久化禁用后的更新。
func TestRemoteCommandSync(t *testing.T) {
	ctx := t.Context()
	store, spec, logs := setup(t)
	registry := adapter.NewRegistry()
	commandPolicy := policy.New(store)
	dispatch := router.New(router.Options{Concurrency: 1, QueueEvents: tune.Defaults().DispatchQueueEvents, QueueBytes: tune.Defaults().DispatchQueueBytes, Adapters: registry, Policy: commandPolicy, CommandState: commandPolicy.CommandState})
	commandPolicy.SetCommandsChanged(dispatch.CommandsChanged)
	dispatch.Start()
	defer dispatch.Close(context.Background())
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: dispatch, RegisterIncoming: openMediaStore(t).RegisterIncoming, StartCommandSync: dispatch.StartCommandSync, StopCommandSync: dispatch.StopCommandSync})
	mux := router.NewMux()
	process.Services(mux)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		process.Unbind()
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	}()
	if err := process.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	audit := filepath.Join(t.TempDir(), "commands.jsonl")
	if err := os.WriteFile(audit, nil, 0600); err != nil {
		t.Fatal(err)
	}
	config, err := json.Marshal(map[string]string{"account": "commands-account", "commands_audit": audit})
	if err != nil {
		t.Fatal(err)
	}
	if err := process.StartInstance(ctx, "menu", config); err != nil {
		t.Fatal(err)
	}
	// 每条审计记录来自 Adapter 的实际 SyncCommands 调用，启动须先清空菜单。
	awaitMenu := func(count int) []sdkadapter.CommandMenuItem {
		t.Helper()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			encoded, err := os.ReadFile(audit)
			if err != nil {
				t.Fatal(err)
			}
			lines := bytes.Split(encoded, []byte{'\n'})
			if len(lines)-1 >= count {
				var menu []sdkadapter.CommandMenuItem
				if err := json.Unmarshal(lines[count-1], &menu); err != nil {
					t.Fatal(err)
				}
				return menu
			}
			select {
			case <-ticker.C:
			case <-timer.C:
				logs.mu.Lock()
				text := logs.b.String()
				logs.mu.Unlock()
				t.Fatalf("menu sync %d timed out; logs=%s", count, text)
			}
		}
	}
	if menu := awaitMenu(1); len(menu) != 0 {
		t.Fatal(menu)
	}
	observer := spec
	observer.PluginID = "command-observer"
	observer.Role = state.RolePlugin
	observer.AdapterID = ""
	observer.GenerationID = "commands-gen"
	observer.PluginDir = t.TempDir()
	observer.Handler = nil
	observer.Bootstrap = &pb.Bootstrap{GenerationId: observer.GenerationID, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	observer.ExtraEnv = map[string]string{helperEnv: "command-plugin", "GORACE": "atexit_sleep_ms=0"}
	observer.I18n = commandCatalog()
	observer.DefaultLocale = "zh-CN"
	observer.CommandsChanged = dispatch.CommandsChanged
	if err := store.AddInstall(ctx, state.Install{PluginID: observer.PluginID, Role: state.RolePlugin, BuildDigest: observer.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, observer.PluginID, observer.GenerationID, "{}", 1); err != nil {
		t.Fatal(err)
	}
	plugin, err := supervisor.Launch(ctx, store, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if result := plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	}()
	dispatch.Attach(router.FromInstance(plugin.Instance), router.NewCaller(observer.PluginID, observer.GenerationID))
	want := []sdkadapter.CommandMenuItem{{Name: "bare", Descriptions: map[string]string{"en-US": "bare", "zh-CN": "bare"}}, {Name: "ping", Descriptions: map[string]string{"en-US": "Ping the bot", "zh-CN": "测试机器人"}}}
	if got := awaitMenu(2); !reflect.DeepEqual(got, want) {
		t.Fatalf("menu=%v want=%v", got, want)
	}
	disabled := false
	if err := commandPolicy.SetOverride(ctx, observer.PluginID, policy.Scope{Kind: policy.ScopeRoute, RouteID: "cmd.ping"}, policy.Override{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}
	if got := awaitMenu(3); !reflect.DeepEqual(got, want[:1]) {
		t.Fatalf("disabled menu=%v want=%v", got, want[:1])
	}
}

// streamCapabilities 让 REGISTER 与实例返回同一份 UTF-16 限额。
func streamCapabilities() []message.Capability {
	return []message.Capability{
		{Name: "message.send", Version: 1, Supported: true, TextLimit: &message.TextLimit{Max: 6, Unit: message.TextUTF16CodeUnits}},
		{Name: "message.edit", Version: 1, Supported: true},
	}
}

type streamReport struct {
	Results      []*message.SendResult
	Capabilities []message.Capability
	Error        string
}

type streamObservation struct {
	Operation, Text, ID string
	Session             string
	At                  time.Time
}

func (i *loopInstance) Edit(_ context.Context, target string, id string, chain message.Chain) (*message.SendResult, *errcode.Error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.streamReceipt("edit", target, id, chain.Text())
}

// streamReceipt 在 Adapter 侧记录真实请求到达时间与消息 ID，返回完整回执。
func (i *loopInstance) streamReceipt(operation string, target string, id, text string) (*message.SendResult, *errcode.Error) {
	observation := streamObservation{Operation: operation, Text: text, ID: id, Session: target, At: time.Now()}
	data, err := json.Marshal(observation)
	if err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	file, err := os.OpenFile(os.Getenv(streamAuditEnv), os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	_, err = file.Write(append(data, '\n'))
	if err = errors.Join(err, file.Close()); err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	return &message.SendResult{Status: message.SendAccepted, Session: target, AdapterInstanceID: i.id,
		PrimaryPlatformMessageID: id, PlatformMessageIDs: []string{id}, ProviderAcceptedAt: time.UnixMilli(1234), CapabilitiesUsed: streamCapabilities(),
		PartResults: []message.PartResult{{PartIndex: 0, SegmentIndexes: []uint32{0}, Status: message.PartAccepted, PlatformMessageIDs: []string{id}}}}, nil
}

// TestRemoteStream 经真实 SDK Adapter、插件、Session、Router 与消息服务验证流式换条和节流。
func TestRemoteStream(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	registry := adapter.NewRegistry()
	dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: tune.Defaults().DispatchQueueEvents, Adapters: registry})
	dispatch.Start()
	defer func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: dispatch, RegisterIncoming: openMediaStore(t).RegisterIncoming})
	mux := router.NewMux()
	process.Services(mux)
	service := coremessage.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout, nil)
	var requestsMu sync.Mutex
	var parents []uint64
	for name, handler := range map[string]router.ServiceFunc{wire.ServiceSend: service.Send, wire.ServiceEdit: service.Edit} {
		mux.Handle(name, func(ctx context.Context, caller router.Caller, in *wire.Inbound) wire.Response {
			requestsMu.Lock()
			parents = append(parents, in.ParentCallID)
			requestsMu.Unlock()
			if _, ok := in.ParentAttach.(*router.Invocation); !ok {
				t.Error("stream request lost trusted parent invocation")
			}
			return handler(ctx, caller, in)
		})
	}
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	audit := filepath.Join(t.TempDir(), "requests.jsonl")
	if err := os.WriteFile(audit, nil, 0600); err != nil {
		t.Fatal(err)
	}
	spec.ExtraEnv = map[string]string{helperEnv: "stream-adapter", streamAuditEnv: audit, "GORACE": "atexit_sleep_ms=0"}
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	capabilities := running.Instance.Adapter.Capabilities
	if len(capabilities) != 2 || capabilities[0].GetTextLimit().GetMax() != 6 || capabilities[0].GetTextLimit().GetUnit() != pb.TextUnit_TEXT_UNIT_UTF16_CODE_UNITS {
		t.Fatalf("REGISTER capabilities=%v", capabilities)
	}
	if err := process.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	if err := process.StartInstance(ctx, "one", []byte(`{"account":"a","stream":true}`)); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(t.TempDir(), "result.json")
	observer := spec
	observer.PluginID = "stream-observer"
	observer.Role = state.RolePlugin
	observer.AdapterID = ""
	observer.GenerationID = "observer-gen"
	observer.PluginDir = t.TempDir()
	observer.Bootstrap = &pb.Bootstrap{GenerationId: observer.GenerationID, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	observerCaller := router.NewCaller(observer.PluginID, observer.GenerationID)
	observer.Handler = mux.For(observerCaller)
	observer.ExtraEnv = map[string]string{helperEnv: "stream-plugin", observerAuditEnv: reportPath, "GORACE": "atexit_sleep_ms=0"}
	if err := store.AddInstall(ctx, state.Install{PluginID: observer.PluginID, Role: state.RolePlugin, BuildDigest: observer.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, observer.PluginID, observer.GenerationID, "{}", 1); err != nil {
		t.Fatal(err)
	}
	plugin, err := supervisor.Launch(ctx, store, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	dispatch.Attach(router.FromInstance(plugin.Instance), observerCaller)
	instance, _ := registry.Get("one")
	target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat", ThreadId: proto.String("topic")}
	_, safe := instance.Send(ctx, target, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "stream-trigger"}}}}})
	if safe != nil {
		t.Fatal(safe)
	}
	// 关闭派发入口并等待真实 Handler 完成后再读取文件，避免把半条 JSON 当成结果。
	closeContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := dispatch.Close(closeContext); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report streamReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Error != "" {
		t.Fatalf("stream error=%s", report.Error)
	}
	if !reflect.DeepEqual(report.Capabilities, streamCapabilities()) {
		t.Fatalf("event capabilities=%+v", report.Capabilities)
	}
	file, err := os.Open(audit)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var observed []streamObservation
	for {
		var call streamObservation
		err := decoder.Decode(&call)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		observed = append(observed, call)
	}
	wantOperations := []string{"send", "edit", "edit", "send"}
	wantText := []string{"a😀", "a😀b", "a😀bcd", "😀"}
	if len(observed) != len(wantOperations) || len(report.Results) != len(observed) {
		t.Fatalf("requests=%+v results=%+v", observed, report.Results)
	}
	for index, call := range observed {
		result := report.Results[index]
		if call.Operation != wantOperations[index] || call.Text != wantText[index] || call.Session != message.SID("loop", "chat", "topic") {
			t.Fatalf("request %d=%+v", index, call)
		}
		if result.PrimaryPlatformMessageID != call.ID || !reflect.DeepEqual(result.PlatformMessageIDs, []string{call.ID}) || result.Ref() != message.Ref(call.Session, call.ID, "one") || result.ProviderAcceptedAt.UnixMilli() != 1234 || !reflect.DeepEqual(result.CapabilitiesUsed, streamCapabilities()) || len(result.PartResults) != 1 || result.PartResults[0].PlatformMessageIDs[0] != call.ID {
			t.Fatalf("receipt %d=%+v", index, result)
		}
	}
	if observed[2].At.Sub(observed[1].At) < time.Second {
		t.Fatalf("edit interval=%v", observed[2].At.Sub(observed[1].At))
	}
	if observed[0].ID != observed[1].ID || observed[1].ID != observed[2].ID || observed[0].ID == observed[3].ID {
		t.Fatalf("message IDs=%v", observed)
	}
	requestsMu.Lock()
	defer requestsMu.Unlock()
	if len(parents) != len(observed) {
		t.Fatalf("parent requests=%v", parents)
	}
	for _, parent := range parents {
		if parent == 0 {
			t.Fatal("stream lost parent call ID")
		}
	}
}

// Action 记录真实跨进程调用次数，回传带平台额外字段的 JSON 原文。
func (i *loopInstance) Action(_ context.Context, to string, name string, version uint32, params json.RawMessage) (json.RawMessage, *errcode.Error) {
	if version != 1 || (name != "loop.echo" && name != "loop.account") {
		return nil, errcode.Unsupported.New("action is not declared")
	}
	if (name == "loop.account") != (to == "") {
		return nil, errcode.InvalidArgument.New("action scope does not match session")
	}
	var input loopActionParams
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, errcode.InvalidArgument.New(err.Error())
	}
	i.mu.Lock()
	i.actionCalls++
	i.mu.Unlock()
	value, err := json.Marshal(input.Value)
	if err != nil {
		return nil, errcode.Internal.New(err.Error())
	}
	return json.RawMessage(` {"value": ` + string(value) + `, "platform_extra": true} `), nil
}

// TestCommandNameOverride 验证同一真实插件进程在保存和删除覆盖后按新名字派发，参数保持一致。
func TestCommandNameOverride(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	commandPolicy := policy.New(store)
	dispatch := router.New(router.Options{Concurrency: 1, QueueEvents: tune.Defaults().DispatchQueueEvents, QueueBytes: tune.Defaults().DispatchQueueBytes, Policy: commandPolicy, CommandNames: commandPolicy.CommandNames})
	commandPolicy.SetCommandsChanged(dispatch.CommandsChanged)
	dispatch.Start()
	defer func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	audit := filepath.Join(t.TempDir(), "names.jsonl")
	if err := os.WriteFile(audit, nil, 0600); err != nil {
		t.Fatal(err)
	}
	spec.PluginID, spec.Role, spec.AdapterID = "names-observer", state.RolePlugin, ""
	spec.GenerationID, spec.PluginDir = "names-gen", t.TempDir()
	spec.Bootstrap = &pb.Bootstrap{GenerationId: spec.GenerationID, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	spec.ExtraEnv = map[string]string{helperEnv: "names-plugin", observerAuditEnv: audit, "GORACE": "atexit_sleep_ms=0"}
	if err := store.AddInstall(ctx, state.Install{PluginID: spec.PluginID, Role: spec.Role, BuildDigest: spec.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, spec.PluginID, spec.GenerationID, "{}", 1); err != nil {
		t.Fatal(err)
	}
	plugin, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if result := plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	}()
	dispatch.Attach(router.FromInstance(plugin.Instance), router.NewCaller(spec.PluginID, spec.GenerationID))
	check := func(name string, matched bool) {
		t.Helper()
		before, err := os.ReadFile(audit)
		if err != nil {
			t.Fatal(err)
		}
		entry, _, ok := dispatch.Commands().Lookup(strings.Fields(name))
		if ok != matched || matched && entry.RouteID != "cmd.echo" {
			t.Fatalf("lookup %s=%+v matched=%v", name, entry, ok)
		}
		event := &pb.Event{EventId: name, EventType: "message.created", SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER, AdapterId: proto.String("loop"), AdapterInstanceId: proto.String("one"), AccountId: proto.String("a"), Actor: &pb.Actor{ExternalUserId: "user"}, Target: &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}, EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "/" + name + ` "hello world"`}}}}}}}}}
		if err := dispatch.Deliver(ctx, event); err != nil {
			t.Fatal(err)
		}
		timer, ticker := time.NewTimer(5*time.Second), time.NewTicker(10*time.Millisecond)
		defer timer.Stop()
		defer ticker.Stop()
		for stats := dispatch.Stats(); stats.Running != 0 || stats.Queued != 0; stats = dispatch.Stats() {
			select {
			case <-ticker.C:
			case <-timer.C:
				t.Fatal("command dispatch timed out")
			}
		}
		after, err := os.ReadFile(audit)
		if err != nil {
			t.Fatal(err)
		}
		decoder := json.NewDecoder(bytes.NewReader(after[len(before):]))
		var commands []string
		for {
			var report observedEvent
			if err := decoder.Decode(&report); err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(report.Route, "cmd.") {
				commands = append(commands, report.Route)
			}
		}
		if matched && !reflect.DeepEqual(commands, []string{"cmd.echo:hello world"}) || !matched && len(commands) != 0 {
			t.Fatalf("name=%s handler reports=%v matched=%v", name, commands, matched)
		}
	}
	check("echo", true)
	if err := commandPolicy.SetCommandNames(ctx, spec.PluginID, "cmd.echo", [][]string{{"renamed", "echo"}, {"another"}}, plugin.Instance.Routes); err != nil {
		t.Fatal(err)
	}
	check("echo", false)
	check("e", false)
	check("renamed echo", true)
	check("another", true)
	if err := commandPolicy.DeleteCommandNames(ctx, spec.PluginID, "cmd.echo"); err != nil {
		t.Fatal(err)
	}
	check("echo", true)
	check("e", true)
	check("renamed echo", false)
}

const objectConfigSchema = `{"root":{"type":"string","default":"whole"},"part":{"type":"object","items":{"prefix":{"type":"string","default":"config"}}}}`

func objectI18n() map[string][]byte {
	return map[string][]byte{"en-US.json": []byte(`{"messages":{"object.reply":"{prefix}|{name}|{count}|{tail}"}}`)}
}

func newObjectPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{ConfigSchema: []byte(objectConfigSchema), I18n: objectI18n()})
	sub := p.Sub("part").Permission(sdk.Anyone)
	values := sub.Config()
	sub.Command("object").Handle(func(c *sdk.Context, args struct {
		Name  string
		Count int
		Tail  string `arg:",raw" default:"fallback"`
	}) (message.Content, error) {
		if c.Config().String("root") != "whole" {
			return nil, errors.New("whole-plugin config is unavailable")
		}
		return message.Text(c.T("object.reply", "prefix", values.String("prefix"), "name", args.Name, "count", args.Count, "tail", args.Tail)), nil
	})
	return p
}

type objectLoopSink struct {
	router *router.Router
	events chan *pb.Event
}

func (s *objectLoopSink) Deliver(ctx context.Context, ev *pb.Event) error {
	select {
	case s.events <- proto.CloneOf(ev):
	case <-ctx.Done():
		return ctx.Err()
	}
	return s.router.Deliver(ctx, ev)
}

// TestPluginObjectCommand 经真实 SDK 插件、Router 与回环 Adapter 验证对象资源、子插件配置和结构体命令。
func TestPluginObjectCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	store, spec, logs := setup(t)
	registry := adapter.NewRegistry()
	sender := coremessage.New(registry, policy.New(store), nil, nil, tune.Defaults().RequestTimeout, nil)
	dispatch := router.New(router.Options{Concurrency: 1, QueueBytes: tune.Defaults().DispatchQueueBytes, QueueEvents: tune.Defaults().DispatchQueueEvents, Adapters: registry, Send: sender.Send})
	dispatch.Start()
	defer func() {
		if err := dispatch.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	sink := &objectLoopSink{router: dispatch, events: make(chan *pb.Event, 4)}
	process := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: sink, RegisterIncoming: openMediaStore(t).RegisterIncoming})
	mux := router.NewMux()
	process.Services(mux)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	if err := process.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	if err := process.StartInstance(ctx, "one", []byte(`{"account":"a"}`)); err != nil {
		t.Fatal(err)
	}
	observer := spec
	observer.PluginID = "object-plugin"
	observer.Role = state.RolePlugin
	observer.AdapterID = ""
	observer.GenerationID = "object-gen"
	observer.PluginDir = t.TempDir()
	observer.ConfigSchema = []byte(objectConfigSchema)
	observer.I18n = objectI18n()
	observer.Bootstrap = &pb.Bootstrap{GenerationId: observer.GenerationID, CandidateConfigJson: `{"part":{"prefix":"prepared"}}`, TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	observer.ExtraEnv = map[string]string{helperEnv: "plugin-object", "GORACE": "atexit_sleep_ms=0"}
	pluginCaller := router.NewCaller(observer.PluginID, observer.GenerationID)
	pluginMux := router.NewMux()
	pluginMux.Handle(wire.ServiceSend, sender.Send)
	observer.Handler = pluginMux.For(pluginCaller)
	if err := store.AddInstall(ctx, state.Install{PluginID: observer.PluginID, Role: state.RolePlugin, BuildDigest: observer.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, observer.PluginID, observer.GenerationID, observer.Bootstrap.CandidateConfigJson, 1); err != nil {
		t.Fatal(err)
	}
	plugin, err := supervisor.Launch(ctx, store, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	dispatch.Attach(router.FromInstance(plugin.Instance), pluginCaller)
	instance, _ := registry.Get("one")
	target := &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: "chat"}
	input := &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "/object Ada 3 raw tail"}}}}}
	if _, safe := instance.Send(ctx, target, input); safe != nil {
		t.Fatal(safe)
	}
	for {
		select {
		case ev := <-sink.events:
			chain := ev.GetEventPayload().GetMessageCreated().GetChain()
			if len(chain.GetSegments()) == 1 && chain.Segments[0].GetText().GetText() == "prepared|Ada|3|raw tail" {
				return
			}
		case <-ctx.Done():
			logs.mu.Lock()
			text := logs.b.String()
			logs.mu.Unlock()
			t.Fatalf("translated command reply not received: %v\n%s", ctx.Err(), text)
		}
	}
}

const optionsConfigSchema = `{"selection":{"type":"object","items":{"choice":{"type":"string","options_from":{"source":"options-plugin/choices","filter":{"group":"all"}}}}}}`

func newOptionsPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{ConfigSchema: []byte(optionsConfigSchema)})
	optionsapi.Serve(p, "choices", func(_ context.Context, q optionsapi.Query) (optionsapi.Page, error) {
		if q.Filter["group"] != "all" {
			return optionsapi.Page{}, fmt.Errorf("missing schema filter")
		}
		page := optionsapi.Page{Options: []optionsapi.Option{}}
		for _, option := range []optionsapi.Option{{Value: "one", Label: "First", Description: "choice one"}, {Value: "two", Label: "Second"}} {
			if len(q.Values) > 0 {
				found := false
				for _, value := range q.Values {
					if value == option.Value {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			} else if !strings.Contains(strings.ToLower(option.Label), strings.ToLower(q.Search)) {
				continue
			}
			page.Options = append(page.Options, option)
		}
		return page, nil
	})
	return p
}

// TestChatDirectory 验证真实 Adapter 交付后，聊天名称可供搜索、筛选并在无名事件后保留。
func TestChatDirectory(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	registry := adapter.NewRegistry()
	p := remote.New(remote.Options{AdapterID: "loop", PluginID: spec.PluginID, Store: store, Registry: registry, Sink: &eventSink{}})
	mux := router.NewMux()
	p.Services(mux)
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); result.Err != nil {
			t.Error(result.Err)
		}
	})
	if err := p.Bind(ctx, running.Instance, caller); err != nil {
		t.Fatal(err)
	}
	if err := p.StartInstance(ctx, "one", []byte(`{"account":"a"}`)); err != nil {
		t.Fatal(err)
	}
	instance, ok := registry.Get("one")
	if !ok {
		t.Fatal("missing adapter instance")
	}
	send := func(chatID, text string) {
		t.Helper()
		_, safe := instance.Send(ctx, &pb.Target{AdapterId: "loop", AdapterInstanceId: "one", AccountId: "a", ChatId: chatID}, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}})
		if safe != nil {
			t.Fatal(safe)
		}
	}
	send("group#%", "chat-directory-group")
	send("person", "chat-directory-private")
	schema, problems := configschema.Parse([]byte(`{
		"all":{"type":"string","options_from":{"source":"core.chats"}},
		"group":{"type":"string","options_from":{"source":"core.chats","filter":{"chat_type":"group"}}},
		"private":{"type":"string","options_from":{"source":"core.chats","filter":{"chat_type":"private"}}}
	}`))
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	service := options.New(options.Options{Schema: func(id string, kind options.SchemaKind) (*configschema.Schema, bool) {
		return schema, id == spec.PluginID && kind == options.ConfigSchema
	}, Chats: store.Chats, Page: tune.Defaults().OptionsPage})
	check := func(field, search, chatID, title, chatType string) {
		t.Helper()
		page, err := service.Query(ctx, spec.PluginID, options.ConfigSchema, field, options.Query{Search: search})
		want := options.Option{Value: string(message.SID("loop", chatID, "")), Label: title, Description: "loop " + chatType}
		if err != nil || page.More || !reflect.DeepEqual(page.Options, []options.Option{want}) {
			t.Fatalf("field=%s search=%s page=%+v error=%v", field, search, page, err)
		}
	}
	check("all", "gArDeN gRoUp", "group#%", "Garden Group", "group")
	check("all", "garden person", "person", "Garden Person", "private")
	check("group", "garden", "group#%", "Garden Group", "group")
	check("private", "garden", "person", "Garden Person", "private")
	send("group#%", "without-title")
	check("all", "garden group", "group#%", "Garden Group", "group")
}

// TestOptionSources 通过真实插件进程、Session、Router 和 Broker 查询配置字段的选项。
func TestOptionSources(t *testing.T) {
	ctx := t.Context()
	store, spec, _ := setup(t)
	spec.PluginID = "options-plugin"
	spec.Role = state.RolePlugin
	spec.AdapterID = ""
	spec.GenerationID = "options-gen"
	spec.PluginDir = t.TempDir()
	spec.ConfigSchema = []byte(optionsConfigSchema)
	spec.ExtraEnv = map[string]string{helperEnv: "options-plugin", "GORACE": "atexit_sleep_ms=0"}
	spec.Bootstrap = &pb.Bootstrap{GenerationId: spec.GenerationID, CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}}
	if err := store.AddInstall(ctx, state.Install{PluginID: spec.PluginID, Role: state.RolePlugin, BuildDigest: spec.BuildDigest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, spec.PluginID, spec.GenerationID, "{}", 1); err != nil {
		t.Fatal(err)
	}
	dispatch := router.New(router.Options{})
	defer dispatch.Close(context.Background())
	serviceBroker := broker.New(broker.Options{Router: dispatch, RequestTimeout: tune.Defaults().RequestTimeout, MaxDepth: tune.Defaults().CallChainMaxDepth})
	schema, problems := configschema.Parse(spec.ConfigSchema)
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	service := options.New(options.Options{Schema: func(id string, kind options.SchemaKind) (*configschema.Schema, bool) {
		return schema, id == spec.PluginID && kind == options.ConfigSchema
	}, Router: dispatch, Broker: serviceBroker, Page: tune.Defaults().OptionsPage})
	caller := router.NewCaller(spec.PluginID, spec.GenerationID)
	mux := router.NewMux()
	mux.Forward(serviceBroker.Call)
	spec.Handler = mux.For(caller)
	running, err := supervisor.Launch(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	defer running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
	dispatch.Attach(router.FromInstance(running.Instance), caller)
	page, err := service.Query(ctx, spec.PluginID, options.ConfigSchema, "selection.choice", options.Query{Search: "First", Locale: "en-US"})
	if err != nil || len(page.Options) != 1 || page.Options[0].Value != "one" || page.Options[0].Label != "First" || page.Options[0].Description != "choice one" || page.More {
		t.Fatalf("search page=%+v error=%v", page, err)
	}
	page, err = service.Query(ctx, spec.PluginID, options.ConfigSchema, "selection.choice", options.Query{Values: []string{"two", "missing"}})
	if err != nil || len(page.Options) != 1 || page.Options[0].Value != "two" || page.Options[0].Label != "Second" {
		t.Fatalf("values page=%+v error=%v", page, err)
	}
	if stopped := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE); stopped.Err != nil {
		t.Fatal(stopped.Err)
	}
	_, err = service.Query(ctx, spec.PluginID, options.ConfigSchema, "selection.choice", options.Query{})
	var invalid *validate.Error
	if !errors.As(err, &invalid) || invalid.Code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		t.Fatalf("inactive source error=%v", err)
	}
}
