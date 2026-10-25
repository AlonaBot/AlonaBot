// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：握手时协商协议版本与能力（WIRE §3.1）。

package validate

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Offer 描述一端在握手中声明的协议版本与能力。
type Offer struct {
	Major     uint32                   // 协议主版本，双方必须相等
	Minor     uint32                   // 支持的最高次版本
	Supported []*pb.ProtocolCapability // 可用的能力集合
	Required  []*pb.ProtocolCapability // 必需能力，必须包含于双方交集
}

// Agreement 保存握手协商出的版本与能力交集。
type Agreement struct {
	Minor        uint32                   // 双方次版本的较小值，之后每一帧都使用它
	Intersection []*pb.ProtocolCapability // 双方都支持的能力，按本端 Supported 的顺序排列
}

// capKey 按名字和版本精确匹配一项能力。
type capKey struct {
	name    string
	version uint32
}

// Negotiate 实现 WIRE §3.1：Major 必须相等，Minor 取较小值；能力以 name+version 精确匹配，
// 交集按 local.Supported 的顺序给出，双方的 required 都必须包含于交集。失败返回 FAILED_PRECONDITION。
func Negotiate(local, peer Offer) (Agreement, error) {
	if local.Major != peer.Major {
		return Agreement{}, errf(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "协议主版本不一致：本端 %d，对端 %d", local.Major, peer.Major)
	}
	peerSet := make(map[capKey]bool, len(peer.Supported))
	for _, c := range peer.Supported {
		peerSet[capKey{c.GetName(), c.GetVersion()}] = true
	}
	var inter []*pb.ProtocolCapability
	interSet := map[capKey]bool{}
	for _, c := range local.Supported {
		k := capKey{c.GetName(), c.GetVersion()}
		if peerSet[k] && !interSet[k] {
			interSet[k] = true
			inter = append(inter, c)
		}
	}
	for _, req := range [][]*pb.ProtocolCapability{local.Required, peer.Required} {
		for _, c := range req {
			if !interSet[capKey{c.GetName(), c.GetVersion()}] {
				return Agreement{}, errf(pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, "必需能力 %s@%d 不在双方交集中", c.GetName(), c.GetVersion())
			}
		}
	}
	return Agreement{Minor: min(local.Minor, peer.Minor), Intersection: inter}, nil
}
