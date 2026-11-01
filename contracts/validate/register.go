package validate

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"path"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// HTTP Listener 只接受这两个 loopback 地址（WIRE-FORMAT §3.3）。
const (
	LoopbackIPv4 = "127.0.0.1"
	LoopbackIPv6 = "::1"
)

// Service 名称长度遵循 Service 描述约定。
const maxServiceNameLength = 64

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
// 命令路径、别名、事件过滤器、HTTP 路由与 Listener、Service 与 Bus 订阅随整份描述一起检查。
func Register(r *pb.Register) error {
	// 1. 先检查描述身份、分支与注册公共字段。
	if r.GetAdapter() != nil {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "plugin registration contains adapter descriptor")
	}
	seen := make(map[string]bool, len(r.GetRoutes()))
	hasHTTP := false
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
		hasHTTP = hasHTTP || rd.GetKind() == pb.RouteKind_ROUTE_KIND_HTTP
	}
	if !IsJSONObject([]byte(r.GetCandidateConfigJson())) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate_config_json is not a JSON object")
	}
	if hasHTTP != (r.GetHttpListener() != nil) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP routes and HTTP listener must be registered together")
	}
	if l := r.GetHttpListener(); l != nil {
		ip := net.ParseIP(l.GetHost())
		if (l.GetHost() != LoopbackIPv4 && l.GetHost() != LoopbackIPv6) || ip == nil || l.GetPort() == 0 || l.GetPort() > 65535 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP listener must be 127.0.0.1 or ::1 with a non-zero port")
		}
	}

	// 2. 按路由类型检查描述；共享路径表覆盖全部命令及其别名，HTTP 按 method/path 检查冲突，Bus 回调可引用后面的路由。
	paths := make(map[string]string)
	httpRoutes := make(map[[2]string]string)
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
		case pb.RouteKind_ROUTE_KIND_HTTP:
			if err := registerHTTPRoute(httpRoutes, route); err != nil {
				return err
			}
		case pb.RouteKind_ROUTE_KIND_SERVICE:
			if err := registerService(route); err != nil {
				return err
			}
		case pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION:
			subscription := route.GetBusSubscription()
			pluginID, eventName, ok := strings.Cut(subscription.GetEventName(), "/")
			if _, err := PluginID(pluginID); err != nil || !ok || eventName == "" || strings.Contains(eventName, "/") {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "bus route %q has invalid event_name %q", route.GetRouteId(), subscription.GetEventName())
			}
			if subscription.GetSchemaId() == "" || subscription.GetSchemaVersion() == 0 {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "bus route %q is missing schema ID or version", route.GetRouteId())
			}
			if !seen[subscription.GetCallbackRouteId()] {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "bus route %q refers to unknown callback %q", route.GetRouteId(), subscription.GetCallbackRouteId())
			}
		case pb.RouteKind_ROUTE_KIND_EVENT:
			filters := route.GetEvent().GetEventTypeFilter()
			if len(filters) == 0 {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q has an empty event_type_filter", route.GetRouteId())
			}
			seenEvents := make(map[string]bool, len(filters))
			for _, eventType := range filters {
				if !IsEventType(eventType) {
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
		if !IsCommandToken(token) {
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

// IsCommandToken 检查命令路径 token：非空、不含空白且不以 / 开头。
func IsCommandToken(token string) bool {
	return token != "" && !strings.ContainsFunc(token, unicode.IsSpace) && !strings.HasPrefix(token, "/")
}

// IsEventType 检查 message.proto 声明的标准事件类型。
func IsEventType(eventType string) bool {
	switch eventType {
	case "message.created", "message.edited", "message.deleted", "button.clicked",
		"member.joined", "member.left", "reaction.added", "scheduled", "bot.online", "bot.offline":
		return true
	}
	return false
}

func registerHTTPRoute(routes map[[2]string]string, route *pb.RouteDescriptor) error {
	spec := route.GetHttp()
	method, rawPath := spec.GetMethod(), spec.GetPath()
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace, http.MethodPatch:
	default:
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP route %q has invalid method %q", route.GetRouteId(), method)
	}
	canonical := path.Clean(rawPath)
	if strings.HasSuffix(rawPath, "/") && canonical != "/" {
		canonical += "/"
	}
	if !strings.HasPrefix(rawPath, "/") || canonical != rawPath || strings.ContainsAny(rawPath, "?#") || strings.Contains(rawPath, "..") {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "HTTP route %q has invalid path %q", route.GetRouteId(), rawPath)
	}
	if spec.GetIsPublic() {
		if spec.GetAuthMode() != pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE && spec.GetAuthMode() != pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_SIGNATURE {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "public HTTP route %q must declare callback verification", route.GetRouteId())
		}
	} else if spec.GetAuthMode() != pb.HttpAuthMode_HTTP_AUTH_MODE_CORE_SESSION {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "private HTTP route %q must use Core session authentication", route.GetRouteId())
	}
	key := [2]string{method, rawPath}
	if previous, exists := routes[key]; exists {
		return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "HTTP %s %s conflicts between routes %q and %q", method, rawPath, previous, route.GetRouteId())
	}
	routes[key] = route.GetRouteId()
	return nil
}

func registerService(route *pb.RouteDescriptor) error {
	service := route.GetService()
	name := service.GetServiceName()
	if !validServiceName(name) || route.GetRouteId() != "svc."+name {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service route %q has invalid name or route_id", route.GetRouteId())
	}
	if service.GetEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON && service.GetEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q has invalid encoding", name)
	}
	if route.GetInputSchemaId() == "" || route.GetInputSchemaVersion() == 0 || route.GetOutputSchemaId() == "" || route.GetOutputSchemaVersion() == 0 {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q is missing input or output schema ID or version", name)
	}
	switch service.GetAccess() {
	case pb.ServiceAccess_SERVICE_ACCESS_PRIVATE, pb.ServiceAccess_SERVICE_ACCESS_PUBLIC:
		if len(service.GetAllowlistPluginIds()) != 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q has an allowlist without ALLOWLIST access", name)
		}
	case pb.ServiceAccess_SERVICE_ACCESS_ALLOWLIST:
		if len(service.GetAllowlistPluginIds()) == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q has an empty allowlist", name)
		}
		for _, pluginID := range service.GetAllowlistPluginIds() {
			if _, err := PluginID(pluginID); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q has invalid allowlist plugin ID: %v", name, err)
			}
		}
	default:
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q has invalid access", name)
	}
	return nil
}

func validServiceName(name string) bool {
	if len(name) == 0 || len(name) > maxServiceNameLength || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
