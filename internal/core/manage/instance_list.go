package manage

import (
	"net/http"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

type instanceSummary struct {
	InstanceID   string `json:"instance_id"`
	AdapterID    string `json:"adapter_id"`
	PackageID    string `json:"package_id"`
	PackageState string `json:"package_state"` // 所属 Adapter 组件：enabled、disabled，或 not_installed（未安装或已卸载）。
	State        string `json:"state"`
	SafeError    string `json:"safe_error"`
	Default      bool   `json:"default"`
}

// 实例所属 Adapter 组件的安装状态。停用与普通卸载保留实例配置，读取方据此提示。
const (
	packageEnabled      = "enabled"
	packageDisabled     = "disabled"
	packageNotInstalled = "not_installed"
)

func (s *Service) listInstances(w http.ResponseWriter, r *http.Request) {
	s.instances.Lock()
	defer s.instances.Unlock()
	installs, err := s.opt.Store.Installs(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	result := make([]instanceSummary, 0)
	settings := s.opt.Settings.Settings()
	defaults := settings.DefaultInstances()
	for adapterID, group := range settings.Adapters {
		owner := group.Package
		if owner == "" {
			for _, in := range installs {
				if in.Role == state.RoleAdapter && in.AdapterID == adapterID && !in.Uninstalled {
					owner = in.PluginID
					break
				}
			}
		}
		packageState := packageNotInstalled
		for _, in := range installs {
			if in.PluginID == owner && !in.Uninstalled {
				enabled, err := s.opt.Store.Enabled(r.Context(), owner)
				if err != nil {
					respond(w, nil, err)
					return
				}
				packageState = packageDisabled
				if enabled {
					packageState = packageEnabled
				}
			}
		}
		for id := range group.Instances {
			item := instanceSummary{InstanceID: id, AdapterID: adapterID, PackageID: owner, PackageState: packageState, State: "stopped", Default: defaults[adapterID] == id}
			if s.opt.InstanceStatus != nil {
				item.State, item.SafeError, err = s.opt.InstanceStatus(r.Context(), owner, id)
				if err != nil {
					respond(w, nil, err)
					return
				}
			} else if _, running := s.opt.Adapters.Get(id); running {
				item.State = "running"
			}
			result = append(result, item)
		}
	}
	slices.SortFunc(result, func(a, b instanceSummary) int { return strings.Compare(a.InstanceID, b.InstanceID) })
	respond(w, struct {
		Instances []instanceSummary `json:"instances"`
	}{result}, nil)
}
