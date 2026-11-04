package policy

import (
	"cmp"
	"context"
	"slices"

	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

// DenyKind 区分被禁止的用户与会话。
type DenyKind string

const (
	// DenyUser 禁止指定用户触发命令。
	DenyUser DenyKind = "user"
	// DenyChat 禁止指定会话触发命令及接收发送。
	DenyChat DenyKind = "chat"
)

// Deny 保存一个实例作用域中的用户或会话禁止。
type Deny struct {
	Kind                  DenyKind
	InstanceID, SubjectID string
}

// Deny 幂等禁止指定用户或会话，持久化提交后才返回成功。
func (p *Policy) Deny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error {
	if err := checkDeny(kind, instanceID, subjectID); err != nil {
		return err
	}
	return p.store.AddGlobalDeny(ctx, state.GlobalDeny{Kind: string(kind), InstanceID: instanceID, SubjectID: subjectID})
}

// Undeny 按完整身份幂等取消用户或会话禁止。
func (p *Policy) Undeny(ctx context.Context, kind DenyKind, instanceID, subjectID string) error {
	if err := checkDeny(kind, instanceID, subjectID); err != nil {
		return err
	}
	return p.store.RemoveGlobalDeny(ctx, state.GlobalDeny{Kind: string(kind), InstanceID: instanceID, SubjectID: subjectID})
}

// Denies 按种类、实例与主体 ID 的字典序读取禁止名单。
func (p *Policy) Denies(ctx context.Context) ([]Deny, error) {
	rows, err := p.store.GlobalDenies(ctx)
	denies := make([]Deny, 0, len(rows))
	for _, row := range rows {
		denies = append(denies, Deny{Kind: DenyKind(row.Kind), InstanceID: row.InstanceID, SubjectID: row.SubjectID})
	}
	slices.SortFunc(denies, func(a, b Deny) int {
		if order := cmp.Compare(a.Kind, b.Kind); order != 0 {
			return order
		}
		if order := cmp.Compare(a.InstanceID, b.InstanceID); order != 0 {
			return order
		}
		return cmp.Compare(a.SubjectID, b.SubjectID)
	})
	return denies, err
}

func checkDeny(kind DenyKind, instanceID, subjectID string) error {
	if kind != DenyUser && kind != DenyChat {
		return invalidArgument("unknown deny kind")
	}
	if instanceID == "" || subjectID == "" {
		return invalidArgument("instance ID and subject ID are required")
	}
	return nil
}
