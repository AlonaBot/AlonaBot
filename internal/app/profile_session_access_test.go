//go:build unix

package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// TestProfileSessionAccess 经管理 API、真实 Adapter/插件 Session 和 Core 路由验证入出站准入。
func TestProfileSessionAccess(t *testing.T) {
	h := pipelineWithPlugin(t, "session-access")
	ctx := t.Context()
	token, err := adminauth.InitToken(ctx, h.store, h.dir)
	if err != nil {
		t.Fatal(err)
	}
	auth := adminauth.New(h.store, time.Now, &url.URL{Scheme: "http", Host: "localhost"}, h.dir)
	handler := manage.New(manage.Options{Store: h.store, Settings: h.assembly.settings, Router: h.assembly.router}).Handler(auth)
	const path = "/api/v1/profiles/limited/session-access"
	request := func(method, body, want string) {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != want {
			t.Fatalf("%s %s: got %d %s, want 200 %s", method, path, w.Code, w.Body.String(), want)
		}
	}
	if err := h.store.CreateProfile(ctx, state.Profile{ID: "limited", Name: "Limited", Color: "red"}, ""); err != nil {
		t.Fatal(err)
	}
	if err := h.store.AssignProfile(ctx, pipelineInstanceID, "", "limited"); err != nil {
		t.Fatal(err)
	}
	const session = "loop:group#topic"
	request(http.MethodGet, "", `{"mode":"all","sessions":[]}`)

	send := &pb.SendRequest{
		To:    &pb.Target{AdapterId: "loop", AdapterInstanceId: pipelineInstanceID, AccountId: pipelineAccountID, ChatId: "group", ThreadId: proto.String("topic")},
		Chain: &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "proactive"}}}}},
	}
	encoded, err := proto.Marshal(send)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, mode, listed, assignment string
		allowed                        bool
	}{
		{name: "default", allowed: true},
		{name: "allow_unlisted", mode: "allow", listed: "loop:group"},
		{name: "allow_listed", mode: "allow", listed: session, allowed: true},
		{name: "deny_listed", mode: "deny", listed: session},
		{name: "session_assignment_wins", assignment: "default", allowed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mode != "" {
				body := fmt.Sprintf(`{"mode":%q,"sessions":[{"instance":%q,"session":%q}]}`, tc.mode, pipelineInstanceID, tc.listed)
				request(http.MethodPut, body, "{}")
				request(http.MethodGet, "", body)
			}
			if tc.assignment != "" {
				if err := h.store.AssignProfile(ctx, pipelineInstanceID, session, tc.assignment); err != nil {
					t.Fatal(err)
				}
			}
			before := len(h.snapshot())
			h.inject(pipelineInput{Text: "/echo incoming", User: "alice"})
			h.idle()
			h.inject(pipelineInput{Text: "event", User: "alice"})
			h.idle()
			var want []string
			if tc.allowed {
				want = []string{"pre:/echo incoming", "handler:/echo incoming", "pre:event", "handler:event"}
			}
			if got := pipelineStages(h.snapshot()[before:], "pre", "handler"); !reflect.DeepEqual(got, want) {
				t.Fatalf("incoming dispatch = %v, want %v", got, want)
			}
			before = len(h.snapshot())
			response := h.assembly.mux.For(h.components["pipeline"].caller)(ctx, &wire.Inbound{Target: wire.ServiceSend,
				Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(send), SchemaVersion: 1, Bytes: encoded}})
			got := pipelineStages(h.snapshot()[before:], "send")
			if tc.allowed {
				if response.Err != nil || !reflect.DeepEqual(got, []string{"send:proactive"}) {
					t.Fatalf("proactive send = %v, %v", got, response.Err)
				}
			} else if response.Err.GetCode() != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || len(got) != 0 {
				t.Fatalf("denied proactive send = %v, %v", got, response.Err)
			}
		})
	}
}
