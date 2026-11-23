package manage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// 源码构建在 data_dir 下使用的目录。
const (
	toolchainDirectory = "toolchain" // 按需下载的 Go 发行版，GOROOT 为其中的 go/。
	cacheDirectory     = "cache"     // GOCACHE、GOMODCACHE、GOTMPDIR 与 HOME。
)

// goRelease 是 Go 下载目录 JSON 列表（?mode=json&include=all）中的一个文件。
type goRelease struct {
	Version string `json:"version"`
	Files   []struct {
		Filename string `json:"filename"`
		SHA256   string `json:"sha256"`
		Size     int64  `json:"size"`
	} `json:"files"`
}

func (s *Service) dataDirectory() string { return filepath.Dir(s.opt.Packages) }

// goVersion 返回 Core 自身的 Go 发行版本，源码构建下载同一版本作为起始工具链。
func goVersion() (string, error) {
	version := strings.Fields(runtime.Version())[0]
	if !strings.HasPrefix(version, "go1.") {
		return "", precondition("Core was built with a development Go toolchain " + version + "; source builds need a released Go version")
	}
	return version, nil
}

// installedGoVersion 返回已下载工具链 VERSION 文件的首行；未下载时为空串。
func (s *Service) installedGoVersion() (string, error) {
	raw, err := os.ReadFile(filepath.Join(s.dataDirectory(), toolchainDirectory, "go", "VERSION"))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	version, _, _ := strings.Cut(string(raw), "\n")
	return version, nil
}

// goToolchain 返回源码构建使用的 go 命令与隔离环境；工具链缺失或版本不同时先下载。
// 调用方持有 s.builds。
func (s *Service) goToolchain(ctx context.Context, log io.Writer) (generate.Toolchain, error) {
	// 1. 复用版本一致的已下载工具链。
	if runtime.GOOS == "windows" {
		return generate.Toolchain{}, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "source builds are not supported on Windows"}
	}
	version, err := goVersion()
	if err != nil {
		return generate.Toolchain{}, err
	}
	data := s.dataDirectory()
	root := filepath.Join(data, toolchainDirectory, "go")
	installed, err := s.installedGoVersion()
	if err != nil {
		return generate.Toolchain{}, err
	}
	if installed != version {
		if err := s.downloadGo(ctx, version, log); err != nil {
			return generate.Toolchain{}, err
		}
	}
	// 2. 缓存与临时目录全部放在 data_dir 下：scratch 镜像没有 /tmp，也不读取宿主的 Go 环境。
	cache := filepath.Join(data, cacheDirectory)
	paths := map[string]string{"GOCACHE": "go-build", "GOMODCACHE": "mod", "GOTMPDIR": "tmp", "HOME": "home"}
	v := s.opt.Tuning
	env := []string{"GOROOT=", "GOPATH=", "GOENV=off", "GOWORK=off", "GOFLAGS=-modcacherw -p=1", "GOTOOLCHAIN=auto", "CGO_ENABLED=0",
		fmt.Sprintf("GOMEMLIMIT=%dB", v.BuildMemorySoftLimit), fmt.Sprintf("GOGC=%d", v.BuildGOGC),
		"GOPROXY=" + s.opt.Settings.Settings().GoProxy, "TMPDIR=" + filepath.Join(cache, "tmp")}
	for name, directory := range paths {
		path := filepath.Join(cache, directory)
		if err := os.MkdirAll(path, 0o700); err != nil {
			return generate.Toolchain{}, err
		}
		env = append(env, name+"="+path)
	}
	return generate.Toolchain{Go: filepath.Join(root, "bin", "go"), Env: env, Stderr: log, StallInterval: v.BuildStallInterval}, nil
}

