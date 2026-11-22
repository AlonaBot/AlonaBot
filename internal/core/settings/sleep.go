package settings

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
)

const defaultIdleTimeout = 15 * time.Minute

// Sleep 保存全局休眠设置与逐插件覆盖；缺失覆盖沿用全局时长和路由准入。
type Sleep struct {
	Mode        string
	IdleTimeout time.Duration
	Plugins     map[string]string // 值只接受 never 或正的 Go duration。
}

// SetSleep 校验并修改全局模式与空闲时长，保留插件覆盖及无关字段原文。
func (d *Document) SetSleep(mode, idleTimeout string) error {
	if err := sleepMode(mode); err != nil {
		return fmt.Errorf("sleep.mode: %w", err)
	}
	if _, err := sleepDuration(idleTimeout); err != nil {
		return fmt.Errorf("sleep.idle_timeout: %w", err)
	}
	if err := d.Set([]string{"sleep", "mode"}, mode); err != nil {
		return err
	}
	return d.Set([]string{"sleep", "idle_timeout"}, idleTimeout)
}

// SetSleepPolicy 写入插件的 never 或正时长覆盖；nil 删除覆盖以恢复自动准入。
// 插件安装身份由调用方核对。
func (d *Document) SetSleepPolicy(pluginID string, policy *string) error {
	path := []string{"sleep", "plugins", pluginID}
	if policy == nil {
		return d.Delete(path)
	}
	if err := sleepPolicy(*policy); err != nil {
		return fmt.Errorf("sleep.plugins[%q]: %w", pluginID, err)
	}
	return d.Set(path, *policy)
}

func sleepMode(mode string) error {
	if mode != "always_on" && mode != "component_idle" {
		return errors.New("expected always_on or component_idle")
	}
	return nil
}

func sleepDuration(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, errors.New("duration must be positive")
	}
	return duration, nil
}

func sleepPolicy(value string) error {
	if value == "never" {
		return nil
	}
	_, err := sleepDuration(value)
	return err
}

func readSleep(n *node, key string, s *Settings) error {
	if key == "sleep.plugins" {
		policies := make(map[string]string)
		if err := n.object(func(id string, value *node) error {
			policy, err := value.text()
			if err == nil {
				err = sleepPolicy(policy)
			}
			if err != nil {
				s.Corrections = append(s.Corrections, configschema.Correction{
					Path: "$.sleep.plugins[" + strconv.Quote(id) + "]", Reason: "reset", Old: value.encode(),
				})
				return nil // 非法覆盖恢复为无覆盖，原文留在 Document 中。
			}
			policies[id] = policy
			return nil
		}); err != nil {
			return err
		}
		s.Sleep.Plugins = policies
		return nil
	}
	value, err := n.text()
	if err != nil {
		return err
	}
	if key == "sleep.mode" {
		if err := sleepMode(value); err != nil {
			return n.fail(err)
		}
		s.Sleep.Mode = value
		return nil
	}
	duration, err := sleepDuration(value)
	if err != nil {
		return n.fail(err)
	}
	s.Sleep.IdleTimeout = duration
	return nil
}
