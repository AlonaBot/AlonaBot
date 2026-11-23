package adapter

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// OperationCatalog 查询真实运行实例的操作，目录随该实例和 Session 失效。
type OperationCatalog interface {
	Operations(context.Context) (*pb.AdapterOperationsResponse, error)
}

type instanceContextKey struct{}
type instanceEpoch struct{ marker byte }
type instanceBinding struct {
	registry *Registry
	id       string
	epoch    *instanceEpoch
}

// Pin 将后续调用绑定到此刻的实例登记代次，替换或撤销登记使本次调用失效。
func (r *Registry) Pin(ctx context.Context, id string) (context.Context, Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	instance, ok := r.instances[id]
	if !ok {
		return ctx, nil, false
	}
	// 嵌套目录查询沿用入口代次，避免替换后重新绑定到另一个实例。
	if binding, yes := ctx.Value(instanceContextKey{}).(instanceBinding); yes && binding.registry == r && binding.id == id && binding.epoch != r.epochs[id] {
		return ctx, nil, false
	}
	return context.WithValue(ctx, instanceContextKey{}, instanceBinding{r, id, r.epochs[id]}), instance, true
}

// Resolve 读取当前实例，并复核可信入口绑定的登记代次。
func (r *Registry) Resolve(ctx context.Context, id string) (Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	instance, ok := r.instances[id]
	if binding, yes := ctx.Value(instanceContextKey{}).(instanceBinding); yes && binding.registry == r && binding.id == id && binding.epoch != r.epochs[id] {
		return nil, false
	}
	return instance, ok
}
