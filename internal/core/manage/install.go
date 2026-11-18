package manage

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
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
)

// InstallRequest 指定本地归档及清单内容，或由 Core 下载的 GitHub 来源。
type InstallRequest struct {
	Archive, ReleaseJSON, Signature string
	PublicKey                       string // release.pub 的文件内容，标准 base64。
	Source                          string // owner/repo；指定时本地文件字段须留空。
	Replace, Confirm                bool
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
	key := candidate.Release.PublisherKeyID
	// prepare 将未签名清单的公钥标识清空，安装记录只信任实际验证过的签名。
	in := state.Install{PluginID: m.ID, Role: state.Role(m.Role), AdapterID: m.AdapterID, Source: m.Source,
		Version: m.Version, PublisherKeyID: key, Entrypoint: m.Entrypoint}
	if in.Source == "dev" {
		return result, invalid("dev source is reserved for library components")
	}
	in.BuildDigest, err = hex.DecodeString(candidate.Asset.SHA256)
	if err != nil {
		return result, err
	}
	previous, err := s.opt.Store.GetInstall(ctx, m.ID)
	if err == nil {
		if !previous.Uninstalled {
			return result, exists("component is already installed: " + m.ID)
		}
		if previous.Source != in.Source || previous.Role != in.Role || previous.AdapterID != in.AdapterID {
			return result, denied("reinstallation source, publisher or identity differs from retained binding")
		}
		if err := checkPublisher(previous.PublisherKeyID, key, req.Confirm); err != nil {
			return result, err
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
	trusted := false
	for _, old := range installs {
		if old.Source == in.Source && old.PublisherKeyID == in.PublisherKeyID {
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
	if in.PublisherKeyID == "" {
		confirmation = append(confirmation, "unsigned package: "+m.ID)
	}
	if !trusted {
		confirmation = append(confirmation, "trust source: "+in.Source)
		if key != "" {
			confirmation = append(confirmation, "trust publisher: "+key)
		}
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
	if previous.Uninstalled && previous.PublisherKeyID != key {
		if err := s.opt.Store.SetPublisherKeyID(ctx, m.ID, key); err != nil {
			return result, errors.Join(err, os.RemoveAll(directory))
		}
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

func (s *Service) prepare(ctx context.Context, req InstallRequest) (*packages.Candidate, error) {
	// 1. 下载只写临时文件；校验完成前不触碰任何已有组件。
	if req.Source != "" && (req.Archive != "" || req.ReleaseJSON != "" || req.Signature != "" || req.PublicKey != "") {
		return nil, invalid("source and local package fields are mutually exclusive")
	}
	if req.Source == "" && (req.Archive == "" || req.ReleaseJSON == "") {
		return nil, invalid("local archive and release.json are required")
	}
	if req.Source != "" {
		if !repositoryName.MatchString(req.Source) {
			return nil, invalid("source must be owner/repo")
		}
		for _, part := range strings.Split(req.Source, "/") {
			if part == "." || part == ".." {
				return nil, invalid("source must be owner/repo")
			}
		}
	}
	staging := filepath.Join(filepath.Dir(s.opt.Packages), "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return nil, err
	}
	if req.Source != "" {
		base := strings.TrimRight(s.opt.ReleaseBaseURL, "/") + "/" + req.Source + "/releases/latest/download/"
		raw, err := s.download(ctx, base+"release.json", false)
		if err != nil {
			return nil, err
		}
		req.ReleaseJSON = string(raw)
		sig, err := s.download(ctx, base+"release.json.sig", true)
		if err != nil {
			return nil, err
		}
		req.Signature = string(sig)
		key, err := s.download(ctx, base+"release.pub", true)
		if err != nil {
			return nil, err
		}
		req.PublicKey = string(key)
	}
	var key ed25519.PublicKey
	if req.PublicKey != "" {
		var err error
		key, err = release.DecodePublicKey([]byte(req.PublicKey))
		if err != nil {
			return nil, denied(err.Error())
		}
	}
	r, err := release.Verify([]byte(req.ReleaseJSON), []byte(req.Signature), key)
	if err != nil {
		return nil, denied(err.Error())
	}
	if req.Source != "" && r.Source != "github.com/"+req.Source {
		return nil, denied("release source differs from requested repository")
	}
	if req.Source != "" {
		asset, ok := r.Asset("any", "any")
		if !ok {
			asset, ok = r.Asset(runtime.GOOS, runtime.GOARCH)
		}
		if !ok {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "release has no asset for this platform"}
		}
		file, err := os.CreateTemp(staging, "download-*.tar.gz")
		if err != nil {
			return nil, err
		}
		req.Archive = file.Name()
		defer os.Remove(req.Archive)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
		if err != nil {
			file.Close()
			return nil, err
		}
		response, err := s.opt.HTTPClient.Do(request)
		if err != nil {
			file.Close()
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			file.Close()
			return nil, fmt.Errorf("download archive: %s", response.Status)
		}
		n, copyErr := io.Copy(file, io.LimitReader(response.Body, s.opt.Limits.MaxArchiveBytes+1))
		if err := errors.Join(copyErr, response.Body.Close(), file.Close()); err != nil {
			return nil, err
		}
		if n > s.opt.Limits.MaxArchiveBytes {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "archive exceeds the size limit"}
		}
	}
	// 2. 同一预检同时覆盖本地与下载包；无签名只跳过真实性检查。
	candidate, err := packages.Preflight(packages.Input{ReleaseJSON: []byte(req.ReleaseJSON), Signature: []byte(req.Signature), PublicKey: key, Source: r.Source,
		OS: runtime.GOOS, Arch: runtime.GOARCH, ArchivePath: req.Archive}, staging, s.opt.Limits)
	if err != nil {
		return nil, err
	}
	if req.Signature == "" {
		candidate.Release.PublisherKeyID = ""
	}
	for _, warning := range candidate.Warnings {
		slog.Warn("package manifest warning", "plugin_id", candidate.Manifest.ID, "warning", warning)
	}
	return candidate, nil
}

func (s *Service) download(ctx context.Context, address string, optional bool) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if optional && response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", address, response.Status)
	}
	return io.ReadAll(response.Body)
}
