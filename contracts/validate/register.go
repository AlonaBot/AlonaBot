package validate

import (
	"bytes"
	"encoding/json"
	"math"
	"net"
	"regexp"
	"strconv"
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
	if r.GetAdapter() != nil {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "plugin registration contains adapter descriptor")
	}
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
			if err := registerCommand(command, route.GetRouteId()); err != nil {
				return err
			}
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

func registerCommand(command *pb.CommandSpec, routeID string) error {
	// 1. 匹配描述在注册边界检查，避免非法模式进入派发快照。
	if command.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH && (len(command.GetPath()) != 0 || len(command.GetAliases()) != 0) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "non-path route %q contains command paths or aliases", routeID)
	}
	switch command.GetMatch() {
	case pb.CommandMatch_COMMAND_MATCH_PATH:
		if command.GetRegex() != "" || len(command.GetPatterns()) != 0 || command.GetIgnoreCase() {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "path route %q contains text matching fields", routeID)
		}
	case pb.CommandMatch_COMMAND_MATCH_REGEX:
		if _, err := regexp.Compile(command.GetRegex()); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has invalid regex: %v", routeID, err)
		}
	case pb.CommandMatch_COMMAND_MATCH_KEYWORD, pb.CommandMatch_COMMAND_MATCH_STARTSWITH,
		pb.CommandMatch_COMMAND_MATCH_ENDSWITH, pb.CommandMatch_COMMAND_MATCH_FULLMATCH:
		if len(command.GetPatterns()) == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has no patterns", routeID)
		}
		for _, pattern := range command.GetPatterns() {
			if pattern == "" {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has an empty pattern", routeID)
			}
		}
	case pb.CommandMatch_COMMAND_MATCH_RAW:
		if command.GetRegex() != "" || len(command.GetPatterns()) != 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "raw route %q contains regex or patterns", routeID)
		}
	default:
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has an invalid match mode", routeID)
	}
	if command.GetMatch() != pb.CommandMatch_COMMAND_MATCH_PATH && len(command.GetParams()) != 0 {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "text or raw route %q contains command parameters", routeID)
	}

	// 2. 参数名对应调用载荷中的绑定键，顺序、类型与缺失值均须可执行。
	names := make(map[string]bool, len(command.GetParams()))
	optional := false
	for i, param := range command.GetParams() {
		name := param.GetName()
		if name == "" || names[name] {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has an empty or duplicate parameter name %q", routeID, name)
		}
		names[name] = true
		if param.GetType() == pb.ParamType_PARAM_TYPE_UNSPECIFIED || !knownEnum(param.GetType()) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q parameter %q has an invalid type", routeID, name)
		}
		if param.GetRequired() && optional {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q required parameter %q follows an optional parameter", routeID, name)
		}
		optional = optional || !param.GetRequired()
		if param.GetRaw() && (param.GetType() != pb.ParamType_PARAM_TYPE_STRING || i != len(command.GetParams())-1) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q raw parameter %q must be the last string parameter", routeID, name)
		}
		if param.DefaultValue != nil {
			if err := commandDefault(param); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q parameter %q has invalid default: %v", routeID, name, err)
			}
		}
	}
	return nil
}

func commandDefault(param *pb.CommandParam) error {
	value := param.GetDefaultValue()
	switch param.GetType() {
	case pb.ParamType_PARAM_TYPE_INT:
		_, err := strconv.ParseInt(value, 10, 64)
		return err
	case pb.ParamType_PARAM_TYPE_FLOAT:
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "float must be finite")
		}
	case pb.ParamType_PARAM_TYPE_BOOL:
		if value != "true" && value != "false" {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "bool must be true or false")
		}
	}
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
