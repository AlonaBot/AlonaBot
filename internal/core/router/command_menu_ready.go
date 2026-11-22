package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"slices"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

type menuReadyKey struct {
	pluginID, identity, chatID string
}

// notifyMenuReady 比较当前准入与已通知 Session，让晚启动、重获许可和新聊天各收到一次通知。
func (r *Router) notifyMenuReady(ctx context.Context, worker *commandSync) error {
	r.mu.Lock()
	sources := r.routeSourcesLocked()
	r.mu.Unlock()
	pluginIDs := make([]string, 0, len(sources))
	for id := range sources {
		pluginIDs = append(pluginIDs, id)
	}
	slices.Sort(pluginIDs)
	chatIDs := []string{""}
	for id, chat := range worker.chats {
		if chat.Active {
			chatIDs = append(chatIDs, id)
		}
	}
	slices.Sort(chatIDs)
	current := make(map[menuReadyKey]bool)
	for _, pluginID := range pluginIDs {
		source := sources[pluginID]
		if !slices.ContainsFunc(source.routes, func(route *pb.RouteDescriptor) bool {
			return slices.Contains(route.GetEvent().GetEventTypeFilter(), menuReadyEvent)
		}) {
			continue
		}
		identity := source.generationID
		if source.session != nil {
			identity = source.session.caller.SessionKey
		}
		for _, chatID := range chatIDs {
			allowed, err := r.menuOwnerAllowed(ctx, worker, pluginID, chatID)
			if err != nil {
				return r.menuFailure(worker, err)
			}
			if !allowed {
				continue
			}
			key := menuReadyKey{pluginID: pluginID, identity: identity, chatID: chatID}
			current[key] = true
			if worker.notified[key] {
				continue
			}
			id := make([]byte, traceIDBytes)
			rand.Read(id)
			event := &pb.Event{EventId: hex.EncodeToString(id), EventType: menuReadyEvent, OccurredAtUnixMs: time.Now().UnixMilli(), SourceKind: pb.EventSourceKind_EVENT_SOURCE_KIND_ADAPTER,
				AdapterId: proto.String(worker.instance.AdapterID()), AdapterInstanceId: proto.String(worker.instance.InstanceID()), AccountId: proto.String(worker.instance.AccountID()),
				EventPayload: &pb.EventPayload{Kind: &pb.EventPayload_CommandMenuReady{CommandMenuReady: &pb.CommandMenuReadyPayload{AdapterId: worker.instance.AdapterID(), AdapterInstanceId: worker.instance.InstanceID(), AccountId: worker.instance.AccountID(), ChatId: chatID}}}}
			if chatID != "" {
				event.Target = menuTarget(worker, chatID)
			}
			if err := r.deliverTo(ctx, event, pluginID); err != nil {
				return r.menuFailure(worker, err)
			}
			worker.notified[key] = true
		}
	}
	worker.notified = current
	return nil
}
