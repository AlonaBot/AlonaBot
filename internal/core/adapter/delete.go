package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Deleter 提供实例可选的撤回入口，按输入顺序保留逐项结果及安全错误。
type Deleter interface {
	Delete(ctx context.Context, target *pb.Target, platformMessageIDs []string) (*pb.DeleteResult, *pb.ErrorInfo)
}
