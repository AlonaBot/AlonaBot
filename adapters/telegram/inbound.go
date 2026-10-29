package telegram

import (
	"encoding/json"
	"strconv"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

type telegramMessage struct {
	ID   int64 `json:"message_id"`
	Date int64 `json:"date"`
	Chat struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
	ThreadID     *int64  `json:"message_thread_id"`
	Text         *string `json:"text"`
	Caption      *string `json:"caption"`
	MediaGroupID *string `json:"media_group_id"`
}

func (i *instance) normalize(raw json.RawMessage) (*pb.Event, error) {
	var update struct {
		ID                int64            `json:"update_id"`
		Message           *telegramMessage `json:"message"`
		ChannelPost       *telegramMessage `json:"channel_post"`
		EditedMessage     *telegramMessage `json:"edited_message"`
		EditedChannelPost *telegramMessage `json:"edited_channel_post"`
	}
	if err := json.Unmarshal(raw, &update); err != nil {
		return nil, err
	}
	message := update.Message
	if message == nil {
		message = update.ChannelPost
	}
	edited := false
	if message == nil {
		message = update.EditedMessage
		edited = message != nil
	}
	if message == nil {
		message = update.EditedChannelPost
		edited = message != nil
	}
	if message == nil {
		return nil, nil
	}
	target := &pb.Target{AdapterId: AdapterID, AdapterInstanceId: i.id, AccountId: i.account.ID, ChatId: strconv.FormatInt(message.Chat.ID, 10)}
	if message.ThreadID != nil {
		target.ThreadId = proto.String(strconv.FormatInt(*message.ThreadID, 10))
	}
	chain := &pb.MessageChain{}
	text := message.Text
	if text == nil {
		text = message.Caption
	}
	if text != nil {
		chain.Segments = []*pb.Segment{{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: *text}}}}
	}
	platformID := strconv.FormatInt(message.ID, 10)
	event := &pb.Event{AdapterId: proto.String(AdapterID), AdapterInstanceId: proto.String(i.id), AccountId: proto.String(i.account.ID), Target: target, OccurredAtUnixMs: message.Date * 1000, SourceEventId: proto.String(strconv.FormatInt(update.ID, 10)), PlatformPayload: raw, Message: &pb.PlatformMessage{PlatformMessageId: platformID, Chain: chain, MediaGroupId: message.MediaGroupID}}
	if edited {
		event.EventType = "message.edited"
		event.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_MessageEdited{MessageEdited: &pb.MessageEditedPayload{PlatformMessageId: platformID, Chain: chain}}}
	} else {
		event.EventType = "message.created"
		event.EventPayload = &pb.EventPayload{Kind: &pb.EventPayload_MessageCreated{MessageCreated: &pb.MessageCreatedPayload{PlatformMessageId: platformID, Chain: chain, ToMe: message.Chat.Type == "private"}}}
	}
	return event, nil
}
