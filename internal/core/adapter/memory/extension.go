package memory

import (
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/google/jsonschema-go/jsonschema"
)

var _ adapter.ExtensionChecker = (*Adapter)(nil)

type extensionKey struct {
	name    string
	version uint32
}

// SetExtensions 原子替换测试实例的声明；解析失败保留原声明。
func (a *Adapter) SetExtensions(declarations []*pb.ExtensionDeclaration) error {
	compiled := make(map[extensionKey]*jsonschema.Resolved, len(declarations))
	for _, declaration := range declarations {
		var schema jsonschema.Schema
		if err := json.Unmarshal([]byte(declaration.GetSchemaJson()), &schema); err != nil {
			return fmt.Errorf("parse adapter extension schema: %w", err)
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			return fmt.Errorf("resolve adapter extension schema: %w", err)
		}
		compiled[extensionKey{declaration.GetName(), declaration.GetVersion()}] = resolved
	}
	a.mu.Lock()
	a.extensions = compiled
	a.mu.Unlock()
	return nil
}

// CheckExtension 按当前测试声明校验载荷。
func (a *Adapter) CheckExtension(name string, version uint32, payloadJSON string) *pb.ErrorInfo {
	a.mu.Lock()
	schema := a.extensions[extensionKey{name, version}]
	a.mu.Unlock()
	if schema == nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, SafeMessage: "adapter extension is not declared"}
	}
	var value any
	if err := json.Unmarshal([]byte(payloadJSON), &value); err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: err.Error()}
	}
	if err := schema.Validate(value); err != nil {
		return &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, SafeMessage: err.Error()}
	}
	return nil
}
