package validate

import (
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/google/jsonschema-go/jsonschema"
)

func messageDeclarations(d *pb.AdapterDescriptor) error {
	standard := map[string]bool{"text": true, "markdown": true, "image": true, "video": true, "voice": true, "file": true, "mention": true, "quote": true, "forward": true, "buttons": true}
	formats := make(map[string]bool)
	for _, f := range d.GetMessageFormats() {
		if !standard[f.GetName()] || f.GetVersion() != 1 {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid standard message format %q version %d", f.GetName(), f.GetVersion())
		}
		if formats[f.Name] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate message format %q", f.Name)
		}
		formats[f.Name] = true
		if _, ok := pb.SupportStatus_name[int32(f.Status)]; !ok {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid message format %q status", f.Name)
		}
		if f.AdapterId != "" && f.AdapterId != d.AdapterId {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "message format %q adapter does not match registration", f.Name)
		}
		for _, source := range f.Sources {
			switch source {
			case "https_url", "media_ref", "incoming", "inline_base64", "platform_file":
			default:
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid message format %q source %q", f.Name, source)
			}
		}
		for _, limit := range []*pb.TextLimit{f.TextLimit, f.ButtonLabelLimit} {
			if limit == nil {
				continue
			}
			if limit.Max == 0 || limit.Unit < pb.TextUnit_TEXT_UNIT_UTF16_CODE_UNITS || limit.Unit > pb.TextUnit_TEXT_UNIT_BYTES {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid message format %q text limit", f.Name)
			}
		}
		if f.SchemaJson != "" {
			var schema jsonschema.Schema
			if err := json.Unmarshal([]byte(f.SchemaJson), &schema); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "parse message format %q schema: %v", f.Name, err)
			}
			if _, err := schema.Resolve(nil); err != nil {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "resolve message format %q schema: %v", f.Name, err)
			}
		}
		if f.ExampleJson != "" && !IsJSONObject([]byte(f.ExampleJson)) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "message format %q example must be a JSON object", f.Name)
		}
	}
	for _, extension := range d.Extensions {
		standard[extension.Name] = true
	}
	combinations := make(map[string]bool)
	for _, rule := range d.GetMessageCombinations() {
		if rule.GetName() == "" || len(rule.GetFormats()) == 0 || rule.GetDescription() == "" {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "message combination requires name, formats and description")
		}
		if combinations[rule.Name] {
			return errf(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "duplicate message combination %q", rule.Name)
		}
		combinations[rule.Name] = true
		for _, name := range rule.Formats {
			if !standard[name] {
				return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "message combination %q references unknown format %q", rule.Name, name)
			}
		}
		if rule.ExampleJson != "" && !IsJSONObject([]byte(rule.ExampleJson)) {
			return errf(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "message combination %q example must be a JSON object", rule.Name)
		}
	}
	return nil
}
