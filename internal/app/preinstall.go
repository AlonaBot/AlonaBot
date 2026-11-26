package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/trust"
)

const (
	bundledPackagesDirectory = "packages"
	officialBuiltin          = "alonabot-plugin-builtin-official"
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
	if trust.Source(id) == "" {
		return req, fmt.Errorf("unknown official component %q", id)
	}
	req.Confirm = true
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
		if slices.ContainsFunc(c.development, func(component Component) bool { return component.manifest.ID == id }) {
			continue
		}
		// 已有记录包括卸载选择；完成标记前中断也保留管理员的状态。
		if _, err := c.store.GetInstall(ctx, id); err == nil {
			continue
		} else if !errors.Is(err, state.ErrNotFound) {
			return err
		}
		req, err := officialBundledPackage(dir, id)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to preinstall bundled component %s: %v", id, err), "plugin_id", id, "error", err)
			continue
		}
		installed, err := management.Install(ctx, req)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to preinstall bundled component %s: %v", id, err), "plugin_id", id, "error", err)
			continue
		}
		if id == officialBuiltin && installed.Role == string(state.RolePlugin) || id == officialWebUI && installed.Role == string(state.RoleWebUI) {
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
adapterGroups:
	for _, adapterID := range adapterIDs {
		group := groups[adapterID]
		if len(group.Instances) == 0 {
			continue
		}
		id := group.Package
		// 开发 Adapter 稍后统一登记，预装阶段保留它对配置块的选择。
		for _, entry := range c.development {
			if entry.role == state.RoleAdapter && entry.manifest.AdapterID == adapterID && (id == "" || id == entry.manifest.ID) {
				configured[entry.manifest.ID] = true
				continue adapterGroups
			}
		}
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
			slog.Error(fmt.Sprintf("Failed to prepare adapter %s from component %s: %v", adapterID, id, err), "plugin_id", id, "adapter_id", adapterID, "error", err)
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

// installOfficialAdapter 经普通在线安装流程准备指定官方 Adapter。
func (c *Core) installOfficialAdapter(ctx context.Context, management *manage.Service, id string) error {
	source := trust.Source(id)
	if source == "" || !strings.HasPrefix(id, officialAdapterPrefix) {
		return fmt.Errorf("unknown official adapter %q", id)
	}
	_, err := management.Install(ctx, manage.InstallRequest{Source: strings.TrimPrefix(source, "github.com/"), ID: id, Confirm: true})
	return err
}
