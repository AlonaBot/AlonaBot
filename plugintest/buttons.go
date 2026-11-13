package plugintest

import (
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// Click 点击已发送链中的动作按钮，以本次用户身份经过 Core 的令牌解析与鉴权。
func (h *Harness) Click(from User, sent Sent, actionID string) Result {
	h.t.Helper()
	if !h.busy.CompareAndSwap(false, true) {
		h.t.Fatal("concurrent event injection on one Harness")
	}
	defer h.busy.Store(false)
	for _, segment := range sent.Chain {
		buttons, ok := segment.(message.ButtonsSegment)
		if !ok {
			continue
		}
		for _, row := range buttons.Rows {
			for _, button := range row {
				if button.ActionID != actionID {
					continue
				}
				if button.URL != "" || button.CoreToken == "" || sent.Result == nil || sent.Result.PrimaryPlatformMessageID == "" {
					h.t.Fatalf("action %q is not a sent callback button", actionID)
				}
				adapterID, chatID, _, ok := sent.To.Parts()
				if !ok {
					h.t.Fatalf("invalid button session %q", sent.To)
				}
				chatType := message.ChatUnspecified
				switch {
				case strings.HasPrefix(chatID, "private:"):
					chatType = message.ChatPrivate
				case strings.HasPrefix(chatID, "group:"):
					chatType = message.ChatGroup
				}
				return h.deliverEvent(from, &message.Event{Type: "button.clicked", AdapterID: adapterID,
					AdapterInstanceID: sent.Result.AdapterInstanceID, AccountID: AccountID, Session: sent.To, ChatType: chatType,
					Payload: &message.ButtonClickedPayload{Value: button.CoreToken, OriginPlatformMessageID: sent.Result.PrimaryPlatformMessageID}})
			}
		}
	}
	h.t.Fatalf("action %q is not present in sent buttons", actionID)
	return Result{}
}
