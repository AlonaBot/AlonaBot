package profile

import (
	"context"
	"encoding/json"

	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// Reply 按键保存回复覆盖；nil 跟随 default。
type Reply struct {
	MentionSender    *bool   `json:"mention_sender"`
	Quote            *bool   `json:"quote"`
	Prefix           *string `json:"prefix"`
	ForwardThreshold *int    `json:"forward_threshold"`
}

// Platform 保存命令前缀和回复设置覆盖，nil 前缀跟随 default。
type Platform struct {
	CommandPrefixes *[]string `json:"command_prefixes"`
	Reply           Reply     `json:"reply"`
}

// Platform 返回已保存的覆盖；default 返回 Core 设置的完整值。
func (s *Service) Platform(id string) (Platform, error) {
	if _, err := s.find(id); err != nil {
		return Platform{}, err
	}
	if id == Default {
		current := s.opt.Settings.Settings()
		return Platform{CommandPrefixes: &current.CommandPrefixes, Reply: Reply{MentionSender: &current.Reply.MentionSender, Quote: &current.Reply.Quote, Prefix: &current.Reply.Prefix, ForwardThreshold: &current.Reply.ForwardThreshold}}, nil
	}
	for _, row := range s.snapshot.Platforms {
		if row.ProfileID == id {
			var platform Platform
			err := json.Unmarshal([]byte(row.Settings), &platform)
			return platform, err
		}
	}
	return Platform{}, nil
}

// EffectivePlatform 按键合并已保存的覆盖与当前 default 平台设置。
func (s *Service) EffectivePlatform(id string) (prefixes []string, reply settings.Reply, err error) {
	platform, err := s.Platform(id)
	if err != nil {
		return nil, settings.Reply{}, err
	}
	current := s.opt.Settings.Settings()
	prefixes, reply = current.CommandPrefixes, current.Reply
	if platform.CommandPrefixes != nil {
		prefixes = *platform.CommandPrefixes
	}
	if platform.Reply.MentionSender != nil {
		reply.MentionSender = *platform.Reply.MentionSender
	}
	if platform.Reply.Quote != nil {
		reply.Quote = *platform.Reply.Quote
	}
	if platform.Reply.Prefix != nil {
		reply.Prefix = *platform.Reply.Prefix
	}
	if platform.Reply.ForwardThreshold != nil {
		reply.ForwardThreshold = *platform.Reply.ForwardThreshold
	}
	return prefixes, reply, nil
}

// PutPlatform 整体替换平台覆盖；default 的完整设置写回 core_config.json。
func (s *Service) PutPlatform(ctx context.Context, id string, platform Platform) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	if platform.CommandPrefixes != nil && len(*platform.CommandPrefixes) == 0 {
		return invalid("command_prefixes must not be empty")
	}
	if threshold := platform.Reply.ForwardThreshold; threshold != nil && *threshold < 0 {
		return invalid("reply.forward_threshold must not be negative")
	}
	if id == Default {
		reply := platform.Reply
		if platform.CommandPrefixes == nil || reply.MentionSender == nil || reply.Quote == nil || reply.Prefix == nil || reply.ForwardThreshold == nil {
			return invalid("default platform settings require all fields and reject null")
		}
		return s.opt.Settings.Update(func(doc *settings.Document) error {
			if err := doc.Set([]string{"command_prefixes"}, *platform.CommandPrefixes); err != nil {
				return err
			}
			return doc.Set([]string{"reply"}, settings.Reply{MentionSender: *reply.MentionSender, Quote: *reply.Quote, Prefix: *reply.Prefix, ForwardThreshold: *reply.ForwardThreshold})
		})
	}
	raw, err := json.Marshal(platform)
	if err != nil {
		return err
	}
	if err := s.opt.Store.PutProfilePlatform(ctx, state.ProfilePlatform{ProfileID: id, Settings: string(raw)}); err != nil {
		return err
	}
	return s.refresh(ctx)
}
