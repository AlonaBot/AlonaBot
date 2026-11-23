package router

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/debugapi"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/profile"
	"google.golang.org/protobuf/proto"
)

const (
	commandInvokedEvent = "command.invoked"
	commandOK           = "ok"
	commandError        = "error"
	commandDenied       = "denied"
)

func (r *Router) commandAllowed(c *candidate, ev *pb.Event, state *dispatchState) bool {
	if !r.available(*c) {
		return false
	}
	if failure := r.commandPolicyFailure(profile.WithView(r.ctx), c, ev); failure != nil {
		r.recordDiagnostic(state, *c, nil, debugapi.PhaseAdmission, failure)
		r.commandInvoked(*c, ev, commandPolicyResult(failure.GetCode()), 0)
		return false
	}
	return true
}

func commandPolicyResult(code pb.ErrorCode) string {
	// FAILED_PRECONDITION 表示命令已被策略停用。
	if code == pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || code == pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
		return commandDenied
	}
	return commandError
}

// commandInvoked 只保留调用身份、结果及原 Target/Actor，沿有界事件队列交付。
func (r *Router) commandInvoked(c candidate, original *pb.Event, result string, duration time.Duration) {
	name := c.route.GetRouteId()
	if c.route.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_PATH {
		names := c.names
		if names == nil {
			names = declaredCommandNames(c.route.GetCommand())
		}
		name = strings.Join(names[0], " ")
	}
	id := make([]byte, traceIDBytes)
	rand.Read(id)
	ev := &pb.Event{
		EventId: hex.EncodeToString(id), EventType: commandInvokedEvent, OccurredAtUnixMs: time.Now().UnixMilli(),
		SourceKind: original.GetSourceKind(), AdapterId: original.AdapterId,
		AdapterInstanceId: original.AdapterInstanceId, AccountId: original.AccountId,
		Target: proto.CloneOf(original.GetTarget()), Actor: proto.CloneOf(original.GetActor()), ChatType: original.GetChatType(),
		EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_CommandInvoked{CommandInvoked: &pb.CommandInvokedPayload{
			PluginId: c.pluginID, RouteId: c.route.GetRouteId(), CommandName: name, Result: result, DurationMs: uint64(duration.Milliseconds()),
		}}},
	}
	// Deliver 不等待空闲分发槽，事件订阅者可以继续调用插件而不占住原命令链。
	if err := r.Deliver(r.ctx, ev); err != nil {
		slog.Warn("deliver command invocation event failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "error", err)
	}
}
