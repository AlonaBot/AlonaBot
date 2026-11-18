package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

const (
	bundledPackagesDirectory = "packages"
	officialSource           = "github.com/AlonaBot/AlonaBot"
	officialHelp             = "alonabot-plugin-help-official"
	officialWebUI            = "alonabot-webui-admin"
	officialAdapterPrefix    = "alonabot-adapter-"
	officialLatestRelease    = "https://api.github.com/repos/AlonaBot/AlonaBot/releases/latest"
)

func bundledDirectory() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable for bundled packages: %w", err)
	}
	return filepath.Join(filepath.Dir(executable), bundledPackagesDirectory), nil
}

func officialBundledPackage(dir, id string) (manage.InstallRequest, error) {
	req, err := manage.BundledPackage(dir, id)
	if err != nil {
		return req, err
	}
	official, ok := trust.Official()
	if !ok {
		return req, fmt.Errorf("official publisher key is unavailable; use alonabot install to install %s manually", id)
	}
	key, err := release.DecodePublicKey([]byte(req.PublicKey))
	if err != nil {
		return req, err
	}
	if !bytes.Equal(key, official) || req.Signature == "" {
		return req, fmt.Errorf("bundled component %s requires the official publisher signature", id)
	}
	publication, err := release.Verify([]byte(req.ReleaseJSON), []byte(req.Signature), official)
	if err != nil {
		return req, err
	}
	if strings.TrimPrefix(publication.Source, "https://") != officialSource {
		return req, fmt.Errorf("bundled component %s has non-official source %q", id, publication.Source)
	}
	return req, nil
}

// preinstall 在启动循环之前只登记安装和默认启用意图；进程交给 startInstalled 统一准备与提交。
func (c *Core) preinstall(ctx context.Context, management *manage.Service, dir string) error {
	done, err := c.store.Preinstalled(ctx)
	if err != nil {
		return fmt.Errorf("read preinstallation state: %w", err)
	}
	if done {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read bundled packages: %w", err)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		// 已有记录包括卸载选择；完成标记前中断也保留管理员的状态。
		if _, err := c.store.GetInstall(ctx, id); err == nil {
			continue
		} else if !errors.Is(err, state.ErrNotFound) {
			return err
		}
		req, err := officialBundledPackage(dir, id)
		if err != nil {
			slog.Error("bundled package preinstallation failed", "plugin_id", id, "error", err)
			continue
		}
		installed, err := management.Install(ctx, req)
		if err != nil {
			slog.Error("bundled package preinstallation failed", "plugin_id", id, "error", err)
			continue
		}
		if id == officialHelp && installed.Role == string(state.RolePlugin) || id == officialWebUI && installed.Role == string(state.RoleWebUI) {
			// 运行循环尚未接收管理回调；持久化意图后沿正常启动路径激活。
			if err := c.store.SetEnabled(ctx, id, true); err != nil {
				return err
			}
		}
	}
	if err := c.store.MarkPreinstalled(ctx); err != nil {
		return fmt.Errorf("mark preinstallation complete: %w", err)
	}
	return nil
}

// configuredAdapters 在运行循环开始前登记启用意图，配置或补装错误只影响所属组件。
func (c *Core) configuredAdapters(ctx context.Context, management *manage.Service, bundles string) (map[string]bool, error) {
	installs, err := c.store.Installs(ctx)
	if err != nil {
		return nil, err
	}
	groups := c.assembly.settings.Settings().Adapters
	adapterIDs := make([]string, 0, len(groups))
	for id := range groups {
		adapterIDs = append(adapterIDs, id)
	}
	slices.Sort(adapterIDs)
	configured := make(map[string]bool)
	for _, adapterID := range adapterIDs {
		group := groups[adapterID]
		if len(group.Instances) == 0 {
			continue
		}
		id := group.Package
		if id == "" {
			// 未绑定时使用当前安装的 Adapter；只有卸载记录时保留卸载选择。
			for _, in := range installs {
				if in.Role == state.RoleAdapter && in.AdapterID == adapterID && (id == "" || !in.Uninstalled) {
					id = in.PluginID
				}
			}
			if id == "" {
				id = officialAdapterPrefix + adapterID
			}
		}
		configured[id] = true
		if err := c.configureAdapter(ctx, management, bundles, id, adapterID); err != nil {
			slog.Error("configured adapter preparation failed", "plugin_id", id, "adapter_id", adapterID, "error", err)
		}
	}
	return configured, nil
}

