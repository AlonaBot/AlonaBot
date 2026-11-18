package manage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
)

type toolPolicy struct {
	PluginID string `json:"plugin_id"`
	Service  string `json:"service"`
	Default  string `json:"default"`
	Override string `json:"override"`
	Effect   string `json:"effect"`
}

type toolPolicyUpdate struct {
	PluginID string `json:"plugin_id"`
	Service  string `json:"service"`
	Policy   string `json:"policy"`
}

func (s *Service) toolRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/tools/policies", func(w http.ResponseWriter, r *http.Request) {
		entries, err := s.opt.Tools.Policies(r.Context())
		result := make([]toolPolicy, 0, len(entries))
		for _, entry := range entries {
			result = append(result, toolPolicy{entry.PluginID, entry.Service, toolPolicyName(entry.Default), toolPolicyName(entry.Override), toolEffectName(entry.Effect)})
		}
		respond(w, struct {
			Policies []toolPolicy `json:"policies"`
		}{result}, err)
	})
	mux.HandleFunc("PUT /api/v1/tools/policies", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			toolPolicyUpdate
			Items *[]json.RawMessage `json:"items"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		items := []toolPolicyUpdate{request.toolPolicyUpdate}
		if request.Items != nil {
			items = make([]toolPolicyUpdate, 0, len(*request.Items))
			for index, raw := range *request.Items {
				var item toolPolicyUpdate
				decoder := json.NewDecoder(bytes.NewReader(raw))
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&item); err != nil {
					respond(w, nil, invalid(fmt.Sprintf("items[%d]: invalid request JSON: %v", index, err)))
					return
				}
				items = append(items, item)
			}
			// 安装替换持有 catalog 后再锁新旧组件，批量写入沿用相同的锁层级。
			s.catalog.Lock()
			defer s.catalog.Unlock()
		}
		// 1. 按固定顺序持有涉及组件的管理锁，避免相反顺序的批量请求互锁。
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.PluginID)
		}
		slices.Sort(ids)
		for _, id := range slices.Compact(ids) {
			unlock := s.lock(id)
			defer unlock()
		}
		// 2. 全部项在写入前校验，数据库事务再保证保存时的原子性。
		rows := make([]state.ToolPolicy, 0, len(items))
		for index, item := range items {
			row, err := s.checkToolPolicy(r, item)
			if err != nil {
				if request.Items != nil {
					var detail *validate.Error
					errors.As(failure(err), &detail)
					code := detail.Code
					if detail.Code == pb.ErrorCode_ERROR_CODE_NOT_FOUND || detail.Code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
						code = pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
					}
					err = &validate.Error{Code: code, Message: fmt.Sprintf("items[%d]: %s", index, detail.Message)}
				}
				respond(w, nil, err)
				return
			}
			rows = append(rows, row)
		}
		respond(w, struct{}{}, s.opt.Tools.SetPolicy(r.Context(), rows...))
	})
}

func (s *Service) checkToolPolicy(r *http.Request, item toolPolicyUpdate) (state.ToolPolicy, error) {
	row := state.ToolPolicy{PluginID: item.PluginID, Service: item.Service}
	if !validate.IsServiceName(item.Service) {
		// 操作名独立于 Service 语法，只接纳目录已记录的完整名字。
		entries, err := s.opt.Tools.Policies(r.Context())
		if err != nil {
			return row, err
		}
		if item.Service == "" || !slices.ContainsFunc(entries, func(entry tool.ToolPolicyEntry) bool {
			return entry.PluginID == item.PluginID && entry.Service == item.Service
		}) {
			return row, invalid("invalid tool service or unknown operation name")
		}
	}
	switch item.Policy {
	case "inherit":
		row.Policy = pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED
	case "allow":
		row.Policy = pb.ToolPolicy_TOOL_POLICY_ALLOW
	case "confirm":
		row.Policy = pb.ToolPolicy_TOOL_POLICY_CONFIRM
	case "deny":
		row.Policy = pb.ToolPolicy_TOOL_POLICY_DENY
	default:
		return row, invalid("unknown tool policy")
	}
	_, err := s.installed(r.Context(), item.PluginID)
	return row, err
}

func toolEffectName(value pb.ToolEffect) string {
	switch value {
	case pb.ToolEffect_TOOL_EFFECT_READ_ONLY:
		return "read_only"
	case pb.ToolEffect_TOOL_EFFECT_WRITE:
		return "write"
	case pb.ToolEffect_TOOL_EFFECT_DESTRUCTIVE:
		return "destructive"
	default:
		return "unspecified"
	}
}

func toolPolicyName(value pb.ToolPolicy) string {
	switch value {
	case pb.ToolPolicy_TOOL_POLICY_ALLOW:
		return "allow"
	case pb.ToolPolicy_TOOL_POLICY_CONFIRM:
		return "confirm"
	case pb.ToolPolicy_TOOL_POLICY_DENY:
		return "deny"
	default:
		return "inherit"
	}
}
