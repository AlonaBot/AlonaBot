package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Editor 提供实例可选的消息编辑入口，完整保留编辑回执与安全错误。
// 未实现此接口的实例在收到 adapter.edit 时返回 UNSUPPORTED。
type Editor interface {
	Edit(ctx context.Context, target *pb.Target, platformMessageID string, chain *pb.MessageChain) (*pb.SendResult, *pb.ErrorInfo)
}
