package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// ActionExecutor 在执行平台动作前检查声明与参数；结果与安全错误可以同时返回。
type ActionExecutor interface {
	CheckAction(name string, version uint32, paramsJSON string) (pb.ActionScope, *pb.ErrorInfo)
	Action(ctx context.Context, target *pb.Target, name string, version uint32, paramsJSON string) (resultJSON string, safe *pb.ErrorInfo)
}

// CompleteAccount 选择显式实例或 Adapter 的缺省实例，只补齐 Adapter、实例与账号。
// 未登记的实例保留给调用方按准入顺序拒绝，chat 与 thread 保持为空。
func (r *Registry) CompleteAccount(adapterID, instanceID string) (*pb.Target, *pb.ErrorInfo) {
	r.mu.RLock()
	if instanceID == "" {
		instanceID = r.defaults[adapterID]
	}
	instance, exists := r.instances[instanceID]
	r.mu.RUnlock()
	if instanceID == "" {
		return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, SafeMessage: "no adapter instance selected; specify one or set default_instance"}
	}
	target := &pb.Target{AdapterId: adapterID, AdapterInstanceId: instanceID}
	if exists {
		target.AccountId = instance.AccountID()
	}
	return target, nil
}
