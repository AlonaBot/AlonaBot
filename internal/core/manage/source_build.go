package manage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"go.yaml.in/yaml/v3"
)

// shortCommitLength 是非标签构建写入版本 build metadata 的提交前缀长度，与 Go 伪版本一致。
const shortCommitLength = 12

var (
	commitName = regexp.MustCompile(`^[0-9a-f]{40}$`)
	// refName 只允许分支、标签与提交常用的字符，拼进 API 路径前在此拒绝其余输入。
	refName = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,255}$`)
)

// builtSource 是一次源码构建产出的组件归档；dir 是 data_dir/staging 下的独占目录。
type builtSource struct {
	dir, archive, version string
}

func (b *builtSource) remove() error { return os.RemoveAll(b.dir) }

// buildSource 下载 GitHub 源码 tarball，并用按需下载的 Go 工具链为本机平台打包。
// 标签构建的版本取标签，并由预检核对 plugin.yaml；其他提交的版本为 plugin.yaml 的版本加
// 「+<12 位提交>」的 build metadata。构建之间串行执行。
func (s *Service) buildSource(ctx context.Context, req InstallRequest, log io.Writer) (_ *builtSource, err error) {
	// 1. 校验请求并解析提交；标签按 GitHub 的 tag ref 判断。
	if req.Archive != "" || req.Source == "" {
		return nil, invalid("source builds require source and no local archive")
	}
	if err := checkRepository(req.Source); err != nil {
		return nil, err
	}
	ref := req.Ref
	if ref == "" {
		ref = "HEAD"
	}
	if !refName.MatchString(ref) || strings.Contains(ref, "..") {
		return nil, invalid("invalid ref")
	}
	apiBase, prefix, err := s.githubEndpoints(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = downloadFailure{cause: err, prefix: prefix}
		}
	}()
	if !s.builds.TryLock() {
		fmt.Fprintln(log, "waiting for another source build")
		s.builds.Lock()
	}
	defer s.builds.Unlock()
	api := strings.TrimRight(apiBase, "/") + "/repos/" + req.Source
	fmt.Fprintf(log, "resolve %s@%s\n", req.Source, ref)
	commit, _, err := s.readGitHubAPI(ctx, api+"/commits/"+ref, "application/vnd.github.sha")
	if err != nil {
		return nil, err
	}
	if !commitName.Match(commit) {
		return nil, fmt.Errorf("GitHub returned an invalid commit for %s: %q", ref, commit)
	}
	version := ""
	if tag := strings.TrimPrefix(req.Ref, "v"); req.Ref != "" && manifest.ValidVersion(tag) {
		if _, status, err := s.readGitHubAPI(ctx, api+"/git/ref/tags/"+req.Ref, "application/vnd.github+json"); err == nil {
			version = tag
		} else if status != http.StatusNotFound {
			return nil, err
		}
	}
	// 2. 下载并解压源码；GitHub 不公布源码 tarball 的摘要。
	staging := filepath.Join(s.dataDirectory(), "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(staging, "source-")
	if err != nil {
		return nil, err
	}
	built := &builtSource{dir: dir}
	defer func() {
		if err != nil {
			err = errors.Join(err, built.remove())
		}
	}()
	tarball := filepath.Join(dir, "source.tar.gz")
	file, err := os.Create(tarball)
	if err != nil {
		return nil, err
	}
	address := "https://github.com/" + req.Source + "/archive/" + string(commit) + ".tar.gz"
	fmt.Fprintf(log, "download %s\n", address)
	if prefix != "" {
		fmt.Fprintln(log, "WARNING: Source archive is fetched through the GitHub download prefix without digest verification.")
	}
	_, err = s.copyGitHubArchive(ctx, prefix, address, file)
	if err := errors.Join(err, file.Close()); err != nil {
		return nil, err
	}
	source := filepath.Join(dir, "source")
	if err := unpackFile(tarball, source, s.opt.Limits.MaxEntries, s.opt.Limits.MaxExpandedBytes); err != nil {
		return nil, fmt.Errorf("unpack source: %w", err)
	}
	if version == "" {
		if version, err = stampCommit(filepath.Join(source, "plugin.yaml"), string(commit)[:shortCommitLength]); err != nil {
			return nil, err
		}
	}
	// 3. 为本机平台打包；cgo 关闭，模块经 goproxy 下载。
	toolchain, err := s.goToolchain(ctx, log)
	if err != nil {
		return nil, err
	}
	if runtime.GOOS == "linux" {
		threshold := s.opt.Tuning.BuildMemoryWarnBelow
		if available, ok := availableBuildMemory(os.DirFS("/")); ok && available < threshold {
			fmt.Fprintf(log, "WARNING: %d MiB of memory is available (below %d MiB). Small plugins build fine; components with large dependencies, such as LLM SDKs, may need about 2 GiB and can fail.\n", available>>20, threshold>>20)
		}
	}
	fmt.Fprintf(log, "build %s %s for %s/%s\n", req.Source, version, runtime.GOOS, runtime.GOARCH)
	built.archive, err = pluginkit.Pack(ctx, generate.Options{Dir: source, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Toolchain: toolchain}, filepath.Join(dir, "out"))
	if err != nil {
		return nil, err
	}
	built.version = version
	fmt.Fprintf(log, "packed %s\n", filepath.Base(built.archive))
	return built, nil
}

// stampCommit 把 plugin.yaml 的 version 改为「原版本（去掉 build metadata）+commit」，返回新版本。
func stampCommit(name, commit string) (string, error) {
	raw, err := os.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read source plugin.yaml: %w", err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return "", invalid("parse source plugin.yaml: " + err.Error())
	}
	if len(document.Content) == 1 && document.Content[0].Kind == yaml.MappingNode {
		fields := document.Content[0].Content
		for i := 0; i+1 < len(fields); i += 2 {
			if fields[i].Value == "version" {
				base, _, _ := strings.Cut(fields[i+1].Value, "+")
				version := base + "+" + commit
				if !manifest.ValidVersion(version) {
					return "", invalid(fmt.Sprintf("source plugin.yaml version %q is not SemVer", fields[i+1].Value))
				}
				fields[i+1].Value = version
				stamped, err := yaml.Marshal(&document)
				if err != nil {
					return "", err
				}
				return version, os.WriteFile(name, stamped, 0o600)
			}
		}
	}
	return "", invalid("source plugin.yaml has no version")
}

// unpackFile 解压 tar.gz 到新建的 dest，去掉所有条目共同的顶层目录（GitHub tarball 与 Go 发行版都有一层）。
// 只接受目录与普通文件，跳过 pax 全局头；路径须是规范相对路径。条目数与普通文件总字节数超过上限时失败。
// 失败时 dest 可能留有部分文件，由调用方删除。
func unpackFile(archivePath, dest string, maxEntries int, maxBytes int64) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	compressed, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	if err := os.Mkdir(dest, 0o700); err != nil {
		return err
	}
	top := ""
	entries, expanded := 0, int64(0)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		if entries++; entries > maxEntries {
			return errors.New("archive has too many entries")
		}
		name := strings.TrimSuffix(header.Name, "/")
		first, rest, _ := strings.Cut(name, "/")
		if top == "" {
			top = first
		}
		if first != top || strings.ContainsAny(name, "\\:\x00") || path.Clean(name) != name || !filepath.IsLocal(filepath.FromSlash(name)) {
			return fmt.Errorf("archive path %q is not under a single top-level directory", header.Name)
		}
		if rest == "" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(rest))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
		case tar.TypeReg:
			if expanded += header.Size; expanded > maxBytes {
				return errors.New("archive expanded size exceeds the limit")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			mode := os.FileMode(0o600)
			if header.Mode&0o111 != 0 {
				mode = 0o700
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, reader)
			if err := errors.Join(copyErr, file.Close()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("archive entry %q is not a regular file or directory", header.Name)
		}
	}
}
