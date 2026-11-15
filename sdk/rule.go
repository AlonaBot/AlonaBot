package sdk

import (
	"fmt"
	"slices"

	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Rule 在本次调用的 Context 中判断是否执行 Handler，错误交由调用边界处理。
type Rule func(*Context) (bool, error)

// Any 按顺序返回首个匹配或错误；空列表不匹配，含 nil 的组合在 collect 报错。
func Any(rules ...Rule) Rule {
	if checkRules(rules) != nil {
		return nil
	}
	rules = slices.Clone(rules)
	return func(c *Context) (bool, error) {
		for _, rule := range rules {
			matched, err := rule(c)
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
	}
}

// ToMe 匹配标准 message.created 中指向本机器人的消息。
func ToMe() Rule {
	return func(c *Context) (bool, error) {
		event := c.Event()
		if event == nil || event.Type != MessageCreated.name {
			return false, nil
		}
		payload, ok := event.Payload.(*message.CreatedPayload)
		return ok && payload != nil && payload.ToMe, nil
	}
}

// NotCommand 匹配尚未被 Core 命令层消费的消息。
func NotCommand() Rule {
	return func(c *Context) (bool, error) { return !c.CommandConsumed(), nil }
}

func checkRules(rules []Rule) error {
	for i, rule := range rules {
		if rule == nil {
			return fmt.Errorf("rule %d is nil", i+1)
		}
	}
	return nil
}
