package broker

import (
	"context"
	"log/slog"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

func (b *Broker) agentID(ctx context.Context) (string, *pb.ErrorInfo) {
	if b.agent == nil {
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "agent is not available"}
	}
	id, err := b.agent(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "resolve agent failed", "error", err)
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "resolve agent failed"}
	}
	if id == "" {
		return "", &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "agent is not available"}
	}
	return id, nil
}
