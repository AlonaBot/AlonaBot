package manage

import (
	"context"
	"net/http"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
)

type instanceAccess struct {
	All       bool     `json:"all"`
	Instances []string `json:"instances"`
}

func (s *Service) accessRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/allowed-instances", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := s.installed(r.Context(), id); err != nil {
			respond(w, nil, err)
			return
		}
		all, instances, err := s.opt.Store.AdapterInstanceAccess(r.Context(), id)
		if instances == nil {
			instances = []string{}
		}
		respond(w, instanceAccess{all, instances}, err)
	})
	mux.HandleFunc("PUT /api/v1/plugins/{id}/allowed-instances", func(w http.ResponseWriter, r *http.Request) {
		request := instanceAccess{All: true}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		for _, id := range request.Instances {
			if !instanceName.MatchString(id) {
				respond(w, nil, invalid("invalid adapter instance name"))
				return
			}
		}
		id := r.PathValue("id")
		unlock := s.lock(id)
		defer unlock()
		_, err := s.installed(r.Context(), id)
		if err == nil {
			err = s.opt.Store.SetAdapterInstanceAccess(r.Context(), id, request.All, request.Instances)
		}
		respond(w, struct{}{}, err)
	})
}

// commandScope 沿用策略的三层作用域，HTTP 用稳定字符串表示层级。
type commandScope struct {
	Kind    string   `json:"kind"`
	Group   []string `json:"group,omitempty"`
	RouteID string   `json:"route_id,omitempty"`
}
type commandOverride struct {
	Scope       commandScope `json:"scope"`
	Permissions []string     `json:"permissions"`
	Enabled     *bool        `json:"enabled"`
}
type invalidCommandName struct {
	RouteID string   `json:"route_id"`
	Name    []string `json:"name"`
	Reason  string   `json:"reason"`
}
type commandPolicy struct {
	RouteID          string               `json:"route_id"`
	Enabled          bool                 `json:"enabled"`
	Permissions      []string             `json:"permissions"`
	DeclaredNames    [][]string           `json:"declared_names"`
	EffectiveNames   [][]string           `json:"effective_names"`
	HasNamesOverride bool                 `json:"has_names_override"`
	InvalidNames     []invalidCommandName `json:"invalid_names"`
}
type commandReference struct {
	PluginID string `json:"plugin_id"`
	RouteID  string `json:"route_id"`
}
type commandShadow struct {
	Path     []string           `json:"path"`
	Winner   commandReference   `json:"winner"`
	Shadowed []commandReference `json:"shadowed"`
}
type pluginPolicies struct {
	Overrides    []commandOverride    `json:"overrides"`
	Commands     []commandPolicy      `json:"commands"`
	Shadows      []commandShadow      `json:"shadows"`
	InvalidNames []invalidCommandName `json:"invalid_names"`
}

func (s *Service) policyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/plugins/{id}/policies", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.readPolicies(r.Context(), r.PathValue("id"))
		respond(w, result, err)
	})
	mux.HandleFunc("PUT /api/v1/plugins/{id}/policies", func(w http.ResponseWriter, r *http.Request) {
		var request commandOverride
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		scope := policy.Scope{Group: request.Scope.Group, RouteID: request.Scope.RouteID}
		switch request.Scope.Kind {
		case "plugin":
			scope.Kind = policy.ScopePlugin
		case "group":
			scope.Kind = policy.ScopeGroup
		case "route":
			scope.Kind = policy.ScopeRoute
		default:
			respond(w, nil, invalid("unknown command scope kind"))
			return
		}
		id := r.PathValue("id")
		unlock := s.lock(id)
		defer unlock()
		_, err := s.installed(r.Context(), id)
		if err == nil {
			err = s.opt.Policy.SetOverride(r.Context(), id, scope, policy.Override{Permissions: request.Permissions, Enabled: request.Enabled})
		}
		respond(w, struct{}{}, err)
	})
	for _, method := range []string{"PUT", "DELETE"} {
		mux.HandleFunc(method+" /api/v1/plugins/{id}/commands/{route_id}/names", func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				Names [][]string `json:"names"`
			}
			if method == "PUT" {
				if err := decode(r, &request); err != nil {
					respond(w, nil, err)
					return
				}
			}
			id, routeID := r.PathValue("id"), r.PathValue("route_id")
			unlock := s.lock(id)
			defer unlock()
			routes, err := s.commandRoutes(r.Context(), id)
			if err == nil {
				if method == "PUT" {
					err = s.opt.Policy.SetCommandNames(r.Context(), id, routeID, request.Names, routes)
				} else {
					index := slices.IndexFunc(routes, func(route *pb.RouteDescriptor) bool { return route.GetRouteId() == routeID })
					if index >= 0 && (routes[index].GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND || routes[index].GetCommand().GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH) {
						err = invalid("route is not a registered path command: " + routeID)
					} else {
						err = s.opt.Policy.DeleteCommandNames(r.Context(), id, routeID)
					}
				}
			}
			respond(w, struct{}{}, err)
		})
	}
}

