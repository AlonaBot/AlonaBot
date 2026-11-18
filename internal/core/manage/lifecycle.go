package manage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Enable 检查必需依赖并启动组件；静态 WebUI 只改变记录。
func (s *Service) Enable(ctx context.Context, id string) (err error) {
	defer finish(&err)
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	return s.enable(ctx, in)
}

func (s *Service) installed(ctx context.Context, id string) (state.Install, error) {
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err == nil && in.Uninstalled {
		err = precondition("component is uninstalled: " + id)
	}
	return in, err
}

func (s *Service) enable(ctx context.Context, in state.Install) error {
	m, err := s.manifest(ctx, in.PluginID)
	if err != nil {
		return err
	}
	if err := checkWebUIVersion(m); err != nil {
		return err
	}
	if err := s.dependencies(ctx, m); err != nil {
		return err
	}
	if in.Role != state.RoleWebUI && s.opt.State(in.PluginID) != "Activated" {
		if err := s.opt.Start(ctx, in.PluginID); err != nil {
			return err
		}
	}
	return s.opt.Store.SetEnabled(ctx, in.PluginID, true)
}

// Disable 在明确确认依赖影响后停止组件，保留安装与数据。
func (s *Service) Disable(ctx context.Context, id string, confirm bool) (err error) {
	defer finish(&err)
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	affected, err := s.affected(ctx, in, "")
	if err != nil {
		return err
	}
	if len(affected) != 0 && !confirm {
		return &ErrNeedsConfirm{Affected: affected}
	}
	if in.Role != state.RoleWebUI {
		if err := s.opt.Stop(ctx, id, pb.StopReason_STOP_REASON_DISABLE); err != nil {
			return err
		}
	}
	return s.opt.Store.SetEnabled(ctx, id, false)
}

// Restart 以当前 Active 建立新 Session，保持代次及其配置快照。
func (s *Service) Restart(ctx context.Context, id string) (err error) {
	defer finish(&err)
	unlock := s.lock(id)
	defer unlock()
	in, err := s.installed(ctx, id)
	if err != nil {
		return err
	}
	if in.Role == state.RoleWebUI {
		return precondition("webui has no process to restart")
	}
	enabled, err := s.opt.Store.Enabled(ctx, id)
	if err != nil {
		return err
	}
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return err
	}
	if !enabled || active == "" {
		return precondition("restart requires an enabled component with an active generation")
	}
	if err := s.opt.Stop(ctx, id, pb.StopReason_STOP_REASON_RESTART); err != nil {
		return err
	}
	return s.opt.Start(ctx, id)
}

// PurgeScope 选择卸载后保留的数据与配置。
type PurgeScope int

const (
	PurgeNone PurgeScope = iota // 保留数据、配置与来源绑定。
	PurgeData                   // 另外删除代次、数据库与运行数据。
	PurgeAll                    // 另外删除配置与安装绑定。
)

// Uninstall 停止组件，确认依赖与清理对象后删除选定范围的受管内容。
func (s *Service) Uninstall(ctx context.Context, id string, purge PurgeScope, confirm bool) (err error) {
	defer finish(&err)
	if purge < PurgeNone || purge > PurgeAll {
		return invalid("invalid purge scope")
	}
	s.catalog.Lock()
	defer s.catalog.Unlock()
	// 与实例编辑采用同一锁顺序，清除 Adapter 配置时不会覆盖并发创建。
	s.instances.Lock()
	defer s.instances.Unlock()
	unlock := s.lock(id)
	defer unlock()
	return s.uninstall(ctx, id, purge, confirm)
}

