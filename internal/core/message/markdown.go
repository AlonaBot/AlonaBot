package message

import (
	"fmt"

	"github.com/AlonaBot/AlonaBot/contracts/markdown"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

const markdownCapability = "message.markdown"

func (s *Service) normalizeMarkdown(chain *pb.MessageChain, target *pb.Target) *pb.ErrorInfo {
	var generic, native bool
	var scan func(*pb.MessageChain)
	scan = func(c *pb.MessageChain) {
		for _, segment := range c.GetSegments() {
			generic = generic || segment.GetMarkdown() != nil
			native = native || segment.GetText().GetFormat() == pb.TextFormat_TEXT_FORMAT_MARKDOWN
			for _, item := range segment.GetForward().GetItems() {
				scan(item.GetChain())
			}
		}
	}
	scan(chain)
	if !generic {
		return nil
	}
	if native {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "generic Markdown and platform-native Markdown cannot be mixed"}
	}
	instance, ok := s.adapters.Get(target.GetAdapterInstanceId())
	if !ok {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	supported := false
	for _, capability := range instance.Capabilities() {
		if capability.Name == markdownCapability && capability.Supported {
			supported = true
		}
	}
	if !supported {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter instance does not support message.markdown"}
	}
	var normalize func(*pb.MessageChain, string) *pb.ErrorInfo
	normalize = func(c *pb.MessageChain, path string) *pb.ErrorInfo {
		for i, segment := range c.GetSegments() {
			at := fmt.Sprintf("%ssegment[%d]", path, i)
			if md := segment.GetMarkdown(); md != nil {
				parsed, err := markdown.RichText(md.Content)
				if err != nil {
					return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: at + ": " + err.Error()}
				}
				segment.Data = &pb.Segment_Text{Text: parsed}
			}
			for j, item := range segment.GetForward().GetItems() {
				if safe := normalize(item.GetChain(), fmt.Sprintf("%s.forward.items[%d].", at, j)); safe != nil {
					return safe
				}
			}
		}
		return nil
	}
	return normalize(chain, "")
}
