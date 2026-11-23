package profile

import (
	"context"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const (
	sessionAccessAll   = "all"
	sessionAccessAllow = "allow"
	sessionAccessDeny  = "deny"
	maxAccessSessions  = 1000
)

// SessionAccess 是配置文件的会话访问模式与完整列表。
type SessionAccess struct {
	Mode     string                 `json:"mode"`
	Sessions []state.ProfileSession `json:"sessions"`
}

// SessionAccess 读取指定配置文件的设置，缺省时返回 all 和空列表。
func (s *Service) SessionAccess(ctx context.Context, id string) (SessionAccess, error) {
	if _, err := s.find(id); err != nil {
		return SessionAccess{}, err
	}
	row, err := s.opt.Store.ProfileSessionAccess(ctx, id)
	if err != nil {
		return SessionAccess{}, err
	}
	if row.Mode == "" {
		return SessionAccess{Mode: sessionAccessAll, Sessions: []state.ProfileSession{}}, nil
	}
	return SessionAccess{Mode: row.Mode, Sessions: row.Sessions}, nil
}

// PutSessionAccess 校验并替换指定配置文件的模式和列表，保留条目的原顺序。
func (s *Service) PutSessionAccess(ctx context.Context, id string, access SessionAccess) error {
	if _, err := s.find(id); err != nil {
		return err
	}
	switch access.Mode {
	case sessionAccessAll, sessionAccessAllow, sessionAccessDeny:
	default:
		return invalid("mode must be all, allow or deny")
	}
	if len(access.Sessions) > maxAccessSessions {
		return invalid("sessions must contain at most 1000 entries")
	}
	unique := make(map[state.ProfileSession]bool, len(access.Sessions))
	sessions := make([]state.ProfileSession, 0, len(access.Sessions))
	for _, entry := range access.Sessions {
		adapterID, rest, _ := strings.Cut(entry.Session, ":")
		chat, _, _ := strings.Cut(rest, "#")
		if strings.TrimSpace(entry.Instance) == "" || strings.TrimSpace(adapterID) == "" || strings.TrimSpace(chat) == "" {
			return invalid("each entry requires a nonempty instance and session")
		}
		if err := validateSession(entry.Session, adapterID); err != nil {
			return err
		}
		if !unique[entry] {
			sessions = append(sessions, entry)
			unique[entry] = true
		}
	}
	return s.opt.Store.PutProfileSessionAccess(ctx, state.ProfileSessionAccess{ProfileID: id, Mode: access.Mode, Sessions: sessions})
}

// SessionDenied 判断可信目标是否被它实际选中的配置文件拒绝。
func SessionDenied(ctx context.Context, store *state.Store, target *pb.Target) (bool, error) {
	if target.GetChatId() == "" {
		return false, nil
	}
	// 显式空话题和未带话题访问同一个聊天。
	entry := state.ProfileSession{Instance: target.GetAdapterInstanceId(), Session: strings.TrimSuffix(TargetSession(target), "#")}
	access, err := store.SessionProfileAccess(ctx, entry.Instance, entry.Session, Default)
	if err != nil {
		return false, err
	}
	switch access.Mode {
	case sessionAccessAllow:
		return !slices.Contains(access.Sessions, entry), nil
	case sessionAccessDeny:
		return slices.Contains(access.Sessions, entry), nil
	default:
		return false, nil
	}
}
