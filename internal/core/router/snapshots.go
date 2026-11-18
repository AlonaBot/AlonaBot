package router

import (
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

type routeSnapshot struct {
	generationID string
	routes       []*pb.RouteDescriptor
}

// SetRouteSnapshot 发布已提交代次的路由定位数据，Session 绑定保持不变。
func (r *Router) SetRouteSnapshot(pluginID, generationID string, routes []*pb.RouteDescriptor) {
	snapshot := routeSnapshot{generationID: generationID, routes: cloneRoutes(routes)}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots[pluginID] = snapshot
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
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.snapshots, pluginID)
}

func cloneRoutes(routes []*pb.RouteDescriptor) []*pb.RouteDescriptor {
	cloned := slices.Clone(routes)
	for i, route := range cloned {
		cloned[i] = proto.CloneOf(route)
	}
	return cloned
}
