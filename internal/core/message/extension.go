package message

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
)

func checkExtensions(instance adapter.Instance, chain *pb.MessageChain) *pb.ErrorInfo {
	for _, segment := range chain.GetSegments() {
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
