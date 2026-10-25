package validate

import (
	"bytes"
	"encoding/json"
	"net"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// routeKindOf 给出 RouteDescriptor 的 spec 分支对应的 kind。
func routeKindOf(r *pb.RouteDescriptor) pb.RouteKind {
	switch r.GetSpec().(type) {
	case *pb.RouteDescriptor_Command:
		return pb.RouteKind_ROUTE_KIND_COMMAND
	case *pb.RouteDescriptor_Event:
		return pb.RouteKind_ROUTE_KIND_EVENT
	case *pb.RouteDescriptor_Hook:
		return pb.RouteKind_ROUTE_KIND_HOOK
	case *pb.RouteDescriptor_Service:
		return pb.RouteKind_ROUTE_KIND_SERVICE
	case *pb.RouteDescriptor_Schedule:
		return pb.RouteKind_ROUTE_KIND_SCHEDULE
	case *pb.RouteDescriptor_Http:
		return pb.RouteKind_ROUTE_KIND_HTTP
	case *pb.RouteDescriptor_BusSubscription:
		return pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION
	}
	return pb.RouteKind_ROUTE_KIND_UNSPECIFIED
}

// IsJSONObject 报告 b 是否为单个合法的 UTF-8 JSON 对象。
func IsJSONObject(b []byte) bool {
	b = bytes.TrimSpace(b)
	return len(b) > 0 && b[0] == '{' && json.Valid(b)
}

// Register 原子校验整份注册集（WIRE §3.1/§3.3）：任一项不合法都整体拒绝，调用方不得保留部分注册。
// 各 kind 的语义细节（命令路径、cron、Schema 等）由各自的工作包在此基础上扩展。
func Register(r *pb.Register) error {
	seen := make(map[string]bool, len(r.GetRoutes()))
	for i, rd := range r.GetRoutes() {
		id := rd.GetRouteId()
		switch {
		case id == "":
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "routes[%d] 缺少 route_id", i)
		case seen[id]:
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "route_id %q 重复", id)
		case rd.GetKind() == pb.RouteKind_ROUTE_KIND_UNSPECIFIED || !knownEnum(rd.GetKind()):
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "route %q 的 kind 非法", id)
		case routeKindOf(rd) != rd.GetKind():
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "route %q 的 kind %s 与描述类型不符", id, rd.GetKind())
		}
		seen[id] = true
	}
	if !IsJSONObject([]byte(r.GetCandidateConfigJson())) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate_config_json 不是 JSON 对象")
	}
	if l := r.GetHttpListener(); l != nil {
		ip := net.ParseIP(l.GetHost())
		if (l.GetHost() != "127.0.0.1" && l.GetHost() != "::1") || ip == nil || l.GetPort() == 0 || l.GetPort() > 65535 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP Listener 只能是 127.0.0.1 或 ::1 的非零端口")
		}
	}
	return nil
}
