// Package message 提供构造 MessageChain 的类型（MESSAGE-MODEL §2）。一条 Chain 是一次逻辑 Send。
package message

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// Segment 是 Chain 中的一个片段。
type Segment interface {
	// toProto 转换为线上的 Segment；未导出以限定片段类型只由本包提供。
	toProto() *pb.Segment
}

// Chain 是一次逻辑 Send 的有序片段。
type Chain []Segment

// Text 构造纯文本片段。
func Text(s string) Segment { return text(s) }

// text 是纯文本片段。
type text string

// toProto 转换为 TextSegment。
func (t text) toProto() *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: string(t)}}}
}

// Proto 把 Chain 转成线上的 MessageChain；结构校验由 Core 端的 contracts/validate 完成。
func (c Chain) Proto() *pb.MessageChain {
	out := &pb.MessageChain{Segments: make([]*pb.Segment, 0, len(c))}
	for _, s := range c {
		out.Segments = append(out.Segments, s.toProto())
	}
	return out
}
