package manage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// Install 先完成下载与预检，再确认信任及同类替换，最后登记不可变安装包。
func (s *Service) Install(ctx context.Context, req InstallRequest) (result Package, err error) {
	defer finish(&err)
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

func exists(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: message}
}

func (s *Service) prepare(ctx context.Context, req InstallRequest) (_ *packages.Candidate, err error) {
	defer func() {
		if req.Source != "" && err != nil {
			err = downloadFailure{err}
		}
	}()
	apiBase, downloadPrefix, err := s.githubEndpoints(req)
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
	staging := filepath.Join(filepath.Dir(s.opt.Packages), "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return nil, err
	}
	input := packages.Input{ID: req.ID, OS: runtime.GOOS, Arch: runtime.GOARCH, ArchivePath: req.Archive}
	var asset release.Download
	if req.Source != "" {
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
		input.Source, input.Version = "github.com/"+req.Source, strings.TrimPrefix(publication.Tag, "v")
		file, err := os.CreateTemp(staging, "download-*.tar.gz")
		if err != nil {
			return nil, err
		}
		input.ArchivePath = file.Name()
		defer os.Remove(input.ArchivePath)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadPrefix+asset.URL, nil)
		if err != nil {
			return nil, errors.Join(err, file.Close())
		}
		response, err := s.opt.HTTPClient.Do(request)
		if err != nil {
			return nil, errors.Join(err, file.Close())
		}
		if response.StatusCode != http.StatusOK {
			raw, readErr := io.ReadAll(response.Body)
			return nil, errors.Join(fmt.Errorf("download archive %s: %s: %s", request.URL, response.Status, raw), readErr, response.Body.Close(), file.Close())
		}
		n, copyErr := io.Copy(file, io.LimitReader(response.Body, s.opt.Limits.MaxArchiveBytes+1))
		if err := errors.Join(copyErr, response.Body.Close(), file.Close()); err != nil {
			return nil, err
		}
		if n > s.opt.Limits.MaxArchiveBytes {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "archive exceeds the size limit"}
		}
		if n != asset.Size {
			return nil, invalid("archive size does not match the GitHub asset")
		}
	}
	// 2. 本地与下载归档共用身份、平台和受控解压检查。
	candidate, err := packages.Preflight(input, staging, s.opt.Limits)
	if err != nil {
		return nil, err
	}
	if req.Source != "" && release.ArchiveName(candidate.Manifest) != asset.Name {
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
