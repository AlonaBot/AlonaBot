package state

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	columnProfileSequence  = "profile_seq"
	columnProfileSession   = "session"
	columnProfileOperation = "operation"
)

// Profile 保存配置文件的身份与创建顺序。
type Profile struct {
	ID       string `gorm:"column:id;primaryKey"`
	Name     string `gorm:"column:name"`
	Color    string `gorm:"column:color"`
	Sequence int64  `gorm:"column:profile_seq"`
}

// TableName 指定配置文件表。
func (Profile) TableName() string { return "profile" }

// ProfileAssignment 保存实例缺省或单独指定；Session 空串表示实例缺省。
type ProfileAssignment struct {
	Instance  string `gorm:"column:instance;primaryKey"`
	Session   string `gorm:"column:session;primaryKey"`
	ProfileID string `gorm:"column:profile_id"`
}

// TableName 指定配置文件的会话指定表。
func (ProfileAssignment) TableName() string { return "profile_assignment" }

// ProfileOverride 保留插件字段覆盖的原始 JSON，包括已失效的项。
type ProfileOverride struct {
	ProfileID string `gorm:"column:profile_id;primaryKey"`
	PluginID  string `gorm:"column:plugin_id;primaryKey"`
	Path      string `gorm:"column:path;primaryKey"`
	Value     string `gorm:"column:value"`
}

// TableName 指定字段覆盖表。
func (ProfileOverride) TableName() string { return "profile_override" }

// ProfilePlugin 保存配置文件内的插件开关；缺行表示启用。
type ProfilePlugin struct {
	ProfileID string `gorm:"column:profile_id;primaryKey"`
	PluginID  string `gorm:"column:plugin_id;primaryKey"`
	Enabled   bool   `gorm:"column:enabled"`
}

// TableName 指定插件开关表。
func (ProfilePlugin) TableName() string { return "profile_plugin" }

// ProfileTool 保存工具或确认操作的策略覆盖。
type ProfileTool struct {
	ProfileID string `gorm:"column:profile_id;primaryKey"`
	PluginID  string `gorm:"column:plugin_id;primaryKey"`
	Name      string `gorm:"column:name;primaryKey"`
	Operation bool   `gorm:"column:operation;primaryKey"`
	Policy    string `gorm:"column:policy"`
}

// TableName 指定工具策略覆盖表。
func (ProfileTool) TableName() string { return "profile_tool" }

// ProfilePlatform 保存配置文件的平台设置，JSON 中的 null 表示继承。
type ProfilePlatform struct {
	ProfileID string `gorm:"column:profile_id;primaryKey"`
	Settings  string `gorm:"column:settings"`
}

// TableName 指定平台设置表。
func (ProfilePlatform) TableName() string { return "profile_platform" }

// ProfileSnapshot 提供同一状态库快照中的配置文件与全部管理设置。
type ProfileSnapshot struct {
	Profiles    []Profile
	Assignments []ProfileAssignment
	Overrides   []ProfileOverride
	Plugins     []ProfilePlugin
	Tools       []ProfileTool
	Platforms   []ProfilePlatform
}

