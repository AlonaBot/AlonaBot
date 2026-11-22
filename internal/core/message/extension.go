package message

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
)

const (
	voiceCapability   = "message.voice"
	forwardCapability = "message.forward"
)

// checkExtensions 递归核对新增片段能力与 Adapter 扩展，先于任何发送副作用。
func checkExtensions(instance adapter.Instance, chain *pb.MessageChain) *pb.ErrorInfo {
	for _, segment := range chain.GetSegments() {
		var required string
		switch segment.GetData().(type) {
		case *pb.Segment_Voice:
			required = voiceCapability
		case *pb.Segment_Forward:
			required = forwardCapability
		}
		if required != "" {
			supported := false
			for _, capability := range instance.Capabilities() {
				if capability.GetName() == required && capability.GetSupported() {
					supported = true
					break
				}
			}
			if !supported {
				return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter instance does not support " + required}
			}
		}
		if extension := segment.GetExtension(); extension != nil {
			checker, ok := instance.(adapter.ExtensionChecker)
			if extension.GetAdapterId() != instance.AdapterID() || !ok {
				return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter instance does not support the extension"}
			}
			if err := checker.CheckExtension(extension.GetName(), extension.GetVersion(), extension.GetPayloadJson()); err != nil {
				return err
			}
		}
		for _, item := range segment.GetForward().GetItems() {
			if err := checkExtensions(instance, item.GetChain()); err != nil {
				return err
			}
		}
	}
	return nil
}