func (s *Service) uninstall(ctx context.Context, id string, purge PurgeScope, confirm bool) (err error) {
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return err
	}
	if in.Source == "dev" {
		return developmentError(id)
	}
	affected, err := s.affected(ctx, in, "")
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(filepath.Dir(s.opt.Packages), "plugins", id)
	group, configured := s.opt.Settings.Settings().Adapters[in.AdapterID]
	ownsInstances := configured && in.Role == state.RoleAdapter && (group.Package == id || group.Package == "" && !in.Uninstalled)
	packageDir := s.directory(in)
	packagePaths := []string{packageDir}
	journals, err := s.opt.Store.Journals(ctx, "")
	if err != nil {
		return err
	}
	for _, record := range journals {
		if record.Phase != "committed" && record.Phase != "cleaned" && (record.PluginID == id || filepath.Clean(record.PackageDir) == filepath.Clean(packageDir)) {
			packagePaths = nil // 恢复既需要候选引用，也需要安装记录中的旧程序。
			break
		}
	}
	var paths []string
	if purge != PurgeNone {
		objects, err := s.opt.Store.PurgeObjects(ctx, id, purge == PurgeAll)
		if err != nil {
			return err
		}
		if purge == PurgeAll {
			profileObjects, err := s.opt.Store.ProfilePurgeObjects(ctx, id)
			if err != nil {
				return err
			}
			objects = append(objects, profileObjects...)
		}
		for _, object := range objects {
			affected = append(affected, id+":"+object)
		}
		// 预览与删除共用受管路径清单，插件工作目录中的其余文件归作者管理。
		paths = append(paths, filepath.Join(pluginDir, "generations"))
		if purge == PurgeAll {
			paths = append(paths, filepath.Join(pluginDir, "config.json"))
			affected = append(affected, id+":installation binding")
			if ownsInstances {
				for name := range group.Instances {
					affected = append(affected, "adapter instance: "+name)
				}
			}
		}
		for _, path := range slices.Concat(paths, packagePaths) {
			if _, err := os.Stat(path); err == nil {
				affected = append(affected, path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		if len(affected) == 0 {
			affected = append(affected, id+":managed data (empty)")
		}
	}
	slices.Sort(affected)
	if len(affected) > 0 && !confirm {
		return &ErrNeedsConfirm{Affected: affected}
	}
	if ownsInstances {
		if _, err := s.opt.Settings.BindAdapter(in.AdapterID, id); err != nil {
			return err
		}
	}
	if in.Role != state.RoleWebUI {
		if err := s.opt.Stop(ctx, id, pb.StopReason_STOP_REASON_DISABLE); err != nil {
			return err
		}
	}
	if purge == PurgeAll && ownsInstances {
		saveErr := s.opt.Settings.Update(func(doc *settings.Document) error { return doc.Delete([]string{"adapters", in.AdapterID}) })
		s.opt.Adapters.SetDefaults(s.opt.Settings.Settings().DefaultInstances())
		if _, retained := s.opt.Settings.Settings().Adapters[in.AdapterID]; saveErr != nil && retained {
			return saveErr
		}
		// 设置已提交：完成剩余卸载，目录同步错误与后续错误一起返回。
		defer func() { err = errors.Join(saveErr, err) }()
		ctx = context.WithoutCancel(ctx)
		instances := make([]string, 0, len(group.Instances))
		for name := range group.Instances {
			instances = append(instances, name)
		}
		if err := s.forgetInstances(ctx, id, instances); err != nil {
			return err
		}
	}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	if err := s.opt.Store.Uninstall(ctx, id, purge != PurgeNone, purge == PurgeAll); err != nil {
		return err
	}
	for _, path := range packagePaths {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

// Pin 持久化允许更新的精确 SemVer，空串解除。
func (s *Service) Pin(ctx context.Context, id, constraint string) (err error) {
	defer finish(&err)
	if constraint != "" && !manifest.ValidVersion(constraint) {
		return invalid("pin must be an exact SemVer version")
	}
	unlock := s.lock(id)
	defer unlock()
	if _, err := s.installed(ctx, id); err != nil {
		return err
	}
	return s.opt.Store.SetPin(ctx, id, constraint)
}

func developmentError(id string) error {
	return precondition("component is a development component: " + id)
}