func (c *Core) configureAdapter(ctx context.Context, management *manage.Service, bundles, id, adapterID string) error {
	in, err := c.store.GetInstall(ctx, id)
	if errors.Is(err, state.ErrNotFound) {
		if _, err := os.Stat(filepath.Join(bundles, id)); errors.Is(err, os.ErrNotExist) {
			err = c.installOfficialAdapter(ctx, management, id)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			req, err := officialBundledPackage(bundles, id)
			if err != nil {
				return err
			}
			if _, err := management.Install(ctx, req); err != nil {
				return err
			}
		}
		in, err = c.store.GetInstall(ctx, id)
	}
	if err != nil {
		return err
	}
	if in.Uninstalled {
		return fmt.Errorf("configured adapter %s is uninstalled; use alonabot install to install it manually", id)
	}
	if in.Role != state.RoleAdapter || in.AdapterID != adapterID {
		return fmt.Errorf("configured adapter %s does not provide adapter_id %s", id, adapterID)
	}
	// 已绑定到其他组件的配置由 Settings 保留；此处只处理它选择的安装身份。
	if err := c.store.SetEnabled(ctx, id, true); err != nil {
		return err
	}
	return nil
}

// installOfficialAdapter 读取官方整合包的外层发行归档，再将选中组件交给普通安装流程。
func (c *Core) installOfficialAdapter(ctx context.Context, management *manage.Service, id string) (err error) {
	if _, ok := trust.Official(); !ok {
		return fmt.Errorf("official publisher key is unavailable; use alonabot install to install %s manually", id)
	}
	// 1. 从 Release 元数据按官方归档名选择当前平台，不使用单组件仓库的下载布局。
	response, err := officialReleaseResponse(ctx, officialLatestRelease)
	if err != nil {
		return err
	}
	var publication struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	decodeErr := json.NewDecoder(response.Body).Decode(&publication)
	if err := errors.Join(decodeErr, response.Body.Close()); err != nil {
		return fmt.Errorf("read official release: %w", err)
	}
	version := strings.TrimPrefix(publication.Tag, "v")
	if !manifest.ValidVersion(version) {
		return fmt.Errorf("official release tag %q is not a SemVer version", publication.Tag)
	}
	root := fmt.Sprintf("alonabot-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	extension := ".tar.gz"
	if runtime.GOOS == "windows" {
		extension = ".zip"
	}
	var address string
	for _, asset := range publication.Assets {
		if asset.Name == root+extension {
			address = asset.URL
			break
		}
	}
	if address == "" {
		return fmt.Errorf("official release %q has no distribution for %s/%s", publication.Tag, runtime.GOOS, runtime.GOARCH)
	}
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("official distribution URL must be absolute HTTPS: %q", address)
	}
	// 2. 临时资源由本次补装独占；沿用安装器既有压缩与展开预算。
	staging := filepath.Join(c.dir, "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(staging, "official-release-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	archive := filepath.Join(dir, "distribution"+extension)
	file, err := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	response, err = officialReleaseResponse(ctx, address)
	if err != nil {
		return errors.Join(err, file.Close())
	}
	limits := packages.NewLimits(c.values)
	n, copyErr := io.Copy(file, io.LimitReader(response.Body, limits.MaxArchiveBytes+1))
	if err := errors.Join(copyErr, response.Body.Close(), file.Close()); err != nil {
		return err
	}
	if n > limits.MaxArchiveBytes {
		return errors.New("official distribution exceeds the archive size limit")
	}
	if runtime.GOOS == "windows" {
		err = extractReleaseZIP(archive, dir, root, id, limits)
	} else {
		err = packages.Extract(archive, dir, limits)
	}
	if err != nil {
		return err
	}
	// 3. 外层解压只定位文件；真实性与包内身份仍经官方公钥和 Install 校验。
	req, err := officialBundledPackage(filepath.Join(dir, root, bundledPackagesDirectory), id)
	if err != nil {
		return err
	}
	_, err = management.Install(ctx, req)
	return err
}

func officialReleaseResponse(ctx context.Context, address string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.Join(fmt.Errorf("download %s: %s", address, response.Status), response.Body.Close())
	}
	return response, nil
}

// extractReleaseZIP 只展开指定组件的平面目录；其他文件由发行构建定义，无需落盘。
func extractReleaseZIP(archive, dir, root, id string, limits packages.Limits) (err error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, reader.Close()) }()
	if len(reader.File) > limits.MaxEntries {
		return errors.New("official distribution has too many entries")
	}
	prefix := root + "/" + bundledPackagesDirectory + "/" + id + "/"
	destination := filepath.Join(dir, root, bundledPackagesDirectory, id)
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	var expanded int64
	for _, entry := range reader.File {
		if !strings.HasPrefix(entry.Name, prefix) || entry.FileInfo().IsDir() {
			continue
		}
		name := strings.TrimPrefix(entry.Name, prefix)
		if !filepath.IsLocal(name) || filepath.Base(name) != name || !entry.Mode().IsRegular() {
			return fmt.Errorf("invalid bundled ZIP entry %q", entry.Name)
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		target, err := os.OpenFile(filepath.Join(destination, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return errors.Join(err, source.Close())
		}
		n, copyErr := io.Copy(target, io.LimitReader(source, limits.MaxExpandedBytes-expanded+1))
		if err := errors.Join(copyErr, source.Close(), target.Close()); err != nil {
			return err
		}
		expanded += n
		if expanded > limits.MaxExpandedBytes {
			return errors.New("official distribution exceeds the expanded size limit")
		}
	}
	return nil
}
