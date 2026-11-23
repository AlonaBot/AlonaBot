package manage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

// InstallRequest 指定本地归档，或由 Core 下载的 GitHub 来源。
type InstallRequest struct {
	GitHubAPIBase        *string `json:"github_api_base,omitempty"`        // nil 沿用 Core 设置，空串使用官方 API。
	GitHubDownloadPrefix *string `json:"github_download_prefix,omitempty"` // nil 沿用 Core 设置，空串直连原资产。
	Archive              string
	Source               string // GitHub 的 owner/repo；本地安装时留空。
	ID                   string // 自动预装时指定期望的组件身份。
	Replace, Confirm     bool
	Build                bool   // 从 Source 的源码构建，而不是下载最新 Release 资产。
	Ref                  string // 构建的分支、标签或提交；空串为默认分支。仅用于 Build。

	built *builtSource // Install 与 Update 在取得目录锁之前完成的构建产物。
	log   io.Writer    // 源码构建日志；nil 时丢弃。
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// Install 先完成下载与预检，再确认信任及同类替换，最后登记不可变安装包。
func (s *Service) Install(ctx context.Context, req InstallRequest) (result Package, err error) {
	defer finish(&err)
	// 源码构建耗时较长，在目录锁外完成；安装仍走同一套预检、信任与替换检查。
	if req.Build {
		if req.built, err = s.buildSource(ctx, req, req.buildLog()); err != nil {
			return result, err
		}
		defer func() { err = errors.Join(err, req.built.remove()) }()
	}
	s.catalog.Lock()
	defer s.catalog.Unlock()
	candidate, err := s.prepare(ctx, req)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, candidate.Discard()) }()
	m := candidate.Manifest
	if err := checkWebUIVersion(m); err != nil {
		return result, err
	}
	s.instances.Lock()
	defer s.instances.Unlock()
	unlock := s.lock(m.ID)
	defer unlock()
	in := state.Install{PluginID: m.ID, Role: state.Role(m.Role), AdapterID: m.AdapterID, Source: m.Source,
		Version: m.Version, Entrypoint: m.Entrypoint, BuildDigest: candidate.Digest}
	if in.Source == "dev" {
		return result, invalid("dev source is reserved for library components")
	}
	previous, err := s.opt.Store.GetInstall(ctx, m.ID)
	if err == nil {
		if !previous.Uninstalled {
			return result, exists("component is already installed: " + m.ID)
		}
		if previous.Source != in.Source || previous.Role != in.Role || previous.AdapterID != in.AdapterID {
			return result, denied("reinstallation source or identity differs from retained binding")
		}
		if in.Role == state.RoleWebUI && in.Entrypoint == "" {
			active, err := s.opt.Store.Active(ctx, in.PluginID)
			if err != nil {
				return result, err
			}
			if active != "" {
				return result, precondition("cannot install a static WebUI while a process generation is retained; reinstall the process WebUI or explicitly remove its data with uninstall --purge-data")
			}
		}
	} else if !errors.Is(err, state.ErrNotFound) {
		return result, err
	}
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return result, err
	}
	var replaced []string
	var confirmation []string
	trusted := trust.MatchesSource(m.ID, m.Source)
	for _, old := range installs {
		if old.Source == in.Source {
			trusted = true
		}
		if old.Uninstalled || old.Role != in.Role {
			continue
		}
		if in.Role != state.RoleAgent && in.Role != state.RoleWebUI && !(in.Role == state.RoleAdapter && old.AdapterID == in.AdapterID) {
			continue
		}
		if !req.Replace {
			return result, exists("component of the same kind is already installed: " + old.PluginID)
		}
		if old.Source == "dev" {
			return result, developmentError(old.PluginID)
		}
		replaced = append(replaced, old.PluginID)
		confirmation = append(confirmation, old.PluginID)
		if old.Role == state.RoleAdapter {
			group := s.opt.Settings.Settings().Adapters[old.AdapterID]
			if group.Package == "" || group.Package == old.PluginID {
				for id := range group.Instances {
					confirmation = append(confirmation, "adapter instance (retained, not migrated): "+id)
				}
			}
		}
		affected, err := s.affected(ctx, old, "")
		if err != nil {
			return result, err
		}
		confirmation = append(confirmation, affected...)
	}
	if !trusted {
		confirmation = append(confirmation, "trust source: "+in.Source)
	}
	if len(confirmation) != 0 && !req.Confirm {
		return result, &ErrNeedsConfirm{Affected: confirmation}
	}
	if err := s.checkCycles(ctx, m, replaced); err != nil {
		return result, err
	}
	wasEnabled := false
	for _, id := range replaced {
		oldUnlock := s.lock(id)
		enabled, readErr := s.opt.Store.Enabled(ctx, id)
		if readErr == nil {
			readErr = s.uninstall(ctx, id, PurgeNone, true)
		}
		oldUnlock()
		if readErr != nil {
			return result, readErr
		}
		wasEnabled = wasEnabled || enabled
	}
	directory, err := candidate.MoveTo(s.opt.Packages)
	if err != nil {
		return result, err
	}
	if err := s.opt.Store.AddInstall(ctx, in); err != nil {
		return result, errors.Join(err, os.RemoveAll(directory))
	}
	if wasEnabled {
		if err := s.enable(ctx, in); err != nil {
			return result, err
		}
	}
	result, err = s.describe(ctx, m.ID)
	result.Warnings = candidate.Warnings
	return result, err
}

