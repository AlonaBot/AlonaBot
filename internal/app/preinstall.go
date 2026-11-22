package app

import (
	"bytes"
	"context"
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

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

const (
	bundledPackagesDirectory = "packages"
	officialHelp             = "alonabot-plugin-help-official"
	officialWebUI            = "alonabot-webui-admin"
	officialAdapterPrefix    = "alonabot-adapter-"
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
	if !trust.MatchesSource(id, publication.Source) {
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

// installOfficialAdapter 从对应官方仓库下载签名组件，再交给普通安装流程。
func (c *Core) installOfficialAdapter(ctx context.Context, management *manage.Service, id string) (err error) {
	key, ok := trust.Official()
	if !ok {
		return fmt.Errorf("official publisher key is unavailable; use alonabot install to install %s manually", id)
	}
	source := trust.Source(id)
	if source == "" || !strings.HasPrefix(id, officialAdapterPrefix) {
		return fmt.Errorf("unknown official adapter %q", id)
	}
	base := "https://" + source + "/releases/latest/download/"
	limits := packages.NewLimits(c.values)
	contents := make(map[string][]byte)
	for _, name := range []string{"release.json", "release.json.sig", "release.pub"} {
		raw, err := officialReleaseBytes(ctx, base+name, limits.MaxArchiveBytes)
		if err != nil {
			return err
		}
		contents[name] = raw
	}
	if len(contents["release.json.sig"]) == 0 {
		return errors.New("official adapter release requires a signature")
	}
	publisher, err := release.DecodePublicKey(contents["release.pub"])
	if err != nil {
		return err
	}
	if !bytes.Equal(publisher, key) {
		return errors.New("official adapter publisher key differs from trusted key")
	}
	publication, err := release.Verify(contents["release.json"], contents["release.json.sig"], key)
	if err != nil {
		return err
	}
	if publication.PluginID != id || !trust.MatchesSource(id, publication.Source) {
		return errors.New("official adapter release identity differs from requested component")
	}
	asset, ok := publication.Asset(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return fmt.Errorf("official adapter has no asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	address, err := url.Parse(asset.URL)
	if err != nil || address.Scheme != "https" || address.Host == "" {
		return fmt.Errorf("official adapter URL must be absolute HTTPS: %q", asset.URL)
	}
	staging := filepath.Join(c.dir, "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(staging, "official-adapter-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	archive := filepath.Join(dir, "package.tar.gz")
	file, err := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	response, err := officialReleaseResponse(ctx, asset.URL)
	if err != nil {
		return errors.Join(err, file.Close())
	}
	n, copyErr := io.Copy(file, io.LimitReader(response.Body, limits.MaxArchiveBytes+1))
	if err := errors.Join(copyErr, response.Body.Close(), file.Close()); err != nil {
		return err
	}
	if n > limits.MaxArchiveBytes {
		return errors.New("official adapter exceeds archive size limit")
	}
	_, err = management.Install(ctx, manage.InstallRequest{Archive: archive, ReleaseJSON: string(contents["release.json"]), Signature: string(contents["release.json.sig"]), PublicKey: string(contents["release.pub"])})
	return err
}

func officialReleaseBytes(ctx context.Context, address string, limit int64) ([]byte, error) {
	response, err := officialReleaseResponse(ctx, address)
	if err != nil {
		return nil, err
	}
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("official asset exceeds archive size limit: %s", address)
	}
	return raw, nil
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
