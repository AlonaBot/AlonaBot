// dist 命令构建 Core 和官方组件，生成签名清单及可重复的五平台发行归档。
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/release"
)

const (
	defaultPlatforms  = "linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64"
	defaultAssetBase  = "https://github.com/AlonaBot/AlonaBot/releases/download/v"
	officialKeySymbol = "github.com/AlonaBot/AlonaBot/internal/core/trust.officialKey"
	versionSymbol     = "github.com/AlonaBot/AlonaBot/internal/app.releaseVersion"
	fileMode          = 0o644
	directoryMode     = 0o755
)

const webuiDir = "web/admin"

type component struct {
	dir, id string
}

var components = []component{
	{"adapters/telegram", "alonabot-adapter-telegram"},
	{"agent", "alonabot-agent-official"},
	{webuiDir, "alonabot-webui-admin"},
	{"plugins/help", "alonabot-plugin-help-official"},
}

type options struct {
	version, key, out, platforms, assetBase string
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		slog.Error("distribution failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) (err error) {
	var opt options
	flags := flag.NewFlagSet("dist", flag.ContinueOnError)
	flags.StringVar(&opt.version, "version", "", "Core release version (required)")
	flags.StringVar(&opt.key, "key", "", "base64 Ed25519 private key file (required)")
	flags.StringVar(&opt.out, "out", "", "output directory (required)")
	flags.StringVar(&opt.platforms, "platforms", defaultPlatforms, "comma-separated OS/arch targets")
	flags.StringVar(&opt.assetBase, "asset-base", "", "absolute https asset URL prefix")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || opt.version == "" || opt.key == "" || opt.out == "" {
		return errors.New("--version, --key and --out are required; positional arguments are not accepted")
	}
	// 1. 所有外部参数与密钥在构建前检查；私钥只留在当前进程。
	if strings.ContainsAny(opt.version, "/\\\x00 \t\n") || opt.version == "." || opt.version == ".." {
		return errors.New("version must be a single path component")
	}
	if opt.assetBase == "" {
		opt.assetBase = defaultAssetBase + opt.version
	}
	base, err := url.Parse(opt.assetBase)
	if err != nil || base.Scheme != "https" || base.Host == "" || base.RawQuery != "" || base.Fragment != "" {
		return errors.New("--asset-base must be an absolute https URL without a query or fragment")
	}
	opt.assetBase = strings.TrimRight(opt.assetBase, "/")
	targets := strings.Split(opt.platforms, ",")
	slices.Sort(targets)
	for i, target := range targets {
		if !slices.Contains(strings.Split(defaultPlatforms, ","), target) {
			return fmt.Errorf("unsupported platform %q", target)
		}
		if i > 0 && targets[i-1] == target {
			return fmt.Errorf("duplicate platform %q", target)
		}
	}
	raw, err := os.ReadFile(opt.key)
	if err != nil {
		return fmt.Errorf("read signing key: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
	if err != nil {
		return fmt.Errorf("decode signing key: %w", err)
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return fmt.Errorf("private key length %d, want %d", len(decoded), ed25519.PrivateKeySize)
	}
	key := ed25519.PrivateKey(decoded)
	public := key.Public().(ed25519.PublicKey)
	out, err := filepath.Abs(opt.out)
	if err != nil {
		return err
	}
	staging, err := os.MkdirTemp("", "alonabot-dist-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()

	// 2. 主机 CLI 驱动所有 pack，交叉编译的程序只放入发行物。
	host := filepath.Join(staging, executableName(runtime.GOOS))
	if err := buildCore(ctx, host, runtime.GOOS, runtime.GOARCH, opt.version, public); err != nil {
		return err
	}
	// WebUI 在暂存副本中安装依赖并构建，不改动工作区的 node_modules 与 dist。
	webui := filepath.Join(staging, "webui")
	if err := copyWebUI(webuiDir, webui); err != nil {
		return err
	}
	if _, err := command(ctx, webui, nil, "npm", "ci"); err != nil {
		return err
	}
	if _, err := command(ctx, webui, nil, "npm", "run", "build"); err != nil {
		return err
	}
	roots := make(map[string]string)
	for _, target := range targets {
		goos, goarch, _ := strings.Cut(target, "/")
		root := filepath.Join(staging, fmt.Sprintf("alonabot-%s-%s-%s", opt.version, goos, goarch))
		if err := os.MkdirAll(filepath.Join(root, "packages"), directoryMode); err != nil {
			return err
		}
		if err := buildCore(ctx, filepath.Join(root, executableName(goos)), goos, goarch, opt.version, public); err != nil {
			return err
		}
		roots[target] = root
	}
	if err := os.MkdirAll(out, directoryMode); err != nil {
		return err
	}
	// 3. 每个组件只调用一次多平台 pack，再签名并将同一清单复制到各平台目录。
	for _, component := range components {
		if component.dir == webuiDir {
			component.dir = webui
		}
		if err := packComponent(ctx, opt, component, host, staging, out, targets, roots, key); err != nil {
			return fmt.Errorf("pack %s: %w", component.id, err)
		}
	}
	var sums strings.Builder
	for _, target := range targets {
		root := roots[target]
		extension := ".tar.gz"
		if strings.HasPrefix(target, "windows/") {
			extension = ".zip"
		}
		name := filepath.Base(root) + extension
		if err := archiveDirectory(root, filepath.Join(out, name)); err != nil {
			return err
		}
		digest, err := archiveDigest(filepath.Join(out, name))
		if err != nil {
			return err
		}
		fmt.Fprintf(&sums, "%x  %s\n", digest, name)
	}
	if err := os.WriteFile(filepath.Join(out, "SHA256SUMS"), []byte(sums.String()), fileMode); err != nil {
		return err
	}
	installer, err := os.ReadFile(filepath.Join("deploy", "install.sh"))
	if err != nil {
		return fmt.Errorf("read install script: %w", err)
	}
	return os.WriteFile(filepath.Join(out, "install.sh"), installer, fileMode)
}

func buildCore(ctx context.Context, output, goos, goarch, version string, public ed25519.PublicKey) error {
	link := "-X " + officialKeySymbol + "=" + base64.StdEncoding.EncodeToString(public) + " -X " + versionSymbol + "=" + version
	_, err := command(ctx, ".", []string{"CGO_ENABLED=0", "GOOS=" + goos, "GOARCH=" + goarch},
		"go", "build", "-trimpath", "-buildvcs=false", "-ldflags", link, "-o", output, "./cmd/alonabot")
	return err
}

func executableName(goos string) string {
	if goos == "windows" {
		return "alonabot.exe"
	}
	return "alonabot"
}

func command(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, output)
	}
	return output, nil
}

func packComponent(ctx context.Context, opt options, component component, host, staging, out string, targets []string, roots map[string]string, key ed25519.PrivateKey) error {
	directory := filepath.Join(staging, "packed", component.id)
	args := []string{"pack", component.dir, "--asset-base", opt.assetBase, "--out", directory}
	for _, target := range targets {
		args = append(args, "--target", target)
	}
	if _, err := command(ctx, ".", nil, host, args...); err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(directory, "release.json"))
	if err != nil {
		return err
	}
	publication, err := release.Verify(raw, nil, nil)
	if err != nil {
		return err
	}
	archives := make(map[string]string)
	for _, target := range targets {
		goos, goarch, _ := strings.Cut(target, "/")
		asset, ok := publication.Asset(goos, goarch)
		if !ok {
			asset, ok = publication.Asset("any", "any")
		}
		if !ok {
			return fmt.Errorf("release has no asset for %s", target)
		}
		address, err := url.Parse(asset.URL)
		if err != nil {
			return err
		}
		archive := filepath.Join(directory, filepath.Base(address.Path))
		if err := checkPrivateKey(archive, key); err != nil {
			return err
		}
		archives[target] = archive
	}
	public := key.Public().(ed25519.PublicKey)
	publication.PublisherKeyID = release.KeyID(public)
	raw, err = json.MarshalIndent(publication, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	signature := []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(key, raw)) + "\n")
	if _, err := release.Verify(raw, signature, public); err != nil {
		return fmt.Errorf("verify release: %w", err)
	}
	metadata := map[string][]byte{
		"release.json":     raw,
		"release.json.sig": signature,
		"release.pub":      []byte(base64.StdEncoding.EncodeToString(public) + "\n"),
	}
	for _, target := range targets {
		for _, directory := range []string{filepath.Join(roots[target], "packages", component.id), filepath.Join(out, "packages", component.id)} {
			if err := os.MkdirAll(directory, directoryMode); err != nil {
				return err
			}
			if err := copyFile(archives[target], filepath.Join(directory, filepath.Base(archives[target])), fileMode); err != nil {
				return err
			}
			for name, content := range metadata {
				if err := os.WriteFile(filepath.Join(directory, name), content, fileMode); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func archiveDigest(path string) (digestBytes []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

// copyWebUI 复制 WebUI 工程到暂存目录，跳过已安装的依赖与旧的构建产物。
func copyWebUI(from, to string) error {
	return filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		if entry.IsDir() && (rel == "node_modules" || rel == "dist") {
			return filepath.SkipDir
		}
		target := filepath.Join(to, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			return os.MkdirAll(target, directoryMode)
		case info.Mode().IsRegular():
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, raw, info.Mode().Perm())
		default:
			return fmt.Errorf("unsupported file in WebUI project: %s", rel)
		}
	})
}
