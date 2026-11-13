package validate

import (
	"bytes"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/schedule"
	"github.com/google/jsonschema-go/jsonschema"
)

// HTTP Listener 只接受这两个 loopback 地址（WIRE-FORMAT §3.3）。
const (
	LoopbackIPv4 = "127.0.0.1"
	LoopbackIPv6 = "::1"
)

// Service 名称长度遵循 Service 描述约定。
const maxServiceNameLength = 64

// maxToolDescriptionBytes 固定工具说明的注册字节上限。
const maxToolDescriptionBytes = 1024

const scheduleRoutePrefix = "job."

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
// 命令路径、别名、事件过滤器与按钮 action、定时任务、HTTP 路由与 Listener、Service 与 Bus 订阅随整份描述一起检查。
func Register(r *pb.Register) error {
	return register(r, false)
}

func register(r *pb.Register, isAdapter bool) error {
	// 1. 先检查描述身份、分支与注册公共字段。
	if !isAdapter && r.GetAdapter() != nil {
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
		permissions := make(map[string]bool, len(rd.GetRequiredPermissions()))
		for _, permission := range rd.GetRequiredPermissions() {
			if !IsPermissionName(permission) || permissions[permission] {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "route %q has invalid or duplicate required permission %q", id, permission)
			}
			permissions[permission] = true
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
	actions := make(map[string]string)
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
		case pb.RouteKind_ROUTE_KIND_HTTP:
			if err := registerHTTPRoute(httpRoutes, route, isAdapter); err != nil {
				return err
			}
		case pb.RouteKind_ROUTE_KIND_SERVICE:
			if err := registerService(route); err != nil {
				return err
			}
		case pb.RouteKind_ROUTE_KIND_SCHEDULE:
			if err := registerSchedule(route); err != nil {
				return err
			}
		case pb.RouteKind_ROUTE_KIND_HOOK:
			filters := route.GetHook().GetEventTypeFilter()
			switch route.GetHook().GetPhase() {
			case pb.HookPhase_HOOK_PHASE_PRE_EVENT:
				if len(filters) != 1 || (!IsEventType(filters[0]) && !IsAdapterEventType(filters[0])) {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "pre-event hook route %q requires exactly one valid event_type_filter", route.GetRouteId())
				}
			case pb.HookPhase_HOOK_PHASE_BEFORE_SEND, pb.HookPhase_HOOK_PHASE_AFTER_SEND:
				if len(filters) != 0 {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "outbound hook route %q must not contain event_type_filter", route.GetRouteId())
				}
			default:
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "hook route %q has an invalid phase", route.GetRouteId())
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
			event := route.GetEvent()
			filters := event.GetEventTypeFilter()
			if len(filters) == 0 {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q has an empty event_type_filter", route.GetRouteId())
			}
			seenEvents := make(map[string]bool, len(filters))
			for _, eventType := range filters {
				if !IsEventType(eventType) && !IsAdapterEventType(eventType) {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q has unknown event type %q", route.GetRouteId(), eventType)
				}
				if seenEvents[eventType] {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "event route %q repeats event type %q", route.GetRouteId(), eventType)
				}
				seenEvents[eventType] = true
			}
			if seenEvents["button.clicked"] {
				action := event.GetActionId()
				if action == "" || event.GetActionSchemaVersion() == 0 {
					return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "button route %q is missing action_id or action_schema_version", route.GetRouteId())
				}
				// 发送只携 action_id，同一插件的一个 action 必须唯一对应一个路由。
				if previous, exists := actions[action]; exists {
					return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "button action %q conflicts between routes %q and %q", action, previous, route.GetRouteId())
				}
				actions[action] = route.GetRouteId()
			}
		}
	}
	return nil
}

func registerSchedule(route *pb.RouteDescriptor) error {
	spec := route.GetSchedule()
	if spec.GetJobId() == "" || route.GetRouteId() != scheduleRoutePrefix+spec.GetJobId() {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "schedule route %q has invalid job_id or route_id", route.GetRouteId())
	}
	if spec.GetJobDeadlineMs() != 0 {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "schedule route %q must use JOB_TIMEOUT instead of job_deadline_ms", route.GetRouteId())
	}
	if _, err := schedule.FromSpec(spec); err != nil {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "schedule route %q has invalid trigger: %v", route.GetRouteId(), err)
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

	if command.GetUsageMode() == pb.UsageMode_USAGE_MODE_CUSTOM && command.GetUsageKey() == "" {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q has no usage_key for custom usage", routeID)
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
		if param.GetType() == pb.ParamType_PARAM_TYPE_MESSAGE && i != len(command.GetParams())-1 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "command route %q message parameter %q must be last", routeID, name)
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
	case pb.ParamType_PARAM_TYPE_USER, pb.ParamType_PARAM_TYPE_IMAGE, pb.ParamType_PARAM_TYPE_MESSAGE:
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parameter type %s does not accept a default", param.GetType())
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

func registerHTTPRoute(routes map[[2]string]string, route *pb.RouteDescriptor, isAdapter bool) error {
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
		switch mode := spec.GetAuthMode(); {
		case mode == pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_STATE, mode == pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_SIGNATURE, mode == pb.HttpAuthMode_HTTP_AUTH_MODE_CALLBACK_NONCE:
		case isAdapter && mode == pb.HttpAuthMode_HTTP_AUTH_MODE_PLATFORM_TOKEN:
		default:
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
	if !IsServiceName(name) || route.GetRouteId() != "svc."+name {
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
	if tool := service.GetTool(); tool != nil {
		if service.GetEncoding() != pb.PayloadEncoding_PAYLOAD_ENCODING_JSON {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q tool requires JSON encoding", name)
		}
		if len(tool.GetDescription()) == 0 || len(tool.GetDescription()) > maxToolDescriptionBytes {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q tool description must be 1 to %d bytes", name, maxToolDescriptionBytes)
		}
		if tool.GetDefaultPolicy() != pb.ToolPolicy_TOOL_POLICY_ALLOW && tool.GetDefaultPolicy() != pb.ToolPolicy_TOOL_POLICY_CONFIRM {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q tool has invalid default policy", name)
		}
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(tool.GetInputSchemaJson()), &schema); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parse service %q tool input schema: %v", name, err)
		}
		if schema.Type != "object" {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "service %q tool input schema root type must be object", name)
		}
		// nil options 保持 Loader 未配置，远程引用由 Resolve 拒绝。
		if _, err := schema.Resolve(nil); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "resolve service %q tool input schema: %v", name, err)
		}
	}
	return nil
}

// IsServiceName 检查 Service 名称是否为 1～64 字节，以小写字母开头且只含小写字母、数字、下划线和点。
func IsServiceName(name string) bool {
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
