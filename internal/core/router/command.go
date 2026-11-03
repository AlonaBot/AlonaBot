package router

import (
	"cmp"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"slices"
	"strings"
	"unicode"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

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

func (r *Router) dispatchCommands(ev *pb.Event) bool {
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
	if tokens := r.commandTokens(text, ev); len(tokens) != 0 {
		if entry, consumed, ok := index.Lookup(tokens); ok {
			for _, route := range routes {
				if route.pluginID != entry.PluginID || route.route.GetRouteId() != entry.RouteID {
					continue
				}
				if !route.plugin.Activated() || !r.permitted(route.candidate, ev) {
					return false
				}
				tail := text
				for range consumed {
					tail = strings.TrimLeftFunc(tail, unicode.IsSpace)
					end := strings.IndexFunc(tail, unicode.IsSpace)
					if end < 0 {
						tail = ""
					} else {
						tail = tail[end:]
					}
				}
				invocation, err := commandArgs(route.route.GetCommand().GetParams(), tail)
				if err != nil {
					r.log.Debug("command arguments rejected", "plugin_id", entry.PluginID, "route_id", entry.RouteID, "error", err)
					if route.route.GetCommand().GetUsageMode() != pb.UsageMode_USAGE_MODE_SILENT {
						r.sendReply(route.candidate, ev, commandUsage(route.route.GetCommand()), nil)
					}
					return false
				}
				invocation.Path = entry.Path
				return r.invokeCommand(route.candidate, ev, invocation)
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
		if matched && r.permitted(route.candidate, ev) && r.invokeCommand(route.candidate, ev, invocation) {
			return true
		}
	}
	return false
}

func commandPhase(route commandRoute) int {
	if route.route.GetCommand().GetMatch() == pb.CommandMatch_COMMAND_MATCH_RAW {
		return 1
	}
	return 0
}

func (r *Router) commandTokens(text string, ev *pb.Event) []string {
	if !strings.HasPrefix(text, "/") {
		return nil
	}
	tokens := strings.Fields(text)
	tokens[0] = strings.TrimPrefix(tokens[0], "/")
	if name, user, hasName := strings.Cut(tokens[0], "@"); hasName {
		instance, ok := r.adapters.Get(ev.GetTarget().GetAdapterInstanceId())
		if !ok {
			return nil
		}
		named, ok := instance.(adapter.Named)
		if !ok || user == "" || !strings.EqualFold(user, named.AccountName()) {
			return nil
		}
		tokens[0] = name
	}
	return tokens
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

func (r *Router) invokeCommand(c candidate, ev *pb.Event, invocation *pb.CommandInvocation) bool {
	if r.ctx.Err() != nil || !c.plugin.Activated() {
		return false
	}
	trace := make([]byte, traceIDBytes)
	rand.Read(trace)
	log := r.log.With("plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "trace_id", hex.EncodeToString(trace))
	response, callError := r.call(c, ev, invocation, trace)
	if callError != nil {
		log.Warn("command handler call failed", "error", callError)
	}
	if response.Err != nil {
		log.Warn("command handler returned an error", "code", response.Err.GetCode(), "error", response.Err.GetSafeMessage())
	}
	outcome, err := decodeOutcome(response.Payload)
	if err != nil {
		log.Warn("decode command handler outcome failed", "error", err)
		return false
	}
	if callError == nil && response.Err == nil && len(outcome.GetChain().GetSegments()) != 0 {
		r.sendReply(c, ev, outcome.Chain, trace)
	}
	return outcome.GetPropagation() == pb.Propagation_PROPAGATION_STOP
}

func (r *Router) sendReply(c candidate, ev *pb.Event, chain *pb.MessageChain, trace []byte) {
	request := &pb.SendRequest{Chain: chain}
	encoded, err := proto.Marshal(request)
	if err != nil {
		r.log.Warn("encode reply failed", "error", err)
		return
	}
	response := r.send(r.ctx, c.caller, &wire.Inbound{TraceID: trace, ParentAttach: &Invocation{PluginID: c.pluginID, Event: ev, DefaultTarget: ev.GetTarget()}, Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(request), SchemaVersion: handlerSchemaVersion, Bytes: encoded}})
	result := &pb.SendResult{}
	if err := proto.Unmarshal(response.Payload.Bytes, result); err != nil {
		r.log.Warn("decode reply result failed", "error", err)
	}
	if response.Err != nil {
		r.log.Warn("reply failed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "error", response.Err, "result", result)
	} else {
		r.log.Debug("reply completed", "plugin_id", c.pluginID, "route_id", c.route.GetRouteId(), "result", result)
	}
}

func commandUsage(spec *pb.CommandSpec) *pb.MessageChain {
	text := spec.GetCustomUsage()
	if text == "" {
		// shortcut: usage 暂用英文，Core 文案目录接入后替换。
		text = "Usage: /" + strings.Join(spec.GetPath(), " ")
		for _, alias := range spec.GetAliases() {
			text += " | /" + strings.Join(alias.GetPath(), " ")
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
			}
			if param.GetRaw() {
				label += "..."
			}
			if param.GetRequired() {
				text += " <" + label + ">"
			} else {
				text += " [" + label + "]"
			}
		}
	}
	return &pb.MessageChain{Segments: []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: text}}}}}
}
