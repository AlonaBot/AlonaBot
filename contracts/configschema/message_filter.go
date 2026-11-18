package configschema

import (
	"fmt"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// CheckMessageFilter 核对消息过滤引用的可信字段；无 schema 时只接受无配置引用的声明。
func (s *Schema) CheckMessageFilter(filter *pb.MessageFilter) error {
	if path := filter.GetWakeWordsConfig(); path != "" {
		if s.messageFilterList(path) == nil {
			return fmt.Errorf("message filter field %q must be a non-secret string list", path)
		}
	}
	if path := filter.GetTriggersConfig(); path != "" {
		f := s.messageFilterList(path)
		if f == nil || !f.attributes.has("options") {
			return fmt.Errorf("message filter field %q must be a non-secret string list with trigger options", path)
		}
		for _, option := range f.attributes.get("options").([]any) {
			switch option {
			case "mention", "private", "wake_word":
			default:
				return fmt.Errorf("message filter field %q has an unsupported trigger option", path)
			}
		}
	}
	return nil
}

func (s *Schema) messageFilterList(path string) *field {
	if s == nil {
		return nil
	}
	fields := s.fields
	for head, tail, more := strings.Cut(path, "."); ; head, tail, more = strings.Cut(tail, ".") {
		f := findField(fields, head)
		if f == nil {
			return nil
		}
		if !more {
			// schema 只允许 string/text 标为 secret；有 items 的 list 保存对象。
			if f.typ == typeList && !f.attributes.has("items") {
				return f
			}
			return nil
		}
		if f.typ != typeObject {
			return nil
		}
		fields = f.items
	}
}
