package state

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrSecretOwnerMismatch 表示引用属于其他插件，调用方应拒绝解析。
var ErrSecretOwnerMismatch = errors.New("secret belongs to another plugin")

// Secret 保存一个不可变的秘密值版本；CreatedAt 由调用方显式赋值，单位为 Unix 毫秒。
type Secret struct {
	RefID     string `gorm:"column:ref_id;primaryKey"`
	PluginID  string `gorm:"column:plugin_id"`
	Value     []byte `gorm:"column:value"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:false"`
}

// TableName 指定不可变秘密值的状态表。
func (Secret) TableName() string { return "secret" }

type generationSecretPath struct {
	GenerationID string `gorm:"column:generation_id;primaryKey"`
	Path         string `gorm:"column:path;primaryKey"`
}

func (generationSecretPath) TableName() string { return "generation_secret_path" }

// AddSecret 创建秘密值版本；引用已存在时返回错误，保留原值。
func (s *Store) AddSecret(ctx context.Context, secret Secret) error {
	if secret.PluginID == "" {
		return errEmptyPluginID
	}
	return s.db.WithContext(ctx).Create(&secret).Error
}

// GetSecret 按引用读取秘密值并核对插件归属；缺失返回 ErrNotFound，越权返回 ErrSecretOwnerMismatch。
func (s *Store) GetSecret(ctx context.Context, pluginID, refID string) (Secret, error) {
	if pluginID == "" {
		return Secret{}, errEmptyPluginID
	}
	var row Secret
	err := s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: refID}).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Secret{}, ErrNotFound
	}
	if err != nil {
		return Secret{}, err
	}
	if row.PluginID != pluginID {
		return Secret{}, ErrSecretOwnerMismatch
	}
	return row, nil
}

// ListSecrets 列出指定插件的秘密值版本，供调用方计算未受保护的引用。
func (s *Store) ListSecrets(ctx context.Context, pluginID string) ([]Secret, error) {
	if pluginID == "" {
		return nil, errEmptyPluginID
	}
	var rows []Secret
	err := s.db.WithContext(ctx).Where(&Secret{PluginID: pluginID}).Find(&rows).Error
	return rows, err
}

// DeleteSecrets 删除指定插件中已由调用方确认不受保护的引用，返回删除数；空集合不执行删除。
func (s *Store) DeleteSecrets(ctx context.Context, pluginID string, refIDs []string) (int, error) {
	if pluginID == "" {
		return 0, errEmptyPluginID
	}
	if len(refIDs) == 0 {
		return 0, nil
	}
	values := make([]any, len(refIDs))
	for i, id := range refIDs {
		values[i] = id
	}
	result := s.db.WithContext(ctx).Where(&Secret{PluginID: pluginID}).
		Where(clause.IN{Column: clause.PrimaryColumn, Values: values}).Delete(&Secret{})
	return int(result.RowsAffected), result.Error
}

// SetGenerationConfig 在同一事务中保存未提交代次的配置快照并替换其 secret 路径。
// 已提交代次保持不可变；路径写入失败时连同快照一起回滚。
func (s *Store) SetGenerationConfig(ctx context.Context, generationID, snapshot string, secretPaths []string) error {
	if generationID == "" {
		return errors.New("generation_id is required")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 按主键读取并检查提交标记，再保存候选快照。
		var row generation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("generation %s: %w", generationID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		if row.CommittedAt != nil {
			return errors.New("generation is already committed")
		}
		row.ConfigSnapshot = snapshot
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		// 2. 整体替换路径集合；任何一条写入失败都回滚快照与路径。
		if err := tx.Where(&generationSecretPath{GenerationID: generationID}).Delete(&generationSecretPath{}).Error; err != nil {
			return err
		}
		for _, path := range secretPaths {
			if err := tx.Create(&generationSecretPath{GenerationID: generationID, Path: path}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SetDevGenerationConfig 在同一事务中覆盖开发代次的配置、版本与 secret 路径，允许刷新已提交代次。
// 仅接受 dev: 前缀且安装来源为 dev 的代次；路径写入失败时回滚全部改动。
func (s *Store) SetDevGenerationConfig(ctx context.Context, generationID, snapshot string, configVersion uint32, secretPaths []string) error {
	if !strings.HasPrefix(generationID, "dev:") {
		return errors.New("generation_id must have the dev: prefix")
	}
	return s.tx(ctx, func(tx *gorm.DB) error {
		// 1. 核对代次所属安装来源，保留代次身份与提交标记。
		var row generation
		err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: generationID}).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("generation %s: %w", generationID, ErrNotFound)
		}
		if err != nil {
			return err
		}
		var install installation
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: row.PluginID}).Take(&install).Error; err != nil {
			return err
		}
		if install.Source != devSource {
			return fmt.Errorf("plugin %s has source %q, expected dev", row.PluginID, install.Source)
		}
		row.ConfigSnapshot, row.ConfigVersion = snapshot, configVersion
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		// 2. 整体替换路径集合，和快照、版本一起提交。
		if err := tx.Where(&generationSecretPath{GenerationID: generationID}).Delete(&generationSecretPath{}).Error; err != nil {
			return err
		}
		for _, path := range secretPaths {
			if err := tx.Create(&generationSecretPath{GenerationID: generationID, Path: path}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GenerationSecretPaths 读取代次保留的 secret 路径模式。
func (s *Store) GenerationSecretPaths(ctx context.Context, generationID string) ([]string, error) {
	if generationID == "" {
		return nil, errors.New("generation_id is required")
	}
	var rows []generationSecretPath
	if err := s.db.WithContext(ctx).Where(&generationSecretPath{GenerationID: generationID}).Find(&rows).Error; err != nil {
		return nil, err
	}
	paths := make([]string, len(rows))
	for i, row := range rows {
		paths[i] = row.Path
	}
	return paths, nil
}
