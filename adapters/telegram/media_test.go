package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AlonaBot/AlonaBot/adapters/telegram/telegramtest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

func TestIncomingMedia(t *testing.T) {
	server := telegramtest.NewServer(t, "local-secret-token")
	content := []byte("a platform file read in chunks")
	server.AddFile("large-photo", "photos/large.jpg", content)
	server.AddFile("document-id", "documents/report.txt", content)
	server.Push(
		json.RawMessage(`{"update_id":1,"message":{"message_id":1,"date":123,"chat":{"id":9,"type":"private"},"photo":[{"file_id":"large-photo","width":100,"height":100,"file_size":30},{"file_id":"small-photo","width":10,"height":10}],"caption":"photo caption"}}`),
		json.RawMessage(`{"update_id":2,"message":{"message_id":2,"date":124,"chat":{"id":9,"type":"private"},"document":{"file_id":"document-id","file_size":30,"mime_type":"text/plain","file_name":"report.txt"},"caption":"document caption"}}`),
	)
	// 代理同时记录 Bot API 和下载请求，证明二者使用同一个实例客户端。
	target, err := url.Parse(server.URL())
	if err != nil {
		t.Fatal(err)
	}
	reverse := httputil.NewSingleHostReverseProxy(target)
	var getFiles, downloads atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/getFile") {
			getFiles.Add(1)
		}
		if strings.HasPrefix(r.URL.Path, "/file/") {
			downloads.Add(1)
		}
		reverse.ServeHTTP(w, r)
	}))
	t.Cleanup(proxy.Close)
	core := newRecordingCore()
	config, err := json.Marshal(map[string]string{"token": "local-secret-token", "api_base_url": "http://telegram.invalid", "proxy": proxy.URL})
	if err != nil {
		t.Fatal(err)
	}
	inst, err := New().Start(context.Background(), adapter.StartRequest{InstanceID: "media", ConfigJSON: config}, core)
	if err != nil {
		t.Fatal(err)
	}
	registerStop(t, inst)
	for range 2 {
		waitSaved(t, core.saved)
	}
	if getFiles.Load() != 0 || downloads.Load() != 0 {
		t.Fatal("received media was downloaded before read")
	}
	core.mu.Lock()
	events := append([]*pb.Event(nil), core.events...)
	core.mu.Unlock()
	if len(events) != 2 {
		t.Fatalf("events = %d", len(events))
	}
	for n, event := range events {
		segments := event.Message.Chain.Segments
		if len(segments) != 1 {
			t.Fatalf("segments = %v", segments)
		}
		var source *pb.MediaSource
		var caption string
		if n == 0 {
			source = segments[0].GetImage().GetSource()
			caption = segments[0].GetImage().GetCaption()
			if caption != "photo caption" {
				t.Fatalf("caption = %q", caption)
			}
		} else {
			file := segments[0].GetFile()
			source = file.GetSource()
			if file.GetFilename() != "report.txt" || file.GetMime() != "text/plain" || file.GetCaption() != "document caption" {
				t.Fatalf("file = %v", file)
			}
		}
		incoming := source.GetIncomingRef()
		want := []string{"large-photo", "document-id"}[n]
		if incoming.GetToken() != want || incoming.GetSize() != uint64(len(content)) {
			t.Fatalf("incoming = %v", incoming)
		}
		reader := inst.(adapter.MediaSource)
		var actual []byte
		for offset := uint64(0); ; {
			data, next, eof, err := reader.ReadPlatformFile(context.Background(), incoming.Token, offset, 7)
			if err != nil {
				t.Fatal(err)
			}
			actual = append(actual, data...)
			if next != offset+uint64(len(data)) {
				t.Fatalf("next = %d", next)
			}
			offset = next
			if eof {
				break
			}
		}
		if !bytes.Equal(actual, content) {
			t.Fatalf("content = %q", actual)
		}
	}
	if getFiles.Load() == 0 || downloads.Load() == 0 {
		t.Fatalf("getFile=%d downloads=%d", getFiles.Load(), downloads.Load())
	}
}

func TestMediaDownloadRedaction(t *testing.T) {
	const token = "local-secret-token"
	for _, stage := range []string{"getFile", "download", "redirect", "transport"} {
		t.Run(stage, func(t *testing.T) {
			var leaked atomic.Bool
			external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Store(true) }))
			defer external.Close()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/getFile") {
					if stage == "getFile" {
						_, err := io.WriteString(w, `{"ok":false,"error_code":400,"description":"bad `+token+`"}`)
						if err != nil {
							t.Error(err)
						}
						return
					}
					_, err := io.WriteString(w, `{"ok":true,"result":{"file_path":"test-file"}}`)
					if err != nil {
						t.Error(err)
					}
					return
				}
				switch stage {
				case "redirect":
					http.Redirect(w, r, external.URL+"/"+token, http.StatusFound)
				case "transport":
					connection, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Error(err)
						return
					}
					if err := connection.Close(); err != nil {
						t.Error(err)
					}
				default:
					http.Error(w, "failed "+token, http.StatusBadGateway)
				}
			}))
			defer server.Close()
			client, err := newBotClient(token, server.URL, "")
			if err != nil {
				t.Fatal(err)
			}
			defer client.http.CloseIdleConnections()
			inst := &instance{client: client}
			_, _, _, err = inst.ReadPlatformFile(context.Background(), "file-id", 0, 10)
			if err == nil || strings.Contains(err.Error(), token) {
				t.Fatalf("unsafe error: %v", err)
			}
			if leaked.Load() {
				t.Fatal("download followed redirect")
			}
		})
	}
}
