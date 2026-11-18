package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// matchMessageFilter 在派发锁外取值。文本来自 PreEvent 后的载荷，toMe 来自原事件。
func (r *Router) matchMessageFilter(pluginID string, filter *pb.MessageFilter, ev *pb.Event, toMe bool) (bool, error) {
	if filter == nil {
		return true, nil
	}
	chatType := ev.GetChatType()
	static := filter.GetToMe() && toMe || slices.Contains(filter.GetChatTypes(), chatType)
	wakePath, triggersPath := filter.GetWakeWordsConfig(), filter.GetTriggersConfig()
	if wakePath == "" {
		return static, nil
	}
	if r.messageConfig == nil {
		return false, errors.New("message filter config reader is unavailable")
	}
	snapshot, err := r.messageConfig(r.ctx, pluginID, ev)
	if err != nil {
		return false, err
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(snapshot, &config); err != nil {
		return false, fmt.Errorf("decode message filter config: %w", err)
	}
	// 选择器决定需要检查哪些条件；没选 wake_word 时不访问词表路径。
	if triggersPath != "" {
		triggers, err := messageConfigList(config, triggersPath)
		if err != nil {
			return false, err
		}
		if chatType == pb.ChatType_CHAT_TYPE_GROUP && toMe && slices.Contains(triggers, "mention") ||
			chatType == pb.ChatType_CHAT_TYPE_PRIVATE && slices.Contains(triggers, "private") {
			return true, nil
		}
		if !slices.Contains(triggers, "wake_word") {
			return false, nil
		}
	}
	words, err := messageConfigList(config, wakePath)
	if err != nil {
		return false, err
	}
	return static || matchesWakeWord(ev.GetEventPayload(), words), nil
}

func messageConfigList(config map[string]json.RawMessage, path string) ([]string, error) {
	parts := strings.Split(path, ".")
	for _, key := range parts[:len(parts)-1] {
		var child map[string]json.RawMessage
		if err := json.Unmarshal(config[key], &child); err != nil {
			return nil, fmt.Errorf("read message filter field %q: %w", path, err)
		}
		config = child
	}
	var list []*string
	if err := json.Unmarshal(config[parts[len(parts)-1]], &list); err != nil {
		return nil, fmt.Errorf("read message filter field %q: %w", path, err)
	}
	if list == nil {
		return nil, fmt.Errorf("message filter field %q is not a string list", path)
	}
	words := make([]string, len(list))
	for i, item := range list {
		if item == nil {
			return nil, fmt.Errorf("message filter field %q contains a non-string element", path)
		}
		words[i] = *item
	}
	return words, nil
}

func matchesWakeWord(payload *pb.EventPayload, words []string) bool {
	chain := payload.GetMessageCreated().GetChain()
	if edited := payload.GetMessageEdited(); edited != nil {
		chain = edited.GetChain()
	}
	var text strings.Builder
	for _, segment := range chain.GetSegments() {
		text.WriteString(segment.GetText().GetText())
	}
	if text.Len() == 0 {
		return false
	}
	folded := foldText(text.String())
	for _, word := range words {
		if word != "" && strings.Contains(folded, foldText(word)) {
			return true
		}
	}
	return false
}
