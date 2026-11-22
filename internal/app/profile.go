package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// profileSchema 只读取 Active 对应声明，文件读取和声明错误交还调用方诊断。
func (a *assembly) profileSchema(ctx context.Context, id string) (*configschema.Schema, error) {
	active, err := a.store.Active(ctx, id)
	if err != nil {
		return nil, err
	}
	return a.profileSchemaAt(ctx, id, active)
}

func (a *assembly) profileSchemaAt(ctx context.Context, id, active string) (*configschema.Schema, error) {
	if active == "" {
		return nil, nil
	}
	a.schemaMu.RLock()
	schemas := a.schemas[id]
	a.schemaMu.RUnlock()
	if schemas.generation == active {
		return schemas.config, nil
	}
	installed, err := a.store.GetInstall(ctx, id)
	if err != nil {
		return nil, err
	}
	var raw []byte
	if installed.Source == devSource {
		raw, _, err = a.core.developmentConfig(id)
	} else {
		directory, directoryErr := installedDirectory(a.dir, installed)
		if directoryErr != nil {
			return nil, directoryErr
		}
		if _, legacyErr := os.Lstat(filepath.Join(directory, "config_schema.json")); legacyErr == nil {
			return nil, fmt.Errorf("component %s: rename config_schema.json to plugin_config_schema.json and rebuild", id)
		} else if !errors.Is(legacyErr, os.ErrNotExist) {
			return nil, legacyErr
		}
		raw, err = os.ReadFile(filepath.Join(directory, "plugin_config_schema.json"))
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid active config schema for %s: %v", id, problems)
	}
	return schema, nil
}

func (a *assembly) profiles(ctx context.Context, candidateID string, candidate *configschema.Schema) (*profile.Service, error) {
	return profile.New(ctx, profile.Options{Store: a.store, Settings: a.settings, Schema: func(ctx context.Context, id string) (*configschema.Schema, error) {
		if id == candidateID {
			return candidate, nil
		}
		return nil, nil
	}})
}

func targetProfile(profiles *profile.Service, target *pb.Target) profile.Profile {
	if target.GetChatId() == "" {
		return profiles.Resolve(target.GetAdapterInstanceId(), "")
	}
	escape := strings.NewReplacer("%", "%25", "#", "%23")
	session := target.GetAdapterId() + ":" + escape.Replace(target.GetChatId())
	if target.ThreadId != nil {
		session += "#" + escape.Replace(target.GetThreadId())
	}
	return profiles.Resolve(target.GetAdapterInstanceId(), session)
}

func (a *assembly) toolProfile(ctx context.Context, target *pb.Target) (*profile.Service, string, error) {
	profiles, err := a.profiles(ctx, "", nil)
	if err != nil {
		return nil, "", err
	}
	return profiles, targetProfile(profiles, target).ID, nil
}

func (a *assembly) resolveProfile(ctx context.Context, id string, target *pb.Target) (string, bool, error) {
	view, err := a.profileView(ctx, id, target)
	return view.ProfileID, view.Enabled, err
}

// profileView 在准入与过滤共用的 context 中冻结选择、开关、Active 和有效覆盖。
func (a *assembly) profileView(ctx context.Context, id string, target *pb.Target) (profile.View, error) {
	return profile.ReadView(ctx, func() (profile.View, error) {
		active, err := a.store.Active(ctx, id)
		if err != nil {
			return profile.View{}, err
		}
		if active == "" {
			return profile.View{}, fmt.Errorf("plugin %s has no active generation", id)
		}
		schema, err := a.profileSchemaAt(ctx, id, active)
		if err != nil {
			return profile.View{}, err
		}
		profiles, err := a.profiles(ctx, id, schema)
		if err != nil {
			return profile.View{}, err
		}
		selected := targetProfile(profiles, target)
		plugins, err := profiles.Plugins(selected.ID)
		if err != nil {
			return profile.View{}, err
		}
		view := profile.View{ProfileID: selected.ID, Enabled: true}
		for _, plugin := range plugins {
			if plugin.PluginID == id {
				view.Enabled = plugin.Enabled
				break
			}
		}
		raw, err := a.store.GenerationConfig(ctx, active)
		if err != nil {
			return profile.View{}, fmt.Errorf("read active generation config: %w", err)
		}
		view.Config = json.RawMessage(raw)
		if overrides := profiles.Overrides(id)[selected.ID]; len(overrides) > 0 {
			view.Config, err = schema.ApplyProfile(view.Config, overrides)
			if err != nil {
				return profile.View{}, err
			}
		}
		return view, nil
	})
}

func (a *assembly) bootstrapProfiles(ctx context.Context, id string, schema *configschema.Schema, boot *pb.Bootstrap) error {
	profiles, err := a.profiles(ctx, id, schema)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(profiles.Overrides(id))
	if err != nil {
		return err
	}
	boot.ProfileOverridesJson = string(raw)
	return nil
}

// notifyProfileOverrides 从新快照通知当前在线 Session；失败关闭原 Session，交还既有退出路径。
func (a *assembly) notifyProfileOverrides(ctx context.Context, id string) (err error) {
	plugin, online := a.router.Plugin(id)
	if !online || !plugin.Activated() {
		return nil
	}
	instance, ok := a.components.Load(id)
	if !ok {
		return nil
	}
	current := instance.(*host.Instance)
	defer func() {
		if err != nil {
			err = errors.Join(err, current.Session.Close())
		}
	}()
	// 1. 读取同一次操作的有效集合，离线插件在下次 Bootstrap 接收。
	schema, err := a.profileSchema(ctx, id)
	if err != nil {
		return err
	}
	profiles, err := a.profiles(ctx, id, schema)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(profiles.Overrides(id))
	if err != nil {
		return err
	}
	request := &pb.ProfileSetOverridesRequest{ProfileOverridesJson: string(raw)}
	encoded, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	// 2. 等待完整替换确认；任何发送、业务或协议错误沿原 Session 退出路径处理。
	response, err := plugin.Call(ctx, wire.Request{Target: wire.ServiceProfileSetOverrides, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: 1, Bytes: encoded}})
	if err != nil {
		return err
	}
	if response.Err != nil {
		return &wire.CallError{ErrorInfo: response.Err}
	}
	reply := &pb.ProfileSetOverridesResponse{}
	if response.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || response.Payload.SchemaID != validate.SchemaID(reply) || response.Payload.SchemaVersion != 1 {
		return errors.New("invalid profile overrides response metadata")
	}
	return proto.Unmarshal(response.Payload.Bytes, reply)
}

func (a *assembly) replySettings(ctx context.Context, target *pb.Target) (settings.Reply, error) {
	profiles, err := a.profiles(ctx, "", nil)
	if err != nil {
		return settings.Reply{}, err
	}
	_, reply, err := profiles.EffectivePlatform(targetProfile(profiles, target).ID)
	return reply, err
}

func (a *assembly) commandPrefixes(ctx context.Context, target *pb.Target) ([]string, error) {
	profiles, err := a.profiles(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	prefixes, _, err := profiles.EffectivePlatform(targetProfile(profiles, target).ID)
	return prefixes, err
}
