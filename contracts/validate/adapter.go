package validate

import (
	"encoding/json"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/google/jsonschema-go/jsonschema"
)

// AdapterRegister 原子校验 Adapter 身份与角色互斥的注册集（WIRE §3.4）。
func AdapterRegister(r *pb.Register, adapterID string) error {
	// 1. 核对安装身份以及角色互斥的注册字段。
	descriptor := r.GetAdapter()
	if descriptor == nil || descriptor.GetAdapterId() != adapterID {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter descriptor does not match installation")
	}
	if r.GetMigration().GetMode() != pb.MigrationMode_MIGRATION_MODE_NONE || len(r.GetMigration().GetCompletedMigrationIds()) != 0 || r.GetMigration().GetSelfMigrateCompleted() {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter registration contains migration")
	}
	for _, route := range r.GetRoutes() {
		if route.GetKind() != pb.RouteKind_ROUTE_KIND_HTTP {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter registration contains non-HTTP plugin routes")
		}
	}
	if err := register(r, true); err != nil {
		return err
	}
	// 2. 校验能力及名称唯一性，并在接纳 extension 声明前解析 Schema 的引用。
	if err := Capabilities(descriptor.GetCapabilities()); err != nil {
		return err
	}
	capabilities := make(map[string]bool)
	for _, capability := range descriptor.GetCapabilities() {
		name := capability.GetName()
		if capabilities[name] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter capability %q", name)
		}
		capabilities[name] = true
	}
	type extensionKey struct {
		name    string
		version uint32
	}
	extensions := make(map[extensionKey]bool)
	for _, extension := range descriptor.GetExtensions() {
		key := extensionKey{extension.GetName(), extension.GetVersion()}
		if !strings.HasPrefix(key.name, adapterID+".") || !IsAdapterEventType(key.name) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter extension name %q", key.name)
		}
		if extensions[key] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter extension %q version %d", key.name, key.version)
		}
		if key.version == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter extension %q has no version", key.name)
		}
		if !IsJSONObject([]byte(extension.GetSchemaJson())) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "extension schema_json is not a JSON object")
		}
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(extension.GetSchemaJson()), &schema); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parse adapter extension %q schema: %v", key.name, err)
		}
		if _, err := schema.Resolve(nil); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "resolve adapter extension %q schema: %v", key.name, err)
		}
		extensions[key] = true
	}
	// 3. 权限先建立完整名称表，再校验引用与无环关系，允许声明顺序任意。
	if err := adapterPermissions(descriptor); err != nil {
		return err
	}
	// 4. 扩展事件按名称唯一注册，并在接纳前解析 Schema 的引用。
	events := make(map[string]bool)
	for _, event := range descriptor.GetEvents() {
		name := event.GetName()
		if !strings.HasPrefix(name, adapterID+".") || !IsAdapterEventType(name) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter event name %q", name)
		}
		if events[name] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter event %q", name)
		}
		events[name] = true
		if event.GetVersion() == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter event %q has no version", name)
		}
		if !IsJSONObject([]byte(event.GetSchemaJson())) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter event %q schema_json is not a JSON object", name)
		}
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(event.GetSchemaJson()), &schema); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parse adapter event %q schema: %v", name, err)
		}
		if _, err := schema.Resolve(nil); err != nil {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "resolve adapter event %q schema: %v", name, err)
		}
	}
	// 5. 动作按名称与版本唯一注册，参数和结果的引用都须在注册时解析。
	actions := make(map[extensionKey]bool)
	for _, action := range descriptor.GetActions() {
		key := extensionKey{action.GetName(), action.GetVersion()}
		if !strings.HasPrefix(key.name, adapterID+".") || !IsAdapterEventType(key.name) || key.version == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter action %q version %d", key.name, key.version)
		}
		if actions[key] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter action %q version %d", key.name, key.version)
		}
		actions[key] = true
		if action.GetScope() != pb.ActionScope_ACTION_SCOPE_CHAT && action.GetScope() != pb.ActionScope_ACTION_SCOPE_ACCOUNT {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter action %q scope", key.name)
		}
		for _, entry := range []struct{ name, raw string }{{"params", action.GetParamsSchemaJson()}, {"result", action.GetResultSchemaJson()}} {
			var schema jsonschema.Schema
			if !IsJSONObject([]byte(entry.raw)) {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter action %q %s schema is not a JSON object", key.name, entry.name)
			}
			if err := json.Unmarshal([]byte(entry.raw), &schema); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parse adapter action %q %s schema: %v", key.name, entry.name, err)
			}
			if _, err := schema.Resolve(nil); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "resolve adapter action %q %s schema: %v", key.name, entry.name, err)
			}
		}
	}
	return nil
}

func adapterPermissions(descriptor *pb.AdapterDescriptor) error {
	permissions := make(map[string]*pb.PermissionDeclaration)
	for _, permission := range descriptor.GetPermissions() {
		name := permission.GetName()
		if !strings.HasPrefix(name, descriptor.GetAdapterId()+".") || !IsAdapterEventType(name) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid adapter permission name %q", name)
		}
		if _, exists := permissions[name]; exists {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter permission %q", name)
		}
		if permission.GetLabelKey() == "" {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter permission %q has no label_key", name)
		}
		permissions[name] = permission
	}
	for _, permission := range descriptor.GetPermissions() {
		for _, implied := range permission.GetImplies() {
			if implied == permission.GetName() || (implied != "chat_admin" && permissions[implied] == nil) {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter permission %q implies invalid permission %q", permission.GetName(), implied)
			}
		}
	}
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter permission implication cycle at %q", name)
		}
		if visited[name] {
			return nil
		}
		visiting[name] = true
		for _, implied := range permissions[name].GetImplies() {
			if implied != "chat_admin" {
				if err := visit(implied); err != nil {
					return err
				}
			}
		}
		delete(visiting, name)
		visited[name] = true
		return nil
	}
	for _, permission := range descriptor.GetPermissions() {
		if err := visit(permission.GetName()); err != nil {
			return err
		}
	}
	return nil
}

// IsPermissionName 检查 Core 内置权限名或 Adapter 权限完整名。
func IsPermissionName(name string) bool {
	return name == "user" || name == "admin" || name == "chat_admin" || IsAdapterEventType(name)
}

// IsAdapterEventType 检查由本地 Adapter ID 和小写字母、数字、下划线组成的完整扩展事件名。
func IsAdapterEventType(eventType string) bool {
	separator := strings.LastIndexByte(eventType, '.')
	if separator < 0 {
		return false
	}
	adapterID, name := eventType[:separator], eventType[separator+1:]
	if _, err := PluginID(adapterID); err != nil || strings.HasPrefix(adapterID, MarketPrefix) || name == "" {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

// Capabilities 校验 Adapter 在注册和实例启动时声明的文本限额。
func Capabilities(capabilities []*pb.Capability) error {
	for _, capability := range capabilities {
		limit := capability.GetTextLimit()
		if limit == nil {
			continue
		}
		if capability.GetName() != "message.send" || !capability.GetSupported() || limit.Max == 0 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "text_limit requires supported message.send and a positive max")
		}
		switch limit.Unit {
		case pb.TextUnit_TEXT_UNIT_UTF16_CODE_UNITS, pb.TextUnit_TEXT_UNIT_UNICODE_CODE_POINTS, pb.TextUnit_TEXT_UNIT_BYTES:
		default:
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid text_limit unit %d", limit.Unit)
		}
	}
	return nil
}
