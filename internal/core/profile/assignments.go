package profile

import (
	"cmp"
	"context"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

var escapeSession = strings.NewReplacer("%", "%25", "#", "%23")
var unescapeSession = strings.NewReplacer("%25", "%", "%23", "#")

// Assignment 表示一个会话的显式指定。
type Assignment struct {
	Instance string `json:"instance"`
	Session  string `json:"session"`
	Profile  string `json:"profile"`
}

// Assignments 列出实例缺省与会话的显式指定，空集合始终编码为 {} 与 []。
type Assignments struct {
	Instances map[string]string `json:"instances"`
	Sessions  []Assignment      `json:"sessions"`
}

// Assignments 返回快照中的显式指定，不展开继承值。
func (s *Service) Assignments() Assignments {
	result := Assignments{Instances: map[string]string{}, Sessions: []Assignment{}}
	for _, a := range s.snapshot.Assignments {
		if a.Session == "" {
			result.Instances[a.Instance] = a.ProfileID
		} else {
			result.Sessions = append(result.Sessions, Assignment{a.Instance, a.Session, a.ProfileID})
		}
	}
	slices.SortFunc(result.Sessions, func(a, b Assignment) int {
		if order := cmp.Compare(a.Instance, b.Instance); order != 0 {
			return order
		}
		return cmp.Compare(a.Session, b.Session)
	})
	return result
}

// Assign 设置实例缺省或单独会话的配置文件；profileID 为 nil 时取消指定。
func (s *Service) Assign(ctx context.Context, instance, session string, profileID *string) error {
	var adapterID string
	for id, group := range s.opt.Settings.Settings().Adapters {
		if _, ok := group.Instances[instance]; ok {
			adapterID = id
			break
		}
	}
	if adapterID == "" {
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "adapter instance not found"}
	}
	if session != "" {
		if err := validateSession(session, adapterID); err != nil {
			return err
		}
	}
	var id string
	if profileID != nil {
		if _, err := s.find(*profileID); err != nil {
			return err
		}
		id = *profileID
	}
	if err := s.opt.Store.AssignProfile(ctx, instance, session, id); err != nil {
		return err
	}
	return s.refresh(ctx)
}

func validateSession(session, adapterID string) error {
	adapter, rest, found := strings.Cut(session, ":")
	if !found || adapter != adapterID {
		return invalid("session adapter does not match the instance")
	}
	chat, thread, topic := strings.Cut(rest, "#")
	if chat == "" || topic && thread == "" || strings.Contains(thread, "#") {
		return invalid("invalid session string")
	}
	for _, part := range []string{chat, thread} {
		if escapeSession.Replace(unescapeSession.Replace(part)) != part {
			return invalid("invalid session string encoding")
		}
	}
	return nil
}
