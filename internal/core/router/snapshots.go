package router

import (
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

type routeSnapshot struct {
	generationID string
	routes       []*pb.RouteDescriptor
	session      *pluginSession // 只保留已加载文案；派发仍须重新取得当前 Caller。
	committed    bool
}

// SetRouteSnapshot 发布已提交代次的路由定位数据，Session 绑定保持不变。
func (r *Router) SetRouteSnapshot(pluginID, generationID string, routes []*pb.RouteDescriptor) {
	defer r.scheduleCommandSync()
	snapshot := routeSnapshot{generationID: generationID, routes: cloneRoutes(routes), committed: true}
	r.mu.Lock()
	defer r.mu.Unlock()
	if session := r.plugins[pluginID]; session != nil && session.caller.GenerationID == generationID {
		snapshot.session = session
	}
	r.snapshots[pluginID] = snapshot
	r.rebuildCommands()
}

// RouteSnapshot 返回供定位的代次与路由副本，无论该插件是否有活 Session。
func (r *Router) RouteSnapshot(pluginID string) (generationID string, routes []*pb.RouteDescriptor, ok bool) {
	r.mu.Lock()
	snapshot, ok := r.snapshots[pluginID]
	r.mu.Unlock()
	return snapshot.generationID, cloneRoutes(snapshot.routes), ok
}

// RemoveRouteSnapshot 在停用或卸载时撤销路由定位；Detach 单独撤销会话。
func (r *Router) RemoveRouteSnapshot(pluginID string) {
	defer r.scheduleCommandSync()
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.snapshots, pluginID)
	r.rebuildCommands()
}

// routeSourcesLocked 合并已提交定位数据与常驻组件；调用方持有 r.mu。
func (r *Router) routeSourcesLocked() map[string]routeSnapshot {
	sources := make(map[string]routeSnapshot, len(r.snapshots)+len(r.plugins))
	if r.ensureActive != nil {
		for id, snapshot := range r.snapshots {
			sources[id] = snapshot
		}
	}
	for id, session := range r.plugins {
		if _, ok := sources[id]; !ok && session.plugin.Activated() {
			sources[id] = routeSnapshot{generationID: session.caller.GenerationID, routes: session.plugin.Routes(), session: session}
		}
	}
	return sources
}

func (r *Router) available(c candidate) bool {
	return r.ensureActive != nil && c.generationID != "" || c.pluginSession != nil && c.plugin.Activated()
}

func cloneRoutes(routes []*pb.RouteDescriptor) []*pb.RouteDescriptor {
	cloned := slices.Clone(routes)
	for i, route := range cloned {
		cloned[i] = proto.CloneOf(route)
	}
	return cloned
}
