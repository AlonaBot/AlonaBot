package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RouteSnapshot 绑定已验证的路由描述及摘要到候选程序包和代次。
type RouteSnapshot struct {
	PluginID, GenerationID string
	Routes                 []*pb.RouteDescriptor
	RouteHash, BuildDigest []byte
}

type routeSnapshotRecord struct {
	GenerationID string `gorm:"column:generation_id;primaryKey"`
	PluginID     string
	Routes       []byte
	RouteHash    []byte
	BuildDigest  []byte
}

func (routeSnapshotRecord) TableName() string { return "route_snapshot" }

func (in Install) usesRouteSnapshot() bool {
	return in.Source != "" && in.Source != devSource && (in.Role == RolePlugin || in.Role == RoleAgent)
}

func checkActiveCommitRoutes(tx *gorm.DB, in Install, generationID string) error {
	if !in.usesRouteSnapshot() {
		return nil
	}
	// 直接登记的配置元数据没有包版本；正式安装总有 manifest 版本。
	if in.Version == "" {
		var row routeSnapshotRecord
		err := tx.Select(columnGenerationID).Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
	}
	_, err := readRouteSnapshot(tx, in.PluginID, generationID, in.BuildDigest)
	return err
}

// SaveRouteSnapshot 保存所属插件的未提交候选路由；已提交的代次保持不可变。
func (s *Store) SaveRouteSnapshot(ctx context.Context, snapshot RouteSnapshot) error {
	encoded, err := encodeRouteSnapshot(snapshot)
	if err != nil {
		return err
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		var candidate generation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: snapshot.GenerationID}).Take(&candidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("generation %s: %w", snapshot.GenerationID, ErrNotFound)
			}
			return err
		}
		if candidate.PluginID != snapshot.PluginID || candidate.CommittedAt != nil {
			return journalPrecondition("route snapshot requires an uncommitted generation owned by the plugin")
		}
		row := routeSnapshotRecord{GenerationID: snapshot.GenerationID, PluginID: snapshot.PluginID,
			Routes: encoded, RouteHash: snapshot.RouteHash, BuildDigest: snapshot.BuildDigest}
		return tx.Save(&row).Error
	})
}

// ActiveRouteSnapshot 读取与当前安装包身份一致的已提交路由快照。
func (s *Store) ActiveRouteSnapshot(ctx context.Context, pluginID string) (RouteSnapshot, error) {
	var snapshot RouteSnapshot
	err := s.tx(ctx, func(tx *gorm.DB) error {
		var installed installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&installed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("plugin %s: %w", pluginID, ErrNotFound)
			}
			return err
		}
		if installed.Uninstalled || !installed.usesRouteSnapshot() {
			return journalPrecondition("route snapshot requires an installed plugin or agent")
		}
		var current pluginState
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: pluginID}).Take(&current).Error; err != nil {
			return err
		}
		var active generation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: current.ActiveGeneration.String}).Take(&active).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || err == nil && (active.PluginID != pluginID || active.CommittedAt == nil) {
			return journalPrecondition("active route snapshot requires a committed generation owned by the plugin")
		}
		if err != nil {
			return err
		}
		snapshot, err = readRouteSnapshot(tx, pluginID, active.GenerationID, installed.BuildDigest)
		return err
	})
	return snapshot, err
}

func readRouteSnapshot(tx *gorm.DB, pluginID, generationID string, buildDigest []byte) (RouteSnapshot, error) {
	var row routeSnapshotRecord
	if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RouteSnapshot{}, journalPrecondition("route snapshot is missing for generation " + generationID)
		}
		return RouteSnapshot{}, err
	}
	if row.PluginID != pluginID || !bytes.Equal(row.BuildDigest, buildDigest) {
		return RouteSnapshot{}, journalPrecondition("route snapshot plugin or build digest differs for generation " + generationID)
	}
	snapshot := RouteSnapshot{PluginID: pluginID, GenerationID: generationID, RouteHash: row.RouteHash, BuildDigest: row.BuildDigest}
	for remaining := row.Routes; len(remaining) > 0; {
		data, size := protowire.ConsumeBytes(remaining)
		if size < 0 {
			return RouteSnapshot{}, journalPrecondition(fmt.Sprintf("decode route snapshot for generation %s: %v", generationID, protowire.ParseError(size)))
		}
		route := new(pb.RouteDescriptor)
		if err := proto.Unmarshal(data, route); err != nil {
			return RouteSnapshot{}, journalPrecondition(fmt.Sprintf("decode route snapshot for generation %s: %v", generationID, err))
		}
		snapshot.Routes = append(snapshot.Routes, route)
		remaining = remaining[size:]
	}
	encoded, err := encodeRouteSnapshot(snapshot)
	if err != nil {
		return RouteSnapshot{}, err
	}
	if !bytes.Equal(encoded, row.Routes) {
		return RouteSnapshot{}, journalPrecondition("route snapshot contains noncanonical route data for generation " + generationID)
	}
	return snapshot, nil
}

func encodeRouteSnapshot(snapshot RouteSnapshot) ([]byte, error) {
	hash := sha256.New()
	// 长度前缀仅用于从 SQLite 还原每条描述；WIRE 摘要只覆盖按序拼接的确定性描述。
	encoded := []byte{}
	marshal := proto.MarshalOptions{Deterministic: true}
	for _, route := range snapshot.Routes {
		data, err := marshal.Marshal(route)
		if err != nil {
			return nil, fmt.Errorf("encode route %q: %w", route.GetRouteId(), err)
		}
		// 丢弃未知字段会改变 PREPARED 摘要；拒绝未进入本版本白名单的数据。
		clean := new(pb.RouteDescriptor)
		if err := (proto.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, clean); err != nil {
			return nil, fmt.Errorf("decode route %q: %w", route.GetRouteId(), err)
		}
		if !proto.Equal(route, clean) {
			return nil, journalPrecondition("route snapshot contains unknown fields for route " + route.GetRouteId())
		}
		hash.Write(data)
		encoded = protowire.AppendBytes(encoded, data)
	}
	if !bytes.Equal(snapshot.RouteHash, hash.Sum(nil)) {
		return nil, journalPrecondition("route snapshot digest differs for generation " + snapshot.GenerationID)
	}
	return encoded, nil
}
