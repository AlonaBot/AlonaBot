package sdk

import (
	"context"

	"github.com/AlonaBot/AlonaBot/contracts/adapterops"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
	"github.com/AlonaBot/AlonaBot/sdk/internal/convert"
)

func (r *adapterRuntime) operations(_ context.Context, in *wire.Inbound) wire.Response {
	request := &pb.AdapterOperationsRequest{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	r.mu.Lock()
	entry, safe := r.operationEntry(request.InstanceId, request.InstanceGeneration)
	if safe != nil {
		r.mu.Unlock()
		return wire.Response{Err: safe}
	}
	instance := entry.instance
	core := entry.core
	r.mu.Unlock()
	response := &pb.AdapterOperationsResponse{}
	for _, op := range adapterops.All() {
		supported := false
		switch op.RPC {
		case wire.ServiceAdapterSend:
			supported = true
		case wire.ServiceAdapterEdit:
			_, supported = instance.(adapter.Editor)
		case wire.ServiceAdapterDelete:
			_, supported = instance.(adapter.Deleter)
		case wire.ServiceAdapterMediaRead:
			_, supported = instance.(adapter.MediaSource)
		case wire.ServiceAdapterReadNodes:
			_, supported = instance.(adapter.NodesReader)
		case wire.ServiceAdapterCommandsSync:
			_, supported = instance.(adapter.CommandSynchronizer)
		case wire.ServiceAdapterCommandMenu:
			_, supported = instance.(adapter.CommandMenuController)
		}
		if op.Capability != "" {
			declared := false
			for _, c := range instance.Capabilities() {
				if c.Name == op.Capability && c.Version == 1 && c.Supported {
					declared = true
				}
			}
			supported = supported && declared
		}
		if supported {
			response.Operations = append(response.Operations, op.Declaration())
		}
	}
	if _, ok := instance.(adapter.ActionExecutor); ok {
		d := r.adapter.Descriptor()
		descriptor, err := convert.DescriptorToProto(&convert.Descriptor{AdapterID: d.AdapterID, Actions: d.Actions})
		if err != nil {
			return dispatchError(pb.ErrorCode_ERROR_CODE_INTERNAL, err.Error())
		}
		response.Operations = append(response.Operations, descriptor.Actions...)
	}
	r.mu.Lock()
	current := r.instances[request.InstanceId]
	available := current == entry && !entry.busy && entry.core == core && !entry.core.stopped.Load()
	r.mu.Unlock()
	if !available {
		return dispatchError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "adapter instance changed during discovery")
	}
	return adapterResponse(response, nil)
}
