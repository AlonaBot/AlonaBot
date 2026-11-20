// release-check 命令核对 dist 候选发行物，并可运行源码外安装与更新链。
//
// 检查依据 CONTRACTS §9.2、§15.2；结果与解包内容保存在独立的新目录中。
// 包的组成：资产核对见 assets.go，受控解包见 archive.go，二进制读取见 binary.go，
// 安装与更新联验见 lifecycle.go，候选进程与公开管理入口见 runtime.go。
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/release"
)

const (
	defaultPlatforms = "linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64"
	resultFile       = "result.json"
	directoryMode    = 0o700
	fileMode         = 0o600
	executableMode   = 0o700
)

type options struct {
	candidate, version, publicKey, out, platforms, assetBase       string
	installCandidate, updateCandidate, wrongSource, wrongPublisher string
	runtimeTimeout                                                 time.Duration // 调用者给出的整条功能联验截止时间。
}

type report struct {
	Scope         string               `json:"scope"`
	RuntimeChecks string               `json:"runtime_checks"`
	Status        string               `json:"status"`
	Candidate     string               `json:"candidate"`
	Version       string               `json:"core_version"`
	Platforms     []string             `json:"platforms"`
	Publisher     string               `json:"trusted_publisher_key_id,omitempty"`
	Checks        []check              `json:"checks"`
	Components    []componentResult    `json:"components"`
	Distributions []distributionResult `json:"distributions"`
	Runtime       *runtimeResult       `json:"runtime,omitempty"`
	Error         string               `json:"error,omitempty"`
}

type check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type componentResult struct {
	ID      string        `json:"id"`
	Version string        `json:"version"`
	Source  string        `json:"source"`
	Assets  []assetResult `json:"assets"`
}

type assetResult struct {
	Platform          string   `json:"platform"`
	Path              string   `json:"path"`
	Size              int64    `json:"size"`   // 归档字节数。
	SHA256            string   `json:"sha256"` // 整个归档的合同摘要。
	EmbeddedResources []string `json:"embedded_resources"`
}

