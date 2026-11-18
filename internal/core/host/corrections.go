package host

import (
	"encoding/json"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

func configCorrections(records []*pb.ConfigCorrection, secretPaths [][]string) ([]configschema.Correction, error) {
	corrections := make([]configschema.Correction, 0, len(records))
	for _, record := range records {
		if !strings.HasPrefix(record.GetPath(), "$") ||
			(record.GetReason() != "converted" && record.GetReason() != "reset" && record.GetReason() != "removed") ||
			(record.GetOriginalJson() != "" && !json.Valid([]byte(record.GetOriginalJson()))) {
			return nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "invalid config correction"}
		}
		correction := configschema.Correction{Path: record.Path, Reason: record.Reason, Old: json.RawMessage(record.OriginalJson)}
		// Schema 字段与模板键不含点号；数组下标独占一段，与 * 对应。
		parts := strings.FieldsFunc(strings.TrimPrefix(correction.Path, "$"), func(c rune) bool { return c == '.' || c == '[' || c == ']' })
		for _, secret := range secretPaths {
			shared := true
			for i := 0; i < min(len(parts), len(secret)); i++ {
				if secret[i] != "*" && secret[i] != parts[i] {
					shared = false
					break
				}
			}
			if shared {
				correction.Old = nil
				break
			}
		}
		record.OriginalJson = string(correction.Old)
		corrections = append(corrections, correction)
	}
	return corrections, nil
}