// Profiles 在一个事务中读取配置文件管理快照。
func (s *Store) Profiles(ctx context.Context) (ProfileSnapshot, error) {
	var result ProfileSnapshot
	err := s.tx(ctx, func(tx *gorm.DB) error {
		for _, rows := range []any{&result.Profiles, &result.Assignments, &result.Overrides, &result.Plugins, &result.Tools, &result.Platforms} {
			if err := tx.Find(rows).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

// CreateProfile 分配创建顺序，并在同一事务内复制来源的设置；不复制指定。
func (s *Store) CreateProfile(ctx context.Context, row Profile, copyFrom string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var last Profile
		if err := tx.Order(clause.OrderByColumn{Column: clause.Column{Name: columnProfileSequence}, Desc: true}).Take(&last).Error; err != nil {
			return err
		}
		row.Sequence = last.Sequence + 1
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if copyFrom == "" {
			return nil
		}
		var source Profile
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: copyFrom}).Take(&source).Error; err != nil {
			return err
		}
		var values []ProfileOverride
		var plugins []ProfilePlugin
		var tools []ProfileTool
		var platforms []ProfilePlatform
		for _, query := range []struct{ key, rows any }{
			{&ProfileOverride{ProfileID: copyFrom}, &values},
			{&ProfilePlugin{ProfileID: copyFrom}, &plugins},
			{&ProfileTool{ProfileID: copyFrom}, &tools},
			{&ProfilePlatform{ProfileID: copyFrom}, &platforms},
		} {
			if err := tx.Where(query.key).Find(query.rows).Error; err != nil {
				return err
			}
		}
		for _, value := range values {
			value.ProfileID = row.ID
			if err := tx.Create(&value).Error; err != nil {
				return err
			}
		}
		for _, plugin := range plugins {
			plugin.ProfileID = row.ID
			if err := tx.Create(&plugin).Error; err != nil {
				return err
			}
		}
		for _, tool := range tools {
			tool.ProfileID = row.ID
			if err := tx.Create(&tool).Error; err != nil {
				return err
			}
		}
		for _, platform := range platforms {
			platform.ProfileID = row.ID
			if err := tx.Create(&platform).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateProfile 在事务中修改名称与颜色，保留创建顺序。
func (s *Store) UpdateProfile(ctx context.Context, id string, name, color *string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var row Profile
		if err := tx.Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Take(&row).Error; err != nil {
			return err
		}
		if name != nil {
			row.Name = *name
		}
		if color != nil {
			row.Color = *color
		}
		return tx.Save(&row).Error
	})
}

// DeleteProfile 删除配置文件并由外键级联清除设置与指定。
func (s *Store) DeleteProfile(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).Delete(&Profile{}).Error
}

// AssignProfile 原子写入或取消指定；空 profileID 取消，空 session 选择实例缺省。
func (s *Store) AssignProfile(ctx context.Context, instance, session, profileID string) error {
	row := ProfileAssignment{Instance: instance, Session: session, ProfileID: profileID}
	if profileID == "" {
		return s.db.WithContext(ctx).Where(&ProfileAssignment{Instance: instance}).Where(clause.Eq{Column: columnProfileSession, Value: session}).Delete(&ProfileAssignment{}).Error
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}

// DeleteProfileAssignments 清除已删除 Adapter 实例的全部指定。
func (s *Store) DeleteProfileAssignments(ctx context.Context, instances []string) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		for _, instance := range instances {
			if err := tx.Where(&ProfileAssignment{Instance: instance}).Delete(&ProfileAssignment{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PatchProfilePlugin 同时保存开关与字段覆盖；null 删除字段，调用方已完成全部校验。
func (s *Store) PatchProfilePlugin(ctx context.Context, profileID, pluginID string, enabled *bool, values map[string]json.RawMessage) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		if enabled != nil {
			row := ProfilePlugin{ProfileID: profileID, PluginID: pluginID, Enabled: *enabled}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		for path, value := range values {
			row := ProfileOverride{ProfileID: profileID, PluginID: pluginID, Path: path, Value: string(value)}
			if string(value) == "null" {
				if err := tx.Delete(&row).Error; err != nil {
					return err
				}
			} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PutProfilePlatform 替换配置文件的平台设置。
func (s *Store) PutProfilePlatform(ctx context.Context, row ProfilePlatform) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error
}

// PutProfileTools 在同一事务中更新所列策略；inherit 删除该项。
func (s *Store) PutProfileTools(ctx context.Context, rows []ProfileTool) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		for _, row := range rows {
			if row.Policy == "inherit" {
				key := &ProfileTool{ProfileID: row.ProfileID, PluginID: row.PluginID, Name: row.Name}
				if err := tx.Where(key).Where(clause.Eq{Column: columnProfileOperation, Value: row.Operation}).Delete(&ProfileTool{}).Error; err != nil {
					return err
				}
			} else if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ProfilePurgeObjects 列出 --purge 将由安装外键级联清除的配置文件记录。
func (s *Store) ProfilePurgeObjects(ctx context.Context, pluginID string) ([]string, error) {
	var result []string
	for _, model := range []interface{ TableName() string }{
		ProfileOverride{PluginID: pluginID}, ProfilePlugin{PluginID: pluginID}, ProfileTool{PluginID: pluginID},
	} {
		var count int64
		if err := s.db.WithContext(ctx).Model(model).Where(model).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			result = append(result, model.TableName())
		}
	}
	return result, nil
}
