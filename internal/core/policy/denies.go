package policy

import (
	"cmp"
	"context"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// DenyKind 区分被禁止的用户与会话。
type DenyKind string

const (
	// DenyUser 禁止指定用户触发命令。
	DenyUser DenyKind = "user"
	// DenyChat 禁止指定会话触发命令及接收发送。
	DenyChat DenyKind = "chat"
)

// Deny 保存一个 Adapter 作用域中的用户或会话禁止。
type Deny struct {
	Kind                 DenyKind
	AdapterID, SubjectID string
}

// Deny 幂等禁止指定用户或会话，持久化提交后才返回成功。
func (p *Policy) Deny(ctx context.Context, kind DenyKind, adapterID, subjectID string) error {
	if err := checkDeny(kind, adapterID, subjectID); err != nil {
		return err
	}
	return p.store.AddGlobalDeny(ctx, state.GlobalDeny{Kind: string(kind), AdapterID: adapterID, SubjectID: subjectID})
}

// Undeny 按完整身份幂等取消用户或会话禁止。
func (p *Policy) Undeny(ctx context.Context, kind DenyKind, adapterID, subjectID string) error {
	if err := checkDeny(kind, adapterID, subjectID); err != nil {
		return err
	}
	return p.store.RemoveGlobalDeny(ctx, state.GlobalDeny{Kind: string(kind), AdapterID: adapterID, SubjectID: subjectID})
}

// Denies 按种类、Adapter 与主体 ID 的字典序读取禁止名单。
func (p *Policy) Denies(ctx context.Context) ([]Deny, error) {
	rows, err := p.store.GlobalDenies(ctx)
	denies := make([]Deny, 0, len(rows))
	for _, row := range rows {
		denies = append(denies, Deny{Kind: DenyKind(row.Kind), AdapterID: row.AdapterID, SubjectID: row.SubjectID})
	}
	slices.SortFunc(denies, func(a, b Deny) int {
		if order := cmp.Compare(a.Kind, b.Kind); order != 0 {
			return order
		}
		if order := cmp.Compare(a.AdapterID, b.AdapterID); order != 0 {
			return order
		}
		return cmp.Compare(a.SubjectID, b.SubjectID)
	})
	return denies, err
}

func checkDeny(kind DenyKind, adapterID, subjectID string) error {
	if kind != DenyUser && kind != DenyChat {
		return invalidArgument("unknown deny kind")
	}
	if _, err := validate.PluginID(adapterID); err != nil || strings.HasPrefix(adapterID, validate.MarketPrefix) {
		return invalidArgument("invalid adapter ID: " + adapterID)
	}
	if subjectID == "" {
		return invalidArgument("subject ID is required")
	}
	return nil
}

func invalidArgument(message string) error {
	return &wire.CallError{ErrorInfo: &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: message}}
}
