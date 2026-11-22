package media

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

const (
	urlTokenBytes = 32 // URL 令牌使用 256-bit 随机值。
	mediaPrefix   = "/media/"
)

type urlGrant struct {
	entry *entry
	host  string // 仅放行签发前缀对应的主机名，支持协议端的容器地址。
}

// AddURLService 注册临时 URL 服务，保存 settings 已校验的缺省前缀。
func (s *Store) AddURLService(m *router.Mux, baseURL *url.URL) {
	prefix := *baseURL
	s.baseURL = &prefix
	s.urls = make(map[string]urlGrant)
	m.Handle(wire.ServiceMediaURL, s.operation(s.mediaURL))
}

func (s *Store) mediaURL(_ context.Context, c router.Caller, in *wire.Inbound) wire.Response {
	request := &pb.MediaURLRequest{}
	if err := decode(in, request); err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err)
	}
	prefix := s.baseURL
	if request.BaseUrl != "" {
		parsed, err := url.Parse(request.BaseUrl)
		if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Opaque != "" {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, errors.New("base_url must be an absolute http(s) URL"))
		}
		prefix = parsed
	}
	// 签发与撤销共用回收锁，URL 始终绑定同一消费租约。
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep()
	e := s.refs[request.Token]
	if e == nil || e.owner != c.SessionKey || e.lease == nil || e.file.Path == "" {
		return failure(pb.ErrorCode_ERROR_CODE_NOT_FOUND, errors.New("send media token not found"))
	}
	random := make([]byte, urlTokenBytes)
	rand.Read(random)
	id := base64.RawURLEncoding.EncodeToString(random)
	s.urls[id] = urlGrant{entry: e, host: prefix.Hostname()}
	e.lease.urls = append(e.lease.urls, id)
	address := *prefix
	address.Path = strings.TrimRight(address.Path, "/") + mediaPrefix + id
	address.RawPath, address.RawQuery, address.Fragment, address.RawFragment = "", "", "", ""
	return response(&pb.MediaURLResponse{Url: address.String(), ExpiresAtUnixMs: e.file.ExpiresAt.UnixMilli()})
}

// ServeHTTP 提供消费租约内的临时媒体，支持 GET、HEAD 和 Range。
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	id, ok := strings.CutPrefix(r.URL.Path, mediaPrefix)
	if !ok {
		http.NotFound(w, r)
		return
	}
	// 1. 在同一回收锁下校验 URL 和打开文件，避免打开已回收的对象。
	s.mu.Lock()
	s.sweep()
	grant, exists := s.urls[id]
	if !exists {
		s.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	address := url.URL{Host: r.Host}
	if !strings.EqualFold(address.Hostname(), grant.host) {
		s.mu.Unlock()
		http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}
	e := grant.entry
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		s.mu.Unlock()
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	f, err := os.Open(e.file.Path)
	if err != nil {
		s.mu.Unlock()
		slog.Error("open media file failed", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// 2. HTTP 在途期间只保留对象计数，慢客户端不会阻塞租约撤销与其他媒体操作。
	lease := e.lease
	activity, err := s.acquire(lease.caller)
	if err != nil {
		s.mu.Unlock()
		_ = f.Close()
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	if lease.upload != nil {
		lease.upload.leases++
	} else {
		lease.incoming.leases++
	}
	s.mu.Unlock()
	defer activity()
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("close media file failed", "err", err)
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.releaseObject(lease)
	}()
	w.Header().Set("Content-Type", e.file.MIME)
	http.ServeContent(w, r, "", time.Time{}, f)
}