func (req InstallRequest) buildLog() io.Writer {
	if req.log == nil {
		return io.Discard
	}
	return req.log
}

func checkRepository(source string) error {
	if !repositoryName.MatchString(source) {
		return invalid("source must be owner/repo")
	}
	for _, part := range strings.Split(source, "/") {
		if part == "." || part == ".." {
			return invalid("source must be owner/repo")
		}
	}
	return nil
}

func exists(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: message}
}

func (s *Service) prepare(ctx context.Context, req InstallRequest) (_ *packages.Candidate, err error) {
	var apiBase, downloadPrefix string
	if req.built == nil {
		apiBase, downloadPrefix, err = s.githubEndpoints(req)
	}
	defer func() {
		if req.Source != "" && err != nil {
			err = downloadFailure{cause: err, prefix: downloadPrefix}
		}
	}()
	if err != nil {
		return nil, err
	}
	// 1. 下载只写临时文件；预检完成前保留已有组件。
	if req.Source != "" && req.Archive != "" {
		return nil, invalid("source and local archive are mutually exclusive")
	}
	if req.Source == "" && req.Archive == "" {
		return nil, invalid("local archive or source is required")
	}
	if req.Ref != "" && !req.Build {
		return nil, invalid("ref requires a source build")
	}
	staging := filepath.Join(filepath.Dir(s.opt.Packages), "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return nil, err
	}
	input := packages.Input{ID: req.ID, OS: runtime.GOOS, Arch: runtime.GOARCH, ArchivePath: req.Archive}
	var asset release.Download
	if req.built != nil {
		input.ArchivePath, input.Source, input.Version = req.built.archive, "github.com/"+req.Source, req.built.version
	} else if req.Source != "" {
		publication, err := s.githubRelease(ctx, req.Source, apiBase)
		if err != nil {
			return nil, err
		}
		asset, err = publication.Select(runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: err.Error()}
		}
		if err := asset.Validate(req.Source, publication.Tag); err != nil {
			return nil, denied(err.Error())
		}
		input.ArchivePath, err = s.downloadArchive(ctx, staging, downloadPrefix, asset)
		if err != nil {
			return nil, err
		}
		defer os.Remove(input.ArchivePath)
		input.Source, input.Version = "github.com/"+req.Source, strings.TrimPrefix(publication.Tag, "v")
	}
	// 2. 本地与下载归档共用身份、平台和受控解压检查。
	candidate, err := packages.Preflight(input, staging, s.opt.Limits)
	if err != nil {
		return nil, err
	}
	if req.Source != "" && req.built == nil && release.ArchiveName(candidate.Manifest) != asset.Name {
		return nil, errors.Join(invalid("archive name does not match the package manifest"), candidate.Discard())
	}
	if req.ID != "" && trust.Source(req.ID) != "" && !trust.MatchesSource(req.ID, candidate.Manifest.Source) {
		return nil, errors.Join(denied("package source differs from the official component repository"), candidate.Discard())
	}
	for _, warning := range candidate.Warnings {
		slog.Warn("package manifest warning", "plugin_id", candidate.Manifest.ID, "warning", warning)
	}
	return candidate, nil
}

func (s *Service) downloadArchive(ctx context.Context, staging, prefix string, asset release.Download) (_ string, err error) {
	// 1. 摘要由已选 API 基址提供；镜像只能提供与该摘要一致的归档。
	var expected []byte
	if asset.Digest == "" {
		if prefix != "" {
			return "", denied("GitHub asset has no digest; clear the download prefix and try again")
		}
	} else {
		algorithm, digest, _ := strings.Cut(asset.Digest, ":")
		if algorithm != "sha256" {
			return "", denied("unsupported GitHub asset digest algorithm")
		}
		expected, err = hex.DecodeString(digest)
		if err != nil || len(expected) != sha256.Size {
			return "", denied("invalid GitHub asset digest")
		}
	}
	// 2. 在同一请求期限内下载、验证并关闭归档，失败时清理暂存文件。
	file, err := os.CreateTemp(staging, "download-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(file.Name()))
		}
	}()
	digest := sha256.New()
	var dest io.Writer = file
	if expected != nil {
		dest = io.MultiWriter(file, digest)
	}
	n, err := s.copyGitHubArchive(ctx, prefix, asset.URL, dest)
	if err := errors.Join(err, file.Close()); err != nil {
		return "", err
	}
	if expected != nil && !bytes.Equal(expected, digest.Sum(nil)) {
		return "", denied("archive digest does not match the GitHub asset")
	}
	if n != asset.Size {
		return "", invalid("archive size does not match the GitHub asset")
	}
	if expected == nil {
		slog.Warn("GitHub asset has no digest; verified archive size only", "url", asset.URL)
	}
	return file.Name(), nil
}
