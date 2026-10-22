// 本文件负责握手时的协议版本与能力协商（WIRE §3.1）。
package validate

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Offer 是一端在握手中声明的协议版本与能力。
type Offer struct {
	// Major 是协议主版本，双方必须相等。
	Major uint32
	// Minor 是支持的最高次版本。
	Minor uint32
	// Supported 是可用的能力集合。
	Supported []*pb.ProtocolCapability
	// Required 是必需能力，必须包含于双方交集。
	Required []*pb.ProtocolCapability
}

// Agreement 是协商结果。
type Agreement struct {
	// Minor 是双方次版本的较小值，之后每一帧都使用它。
	Minor uint32
	// Intersection 是双方都支持的能力，按本端 Supported 的顺序排列。
	Intersection []*pb.ProtocolCapability
}

// capKey 是能力的精确匹配键。
type capKey struct {
	// name 是能力名。
	name string
	// version 是能力版本。
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
