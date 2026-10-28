package validate

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// HTTP Listener 只接受这两个 loopback 地址（WIRE-FORMAT §3.3）。
const (
	LoopbackIPv4 = "127.0.0.1"
	LoopbackIPv6 = "::1"
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
// 命令路径、别名与事件过滤器随整份描述一起检查。
func Register(r *pb.Register) error {
	// 1. 先检查描述身份、分支与注册公共字段。
	seen := make(map[string]bool, len(r.GetRoutes()))
	for i, rd := range r.GetRoutes() {
		id := rd.GetRouteId()
		switch {
		case id == "":
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "routes[%d] is missing route_id", i)
		case seen[id]:
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate route_id %q", id)
		case rd.GetKind() == pb.RouteKind_ROUTE_KIND_UNSPECIFIED || !knownEnum(rd.GetKind()):
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "route %q has an invalid kind", id)
		case routeKindOf(rd) != rd.GetKind():
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "route %q has kind %s, which does not match its spec", id, rd.GetKind())
		}
		seen[id] = true
	}
	if !IsJSONObject([]byte(r.GetCandidateConfigJson())) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate_config_json is not a JSON object")
	}
	if l := r.GetHttpListener(); l != nil {
		ip := net.ParseIP(l.GetHost())
		if (l.GetHost() != LoopbackIPv4 && l.GetHost() != LoopbackIPv6) || ip == nil || l.GetPort() == 0 || l.GetPort() > 65535 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP listener must be 127.0.0.1 or ::1 with a non-zero port")
		}
	}

	// 2. 共享路径表覆盖全部命令及其别名，任一冲突都拒绝整份注册。
	paths := make(map[string]string)
	for _, route := range r.GetRoutes() {
		switch route.GetKind() {
		case pb.RouteKind_ROUTE_KIND_COMMAND:
			command := route.GetCommand()
			if command.GetMatch() == pb.CommandMatch_COMMAND_MATCH_PATH && len(command.GetPath()) == 0 {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has an empty path", route.GetRouteId())
			}
			if len(command.GetPath()) != 0 {
				if err := registerCommandPath(paths, command.GetPath(), route.GetRouteId()); err != nil {
					return err
				}
			}
			for _, alias := range command.GetAliases() {
				if err := registerCommandPath(paths, alias.GetPath(), route.GetRouteId()); err != nil {
					return err
				}
			}
		case pb.RouteKind_ROUTE_KIND_EVENT:
			filters := route.GetEvent().GetEventTypeFilter()
			if len(filters) == 0 {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q has an empty event_type_filter", route.GetRouteId())
			}
			seenEvents := make(map[string]bool, len(filters))
			for _, eventType := range filters {
				switch eventType {
				case "message.created", "message.edited", "message.deleted", "button.clicked",
					"member.joined", "member.left", "reaction.added", "scheduled":
				default:
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q has unknown event type %q", route.GetRouteId(), eventType)
				}
				if seenEvents[eventType] {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q repeats event type %q", route.GetRouteId(), eventType)
				}
				seenEvents[eventType] = true
			}
		}
	}
	return nil
}

func registerCommandPath(paths map[string]string, path []string, routeID string) error {
	if len(path) == 0 {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has an empty alias path", routeID)
	}
	for _, token := range path {
		if token == "" || strings.ContainsFunc(token, unicode.IsSpace) || strings.HasPrefix(token, "/") {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has invalid path token %q", routeID, token)
		}
	}
	key := strings.Join(path, " ")
	if previous, exists := paths[key]; exists {
		return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "command path %q conflicts between routes %q and %q", key, previous, routeID)
	}
	paths[key] = routeID
	return nil
}
