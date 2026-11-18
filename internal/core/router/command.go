package router

import (
	"cmp"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	corei18n "github.com/AlonaBot/AlonaBot/internal/core/i18n"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

const commandUsageKey = "core.usage.command"

type commandRoute struct {
	candidate
	regex *regexp.Regexp
}

// cacheCommands 在持有 mu 时构建插件路由快照；描述已经过 REGISTER 校验。
func (r *Router) cacheCommands(session *pluginSession) {
	p := session.plugin
	var routes []commandRoute
	for _, route := range p.Routes() {
		if route.GetKind() != pb.RouteKind_ROUTE_KIND_COMMAND {
			continue
		}
		snapshot := proto.CloneOf(route)
		next := commandRoute{candidate: candidate{pluginSession: session, pluginID: p.PluginID(), route: snapshot}}
		if snapshot.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_REGEX {
			next.regex = regexp.MustCompile(snapshot.GetCommand().GetRegex())
		}
		routes = append(routes, next)
	}
	r.matches[p.PluginID()] = routes
}

func (r *Router) dispatchCommands(ev *pb.Event, event *trustedEvent) bool {
	// 1. 同时读取路径索引和插件路由快照，插件调用期间不持注册锁。
	r.mu.Lock()
	index := r.commands
	var routes []commandRoute
	for _, cached := range r.matches {
		routes = append(routes, cached...)
	}
	r.mu.Unlock()
	var content strings.Builder
	for _, segment := range ev.GetEventPayload().GetMessageCreated().GetChain().GetSegments() {
		content.WriteString(segment.GetText().GetText())
	}
	text := content.String()
	head := commandHead(ev.GetEventPayload().GetMessageCreated().GetChain(), ev.GetAccountId())
	if tokens, prefix := r.commandTokens(head, ev); len(tokens) != 0 {
		if entry, consumed, ok := index.Lookup(tokens); ok {
			event.commandConsumed = true
			for _, route := range routes {
				if route.pluginID != entry.PluginID || route.route.GetRouteId() != entry.RouteID {
					continue
				}
				route.names = entry.Names
				if !r.commandAllowed(route.candidate, ev) {
					return false
				}
				tail := strings.TrimPrefix(head, prefix)
				for range consumed {
					tail = strings.TrimLeftFunc(tail, unicode.IsSpace)
					end := strings.IndexFunc(tail, unicode.IsSpace)
					if end < 0 {
						tail = ""
					} else {
						tail = tail[end:]
					}
				}
				skipped := len(head) - len(tail)
				bind := func(prepared *pb.Event) (*pb.CommandInvocation, error) {
					cursor := commandStart(prepared.GetEventPayload().GetMessageCreated().GetChain(), prepared.GetAccountId())
					cursor.skipText(skipped)
					invocation, err := commandArgs(route.route.GetCommand().GetParams(), cursor)
					if err != nil {
						return nil, err
					}
					invocation.Path = entry.Path
					return invocation, nil
				}
				return r.invokeCommand(route.candidate, ev, event, bind, prefix)
			}
			return false
		}
	}
	// 2. 文本匹配统一排序，Raw 排在全部文本匹配之后；已识别命令在上面直接结束本层。
	slices.SortFunc(routes, func(a, b commandRoute) int {
		if order := cmp.Compare(commandPhase(a), commandPhase(b)); order != 0 {
			return order
		}
		if order := cmp.Compare(a.route.GetPriority(), b.route.GetPriority()); order != 0 {
			return order
		}
		if order := cmp.Compare(a.pluginID, b.pluginID); order != 0 {
			return order
		}
		return cmp.Compare(a.route.GetRouteId(), b.route.GetRouteId())
	})
	for _, route := range routes {
		spec := route.route.GetCommand()
		invocation := &pb.CommandInvocation{}
		matched := false
		switch spec.GetMatch() {
		case pb.CommandMatch_COMMAND_MATCH_PATH:
			continue
		case pb.CommandMatch_COMMAND_MATCH_REGEX:
			invocation.RegexGroups = route.regex.FindStringSubmatch(text)
			matched = invocation.RegexGroups != nil
		case pb.CommandMatch_COMMAND_MATCH_RAW:
			matched = true
		default:
			for _, pattern := range spec.GetPatterns() {
				if matchPattern(text, pattern, spec.GetMatch(), spec.GetIgnoreCase()) {
					matched = true
					break
				}
			}
		}
		if matched && r.commandAllowed(route.candidate, ev) && r.invokeCommand(route.candidate, ev, event, func(*pb.Event) (*pb.CommandInvocation, error) { return invocation, nil }, "") {
			return true
		}
	}
	return false
}

func (r *Router) commandPolicyFailure(c candidate, ev *pb.Event) *pb.ErrorInfo {
	if r.policy == nil {
		if r.permitted(c, ev) {
			return nil
		}
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED}
	}
	failure := r.policy.CheckCommand(r.ctx, c.pluginID, c.route, ev.GetActor(), ev.GetTarget())
	if failure == nil {
		return nil
	}
	if commandPolicyResult(failure.GetCode()) == commandDenied {
		r.metrics.Add(c.pluginID, "event", "denied")
		slog.Debug("command policy denied", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "code", failure.GetCode())
	} else {
		slog.Warn("command policy check failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "code", failure.GetCode(), "error", failure.GetSafeMessage())
	}
	return failure
}

func commandPhase(route commandRoute) int {
	if route.route.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_RAW {
		return 1
	}
	return 0
}

func commandStart(chain *pb.MessageChain, accountID string) argumentCursor {
	cursor := argumentCursor{segments: chain.GetSegments()}
	for cursor.segment < len(cursor.segments) {
		segment := cursor.segments[cursor.segment]
		if segment.GetReference() != nil || (segment.GetMention() != nil && segment.GetMention().GetExternalUserId() == accountID) || (segment.GetText() != nil && strings.TrimSpace(segment.GetText().GetText()) == "") {
			cursor.segment++
			continue
		}
		break
	}
	cursor.trimSpace()
	return cursor
}

// commandHead 只读取连续的开头文本；非文本参数后面的文字不参与路径选择。
func commandHead(chain *pb.MessageChain, accountID string) string {
	cursor := commandStart(chain, accountID)
	var head strings.Builder
	for _, segment := range cursor.segments[cursor.segment:] {
		if segment.GetText() == nil {
			break
		}
		head.WriteString(segment.GetText().GetText()[cursor.offset:])
		cursor.offset = 0
	}
	return head.String()
}

func (r *Router) commandTokens(text string, ev *pb.Event) ([]string, string) {
	prefix := ""
	matched := false
	for _, candidate := range r.prefixes(ev.GetAdapterInstanceId()) {
		if strings.HasPrefix(text, candidate) && (!matched || len(candidate) > len(prefix)) {
			prefix, matched = candidate, true
		}
	}
	if !matched {
		return nil, ""
	}
	remainder := strings.TrimPrefix(text, prefix)
	// 路径紧接前缀，保留 / cmd 不被识别为 /cmd 的既有行为。
	if remainder != strings.TrimLeftFunc(remainder, unicode.IsSpace) {
		return nil, ""
	}
	tokens := strings.Fields(remainder)
	if len(tokens) == 0 {
		return nil, ""
	}
	if name, user, hasName := strings.Cut(tokens[0], "@"); hasName {
		instance, ok := r.adapters.Get(ev.GetTarget().GetAdapterInstanceId())
		if !ok {
			return nil, ""
		}
		named, ok := instance.(adapter.Named)
		if !ok || user == "" || !strings.EqualFold(user, named.AccountName()) {
			return nil, ""
		}
		tokens[0] = name
	}
	return tokens, prefix
}

func matchPattern(text, pattern string, match pb.CommandMatch, ignoreCase bool) bool {
	if ignoreCase {
		text = foldText(text)
		pattern = foldText(pattern)
	}
	switch match {
	case pb.CommandMatch_COMMAND_MATCH_KEYWORD:
		return strings.Contains(text, pattern)
	case pb.CommandMatch_COMMAND_MATCH_STARTSWITH:
		return strings.HasPrefix(text, pattern)
	case pb.CommandMatch_COMMAND_MATCH_ENDSWITH:
		return strings.HasSuffix(text, pattern)
	case pb.CommandMatch_COMMAND_MATCH_FULLMATCH:
		return text == pattern
	}
	return false
}

// foldText 取每个 SimpleFold 环的最小 rune，覆盖 Kelvin sign 与长 s 等非小写等价类。
func foldText(text string) string {
	return strings.Map(func(r rune) rune {
		canonical := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < canonical {
				canonical = next
			}
		}
		return canonical
	}, text)
}

