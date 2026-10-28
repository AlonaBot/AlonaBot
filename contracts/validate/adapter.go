package validate

import pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"

// AdapterRegister 原子校验 Adapter 身份与角色互斥的注册集（WIRE §3.4）。
func AdapterRegister(r *pb.Register, adapterID string) error {
	// 1. 核对安装身份以及角色互斥的注册字段。
	descriptor := r.GetAdapter()
	if descriptor == nil || descriptor.GetAdapterId() != adapterID {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter descriptor does not match installation")
	}
	if len(r.GetRoutes()) != 0 || r.GetHttpListener() != nil || r.GetMigration().GetMode() != pb.MigrationMode_MIGRATION_MODE_NONE || len(r.GetMigration().GetCompletedMigrationIds()) != 0 || r.GetMigration().GetSelfMigrateCompleted() {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "adapter registration contains plugin routes, listener or migration")
	}
	if !IsJSONObject([]byte(r.GetCandidateConfigJson())) {
		return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "candidate_config_json is not a JSON object")
	}
	// 2. 检查能力名称唯一性与 extension 声明的基本结构。
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
		if extensions[key] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate adapter extension %q version %d", key.name, key.version)
		}
		if !IsJSONObject([]byte(extension.GetSchemaJson())) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "extension schema_json is not a JSON object")
		}
		extensions[key] = true
	}
	return nil
}
