package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

// InstanceProblem 定位无法启动的实例配置；Reason 保留完整原因，省略配置取值。
type InstanceProblem struct {
	InstanceID string
	Path       string // core_config.json 内的 JSON 路径。
	Reason     string
}

type instancePreparation struct {
	ready          map[string]json.RawMessage
	problems       []InstanceProblem
	corrections    map[string][]configschema.Correction
	unknown        map[string][]string
	configurations map[string]json.RawMessage
}

var errNoInstanceWrite = errors.New("instance configuration does not require a write")

// BindAdapter 为已有的无归属配置记录当前组件；其他组件的配置保持原样并返回 false。
// 调用方先检查返回值，再准备实例，避免把保留的凭证交给替换组件。
func (f *File) BindAdapter(adapterID, packageID string) (bool, error) {
	group, exists := f.Settings().Adapters[adapterID]
	if !exists {
		return true, nil
	}
	if group.Package != "" {
		return group.Package == packageID, nil
	}
	owned := false
	err := f.Update(func(doc *Document) error {
		current, err := doc.settings()
		if err != nil {
			return err
		}
		group, exists := current.Adapters[adapterID]
		owned = !exists || group.Package == "" || group.Package == packageID
		if !exists || group.Package != "" {
			return errNoInstanceWrite
		}
		return doc.Set([]string{"adapters", adapterID, "package"}, packageID)
	})
	if errors.Is(err, errNoInstanceWrite) {
		err = nil
	}
	return owned, err
}

// PrepareInstances 校正一个 Adapter 的运行副本，保存内存记录，不自动写回原文件。
// replacement 是管理入口已经校验的显式编辑，经 Update 提交；失败实例独立列入 problems。
func (f *File) PrepareInstances(adapterID string, schema *configschema.Schema, replacement ...map[string]json.RawMessage) (map[string]json.RawMessage, []InstanceProblem, error) {
	if len(replacement) != 0 {
		var prepared instancePreparation
		err := f.Update(func(doc *Document) error {
			prepared = prepareInstances(doc, adapterID, schema, replacement[0])
			if len(prepared.problems) != 0 {
				return errNoInstanceWrite
			}
			for id, raw := range replacement[0] {
				if err := doc.Set([]string{"adapters", adapterID, "instances", id}, raw); err != nil {
					return err
				}
			}
			return nil
		})
		if errors.Is(err, errNoInstanceWrite) {
			return prepared.ready, prepared.problems, nil
		}
		if err != nil {
			return prepared.ready, prepared.problems, err
		}
	}
	f.mu.Lock() // 记录和准备结果使用同一份原文，和 Update/Reload 串行。
	defer f.mu.Unlock()
	if _, found := f.current.Adapters[adapterID]; !found {
		return map[string]json.RawMessage{}, nil, nil
	}
	prepared := prepareInstances(f.document, adapterID, schema, nil)
	if group, found := f.current.Adapters[adapterID]; found {
		group.Corrections, group.Unknown = prepared.corrections, prepared.unknown
		group.Configurations = prepared.configurations
		f.current.Adapters[adapterID] = group
	}
	for _, records := range prepared.corrections {
		logCorrections(records)
	}
	return prepared.ready, prepared.problems, nil
}

// TrialInstances 只读试算实例配置，供更新预检使用；不写回、备份或发布运行记录。
func (f *File) TrialInstances(adapterID string, schema *configschema.Schema) []InstanceProblem {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return prepareInstances(f.document, adapterID, schema, nil).problems
}

func prepareInstances(document *Document, adapterID string, schema *configschema.Schema, replacements map[string]json.RawMessage) instancePreparation {
	prepared := instancePreparation{ready: make(map[string]json.RawMessage), corrections: make(map[string][]configschema.Correction), unknown: make(map[string][]string), configurations: make(map[string]json.RawMessage)}
	if len(replacements) != 0 {
		var err error
		document, err = parseDocument(document.root.encode())
		if err != nil {
			prepared.problems = append(prepared.problems, InstanceProblem{Reason: err.Error()})
			return prepared
		}
		for id, raw := range replacements {
			if err := document.Set([]string{"adapters", adapterID, "instances", id}, raw); err != nil {
				prepared.problems = append(prepared.problems, InstanceProblem{InstanceID: id, Reason: err.Error()})
				return prepared
			}
		}
	}
	index := document.root.index("adapters")
	if index < 0 {
		return prepared
	}
	adapters := document.root.members[index].value
	index = adapters.index(adapterID)
	if index < 0 {
		return prepared
	}
	group := adapters.members[index].value
	index = group.index("instances")
	if index < 0 {
		return prepared
	}
	for _, instance := range group.members[index].value.members {
		path := fmt.Sprintf("$.adapters[%q].instances[%q]", adapterID, instance.key)
		if !instanceName.MatchString(instance.key) {
			prepared.problems = append(prepared.problems, InstanceProblem{InstanceID: instance.key, Path: path, Reason: "invalid adapter instance name"})
			continue
		}
		raw := instance.value.encode()
		records := []configschema.Correction{}
		if instance.value.delimiter != '{' {
			record := configschema.Correction{Path: "$", Reason: "reset"}
			if len(schema.SecretPaths()) == 0 {
				record.Old = raw
			}
			records = append(records, record)
			raw = []byte("{}")
		}
		complete, corrections, unknown, err := correctConfiguration(raw, schema)
		prepared.configurations[instance.key] = complete
		records = append(records, corrections...)
		for i := range records {
			records[i].Path = path + strings.TrimPrefix(records[i].Path, "$")
		}
		prepared.corrections[instance.key] = records
		for i := range unknown {
			unknown[i] = path + strings.TrimPrefix(unknown[i], "$")
		}
		prepared.unknown[instance.key] = unknown
		if err != nil {
			problemPath := path
			if _, tail, found := strings.Cut(err.Error(), "config validation failed at $"); found {
				suffix, _, _ := strings.Cut(tail, ": ")
				problemPath += suffix
			}
			prepared.problems = append(prepared.problems, InstanceProblem{InstanceID: instance.key, Path: problemPath, Reason: err.Error()})
			continue
		}
		prepared.ready[instance.key] = complete
	}
	return prepared
}

// PruneInstance 删除实例 schema 之外的键；由 File.Update 提交并保留其他值的原文。
func (d *Document) PruneInstance(adapterID, instanceID string, schema *configschema.Schema) error {
	var fields map[string]configField
	if err := json.Unmarshal(schema.Canonical(), &fields); err != nil {
		return err
	}
	parent, err := d.parent([]string{"adapters", adapterID, "instances", instanceID}, false)
	if err != nil || parent == nil {
		return err
	}
	if index := parent.index(instanceID); index >= 0 {
		pruneFields(parent.members[index].value, fields)
	}
	return nil
}
