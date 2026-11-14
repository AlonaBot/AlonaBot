package manage

import (
	"net/http"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
)

type toolPolicy struct {
	PluginID string `json:"plugin_id"`
	Service  string `json:"service"`
	Default  string `json:"default"`
	Override string `json:"override"`
}

func (s *Service) toolRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/tools/policies", func(w http.ResponseWriter, r *http.Request) {
		entries, err := s.opt.Tools.Policies(r.Context())
		result := make([]toolPolicy, 0, len(entries))
		for _, entry := range entries {
			result = append(result, toolPolicy{entry.PluginID, entry.Service, toolPolicyName(entry.Default), toolPolicyName(entry.Override)})
		}
		respond(w, struct {
			Policies []toolPolicy `json:"policies"`
		}{result}, err)
	})
	mux.HandleFunc("PUT /api/v1/tools/policies", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			PluginID string `json:"plugin_id"`
			Service  string `json:"service"`
			Policy   string `json:"policy"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if !validate.IsServiceName(request.Service) {
			// ConfirmOperation 的操作名独立于 Service 语法，只接纳目录已记录的完整名字。
			entries, err := s.opt.Tools.Policies(r.Context())
			if err != nil {
				respond(w, nil, err)
				return
			}
			if request.Service == "" || !slices.ContainsFunc(entries, func(entry tool.ToolPolicyEntry) bool {
				return entry.PluginID == request.PluginID && entry.Service == request.Service
			}) {
				respond(w, nil, invalid("invalid tool service or unknown operation name"))
				return
			}
		}
		var value pb.ToolPolicy
		switch request.Policy {
		case "inherit":
			value = pb.ToolPolicy_TOOL_POLICY_UNSPECIFIED
		case "allow":
			value = pb.ToolPolicy_TOOL_POLICY_ALLOW
		case "confirm":
			value = pb.ToolPolicy_TOOL_POLICY_CONFIRM
		case "deny":
			value = pb.ToolPolicy_TOOL_POLICY_DENY
		default:
			respond(w, nil, invalid("unknown tool policy"))
			return
		}
		unlock := s.lock(request.PluginID)
		defer unlock()
		_, err := s.installed(r.Context(), request.PluginID)
		if err == nil {
			err = s.opt.Tools.SetPolicy(r.Context(), request.PluginID, request.Service, value)
		}
		respond(w, struct{}{}, err)
	})
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
