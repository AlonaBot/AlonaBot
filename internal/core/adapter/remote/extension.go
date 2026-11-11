package remote

import (
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
)

var _ adapter.ExtensionChecker = (*instance)(nil)

type extensionKey struct {
	name    string
	version uint32
}

// CheckExtension 使用当前 Session 的只读声明；旧实例不会采用新 Session 的声明。
func (i *instance) CheckExtension(name string, version uint32, payloadJSON string) *pb.ErrorInfo {
	p := i.process
	p.mu.Lock()
	schema := p.declarations.extensions[extensionKey{name, version}]
	current := p.session != nil && p.session.Session == i.session
	p.mu.Unlock()
	if !current || schema == nil {
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
