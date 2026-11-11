package memory

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"google.golang.org/protobuf/proto"
)

const editCapability = "message.edit"

var _ adapter.Editor = (*Adapter)(nil)

// Edit 将编辑快照追加到 Sent 记录，并为原消息返回一个 ACCEPTED 回执。
func (a *Adapter) Edit(_ context.Context, target *pb.Target, platformMessageID string, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo) {
	a.mu.Lock()
	a.sent = append(a.sent, Sent{Target: proto.CloneOf(target), Chain: proto.CloneOf(chain)})
	a.mu.Unlock()
	indexes := make([]uint32, len(chain.GetSegments()))
	for i := range indexes {
		indexes[i] = uint32(i)
	}
	return &pb.SendResult{
		Status: pb.SendStatus_SEND_STATUS_ACCEPTED, Target: proto.CloneOf(target),
		PlatformMessageIds: []string{platformMessageID}, PrimaryPlatformMessageId: proto.String(platformMessageID),
		CapabilitiesUsed: []*pb.Capability{{Name: editCapability, Version: 1, Supported: true}},
		PartResults:      []*pb.PartResult{{SegmentIndexes: indexes, Status: pb.PartStatus_PART_STATUS_ACCEPTED, PlatformMessageIds: []string{platformMessageID}}},
	}, nil
}
