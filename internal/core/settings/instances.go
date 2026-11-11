package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

// InstanceProblem 定位无法启动的实例配置；Reason 保留原因和版本，不含配置取值。
type InstanceProblem struct {
	InstanceID string
	Path       string // core_config.json 内的 JSON 路径。
	Reason     string
}

type instancePreparation struct {
	ready    map[string]json.RawMessage
	migrated map[string]json.RawMessage // 只保存迁移后的原文，环境变量保持未展开。
	problems []InstanceProblem
	version  int
}

var errNoInstanceMigration = errors.New("instance configuration does not require a write")

// PrepareInstances 迁移、补齐并校验一个 Adapter 的全部实例，返回可启动实例的完整配置。
// 低版本的全部实例成功后才备份并经 Update 写回；实例失败只列入 problems。
// 返回的 error 表示备份或写回失败，原文中的环境变量引用始终保留。
func (f *File) PrepareInstances(adapterID string, schema *configschema.Schema) (map[string]json.RawMessage, []InstanceProblem, error) {
	// 1. 当前文档只读试算，无需迁移或有失败实例时不触碰磁盘。
	f.mu.RLock()
	prepared := prepareInstances(f.document, adapterID, schema)
	f.mu.RUnlock()
	if len(prepared.problems) != 0 || prepared.version >= schema.Version() {
		return prepared.ready, prepared.problems, nil
	}
	// 2. Update 持锁重读磁盘，在同一份候选上试算和编辑，保留期间的手动修改。
	err := f.Update(func(document *Document) error {
		prepared = prepareInstances(document, adapterID, schema)
		if len(prepared.problems) != 0 || prepared.version >= schema.Version() {
			return errNoInstanceMigration
		}
		for id, raw := range prepared.migrated {
			if err := document.Set([]string{"adapters", adapterID, "instances", id}, raw); err != nil {
				return err
			}
		}
		if err := document.Set([]string{"adapters", adapterID, "$version"}, schema.Version()); err != nil {
			return err
		}
		return backupSettings(f.name, currentVersion)
	})
	if errors.Is(err, errNoInstanceMigration) {
		err = nil
	}
	return prepared.ready, prepared.problems, err
}

// TrialInstances 只读试算实例配置，供更新预检在停止旧 Adapter 前检查；不写回或备份。
func (f *File) TrialInstances(adapterID string, schema *configschema.Schema) []InstanceProblem {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return prepareInstances(f.document, adapterID, schema).problems
}

func prepareInstances(document *Document, adapterID string, schema *configschema.Schema) instancePreparation {
	prepared := instancePreparation{ready: make(map[string]json.RawMessage), migrated: make(map[string]json.RawMessage), version: schema.Version()}
	adaptersIndex := document.root.index("adapters")
	if adaptersIndex < 0 {
		return prepared
	}
	adapters := document.root.members[adaptersIndex].value
	index := adapters.index(adapterID)
	if index < 0 {
		return prepared
	}
	group := adapters.members[index].value
	prepared.version = defaultInstanceVersion
	path := "$.adapters[" + strconv.Quote(adapterID) + "]"
	if index := group.index("$version"); index >= 0 {
		version, err := positiveVersion(group.members[index].value)
		if err != nil {
			prepared.problems = append(prepared.problems, InstanceProblem{Path: path + `["$version"]`, Reason: err.Error()})
			return prepared
		}
		prepared.version = version
	}
	if prepared.version > schema.Version() {
		prepared.problems = append(prepared.problems, InstanceProblem{Path: path + `["$version"]`,
			Reason: fmt.Sprintf("%s: instance config version %d is newer than schema version %d", FileName, prepared.version, schema.Version())})
		return prepared
	}
	index = group.index("instances")
	if index < 0 {
		return prepared
	}
	for _, instance := range group.members[index].value.members {
		basePath := path + ".instances[" + strconv.Quote(instance.key) + "]"
		raw, complete, err := prepareInstance(instance.value, prepared.version, schema)
		if err != nil {
			// configschema 的错误只包含声明路径和原因，保留完整错误并补文件与版本。
			problemPath := basePath
			if _, tail, ok := strings.Cut(err.Error(), "config validation failed at $"); ok {
				suffix, _, _ := strings.Cut(tail, ": ")
				problemPath += suffix
			}
			prepared.problems = append(prepared.problems, InstanceProblem{InstanceID: instance.key, Path: problemPath,
				Reason: fmt.Sprintf("%s: instance config version %d to %d: %s", FileName, prepared.version, schema.Version(), err)})
			continue
		}
		prepared.migrated[instance.key], prepared.ready[instance.key] = raw, complete
	}
	return prepared
}

func prepareInstance(original *node, version int, schema *configschema.Schema) (json.RawMessage, json.RawMessage, error) {
	// 1. 在独立对象上迁移原文，保留数字精度、未知键和环境变量引用。
	var configuration map[string]any
	decoder := json.NewDecoder(bytes.NewReader(original.encode()))
	decoder.UseNumber()
	if err := decoder.Decode(&configuration); err != nil {
		return nil, nil, err
	}
	if err := schema.Migrate(configuration, version); err != nil {
		return nil, nil, err
	}
	raw, err := marshal(configuration)
	if err != nil {
		return nil, nil, err
	}
	// 2. 展开运行副本后补缺失值并校验，写回仍使用迁移后的原文。
	candidate, err := parseNode(raw, original.path)
	if err != nil {
		return nil, nil, err
	}
	expanded, err := candidate.expanded()
	if err != nil {
		return nil, nil, err
	}
	decoder = json.NewDecoder(bytes.NewReader(expanded.encode()))
	decoder.UseNumber()
	if err := decoder.Decode(&configuration); err != nil {
		return nil, nil, err
	}
	schema.Complete(configuration)
	if err := schema.Validate(configuration); err != nil {
		return nil, nil, err
	}
	complete, err := marshal(configuration)
	return raw, complete, err
}