func (s *Service) commandRoutes(ctx context.Context, id string) ([]*pb.RouteDescriptor, error) {
	if _, err := s.installed(ctx, id); err != nil {
		return nil, err
	}
	plugin, ok := s.opt.Router.Plugin(id)
	if !ok {
		return []*pb.RouteDescriptor{}, nil
	}
	return plugin.Routes(), nil
}

func (s *Service) readPolicies(ctx context.Context, id string) (pluginPolicies, error) {
	unlock := s.lock(id)
	defer unlock()
	result := pluginPolicies{Overrides: []commandOverride{}, Commands: []commandPolicy{}, Shadows: []commandShadow{}, InvalidNames: []invalidCommandName{}}
	routes, err := s.commandRoutes(ctx, id)
	if err != nil {
		return result, err
	}
	// 1. 保留所有作用域的显式覆盖，包含当前已无声明的路由。
	overrides, err := s.opt.Policy.Overrides(ctx, id)
	if err != nil {
		return result, err
	}
	for _, entry := range overrides {
		scope := commandScope{Group: entry.Scope.Group, RouteID: entry.Scope.RouteID}
		switch entry.Scope.Kind {
		case policy.ScopePlugin:
			scope.Kind = "plugin"
		case policy.ScopeGroup:
			scope.Kind = "group"
		case policy.ScopeRoute:
			scope.Kind = "route"
		}
		result.Overrides = append(result.Overrides, commandOverride{scope, entry.Override.Permissions, entry.Override.Enabled})
	}
	// 2. 权限复用派发规则，名称与失效诊断读取 Router 的生效索引。
	index := s.opt.Router.Commands()
	names := s.opt.Policy.CommandNames(id)
	for _, entry := range index.InvalidNames() {
		if entry.PluginID == id {
			result.InvalidNames = append(result.InvalidNames, invalidCommandName{entry.RouteID, entry.Name, entry.Reason})
		}
	}
	for _, route := range routes {
		if route.GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND {
			continue
		}
		enabled, permissions, err := s.opt.Policy.CommandState(ctx, id, route)
		if err != nil {
			return result, err
		}
		view := commandPolicy{RouteID: route.GetRouteId(), Enabled: enabled, Permissions: permissions, DeclaredNames: [][]string{}, EffectiveNames: [][]string{}, InvalidNames: []invalidCommandName{}}
		_, view.HasNamesOverride = names[view.RouteID]
		if route.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_PATH {
			view.DeclaredNames = append(view.DeclaredNames, route.GetCommand().GetPath())
			for _, alias := range route.GetCommand().GetAliases() {
				view.DeclaredNames = append(view.DeclaredNames, alias.GetPath())
			}
			view.EffectiveNames = effectiveCommandNames(index, id, view.RouteID, append(slices.Clone(view.DeclaredNames), names[view.RouteID]...))
		}
		for _, entry := range result.InvalidNames {
			if entry.RouteID == view.RouteID {
				view.InvalidNames = append(view.InvalidNames, entry)
			}
		}
		result.Commands = append(result.Commands, view)
	}
	// 3. 只返回与该插件有关的跨插件遮蔽关系。
	for _, shadow := range index.Shadows() {
		related := shadow.Winner.PluginID == id
		view := commandShadow{Path: shadow.Path, Winner: commandReference{shadow.Winner.PluginID, shadow.Winner.RouteID}, Shadowed: []commandReference{}}
		for _, entry := range shadow.Shadowed {
			related = related || entry.PluginID == id
			view.Shadowed = append(view.Shadowed, commandReference{entry.PluginID, entry.RouteID})
		}
		if related {
			result.Shadows = append(result.Shadows, view)
		}
	}
	return result, nil
}

// effectiveCommandNames 从胜出索引或遮蔽诊断取得 Router 的完整生效名字，沿用它的失效覆盖规则。
func effectiveCommandNames(index *router.CommandIndex, id, routeID string, paths [][]string) [][]string {
	for _, path := range paths {
		entry, _, ok := index.Lookup(path)
		if ok && entry.PluginID == id && entry.RouteID == routeID {
			return entry.Names
		}
	}
	for _, shadow := range index.Shadows() {
		for _, entry := range shadow.Shadowed {
			if entry.PluginID == id && entry.RouteID == routeID {
				return entry.Names
			}
		}
	}
	return [][]string{}
}
