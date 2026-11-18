package validate

import (
	"regexp"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

var messageConfigPath = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*$`)

func registerMessageFilter(routeID string, event *pb.EventSpec) error {
	filter := event.GetMessageFilter()
	if filter == nil {
		return nil
	}
	const invalid = pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
	for _, eventType := range event.GetEventTypeFilter() {
		switch eventType {
		case "message.created", "message.edited", "message.deleted":
		default:
			return errf(invalid, "event route %q cannot filter event type %q", routeID, eventType)
		}
	}
	static := filter.GetToMe() || len(filter.GetChatTypes()) != 0
	wake, triggers := filter.GetWakeWordsConfig(), filter.GetTriggersConfig()
	if !static && wake == "" && triggers == "" {
		return errf(invalid, "event route %q has an empty message_filter", routeID)
	}
	if triggers != "" && (static || wake == "") {
		return errf(invalid, "event route %q requires wake_words_config and no static conditions with triggers_config", routeID)
	}
	for _, chatType := range filter.GetChatTypes() {
		switch chatType {
		case pb.ChatType_CHAT_TYPE_PRIVATE, pb.ChatType_CHAT_TYPE_GROUP, pb.ChatType_CHAT_TYPE_CHANNEL:
		default:
			return errf(invalid, "event route %q has invalid message_filter chat type %v", routeID, chatType)
		}
	}
	for _, path := range []string{wake, triggers} {
		if path != "" && !messageConfigPath.MatchString(path) {
			return errf(invalid, "event route %q has invalid message_filter config path %q", routeID, path)
		}
	}
	return nil
}