type distributionResult struct {
	Platform string `json:"platform"`
	Path     string `json:"path"`
	Size     int64  `json:"size"` // 发行归档字节数。
	SHA256   string `json:"sha256"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		slog.Error("release check failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) (err error) {
	var opt options
	flags := flag.NewFlagSet("release-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opt.candidate, "candidate", "", "dist candidate directory (required)")
	flags.StringVar(&opt.version, "version", "", "Core archive version (required)")
	flags.StringVar(&opt.publicKey, "public-key", "", "independently trusted base64 Ed25519 public key file (required)")
	flags.StringVar(&opt.out, "out", "", "new evidence directory, outside candidate (required)")
	flags.StringVar(&opt.platforms, "platforms", defaultPlatforms, "comma-separated OS/arch targets; must match the candidate")
	flags.StringVar(&opt.assetBase, "asset-base", "", "expected https asset URL prefix (default: official download URL for version)")
	flags.StringVar(&opt.installCandidate, "install-candidate", "", "signed configdb example package directory; enables host runtime checks")
	flags.StringVar(&opt.updateCandidate, "update-candidate", "", "same-source, same-publisher package with a higher component SemVer")
	flags.StringVar(&opt.wrongSource, "wrong-source-candidate", "", "valid signed package with the same ID and publisher but a different source")
	flags.StringVar(&opt.wrongPublisher, "wrong-publisher-candidate", "", "valid signed package with the same ID and source but a different publisher")
	flags.DurationVar(&opt.runtimeTimeout, "runtime-timeout", 0, "required overall deadline for runtime checks, including process shutdown (e.g. 2m)")
	r := report{Scope: "asset-integrity", RuntimeChecks: "not_run", Status: "fail", Checks: []check{}, Components: []componentResult{}, Distributions: []distributionResult{}}
	outCreated := false
	// 参数错误也输出 JSON；已有证据目录始终保持原样。
	defer func() {
		if errors.Is(err, flag.ErrHelp) {
			err = nil
			return
		}
		if err != nil {
			r.Error = err.Error()
		} else {
			r.Status = "pass"
		}
		raw, marshalErr := json.MarshalIndent(r, "", "  ")
		err = errors.Join(err, marshalErr)
		if marshalErr != nil {
			return
		}
		raw = append(raw, '\n')
		if outCreated {
			if writeErr := os.WriteFile(filepath.Join(opt.out, resultFile), raw, fileMode); writeErr != nil {
				err = errors.Join(err, fmt.Errorf("write result JSON: %w", writeErr))
				r.Status, r.Error = "fail", err.Error()
				raw, _ = json.MarshalIndent(r, "", "  ") // 同一结构已经编码成功。
				raw = append(raw, '\n')
			}
		}
		_, writeErr := stdout.Write(raw)
		err = errors.Join(err, writeErr)
	}()
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || opt.candidate == "" || opt.version == "" || opt.publicKey == "" || opt.out == "" {
		return errors.New("--candidate, --version, --public-key and --out are required; positional arguments are not accepted")
	}
	runtimeRequested := opt.installCandidate != "" || opt.updateCandidate != "" || opt.wrongSource != "" || opt.wrongPublisher != "" || opt.runtimeTimeout != 0
	if runtimeRequested && (opt.installCandidate == "" || opt.updateCandidate == "" || opt.wrongSource == "" || opt.wrongPublisher == "" || opt.runtimeTimeout <= 0) {
		return errors.New("runtime checks require --install-candidate, --update-candidate, --wrong-source-candidate, --wrong-publisher-candidate and a positive --runtime-timeout")
	}
	if strings.ContainsAny(opt.version, "/\\\x00 \t\r\n") || opt.version == "." || opt.version == ".." {
		return errors.New("version must be a single path component")
	}
	r.Version = opt.version
	r.Platforms = strings.Split(opt.platforms, ",")
	slices.Sort(r.Platforms)
	for i, target := range r.Platforms {
		if !slices.Contains(strings.Split(defaultPlatforms, ","), target) {
			return fmt.Errorf("unsupported platform %q", target)
		}
		if i > 0 && r.Platforms[i-1] == target {
			return fmt.Errorf("duplicate platform %q", target)
		}
	}
	opt.candidate, err = filepath.Abs(opt.candidate)
	if err != nil {
		return err
	}
	opt.candidate, err = filepath.EvalSymlinks(opt.candidate)
	if err != nil {
		return err
	}
	opt.out, err = filepath.Abs(opt.out)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(opt.out))
	if err != nil {
		return fmt.Errorf("resolve output parent: %w", err)
	}
	opt.out = filepath.Join(parent, filepath.Base(opt.out))
	candidateInfo, err := os.Stat(opt.candidate)
	if err != nil {
		return err
	}
	if !candidateInfo.IsDir() {
		return errors.New("candidate must be a directory")
	}
	// 用文件身份覆盖 macOS 大小写别名；输出不得写回任何候选子目录。
	for dir := parent; ; dir = filepath.Dir(dir) {
		info, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if os.SameFile(candidateInfo, info) {
			return errors.New("output directory must be outside the candidate")
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}
	r.Candidate = opt.candidate
	if err := os.Mkdir(opt.out, directoryMode); err != nil {
		return fmt.Errorf("create evidence directory: %w", err)
	}
	outCreated = true
	raw, err := readRegular(opt.publicKey)
	if err != nil {
		return err
	}
	public, err := release.DecodePublicKey(raw)
	if err != nil {
		return err
	}
	r.Publisher = release.KeyID(public)
	opt.assetBase = strings.TrimRight(opt.assetBase, "/")
	if err := verifyAssets(opt, public, &r); err != nil {
		return err
	}
	if runtimeRequested {
		r.Scope, r.RuntimeChecks = "asset-integrity-and-lifecycle", "fail"
		if err := verifyLifecycle(opt, &r); err != nil {
			return err
		}
		r.RuntimeChecks = "pass"
	}
	return nil
}

func (r *report) perform(name string, fn func() error) error {
	err := fn()
	c := check{Name: name, Status: "pass"}
	if err != nil {
		c.Status, c.Error = "fail", err.Error()
	}
	r.Checks = append(r.Checks, c)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func readRegular(name string) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", name)
	}
	return os.ReadFile(name)
}
