package remote

import (
	"context"
	"errors"

	"github.com/AlonaBot/AlonaBot/contracts/adapterops"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// DescribeMessage 读取本次运行所属 REGISTER 声明及实例能力快照。
func (i *instance) DescribeMessage(ctx context.Context) (*pb.MessageDescription, *pb.ErrorInfo) {
	if err := ctx.Err(); err != nil {
		return nil, messageCheckError(err)
	}
	p := i.process
	p.mu.Lock()
	defer p.mu.Unlock()
	if !i.running || p.session == nil || p.session.Session != i.session || p.instances[i.id] != i {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance is unavailable"}
	}
	descriptor := p.session.Adapter
	description := &pb.MessageDescription{}
	for _, f := range descriptor.GetMessageFormats() {
		format := proto.CloneOf(f)
		capabilityName := "message." + format.Name
		switch format.Name {
		case "text":
			capabilityName = "message.send"
		case "quote":
			capabilityName = "message.reference"
		}
		for _, capability := range i.capabilities {
			if capability.Name == capabilityName && !capability.Supported {
				format.Status = pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED
			}
			if capability.Name == "message.send" && capability.TextLimit != nil && (format.Name == "text" || format.Name == "markdown") {
				format.TextLimit = proto.CloneOf(capability.TextLimit)
			}
		}
		description.Formats = append(description.Formats, format)
	}
	for _, e := range descriptor.GetExtensions() {
		description.Extensions = append(description.Extensions, &pb.MessageFormat{Name: e.Name, Version: e.Version, Status: pb.SupportStatus_SUPPORT_STATUS_SUPPORTED, Description: e.Description, SchemaJson: e.SchemaJson, ExampleJson: e.ExampleJson, AdapterId: descriptor.AdapterId})
	}
	for _, r := range descriptor.GetMessageCombinations() {
		description.Combinations = append(description.Combinations, proto.CloneOf(r))
	}
	// 旧声明仍保留实例已知的文本限额，格式支持状态继续为 unknown。
	for _, capability := range i.capabilities {
		if capability.Name != "message.send" || capability.TextLimit == nil {
			continue
		}
		for _, name := range []string{"text", "markdown"} {
			found := false
			for _, format := range description.Formats {
				if format.Name == name {
					found = true
					break
				}
			}
			if !found {
				description.Formats = append(description.Formats, &pb.MessageFormat{Name: name, Version: 1, TextLimit: proto.CloneOf(capability.TextLimit)})
			}
		}
	}
	return description, nil
}

// CheckMessage 从真实操作目录确认可选接口，并在原 Session 与代次内调用。
func (i *instance) CheckMessage(ctx context.Context, target *pb.Target, chain *pb.MessageChain) (pb.SupportStatus, *pb.ErrorInfo) {
	operations, err := i.Operations(ctx)
	if err != nil {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, messageCheckError(err)
	}
	operation, _ := adapterops.Find(adapterops.CheckMessageName)
	supported := false
	for _, declared := range operations.Operations {
		if declared.Name == operation.Name && declared.Version == 1 {
			supported = true
			break
		}
	}
	if !supported {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, nil
	}
	request, err := payload(&pb.AdapterSendRequest{InstanceId: i.id, InstanceGeneration: i.generation, Target: target, Chain: chain})
	if err != nil {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, messageCheckError(err)
	}
	response, callError := i.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterCheckMessage, Payload: request})
	result := &pb.CheckMessageResponse{}
	if response.Payload.SchemaID != "" {
		if err := decode(response.Payload, result); err != nil {
			return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, messageCheckError(err)
		}
	} else if response.Err == nil && callError == nil {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INTERNAL, SafeMessage: "adapter check response is missing"}
	}
	if response.Err != nil {
		return result.Status, response.Err
	}
	if callError != nil {
		return result.Status, messageCheckError(callError)
	}
	if _, ok := pb.SupportStatus_name[int32(result.Status)]; !ok {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: "invalid adapter check status"}
	}
	p := i.process
	p.mu.Lock()
	current := i.running && p.session != nil && p.session.Session == i.session && p.instances[i.id] == i
	p.mu.Unlock()
	if !current {
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "adapter instance changed during message check"}
	}
	return result.Status, nil
}

func messageCheckError(err error) *pb.ErrorInfo {
	var call *wire.CallError
	if errors.As(err, &call) {
		return call.ErrorInfo
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	if errors.Is(err, context.Canceled) {
		code = pb.ErrorCode_ERROR_CODE_CANCELLED
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED
	}
	return &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}
}
