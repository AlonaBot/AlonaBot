package service

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

const (
	// CapabilityDiscover 标识身份过滤的服务发现协议。
	CapabilityDiscover = "service_discover"
	// CapabilityRoleAccess 标识 allowlist 中的角色别名授权。
	CapabilityRoleAccess = "service_role_access"
	// CapabilityReentrant 标识受控回调重入。
	CapabilityReentrant = "service_reentrant"
	// CapabilityStream 标识有背压的服务增量传输。
	CapabilityStream = "service_stream"
)

// Filter 按 Schema ID 及版本筛选；空 ID 或零版本分别表示不限该字段。
type Filter struct {
	InputSchemaID       string
	InputSchemaVersion  uint32
	OutputSchemaID      string
	OutputSchemaVersion uint32
}

// Descriptor 保留完整目标、身份及原始注册声明，包含 priority、编码、Schema 和重入／增量信息。
type Descriptor struct {
	Target              string
	PluginID            string
	RouteID             string
	Priority            int32
	Encoding            pb.PayloadEncoding
	InputSchemaID       string
	InputSchemaVersion  uint32
	OutputSchemaID      string
	OutputSchemaVersion uint32
	DeltaSchemaID       string
	DeltaSchemaVersion  uint32
	Reentrant           bool
}

// Discover 查询当前调用身份可访问的有效服务，包括可按需唤醒的注册快照。
func Discover(ctx context.Context, filter Filter) ([]Descriptor, error) {
	if err := registry.RequireCapability(ctx, CapabilityDiscover); err != nil {
		return nil, err
	}
	response, err := registry.Proto[*pb.ServiceDiscoverRequest, *pb.ServiceDiscoverResponse]("core.service.discover").Call(ctx, &pb.ServiceDiscoverRequest{InputSchemaId: filter.InputSchemaID, InputSchemaVersion: filter.InputSchemaVersion, OutputSchemaId: filter.OutputSchemaID, OutputSchemaVersion: filter.OutputSchemaVersion})
	if err != nil {
		return nil, err
	}
	result := make([]Descriptor, 0, len(response.Services))
	for _, item := range response.Services {
		r := item.GetRoute()
		s := r.GetService()
		result = append(result, Descriptor{Target: item.Target, PluginID: item.PluginId, RouteID: r.GetRouteId(), Priority: r.GetPriority(), Encoding: s.GetEncoding(), InputSchemaID: r.GetInputSchemaId(), InputSchemaVersion: r.GetInputSchemaVersion(), OutputSchemaID: r.GetOutputSchemaId(), OutputSchemaVersion: r.GetOutputSchemaVersion(), DeltaSchemaID: s.GetDeltaSchemaId(), DeltaSchemaVersion: s.GetDeltaSchemaVersion(), Reentrant: s.GetReentrant()})
	}
	return result, nil
}