// downloadGo 按下载目录的 JSON 列表核对大小与 SHA-256，解压后替换 data_dir/toolchain/go。
func (s *Service) downloadGo(ctx context.Context, version string, log io.Writer) (err error) {
	// 1. 从列表取得本机平台归档的摘要。
	base := s.opt.Settings.Settings().GoDownloadBase
	name := version + "." + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	listing := base + "?mode=json&include=all"
	response, err := s.get(ctx, listing)
	if err != nil {
		return err
	}
	var releases []goRelease
	decodeErr := json.NewDecoder(response.Body).Decode(&releases)
	if err := errors.Join(decodeErr, response.Body.Close()); err != nil {
		return fmt.Errorf("read Go download list %s: %w", listing, err)
	}
	var want []byte
	var size int64
	for _, release := range releases {
		for _, file := range release.Files {
			if release.Version == version && file.Filename == name {
				want, err = hex.DecodeString(file.SHA256)
				if err != nil || len(want) != sha256.Size {
					return fmt.Errorf("invalid SHA-256 for %s in %s", name, listing)
				}
				size = file.Size
			}
		}
	}
	if want == nil {
		return fmt.Errorf("Go download list %s has no %s", listing, name)
	}
	// 2. 下载到工具链目录下的临时文件，核对后解压到临时目录，最后整体替换。
	toolchains := filepath.Join(s.dataDirectory(), toolchainDirectory)
	if err := os.MkdirAll(toolchains, 0o700); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(toolchains, ".download-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()
	fmt.Fprintf(log, "download Go %s from %s (%d bytes)\n", version, base+name, size)
	response, err = s.get(ctx, base+name)
	if err != nil {
		return err
	}
	archive := filepath.Join(staging, name)
	file, err := os.Create(archive)
	if err != nil {
		return errors.Join(err, response.Body.Close())
	}
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(file, hash), response.Body)
	if err := errors.Join(copyErr, response.Body.Close(), file.Close()); err != nil {
		return fmt.Errorf("download %s: %w", base+name, err)
	}
	if n != size || !bytes.Equal(hash.Sum(nil), want) {
		return denied(fmt.Sprintf("%s does not match the size and SHA-256 in the Go download list", name))
	}
	tree := filepath.Join(staging, "go")
	if err := unpackFile(archive, tree, math.MaxInt, math.MaxInt64); err != nil {
		return fmt.Errorf("unpack %s: %w", name, err)
	}
	root := filepath.Join(toolchains, "go")
	if err := os.RemoveAll(root); err != nil {
		return err
	}
	if err := os.Rename(tree, root); err != nil {
		return err
	}
	fmt.Fprintf(log, "installed Go %s at %s\n", version, root)
	return nil
}

// get 发送 GET 请求，非 200 响应返回带状态的错误；成功时由调用方关闭响应体。
func (s *Service) get(ctx context.Context, address string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, errors.Join(fmt.Errorf("download %s: %s: %s", address, response.Status, body), response.Body.Close())
	}
	return response, nil
}

// buildEnvironment 是源码构建工具链与缓存的磁盘占用。
type buildEnvironment struct {
	GoVersion      string `json:"go_version"` // 已下载工具链的版本；未下载时为空串。
	ToolchainBytes int64  `json:"toolchain_bytes"`
	CacheBytes     int64  `json:"cache_bytes"`
}

func (s *Service) buildEnvironment() (buildEnvironment, error) {
	version, err := s.installedGoVersion()
	if err != nil {
		return buildEnvironment{}, err
	}
	toolchain, err := directorySize(filepath.Join(s.dataDirectory(), toolchainDirectory))
	if err != nil {
		return buildEnvironment{}, err
	}
	cache, err := directorySize(filepath.Join(s.dataDirectory(), cacheDirectory))
	return buildEnvironment{GoVersion: version, ToolchainBytes: toolchain, CacheBytes: cache}, err
}

func directorySize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil || !entry.Type().IsRegular() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

func (s *Service) buildEnvironmentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/core/build-environment", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.buildEnvironment()
		respond(w, result, err)
	})
	// part=toolchain 或 cache 只删除一部分，省略时两者都删除；源码构建运行时返回 412。
	mux.HandleFunc("DELETE /api/v1/core/build-environment", func(w http.ResponseWriter, r *http.Request) {
		parts := []string{toolchainDirectory, cacheDirectory}
		switch part := r.URL.Query().Get("part"); part {
		case "":
		case toolchainDirectory, cacheDirectory:
			parts = []string{part}
		default:
			respond(w, nil, invalid("part must be toolchain or cache"))
			return
		}
		if !s.builds.TryLock() {
			respond(w, nil, precondition("a source build is running"))
			return
		}
		var err error
		for _, part := range parts {
			err = errors.Join(err, os.RemoveAll(filepath.Join(s.dataDirectory(), part)))
		}
		s.builds.Unlock()
		if err != nil {
			respond(w, nil, err)
			return
		}
		result, err := s.buildEnvironment()
		respond(w, result, err)
	})
}
