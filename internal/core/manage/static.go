package manage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// AdminAPIMajor 标识当前管理 HTTP 契约的主版本（CONTRACTS §7.5）。
const AdminAPIMajor uint32 = 1

const (
	webUIIndex = "index.html"
	iconCSP    = "sandbox; default-src 'none'; style-src 'unsafe-inline'"
)

var (
	errNoWebUI        = errors.New("no enabled WebUI; use alonabot install to install a WebUI component")
	errStaticNotFound = errors.New("static file not found")
	errStaticSymlink  = errors.New("static path is a symbolic link")
)

func checkWebUIVersion(m *manifest.Manifest) error {
	if m.Role == "webui" && (AdminAPIMajor < m.AdminAPIMinMajor || AdminAPIMajor > m.AdminAPIMaxMajor) {
		return precondition(fmt.Sprintf("Core Admin API major %d is outside WebUI supported range [%d, %d]", AdminAPIMajor, m.AdminAPIMinMajor, m.AdminAPIMaxMajor))
	}
	return nil
}

// WebUI 提供本次运行的开发静态根目录，或每次请求时选中的已启用 WebUI 包。
// 静态资源允许登录前读取，管理 API 与图标使用 Handler 的认证入口。
func (s *Service) WebUI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if reservedStaticPath(r.URL.Path) || !validStaticPath(name) {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, info, err := s.webUIContent(r.Context(), name)
		if err != nil {
			var detail *validate.Error
			switch {
			case errors.Is(err, errNoWebUI):
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			case errors.As(err, &detail) && detail.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION:
				http.Error(w, detail.Message, http.StatusServiceUnavailable)
			case errors.Is(err, errStaticNotFound):
				http.NotFound(w, r)
			default:
				staticError(w, err)
			}
			return
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(body))
	})
}

func reservedStaticPath(p string) bool {
	for _, prefix := range []string{"/api", "/plugins", "/media", "/admin", "/healthz", "/internal", "/auth"} {
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			return true
		}
	}
	return false
}

func validStaticPath(name string) bool {
	if name == "" {
		return true
	}
	return fs.ValidPath(strings.TrimSuffix(name, "/")) && !strings.ContainsAny(name, "\\:\x00")
}

func (s *Service) webUIContent(ctx context.Context, name string) ([]byte, fs.FileInfo, error) {
	if s.opt.DevelopmentWebUI != nil {
		return staticContent(s.opt.DevelopmentWebUI, "", name)
	}
	// 安装与替换共用 catalog → 包锁顺序，文件读完后释放，网络写出期间不持锁。
	s.catalog.Lock()
	defer s.catalog.Unlock()
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, in := range installs {
		if in.Role != state.RoleWebUI || in.Uninstalled {
			continue
		}
		unlock := s.lock(in.PluginID)
		defer unlock()
		enabled, err := s.opt.Store.Enabled(ctx, in.PluginID)
		if err != nil {
			return nil, nil, err
		}
		if !enabled {
			continue
		}
		m, err := s.manifest(ctx, in.PluginID)
		if err != nil {
			return nil, nil, fmt.Errorf("read WebUI manifest: %w", err)
		}
		if m.Role != string(state.RoleWebUI) {
			return nil, nil, errors.New("installed WebUI manifest has a different role")
		}
		if err := checkWebUIVersion(m); err != nil {
			return nil, nil, err
		}
		root, err := os.OpenRoot(s.directory(in))
		if err != nil {
			return nil, nil, err
		}
		defer root.Close()
		// 已安装入口损坏应报告安装故障，避免资源请求把缺失静态根当作普通 404。
		index, err := staticStat(root.FS(), m.Assets+"/"+webUIIndex)
		if err != nil {
			return nil, nil, err
		}
		if !index.Mode().IsRegular() {
			return nil, nil, errors.New("installed WebUI index is not a regular file")
		}
		return staticContent(root.FS(), m.Assets+"/", name)
	}
	return nil, nil, errNoWebUI
}

func staticContent(fsys fs.FS, prefix, name string) ([]byte, fs.FileInfo, error) {
	if name == "" {
		return readStaticFile(fsys, prefix+webUIIndex)
	}
	info, err := staticStat(fsys, prefix+strings.TrimSuffix(name, "/"))
	if err == nil {
		if info.IsDir() {
			body, index, err := readStaticFile(fsys, prefix+strings.TrimSuffix(name, "/")+"/"+webUIIndex)
			if errors.Is(err, fs.ErrNotExist) {
				err = errors.Join(errStaticNotFound, err)
			}
			return body, index, err
		}
		if strings.HasSuffix(name, "/") || !info.Mode().IsRegular() {
			return nil, nil, errStaticNotFound
		}
		return readStaticFile(fsys, prefix+name)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, err
	}
	if path.Ext(path.Base(strings.TrimSuffix(name, "/"))) != "" {
		return nil, nil, errors.Join(errStaticNotFound, err)
	}
	return readStaticFile(fsys, prefix+webUIIndex)
}

// staticStat 在打开文件前逐级拒绝链接，保留实际 I/O 错误供响应层记录。
func staticStat(fsys fs.FS, name string) (fs.FileInfo, error) {
	if err := manifest.ValidateRelPath(name); err != nil {
		return nil, err
	}
	var info fs.FileInfo
	for i := 0; i <= len(name); i++ {
		if i < len(name) && name[i] != '/' {
			continue
		}
		var err error
		info, err = fs.Lstat(fsys, name[:i])
		if err != nil {
			return nil, err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("%w: %q", errStaticSymlink, name[:i])
		}
	}
	return info, nil
}

func readStaticFile(fsys fs.FS, name string) ([]byte, fs.FileInfo, error) {
	info, err := staticStat(fsys, name)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("static path %q is not a regular file", name)
	}
	file, err := fsys.Open(name)
	if err != nil {
		return nil, nil, err
	}
	body, readErr := io.ReadAll(file)
	return body, info, errors.Join(readErr, file.Close())
}

func (s *Service) icon(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := validate.PluginID(id); err != nil {
		http.NotFound(w, r)
		return
	}
	body, info, contentType, err := s.iconContent(r.Context(), id)
	if errors.Is(err, errStaticNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		staticError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Security-Policy", iconCSP)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(body))
}

func (s *Service) iconContent(ctx context.Context, id string) ([]byte, fs.FileInfo, string, error) {
	unlock := s.lock(id)
	defer unlock()
	in, err := s.opt.Store.GetInstall(ctx, id)
	if errors.Is(err, state.ErrNotFound) || err == nil && (in.Uninstalled || in.Source == "dev") {
		return nil, nil, "", errStaticNotFound
	}
	if err != nil {
		return nil, nil, "", err
	}
	m, err := s.manifest(ctx, id)
	if err != nil {
		return nil, nil, "", err
	}
	contentType := ""
	switch path.Ext(m.Icon) {
	case ".svg":
		contentType = "image/svg+xml"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	default:
		return nil, nil, "", errStaticNotFound
	}
	root, err := os.OpenRoot(s.directory(in))
	if err != nil {
		return nil, nil, "", err
	}
	defer root.Close()
	info, err := staticStat(root.FS(), m.Icon)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, errStaticSymlink) || err == nil && !info.Mode().IsRegular() {
		return nil, nil, "", errStaticNotFound
	}
	if err != nil {
		return nil, nil, "", err
	}
	body, info, err := readStaticFile(root.FS(), m.Icon)
	return body, info, contentType, err
}

func staticError(w http.ResponseWriter, err error) {
	slog.Error("read component static file failed", "error", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
