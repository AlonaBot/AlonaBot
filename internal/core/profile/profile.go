// Package profile 保存并管理会话的配置文件与覆盖设置。
//
// [New] 从状态库与 Active schema 建立一次管理操作的快照；[Service.Resolve]
// 与 [Service.Overrides] 在内存中解析该快照。管理调用方串行化写入，读取 Active
// schema 到提交覆盖值期间持有对应插件的管理锁。
// 本包提供存储与管理，调用派发和 SDK 下发由调用方接入。
package profile

import (
	"cmp"
	"context"
	"crypto/rand"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Default 指定固定存在的基础配置文件 ID。
const Default = "default"
const maxNameRunes = 32

// Profile 标识配置文件的 ID、显示名称与颜色。
type Profile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Summary 为配置文件附上聊天目录中使用它的会话数。
type Summary struct {
	Profile
	Sessions int `json:"sessions"`
}

// Options 绑定同一个 Core 的状态、设置与 Active schema 读取入口。
type Options struct {
	Store    *state.Store
	Settings *settings.File
	Schema   func(context.Context, string) (*configschema.Schema, error) // 无 Active schema 时返回 nil, nil。
}

// Service 持有一次操作的配置文件快照；操作完成后重新构造以观察外部管理变更。
type Service struct {
	opt      Options
	snapshot state.ProfileSnapshot
	installs []state.Install
	schemas  map[string]*configschema.Schema
}

// New 读取管理快照和已安装插件的 Active schema，读取错误原样返回。
func New(ctx context.Context, opt Options) (*Service, error) {
	s := &Service{opt: opt, schemas: make(map[string]*configschema.Schema)}
	if err := s.refresh(ctx); err != nil {
		return nil, err
	}
	installs, err := opt.Store.Installs(ctx)
	if err != nil {
		return nil, err
	}
	s.installs = installs
	for _, in := range installs {
		if in.Uninstalled || in.Role != state.RolePlugin && in.Role != state.RoleAgent {
			continue
		}
		schema, err := opt.Schema(ctx, in.PluginID)
		if err != nil {
			return nil, err
		}
		s.schemas[in.PluginID] = schema
	}
	return s, nil
}

func (s *Service) refresh(ctx context.Context) error {
	snapshot, err := s.opt.Store.Profiles(ctx)
	if err != nil {
		return err
	}
	slices.SortFunc(snapshot.Profiles, func(a, b state.Profile) int { return cmp.Compare(a.Sequence, b.Sequence) })
	s.snapshot = snapshot
	return nil
}

func (s *Service) find(id string) (Profile, error) {
	for _, p := range s.snapshot.Profiles {
		if p.ID == id {
			return Profile{p.ID, p.Name, p.Color}, nil
		}
	}
	return Profile{}, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "profile not found: " + id}
}

// Resolve 按单独指定、实例缺省、default 的顺序解析快照中的配置文件。
func (s *Service) Resolve(instance, session string) Profile {
	id := state.ResolveProfileAssignment(s.snapshot.Assignments, instance, session)
	if id == "" {
		id = Default
	}
	for _, p := range s.snapshot.Profiles {
		if p.ID == id {
			return Profile{p.ID, p.Name, p.Color}
		}
	}
	panic("profile snapshot is missing a referenced profile") // 状态库外键与初始 default 行保证引用完整。
}

// List 按创建顺序列出配置文件，default 在首位；会话数只来自聊天目录。
func (s *Service) List(ctx context.Context) ([]Summary, error) {
	chats, _, err := s.opt.Store.Chats(ctx, state.ChatQuery{})
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, chat := range chats {
		counts[s.Resolve(chat.InstanceID, chat.AdapterID+":"+escapeSession.Replace(chat.ChatID)).ID]++
	}
	result := make([]Summary, 0, len(s.snapshot.Profiles))
	for _, p := range s.snapshot.Profiles {
		result = append(result, Summary{Profile: Profile{p.ID, p.Name, p.Color}, Sessions: counts[p.ID]})
	}
	return result, nil
}

func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 1 || n > maxNameRunes {
		return "", invalid("name must contain 1 to 32 Unicode code points")
	}
	return name, nil
}
func validColor(color string) error {
	switch color {
	case "teal", "navy", "amber", "pink", "green", "purple", "red", "gray":
		return nil
	}
	return invalid("invalid profile color")
}
func (s *Service) uniqueName(name, except string) error {
	for _, p := range s.snapshot.Profiles {
		if p.Name == name && p.ID != except {
			return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, Message: "profile name already exists"}
		}
	}
	return nil
}

// Create 新建配置文件，可复制已有设置，保留来源配置文件的会话指定。
func (s *Service) Create(ctx context.Context, name, color, copyFrom string) (Profile, error) {
	name, err := validName(name)
	if err != nil {
		return Profile{}, err
	}
	if err := validColor(color); err != nil {
		return Profile{}, err
	}
	if err := s.uniqueName(name, ""); err != nil {
		return Profile{}, err
	}
	if copyFrom != "" {
		if _, err := s.find(copyFrom); err != nil {
			return Profile{}, err
		}
	}
	row := state.Profile{ID: rand.Text(), Name: name, Color: color}
	if copyFrom != "" {
		for pluginID := range s.schemas {
			all := s.Overrides(pluginID)
			if values := all[copyFrom]; len(values) > 0 {
				all[row.ID] = values
			}
			if err := checkOverridesLimit(all); err != nil {
				return Profile{}, err
			}
		}
	}
	if err := s.opt.Store.CreateProfile(ctx, row, copyFrom); err != nil {
		return Profile{}, err
	}
	return Profile{row.ID, row.Name, row.Color}, s.refresh(ctx)
}

// Rename 修改给出的名称或颜色，省略的项保持原值。
func (s *Service) Rename(ctx context.Context, id string, name, color *string) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	if name != nil {
		trimmed, err := validName(*name)
		if err != nil {
			return err
		}
		if err := s.uniqueName(trimmed, id); err != nil {
			return err
		}
		name = &trimmed
	}
	if color != nil {
		if err := validColor(*color); err != nil {
			return err
		}
	}
	if err := s.opt.Store.UpdateProfile(ctx, id, name, color); err != nil {
		return err
	}
	return s.refresh(ctx)
}

// Delete 删除配置文件及其指定，使会话回到下一级选择；default 返回 INVALID_ARGUMENT。
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	if id == Default {
		return invalid("default profile cannot be deleted")
	}
	if err := s.opt.Store.DeleteProfile(ctx, id); err != nil {
		return err
	}
	return s.refresh(ctx)
}

func (s *Service) plugin(id string) (state.Install, error) {
	for _, in := range s.installs {
		if in.PluginID == id && !in.Uninstalled {
			if in.Role != state.RolePlugin && in.Role != state.RoleAgent {
				return in, invalid("profile settings require a plugin or agent")
			}
			return in, nil
		}
	}
	return state.Install{}, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "plugin not found: " + id}
}
func invalid(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: message}
}
