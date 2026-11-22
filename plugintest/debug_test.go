package plugintest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/plugintest/internal/probe"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/debug"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/media"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// TestDebugHarness 经真实 Core、认证 HTTP、SDK Session 和媒体存储验证两种调试路径。
func TestDebugHarness(t *testing.T) {
	p := probe.New()
	gate := make(chan struct{})
	p.Command("gate").Handle(func(c *sdk.Context) (message.Content, error) {
		select {
		case <-gate:
			return message.Text("released"), nil
		case <-c.Done():
			return nil, c.Err()
		}
	})
	p.Command("echo").Args("text").Handle(func(c *sdk.Context, text string) (message.Content, error) { return message.Text(text), nil })
	p.Command("invalidhandler").Handle(func(c *sdk.Context) error { return errcode.InvalidArgument.New("handler rejected input") })
	p.FullMatch("media").Handle(func(c *sdk.Context) error {
		source := c.Event().Payload.(*message.CreatedPayload).Chain[1].(message.FileSegment).Source
		reader, err := media.Open(c, source)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		ref, err := media.Upload(c, bytes.NewReader(data), "image/png")
		if err != nil {
			return err
		}
		_, err = c.Send(message.Chain{message.Text("first"), message.Image(ref.Source()), message.File(ref.Source(), "pixel.png"), message.Text("last")})
		return err
	})
	type params struct {
		Value string `json:"value"`
	}
	behavior := plugintest.Behavior{
		Send: func(_ context.Context, c plugintest.Call) (*message.SendResult, *errcode.Error) {
			result := &message.SendResult{Status: message.SendAccepted, Session: c.To, AdapterInstanceID: c.InstanceID, PlatformMessageIDs: c.MessageIDs, PrimaryPlatformMessageID: c.MessageIDs[0]}
			if c.Chain.Text() == "partial" {
				result.Status = message.SendPartial
				return result, errcode.Unavailable.New("platform accepted only part")
			}
			return result, nil
		},
		Action: func(_ context.Context, c plugintest.Call) (json.RawMessage, *errcode.Error) {
			return c.Params, errcode.Unavailable.New("platform response contains warning")
		},
	}
	h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Operations(behavior), plugintest.Adapter(plugintest.AdapterID, adapter.Action[params, params]("plugintest.echo", 1, adapter.ActionScopeChat)))
	user := plugintest.User{ID: "alice"}
	chat := plugintest.Private(user.ID)
	imageData := image.NewRGBA(image.Rect(0, 0, 2, 1))
	imageData.Set(0, 0, color.RGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, imageData); err != nil {
		t.Fatal(err)
	}
	fixture := h.Fixture(plugintest.InstanceID, "pixel", plugintest.Media{Data: encoded.Bytes(), MIME: "image/png", Filename: "pixel.png"})
	result := h.Message(chat, user, message.Chain{message.Text("media"), message.File(fixture, "in.png")})
	if len(result.Sends) != 1 || len(result.Sends[0].Media) != 2 || len(result.Sends[0].Chain) != 4 {
		t.Fatalf("media result=%+v", result)
	}
	for _, captured := range result.Sends[0].Media {
		decoded, format, err := captured.Image()
		if err != nil || format != "png" || decoded.Bounds() != imageData.Bounds() || !bytes.Equal(captured.Data, encoded.Bytes()) || captured.MIME != "image/png" {
			t.Fatalf("captured=%+v format=%s error=%v", captured, format, err)
		}
	}
	if result.Sends[0].Media[1].Filename != "pixel.png" {
		t.Fatalf("filename=%q", result.Sends[0].Media[1].Filename)
	}
	foundRead := false
	for _, c := range h.Calls() {
		if c.Operation == "media.read" && c.Name == "pixel" {
			foundRead = true
		}
	}
	if !foundRead {
		t.Fatal("incoming fixture bypassed adapter media read")
	}
	first := h.Submit(chat, user, message.Text("/echo first"))
	second := h.Submit(plugintest.Private("bob"), plugintest.User{ID: "bob"}, message.Text("/echo second"))
	for _, entry := range []struct {
		pending *plugintest.Pending
		want    string
	}{{first, "first"}, {second, "second"}} {
		result, err := entry.pending.Wait(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		result.RequireReply(t, entry.want)
		if len(result.TraceIDs) != 1 || result.Sends[0].TraceID != result.TraceIDs[0] {
			t.Fatalf("trace ownership=%+v", result)
		}
	}
	pending := h.Submit(chat, user, message.Text("/gate"))
	waitContext, cancelWait := context.WithCancel(t.Context())
	cancelWait()
	if _, err := pending.Wait(waitContext); err != context.Canceled {
		t.Fatalf("cancelled waiter=%v", err)
	}
	close(gate)
	resumed, err := pending.Wait(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	resumed.RequireReply(t, "released")
	client := h.DebugClient()
	invalid := *client
	invalid.Token = "invalid"
	identity := plugintest.InstanceConfig{AdapterID: plugintest.AdapterID, InstanceID: plugintest.InstanceID, AccountID: plugintest.AccountID}
	if connection, err := plugintest.Connect(t.Context(), &invalid, identity); err == nil {
		connection.Close()
		t.Fatal("invalid management token accepted")
	}
	connection, err := plugintest.Connect(t.Context(), client, identity)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	result, err = connection.Text(t.Context(), chat, user, "/echo actual")
	if err != nil {
		t.Fatal(err)
	}
	result.RequireReply(t, "actual")
	if len(result.Outcomes) != 1 || len(result.TraceIDs) != 1 || result.Outcomes[0].RouteID != "cmd.echo" || result.Outcomes[0].TraceID != result.TraceIDs[0] || result.Sends[0].TraceID != result.TraceIDs[0] {
		t.Fatalf("connected dispatch=%+v", result)
	}
	invalidArgs, err := connection.Text(t.Context(), chat, user, "/echo")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []plugintest.Result{h.Text(chat, user, "/echo"), invalidArgs} {
		result.RequireReply(t, "Usage: /echo <text>")
		if len(result.Outcomes) != 0 || len(result.Diagnostics) != 1 || result.Diagnostics[0].Phase != plugintest.PhaseArguments || errcode.Of(result.Diagnostics[0].Err) != errcode.InvalidArgument || len(result.TraceIDs) != 1 || result.Diagnostics[0].TraceID != result.TraceIDs[0] || result.Sends[0].TraceID != result.TraceIDs[0] {
			t.Fatalf("argument parsing=%+v", result)
		}
	}
	denied, err := connection.Text(t.Context(), chat, user, "/admin")
	if err != nil || len(denied.Sends) != 0 || len(denied.Outcomes) != 0 || len(denied.Diagnostics) != 1 || denied.Diagnostics[0].Phase != plugintest.PhaseAdmission || errcode.Of(denied.Diagnostics[0].Err) != errcode.PermissionDenied || len(denied.TraceIDs) != 1 || denied.Diagnostics[0].TraceID != denied.TraceIDs[0] {
		t.Fatalf("permission denial=%+v error=%v", denied, err)
	}
	failedHandler, err := connection.Text(t.Context(), chat, user, "/invalidhandler")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []plugintest.Result{h.Text(chat, user, "/invalidhandler"), failedHandler} {
		if len(result.Outcomes) != 1 || len(result.Diagnostics) != 0 || errcode.Of(result.Outcomes[0].Err) != errcode.InvalidArgument {
			t.Fatalf("actual Handler error=%+v", result)
		}
	}
	result, err = connection.Text(t.Context(), chat, user, "/echo partial")
	if err != nil || len(result.Sends) != 1 || result.Sends[0].Err == nil || result.Sends[0].Result.Status != message.SendPartial {
		t.Fatalf("partial reply=%+v error=%v", result, err)
	}
	action, err := connection.Client.Invoke(t.Context(), debug.Call{Instance: plugintest.InstanceID, Name: "plugintest.echo", Version: 1, To: json.RawMessage(`{"chatId":"private:alice"}`), Params: json.RawMessage(`{"value":"actual"}`)})
	if err == nil || string(action.Payload) != `{"value":"actual"}` || action.TraceID == "" {
		t.Fatalf("partial action=%+v error=%v", action, err)
	}
	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}
	h.Text(chat, user, "/ping").RequireReply(t, "pong")
	if _, err := client.Operations(t.Context(), plugintest.InstanceID, ""); err != nil {
		t.Fatalf("Core unavailable after client close: %v", err)
	}
}
