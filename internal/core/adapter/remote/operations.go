package remote

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Operations 从本次运行所属 Session 查询，响应返回后复核实例身份。
func (i *instance) Operations(ctx context.Context) (*pb.AdapterOperationsResponse, error) {
	p := i.process
	p.mu.Lock()
	running := i.running && p.session != nil && p.session.Session == i.session && p.instances[i.id] == i
	p.mu.Unlock()
	if !running {
		return nil, failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	out := &pb.AdapterOperationsResponse{}
	request, err := payload(&pb.AdapterOperationsRequest{InstanceId: i.id, InstanceGeneration: i.generation})
	if err != nil {
		return nil, err
	}
	response, err := i.session.Call(ctx, wire.Request{Target: wire.ServiceAdapterOperations, Payload: request})
	if err != nil {
		return nil, err
	}
	if err := decode(response.Payload, out); err != nil {
		return nil, err
	}
	p.mu.Lock()
	running = i.running && p.session != nil && p.session.Session == i.session && p.instances[i.id] == i
	p.mu.Unlock()
	if !running {
		return nil, failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance changed during operation discovery")
	}
	return out, nil
}

// Inject 复用所属 Session 的标准入站验证，保留当前 Adapter 游标。
func (i *instance) Inject(ctx context.Context, event *pb.Event) error {
	p := i.process
	p.mu.Lock()
	valid := i.running && p.session != nil && p.session.Session == i.session && p.instances[i.id] == i
	caller := p.caller
	p.mu.Unlock()
	if !valid {
		return failure(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance is unavailable")
	}
	request, err := payload(&pb.AdapterDeliverRequest{InstanceId: i.id, Event: event})
	if err != nil {
		return err
	}
	response := p.serviceRequest(ctx, caller, &wire.Inbound{Target: wire.ServiceCoreAdapterDeliver, Payload: request}, i)
	if response.Err != nil {
		return &wire.CallError{ErrorInfo: response.Err}
	}
	accepted := &pb.AdapterDeliverResponse{}
	if err := decode(response.Payload, accepted); err != nil {
		return err
	}
	if !accepted.Accepted {
		return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "event dispatch queue is full")
	}
	return nil
}
