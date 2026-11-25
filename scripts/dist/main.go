// dist 命令构建五平台 Core 归档，整合模式另收录官方组件。
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/releasenotes"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

const (
	defaultPlatforms = "linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64"
	versionSymbol    = "github.com/AlonaBot/AlonaBot/internal/app.releaseVersion"
	fileMode         = 0o644
	directoryMode    = 0o755
)

type component struct {
	dir, id, module string
}

var components = []component{
	{"agent", "alonabot-agent-official", "github.com/AlonaBot/alonabot-agent"},
	{"webui", "alonabot-webui-admin", "github.com/AlonaBot/alonabot-webui"},
}

type componentSource struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

type options struct {
	version, out, platforms string
	bundle                  bool
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
	flags.StringVar(&opt.out, "out", "", "output directory (required)")
	flags.StringVar(&opt.platforms, "platforms", defaultPlatforms, "comma-separated OS/arch targets")
	flags.BoolVar(&opt.bundle, "bundle", false, "include WebUI and Agent packages for the integrated image")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || opt.version == "" || opt.out == "" {
		return errors.New("--version and --out are required; positional arguments are not accepted")
	}
	// 1. 构建前检查版本与目标。
	if strings.ContainsAny(opt.version, "/\\\x00 \t\n") || opt.version == "." || opt.version == ".." {
		return errors.New("version must be a single path component")
	}
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
	out, err := filepath.Abs(opt.out)
	if err != nil {
		return err
	}
	staging, err := os.MkdirTemp("", "alonabot-dist-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()

	// 2. 整合构建为每个组件解析一次最新正式发行，让所有目标平台使用同一版本。
	var sources map[string]componentSource
	if opt.bundle {
		sources = make(map[string]componentSource)
		for _, c := range components {
			version, err := latestReleaseTag(ctx, c.module)
			if err != nil {
				return fmt.Errorf("resolve latest release for %s: %w", c.id, err)
			}
			sources[c.dir] = componentSource{Module: c.module, Version: version}
		}
	}
	roots := make(map[string]string)
	for _, target := range targets {
		goos, goarch, _ := strings.Cut(target, "/")
		root := filepath.Join(staging, fmt.Sprintf("alonabot-%s-%s-%s", opt.version, goos, goarch))
		if err := os.MkdirAll(root, directoryMode); err != nil {
			return err
		}
		if err := buildCore(ctx, filepath.Join(root, executableName(goos)), goos, goarch, opt.version); err != nil {
			return err
		}
		if err := copyFile("LICENSE", filepath.Join(root, "LICENSE"), fileMode); err != nil {
			return err
		}
		provenance, err := json.MarshalIndent(struct {
			Source     string                     `json:"source"`
			Version    string                     `json:"version"`
			Components map[string]componentSource `json:"components,omitempty"`
		}{"https://github.com/AlonaBot/AlonaBot/tree/v" + opt.version, opt.version, sources}, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "source.json"), append(provenance, '\n'), fileMode); err != nil {
			return err
		}
		roots[target] = root
	}
	if err := os.MkdirAll(out, directoryMode); err != nil {
		return err
	}
	// 3. 原样收录组件平台包，公开安装和预装使用同一份资产。
	if opt.bundle {
		for _, c := range components {
			if err := collectComponent(ctx, c, sources[c.dir].Version, out, targets, roots); err != nil {
				return fmt.Errorf("collect %s: %w", c.id, err)
			}
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
	if err := os.WriteFile(filepath.Join(out, "install.sh"), installer, fileMode); err != nil {
		return err
	}
	compose, err := os.ReadFile("docker-compose.yml")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "docker-compose.yml"), compose, fileMode); err != nil {
		return err
	}
	notes, err := releasenotes.Render("changelog", "v"+opt.version)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "release-notes.md"), notes, fileMode)
}

func buildCore(ctx context.Context, output, goos, goarch, version string) error {
	link := "-X " + versionSymbol + "=" + version + " -X main.version=" + version +
		" -X main.buildDate=" + time.Now().UTC().Format(time.DateOnly)
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

// latestReleaseTag 通过 GitHub 查询最新正式发行；后续下载使用返回的具体 tag。
func latestReleaseTag(ctx context.Context, module string) (string, error) {
	repo := strings.TrimPrefix(module, "github.com/")
	raw, err := download(ctx, "https://api.github.com/repos/"+repo+"/releases/latest")
	if err != nil {
		return "", err
	}
	var publication release.GitHubRelease
	if err := json.Unmarshal(raw, &publication); err != nil {
		return "", fmt.Errorf("parse GitHub release: %w", err)
	}
	if err := publication.Validate(repo); err != nil {
		return "", err
	}
	return publication.Tag, nil
}

// collectComponent 核对指定版本的包内身份与平台，再原样收录到发行树。
func collectComponent(ctx context.Context, component component, version, out string, targets []string, roots map[string]string) error {
	directory := filepath.Join(out, "packages", component.id)
	if err := os.MkdirAll(directory, directoryMode); err != nil {
		return err
	}
	staging, err := os.MkdirTemp("", "alonabot-collect-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	for _, target := range targets {
		goos, goarch, _ := strings.Cut(target, "/")
		name := fmt.Sprintf("%s-%s-%s-%s.tar.gz", component.id, strings.TrimPrefix(version, "v"), goos, goarch)
		raw, err := download(ctx, "https://"+component.module+"/releases/download/"+version+"/"+name)
		if err != nil {
			return err
		}
		archive := filepath.Join(staging, name)
		if err := os.WriteFile(archive, raw, fileMode); err != nil {
			return err
		}
		candidate, err := packages.Preflight(packages.Input{ID: component.id, Version: strings.TrimPrefix(version, "v"), Source: component.module, OS: goos, Arch: goarch, ArchivePath: archive}, staging, packages.NewLimits(tune.Defaults()))
		if err != nil {
			return err
		}
		expected := release.ArchiveName(candidate.Manifest)
		if err := candidate.Discard(); err != nil {
			return err
		}
		if expected != name {
			return fmt.Errorf("component archive name %q differs from manifest %q", name, expected)
		}
		for _, root := range []string{out, roots[target]} {
			destination := filepath.Join(root, "packages", component.id)
			if err := os.MkdirAll(destination, directoryMode); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(destination, name), raw, fileMode); err != nil {
				return err
			}
		}
	}
	return nil
}

func download(ctx context.Context, address string) (_ []byte, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s: %s", address, response.Status, raw)
	}
	return raw, nil
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
