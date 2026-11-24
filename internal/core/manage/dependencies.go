package manage

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

func (s *Service) dependencies(ctx context.Context, m *manifest.Manifest) error {
	in, err := DependencySnapshot(ctx, s.opt.Store, s.opt.Descriptor)
	if err != nil {
		return err
	}
	return CheckDependencies(m, in)
}

// DependencySnapshot 读取已启用组件的版本与当前运行 Adapter 的声明，供启动和管理共用。
func DependencySnapshot(ctx context.Context, store *state.Store, descriptor func(string) *pb.AdapterDescriptor) (manifest.Installed, error) {
	installs, err := store.Installs(ctx)
	if err != nil {
		return manifest.Installed{}, failure(err)
	}
	in := manifest.Installed{Plugins: map[string]string{}, Adapters: map[string][]manifest.ExtensionRequirement{}, Actions: map[string][]manifest.ActionRequirement{}}
	for _, installed := range installs {
		if installed.Uninstalled {
			continue
		}
		enabled, err := store.Enabled(ctx, installed.PluginID)
		if err != nil {
			return manifest.Installed{}, failure(err)
		}
		if !enabled {
			continue
		}
		if installed.Role != state.RoleAdapter {
			in.Plugins[installed.PluginID] = installed.Version
			continue
		}
		if descriptor := descriptor(installed.PluginID); descriptor != nil {
			in.Adapters[installed.AdapterID] = nil
			for _, extension := range descriptor.Extensions {
				in.Adapters[installed.AdapterID] = append(in.Adapters[installed.AdapterID], manifest.ExtensionRequirement{Name: extension.Name, Version: extension.Version})
			}
			for _, action := range descriptor.Actions {
				in.Actions[installed.AdapterID] = append(in.Actions[installed.AdapterID], manifest.ActionRequirement{Name: action.Name, Version: action.Version})
			}
		}
	}
	return in, nil
}

// CheckDependencies 汇总必需依赖的问题并返回错误，可选依赖问题按统一格式记录 WARN。
func CheckDependencies(m *manifest.Manifest, in manifest.Installed) error {
	var required []string
	for _, problem := range manifest.CheckDependencies(m, in) {
		if problem.Required {
			required = append(required, problem.Dependency+": "+problem.Reason)
		} else {
			slog.Warn(fmt.Sprintf("Optional dependency %s is unavailable for component %s: %s", problem.Dependency, m.ID, problem.Reason), "plugin_id", m.ID, "dependency", problem.Dependency, "reason", problem.Reason)
		}
	}
	if len(required) != 0 {
		return precondition("required dependencies are unavailable: " + strings.Join(required, "; "))
	}
	return nil
}

func (s *Service) affected(ctx context.Context, target state.Install, version string) ([]string, error) {
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return nil, failure(err)
	}
	var affected []string
	for _, in := range installs {
		if in.Uninstalled || in.PluginID == target.PluginID {
			continue
		}
		enabled, err := s.opt.Store.Enabled(ctx, in.PluginID)
		if err != nil {
			return nil, failure(err)
		}
		if !enabled || in.Role == state.RoleAdapter || in.Role == state.RoleWebUI && in.Entrypoint == "" {
			continue
		}
		m, err := s.manifest(ctx, in.PluginID)
		if err != nil {
			return nil, err
		}
		broken := false
		for _, dependency := range m.AdapterDependencies {
			if target.Role == state.RoleAdapter && version == "" && dependency.Required && dependency.AdapterID == target.AdapterID {
				broken = true
			}
		}
		for _, dependency := range m.PluginDependencies {
			if dependency.PluginID != target.PluginID || !dependency.Required {
				continue
			}
			if version == "" {
				broken = true
				continue
			}
			check := &manifest.Manifest{PluginDependencies: []manifest.PluginDependency{dependency}}
			if len(manifest.CheckDependencies(check, manifest.Installed{Plugins: map[string]string{target.PluginID: version}})) != 0 {
				broken = true
			}
		}
		if broken {
			affected = append(affected, in.PluginID)
		}
	}
	slices.Sort(affected)
	return affected, nil
}

func (s *Service) checkCycles(ctx context.Context, candidate *manifest.Manifest, replacing []string) error {
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return failure(err)
	}
	graph := map[string]*manifest.Manifest{candidate.ID: candidate}
	for _, in := range installs {
		if in.Uninstalled || in.PluginID == candidate.ID || slices.Contains(replacing, in.PluginID) || in.Role == state.RoleWebUI && in.Entrypoint == "" {
			continue
		}
		m, err := s.manifest(ctx, in.PluginID)
		if err != nil {
			return err
		}
		graph[m.ID] = m
	}
	visited := map[string]int{}
	var visit func(string, []string) error
	visit = func(id string, path []string) error {
		if visited[id] == 1 {
			return precondition(fmt.Sprintf("required plugin dependency cycle: %s", strings.Join(append(path, id), " -> ")))
		}
		if visited[id] == 2 || graph[id] == nil {
			return nil
		}
		visited[id] = 1
		for _, dependency := range graph[id].PluginDependencies {
			if dependency.Required {
				if err := visit(dependency.PluginID, append(path, id)); err != nil {
					return err
				}
			}
		}
		visited[id] = 2
		return nil
	}
	for id := range graph {
		if err := visit(id, nil); err != nil {
			return err
		}
	}
	return nil
}
