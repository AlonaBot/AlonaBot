package message

import (
	"fmt"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

// replySources 将 Core 修饰后链的每个顶层片段映射到调用方原链；自动插入片段没有来源。
type replySources [][]uint32

func newReplySources(chain *pb.MessageChain) replySources {
	sources := make(replySources, len(chain.Segments))
	for i := range sources {
		sources[i] = []uint32{uint32(i)}
	}
	return sources
}

func (s replySources) insert(index int) replySources {
	return slices.Insert(s, index, []uint32(nil))
}

func (s replySources) forward() replySources {
	var original []uint32
	for _, indexes := range s {
		original = append(original, indexes...)
	}
	return replySources{original}
}

// result 只改返回调用方的副本；观察、Hook 与回复修饰记账继续使用平台原回执。
func (s replySources) result(result *pb.SendResult) (*pb.SendResult, error) {
	if s == nil || result == nil {
		return result, nil
	}
	// Adapter 回执属于协议边界；错误下标保留完整回执并报告，避免映射时越界。
	for _, part := range result.PartResults {
		for _, index := range part.SegmentIndexes {
			if uint64(index) >= uint64(len(s)) {
				return result, fmt.Errorf("map reply receipt: adapter segment index %d exceeds sent chain length %d", index, len(s))
			}
		}
	}
	mapped := proto.CloneOf(result)
	for _, part := range mapped.PartResults {
		var original []uint32
		for _, index := range part.SegmentIndexes {
			for _, source := range s[index] {
				if !slices.Contains(original, source) {
					original = append(original, source)
				}
			}
		}
		part.SegmentIndexes = original
	}
	return mapped, nil
}
