package adapter

import (
	"context"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// MessageDescriber 返回当前运行实例的出站声明，保留 Session 与代次错误。
type MessageDescriber interface {
	DescribeMessage(context.Context) (*pb.MessageDescription, *pb.ErrorInfo)
}

// MessageChecker 只读检查消息和 Core 已授权的来源元数据。
type MessageChecker interface {
	CheckMessage(context.Context, *pb.Target, *pb.MessageChain) (pb.SupportStatus, *pb.ErrorInfo)
}