func (r *Router) invokeCommand(c candidate, ev *pb.Event, event *trustedEvent, bind func(*pb.Event) (*pb.CommandInvocation, error), prefix string) bool {
	if r.ctx.Err() != nil || !c.plugin.Activated() {
		return false
	}
	trace := make([]byte, traceIDBytes)
	rand.Read(trace)
	started := time.Now()
	response, callError := r.call(c, event, bind, trace)
	duration := time.Since(started)
	if errors.Is(callError, errCommandArguments) {
		slog.With("plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "trace_id", hex.EncodeToString(trace)).Debug("command arguments rejected", "error", callError)
		if c.route.GetCommand().GetUsageMode() != pb.UsageMode_USAGE_MODE_SILENT {
			usage, err := commandUsage(c, ev, prefix)
			if err != nil {
				slog.Warn("format command usage failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "error", err)
			} else {
				r.sendReply(c, ev, &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: usage}}}}}, trace)
			}
		}
		return false
	}
	result := commandOK
	if callError != nil || response.Err != nil {
		result = commandError
		if wire.CodeOf(callError) == pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED || response.Err.GetCode() == pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
			result = commandDenied
		}
	}
	if callError != nil {
		slog.With("plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "trace_id", hex.EncodeToString(trace)).Warn("command handler call failed", "error", callError)
	}
	if response.Err != nil {
		slog.With("plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "trace_id", hex.EncodeToString(trace)).Warn("command handler returned an error", "code", response.Err.GetCode(), "error", response.Err.GetSafeMessage())
	}
	outcome, err := decodeOutcome(response.Payload)
	if err != nil {
		if result == commandOK {
			result = commandError
		}
		r.commandInvoked(c, ev, result, duration)
		slog.With("plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "trace_id", hex.EncodeToString(trace)).Warn("decode command handler outcome failed", "error", err)
		return false
	}
	// SDK Rule 返回 false 时没有调用 Handler；保留实际调用错误与 Core 拒绝通知。
	if outcome.GetMatched() || result != commandOK {
		r.commandInvoked(c, ev, result, duration)
	}
	event.commandConsumed = event.commandConsumed || outcome.GetMatched()
	if callError == nil && response.Err == nil && outcome.GetMatched() && c.route.GetCommand().GetHandlerMode() == pb.HandlerMode_HANDLER_MODE_AUTO_REPLY && len(outcome.GetChain().GetSegments()) != 0 {
		r.sendReply(c, ev, outcome.Chain, trace)
	}
	return outcome.GetPropagation() == pb.Propagation_PROPAGATION_STOP
}

func (r *Router) sendReply(c candidate, ev *pb.Event, chain *pb.MessageChain, trace []byte) {
	request := &pb.SendRequest{Chain: chain}
	encoded, err := proto.Marshal(request)
	if err != nil {
		slog.Warn("encode reply failed", "error", err)
		return
	}
	response := r.send(r.ctx, c.caller, &wire.Inbound{TraceID: trace, ParentAttach: &Invocation{Depth: 1, PluginID: c.pluginID, Event: ev, DefaultTarget: ev.GetTarget(), TraceID: trace}, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded}})
	result := &pb.SendResult{}
	if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil {
		slog.Warn("decode reply result failed", "error", err)
	}
	if response.Err != nil {
		slog.Warn("reply failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "error", response.Err, "result", result)
	} else {
		slog.Debug("reply completed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "result", result)
	}
}

func commandUsage(c candidate, ev *pb.Event, prefix string) (string, error) {
	spec := c.route.GetCommand()
	locales := []string{ev.GetActor().GetLocale(), c.plugin.DefaultLocale()}
	names := c.names
	if names == nil {
		names = declaredCommandNames(spec)
	}
	command := strings.Join(names[0], " ")
	if spec.GetUsageMode() == pb.UsageMode_USAGE_MODE_CUSTOM {
		var text string
		var err error
		if catalog := c.plugin.Messages(); catalog == nil {
			err = errors.New("plugin has no message catalog")
		} else {
			text, err = catalog.Format(locales, spec.GetUsageKey(), map[string]string{"prefix": prefix, "command": prefix + command})
		}
		if err == nil {
			return text, nil
		}
		slog.Warn("format custom command usage failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "error", err)
	}
	var aliases, arguments string
	for _, alias := range names[1:] {
		aliases += " | " + prefix + strings.Join(alias, " ")
	}
	for _, param := range spec.GetParams() {
		label := param.GetName()
		switch param.GetType() {
		case pb.ParamType_PARAM_TYPE_INT:
			label += ":int"
		case pb.ParamType_PARAM_TYPE_FLOAT:
			label += ":float"
		case pb.ParamType_PARAM_TYPE_BOOL:
			label += ":bool"
		case pb.ParamType_PARAM_TYPE_USER:
			label += ":user"
		case pb.ParamType_PARAM_TYPE_IMAGE:
			label += ":image"
		case pb.ParamType_PARAM_TYPE_MESSAGE:
			label += ":message"
		}
		if param.GetRaw() {
			label += "..."
		}
		if param.GetRequired() {
			arguments += " <" + label + ">"
		} else {
			arguments += " [" + label + "]"
		}
	}
	return corei18n.Messages().Format(locales, commandUsageKey, map[string]string{"prefix": prefix, "command": command, "aliases": aliases, "arguments": arguments})
}
