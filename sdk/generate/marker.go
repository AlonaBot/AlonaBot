package generate

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
	"unicode"
)

func (s *scan) marker(comment *ast.Comment, function, pkg string) {
	fields := strings.Fields(strings.TrimPrefix(comment.Text, "//alonabot:"))
	if len(fields) == 0 {
		s.diagnostic(comment.Pos(), "missing marker kind")
		return
	}
	kind := fields[0]
	if kind != "command" && kind != "event" {
		s.diagnostic(comment.Pos(), "unknown marker "+strconv.Quote(kind))
		return
	}
	var values []string
	var aliases [][]string
	var priority int32
	optionsStarted := false
	for _, field := range fields[1:] {
		key, value, option := strings.Cut(field, "=")
		if !option {
			if optionsStarted {
				s.diagnostic(comment.Pos(), "positional token after options: "+strconv.Quote(field))
			}
			values = append(values, field)
			continue
		}
		optionsStarted = true
		switch key {
		case "priority":
			number, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				s.diagnostic(comment.Pos(), "invalid int32 priority "+strconv.Quote(value))
			} else {
				priority = int32(number)
			}
		case "alias":
			if kind != "command" {
				s.diagnostic(comment.Pos(), "unknown event option \"alias\"")
				continue
			}
			alias := strings.Split(value, "/")
			s.checkTokens(comment, alias)
			aliases = append(aliases, alias)
		default:
			s.diagnostic(comment.Pos(), "unknown option "+strconv.Quote(key))
		}
	}
	if kind == "command" {
		if len(values) == 0 {
			s.diagnostic(comment.Pos(), "command requires a path")
			return
		}
		s.checkTokens(comment, values)
		duplicatePath := false
		for _, path := range append([][]string{values}, aliases...) {
			key := strings.Join(path, " ")
			if previous, ok := s.paths[key]; ok {
				duplicatePath = true
				s.diagnostic(comment.Pos(), fmt.Sprintf("duplicate command path %q; first declared at %s", key, previous))
			} else {
				s.paths[key] = s.position(comment.Pos())
			}
		}
		if !duplicatePath {
			s.checkRoute(comment, "cmd."+strings.Join(values, "."))
		}
		s.result.Commands = append(s.result.Commands, Command{Path: values, Aliases: aliases, Priority: priority, Package: pkg, Func: function})
		return
	}
	if len(values) != 2 {
		s.diagnostic(comment.Pos(), "event requires route_id and event_type")
		return
	}
	route, event := values[0], values[1]
	if route == "" || route == `""` || strings.HasPrefix(route, "cmd.") {
		s.diagnostic(comment.Pos(), "event route_id must be non-empty and must not start with cmd.")
	}
	// 与 contracts/validate.IsEventType 保持一致；生成器的依赖范围限于标准库。
	switch event {
	case "message.created", "message.edited", "message.deleted", "button.clicked", "member.joined", "member.left", "reaction.added", "scheduled", "bot.online", "bot.offline":
	default:
		s.diagnostic(comment.Pos(), "unknown event_type "+strconv.Quote(event))
	}
	s.checkRoute(comment, route)
	s.result.Events = append(s.result.Events, Event{RouteID: route, EventType: event, Priority: priority, Package: pkg, Func: function})
}

// checkTokens 复用 contracts/validate.IsCommandToken 的规则；标准库依赖约束要求在此实现。
func (s *scan) checkTokens(comment *ast.Comment, path []string) {
	for _, value := range path {
		if value == "" || strings.ContainsFunc(value, unicode.IsSpace) || strings.HasPrefix(value, "/") {
			s.diagnostic(comment.Pos(), "invalid command token "+strconv.Quote(value))
		}
	}
}

func (s *scan) checkRoute(comment *ast.Comment, route string) {
	if previous, ok := s.routes[route]; ok {
		s.diagnostic(comment.Pos(), fmt.Sprintf("duplicate route_id %q; first declared at %s", route, previous))
	} else {
		s.routes[route] = s.position(comment.Pos())
	}
}
