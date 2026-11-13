package remote

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/google/jsonschema-go/jsonschema"
)

type declaredEvent struct {
	version uint32
	schema  *jsonschema.Resolved
}

type actionKey struct {
	name    string
	version uint32
}

type declaredAction struct {
	scope  pb.ActionScope
	params *jsonschema.Resolved
	result *jsonschema.Resolved
}

// declarations 缓存当前 Session 已校验的声明；由 Process.mu 保护其发布，发布后只读。
type declarations struct {
	permissions map[string][]string // 每个已声明权限的传递闭包，含其自身
	events      map[string]declaredEvent
	extensions  map[extensionKey]*jsonschema.Resolved
	actions     map[actionKey]declaredAction
}

func compileDeclarations(descriptor *pb.AdapterDescriptor) (declarations, error) {
	result := declarations{permissions: make(map[string][]string), events: make(map[string]declaredEvent), extensions: make(map[extensionKey]*jsonschema.Resolved), actions: make(map[actionKey]declaredAction)}
	graph := make(map[string][]string)
	for _, permission := range descriptor.GetPermissions() {
		graph[permission.GetName()] = permission.GetImplies()
	}
	// 注册边界已经保证无环且引用有效；在绑定时展开，交付只做集合合并。
	for name := range graph {
		seen := make(map[string]bool)
		var visit func(string)
		visit = func(name string) {
			if seen[name] {
				return
			}
			seen[name] = true
			for _, implied := range graph[name] {
				visit(implied)
			}
		}
		visit(name)
		for implied := range seen {
			result.permissions[name] = append(result.permissions[name], implied)
		}
		slices.Sort(result.permissions[name])
	}
	for _, event := range descriptor.GetEvents() {
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(event.GetSchemaJson()), &schema); err != nil {
			return declarations{}, fmt.Errorf("parse adapter event %q schema: %w", event.GetName(), err)
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			return declarations{}, fmt.Errorf("resolve adapter event %q schema: %w", event.GetName(), err)
		}
		result.events[event.GetName()] = declaredEvent{version: event.GetVersion(), schema: resolved}
	}
	for _, extension := range descriptor.GetExtensions() {
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(extension.GetSchemaJson()), &schema); err != nil {
			return declarations{}, fmt.Errorf("parse adapter extension %q schema: %w", extension.GetName(), err)
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			return declarations{}, fmt.Errorf("resolve adapter extension %q schema: %w", extension.GetName(), err)
		}
		result.extensions[extensionKey{extension.GetName(), extension.GetVersion()}] = resolved
	}
	for _, action := range descriptor.GetActions() {
		resolved := make([]*jsonschema.Resolved, 0, 2)
		for _, raw := range []string{action.GetParamsSchemaJson(), action.GetResultSchemaJson()} {
			var schema jsonschema.Schema
			if err := json.Unmarshal([]byte(raw), &schema); err != nil {
				return declarations{}, fmt.Errorf("parse adapter action %q schema: %w", action.GetName(), err)
			}
			compiled, err := schema.Resolve(nil)
			if err != nil {
				return declarations{}, fmt.Errorf("resolve adapter action %q schema: %w", action.GetName(), err)
			}
			resolved = append(resolved, compiled)
		}
		result.actions[actionKey{action.GetName(), action.GetVersion()}] = declaredAction{scope: action.GetScope(), params: resolved[0], result: resolved[1]}
	}
	return result, nil
}

func (p *Process) platformPermissions(instanceID string, permissions []string) []string {
	var expanded []string
	for _, name := range permissions {
		implied, declared := p.declarations.permissions[name]
		if !declared {
			slog.Warn("discard undeclared adapter permission", "plugin_id", p.opt.PluginID, "instance_id", instanceID, "name", name)
			continue
		}
		expanded = append(expanded, implied...)
	}
	slices.Sort(expanded)
	return slices.Compact(expanded)
}

func (d declarations) validateEvent(ev *pb.Event) error {
	declared, ok := d.events[ev.GetEventType()]
	payload := ev.GetEventPayload().GetAdapterEvent()
	if !ok {
		return fmt.Errorf("undeclared adapter event %q", ev.GetEventType())
	}
	if payload == nil {
		return fmt.Errorf("adapter event %q requires adapter_event payload", ev.GetEventType())
	}
	if payload.GetVersion() != declared.version {
		return fmt.Errorf("adapter event %q version %d differs from declared version %d", ev.GetEventType(), payload.GetVersion(), declared.version)
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(payload.GetPayloadJson()), &value); err != nil {
		return fmt.Errorf("decode adapter event %q payload: %w", ev.GetEventType(), err)
	}
	if value == nil {
		return fmt.Errorf("adapter event %q payload_json is not a JSON object", ev.GetEventType())
	}
	if err := declared.schema.Validate(value); err != nil {
		return fmt.Errorf("validate adapter event %q payload: %w", ev.GetEventType(), err)
	}
	return nil
}

// CheckEvent 按事件所属实例当前 Session 的声明校验扩展载荷，供 Core 改写后复检。
func (i *instance) CheckEvent(ev *pb.Event) error {
	p := i.process
	p.mu.Lock()
	current := i.id == ev.GetAdapterInstanceId() && i.running && p.session != nil && i.session == p.session.Session
	declared := p.declarations
	p.mu.Unlock()
	if !current {
		return fmt.Errorf("adapter event instance %q has no current session", ev.GetAdapterInstanceId())
	}
	return declared.validateEvent(ev)
}
