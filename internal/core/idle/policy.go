// Package idle 根据生效设置、已提交声明和活动快照判定插件的自动休眠准入与空闲期限。
//
// Controller 由 Core 生命周期循环串行调用；进程停止与唤醒仍由 supervisor 执行。
// 路由判定只读已验证的描述，不启动插件，也不参与消息过滤和权限检查。
package idle

import (
	"fmt"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

// Policy 保存原始覆盖、生效时长与当前准入原因，供决策和管理查询共享。
type Policy struct {
	Policy          *string // nil 表示没有逐插件覆盖。
	EffectivePolicy string
	SleepCapable    bool
	Eligible        bool
	Reason          string
	Warnings        []string
}

// Resolve 按固定原因顺序计算准入；显式时长仅放行路由的常驻要求。
func Resolve(s settings.Sleep, m *manifest.Manifest, independent bool, routes []*pb.RouteDescriptor) Policy {
	p := Policy{SleepCapable: m.SleepCapable, EffectivePolicy: "never"}
	if value, ok := s.Plugins[m.ID]; ok {
		p.Policy = &value
	}
	switch {
	case s.Mode == "always_on":
		p.Reason = "always_on"
	case !m.SleepCapable || m.Role != "plugin" && m.Role != "agent":
		p.Reason = "manifest_requires_resident"
	case !independent:
		p.Reason = "in_process"
	case p.Policy != nil && *p.Policy == "never":
		p.Reason = "policy_never"
	default:
		for _, route := range routes {
			if reason := residentRoute(route); reason != "" {
				if p.Policy == nil {
					p.Reason = "route_requires_resident"
					return p
				}
				p.Warnings = append(p.Warnings, fmt.Sprintf("route %s: %s may cause frequent wakeups", route.GetRouteId(), reason))
			}
		}
		p.Eligible = true
		p.EffectivePolicy = s.IdleTimeout.String()
		if p.Policy != nil {
			duration, _ := time.ParseDuration(*p.Policy) // settings 已校验，只规范化显示形式。
			p.EffectivePolicy = duration.String()
		}
	}
	return p
}

func residentRoute(route *pb.RouteDescriptor) string {
	switch route.GetKind() {
	case pb.RouteKind_ROUTE_KIND_HTTP:
		return "HTTP"
	case pb.RouteKind_ROUTE_KIND_HOOK:
		return "Hook"
	case pb.RouteKind_ROUTE_KIND_BUS_SUBSCRIPTION:
		return "Bus subscription"
	case pb.RouteKind_ROUTE_KIND_COMMAND:
		if route.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_RAW {
			return "Raw"
		}
	case pb.RouteKind_ROUTE_KIND_EVENT:
		event := route.GetEvent()
		for _, name := range event.GetEventTypeFilter() {
			switch name {
			case "button.clicked":
			case "message.created", "message.edited", "message.deleted":
				if event.GetMessageFilter() == nil {
					return "unfiltered message event"
				}
			default:
				return "continuous event listener"
			}
		}
	}
	return ""
}
