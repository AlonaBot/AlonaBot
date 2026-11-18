package settings

import "errors"

// Reply 保存 default 配置文件的回复设置；零值关闭全部回复修饰。
type Reply struct {
	MentionSender    bool   `json:"mention_sender"`
	Quote            bool   `json:"quote"`
	Prefix           string `json:"prefix"`
	ForwardThreshold int    `json:"forward_threshold"` // Unicode 码点数；0 关闭合并转发。
}

func readReply(n *node, reply *Reply) error {
	return n.object(func(key string, value *node) error {
		switch key {
		case "mention_sender", "quote":
			enabled, ok := value.value.(bool)
			if !ok {
				return value.fail(errors.New("expected a boolean"))
			}
			if key == "mention_sender" {
				reply.MentionSender = enabled
			} else {
				reply.Quote = enabled
			}
		case "prefix":
			prefix, err := value.text()
			if err != nil {
				return err
			}
			reply.Prefix = prefix
		case "forward_threshold":
			threshold, err := value.integer()
			if err != nil {
				return err
			}
			if threshold < 0 {
				return value.fail(errors.New("must not be negative"))
			}
			reply.ForwardThreshold = threshold
		default:
			return value.fail(errors.New("unknown key"))
		}
		return nil
	})
}
