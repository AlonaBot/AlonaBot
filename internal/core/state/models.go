package state

import "database/sql"

// clause 使用的列名与下方模型标签保持一致。
const (
	columnSeq        = "seq"
	columnInstallSeq = "install_seq"
)

type schemaVersion struct {
	Version int `gorm:"column:version"`
}

func (schemaVersion) TableName() string { return "schema_version" }

type installation struct {
	PluginID    string `gorm:"column:plugin_id;primaryKey"`
	InstallSeq  int64  `gorm:"column:install_seq"`
	Source      string `gorm:"column:source"`
	Entrypoint  string `gorm:"column:entrypoint"`
	BuildDigest []byte `gorm:"column:build_digest"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:false"` // 毫秒，由调用方显式赋值
	Role        Role   `gorm:"column:role"`
	AdapterID   string `gorm:"column:adapter_id"`
}

func (installation) TableName() string { return "installs" }

type generation struct {
	GenerationID   string `gorm:"column:generation_id;primaryKey"`
	PluginID       string `gorm:"column:plugin_id"`
	ConfigSnapshot string `gorm:"column:config_snapshot"`
	ConfigVersion  uint32 `gorm:"column:config_version"`
	CreatedAt      int64  `gorm:"column:created_at;autoCreateTime:false"` // 毫秒
	CommittedAt    *int64 `gorm:"column:committed_at"`                    // NULL 表示尚未提交
}

func (generation) TableName() string { return "generations" }

type pluginState struct {
	PluginID         string         `gorm:"column:plugin_id;primaryKey"`
	ActiveGeneration sql.NullString `gorm:"column:active_generation"`
	Enabled          bool           `gorm:"column:enabled"`
}

func (pluginState) TableName() string { return "plugin_state" }

type pendingConfig struct {
	PluginID       string `gorm:"column:plugin_id;primaryKey"`
	BaseGeneration string `gorm:"column:base_generation"`
	ConfigJSON     string `gorm:"column:config_json"`
	UpdatedAt      int64  `gorm:"column:updated_at;autoUpdateTime:false"` // 毫秒，保留显式零值
}

func (pendingConfig) TableName() string { return "pending_configs" }

type adapterCursor struct {
	InstanceID string `gorm:"column:instance_id;primaryKey"`
	PluginID   string `gorm:"column:plugin_id"`
	Cursor     []byte `gorm:"column:cursor"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoUpdateTime:false"` // 毫秒
}

func (adapterCursor) TableName() string { return "adapter_cursors" }

type adapterPermission struct {
	PluginID          string `gorm:"column:plugin_id;primaryKey"`
	AdapterInstanceID string `gorm:"column:adapter_instance_id;primaryKey"`
}

func (adapterPermission) TableName() string { return "plugin_adapter_permissions" }

type conversationEntry struct {
	PluginID        string `gorm:"column:plugin_id;primaryKey"`
	ConversationKey string `gorm:"column:conversation_key;primaryKey"`
	Seq             uint64 `gorm:"column:seq;primaryKey;autoIncrement:false"`
	EntryJSON       string `gorm:"column:entry_json"`
	CreatedAt       int64  `gorm:"column:created_at;autoCreateTime:false"` // 毫秒
}

func (conversationEntry) TableName() string { return "conversation_entries" }
