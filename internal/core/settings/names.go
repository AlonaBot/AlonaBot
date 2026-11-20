package settings

import (
	"errors"
	"regexp"
)

// CONTRACTS §15.1 规定实例名的字符集和长度，且跨 Adapter 唯一。
const maxInstanceNameBytes = 64

// 各名称入口先检查长度，避免编译展开有界重复。
var instanceName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func validateNames(s Settings, locations map[string]*node) error {
	seen := make(map[string]bool)
	for id, adapter := range s.Adapters {
		path := "adapters." + id
		for name := range adapter.Instances {
			key := path + ".instances." + name
			if len(name) > maxInstanceNameBytes || !instanceName.MatchString(name) {
				return at(locations, key, errors.New("invalid adapter instance name"))
			}
			if seen[name] {
				return at(locations, key, errors.New("duplicate adapter instance name"))
			}
			seen[name] = true
		}
		if adapter.DefaultInstance != "" {
			if _, ok := adapter.Instances[adapter.DefaultInstance]; !ok {
				return at(locations, path+".default_instance", errors.New("default_instance must name an instance of this adapter"))
			}
		}
	}
	return nil
}

// DefaultInstances 返回各 Adapter 的缺省实例新表，显式设置优先于唯一实例。
func (s Settings) DefaultInstances() map[string]string {
	defaults := make(map[string]string)
	for id, adapter := range s.Adapters {
		if adapter.DefaultInstance != "" {
			defaults[id] = adapter.DefaultInstance
		} else if len(adapter.Instances) == 1 {
			for name := range adapter.Instances {
				defaults[id] = name
			}
		}
	}
	return defaults
}
