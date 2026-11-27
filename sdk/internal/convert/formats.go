package convert

import (
	"encoding/json"
	"slices"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// SupportStatusFromProto 转换支持状态；协议零值表示 unknown。
func SupportStatusFromProto(status pb.SupportStatus) message.SupportStatus {
	switch status {
	case pb.SupportStatus_SUPPORT_STATUS_SUPPORTED:
		return message.SupportSupported
	case pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED:
		return message.SupportUnsupported
	default:
		return message.SupportUnknown
	}
}

// SupportStatusToProto 编码支持状态。
func SupportStatusToProto(status message.SupportStatus) pb.SupportStatus {
	switch status {
	case message.SupportSupported:
		return pb.SupportStatus_SUPPORT_STATUS_SUPPORTED
	case message.SupportUnsupported:
		return pb.SupportStatus_SUPPORT_STATUS_UNSUPPORTED
	case "", message.SupportUnknown:
		return pb.SupportStatus_SUPPORT_STATUS_UNSPECIFIED
	default:
		return pb.SupportStatus(-1)
	}
}

func textLimitToProto(limit *message.TextLimit) *pb.TextLimit {
	if limit == nil {
		return nil
	}
	return &pb.TextLimit{Max: limit.Max, Unit: pb.TextUnit(limit.Unit)}
}
func textLimitFromProto(limit *pb.TextLimit) *message.TextLimit {
	if limit == nil {
		return nil
	}
	return &message.TextLimit{Max: limit.Max, Unit: message.TextUnit(limit.Unit)}
}

// MessageFormatsToProto 编码格式及可选限制。
func MessageFormatsToProto(formats []message.MessageFormat) []*pb.MessageFormat {
	var out []*pb.MessageFormat
	for _, f := range formats {
		out = append(out, &pb.MessageFormat{Name: f.Name, Version: f.Version, Status: SupportStatusToProto(f.Status), Description: f.Description, SchemaJson: string(f.SchemaJSON), ExampleJson: string(f.ExampleJSON), Sources: slices.Clone(f.Sources), TextLimit: textLimitToProto(f.TextLimit), ButtonLabelLimit: textLimitToProto(f.ButtonLabelLimit), AdapterId: f.AdapterID})
	}
	return out
}
func messageFormatsFromProto(formats []*pb.MessageFormat) []message.MessageFormat {
	var out []message.MessageFormat
	for _, f := range formats {
		out = append(out, message.MessageFormat{Name: f.Name, Version: f.Version, Status: SupportStatusFromProto(f.Status), Description: f.Description, SchemaJSON: json.RawMessage(f.SchemaJson), ExampleJSON: json.RawMessage(f.ExampleJson), Sources: slices.Clone(f.Sources), TextLimit: textLimitFromProto(f.TextLimit), ButtonLabelLimit: textLimitFromProto(f.ButtonLabelLimit), AdapterID: f.AdapterId})
	}
	return out
}

// CombinationsToProto 编码说明性质的组合规则。
func CombinationsToProto(rules []message.CombinationRule) []*pb.MessageCombination {
	var out []*pb.MessageCombination
	for _, r := range rules {
		out = append(out, &pb.MessageCombination{Name: r.Name, Formats: slices.Clone(r.Formats), Description: r.Description, ExampleJson: string(r.ExampleJSON)})
	}
	return out
}

// OutboundDescriptionFromProto 补齐标准格式的 unknown 和 SDK 共用 Schema、例子。
func OutboundDescriptionFromProto(d *pb.MessageDescription) *message.OutboundDescription {
	out := &message.OutboundDescription{Formats: message.StandardFormats(), Extensions: messageFormatsFromProto(d.Extensions)}
	for _, declared := range messageFormatsFromProto(d.Formats) {
		for i, standard := range out.Formats {
			if standard.Name != declared.Name {
				continue
			}
			if declared.Description == "" {
				declared.Description = standard.Description
			}
			if len(declared.SchemaJSON) == 0 {
				declared.SchemaJSON = standard.SchemaJSON
			}
			if len(declared.ExampleJSON) == 0 {
				declared.ExampleJSON = standard.ExampleJSON
			}
			out.Formats[i] = declared
			break
		}
	}
	for _, r := range d.Combinations {
		out.Combinations = append(out.Combinations, message.CombinationRule{Name: r.Name, Formats: slices.Clone(r.Formats), Description: r.Description, ExampleJSON: json.RawMessage(r.ExampleJson)})
	}
	return out
}
