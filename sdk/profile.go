package sdk

import (
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/config"
)

type profileViews struct {
	values map[string]config.Values
}

func (p *Plugin) setProfileOverrides(raw string) error {
	var overrides map[string]map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &overrides); err != nil {
		return fmt.Errorf("decode profile overrides: %w", err)
	}
	if overrides == nil {
		return fmt.Errorf("profile overrides must be a JSON object")
	}
	views := &profileViews{values: make(map[string]config.Values, len(overrides))}
	for id, fields := range overrides {
		if id == "" || id == "default" {
			return fmt.Errorf("profile %q cannot have overrides", id)
		}
		if fields == nil {
			return fmt.Errorf("profile %q overrides must be a JSON object", id)
		}
		values, err := p.Config().WithOverrides(fields)
		if err != nil {
			return fmt.Errorf("profile %q: %w", id, err)
		}
		views.values[id] = values
	}
	// 全部视图构造成功后一次发布，读取方始终只观察到一个完整集合。
	p.profileViews.Store(views)
	return nil
}

func (p *Plugin) changeProfileOverrides(in *wire.Inbound) wire.Response {
	request := &pb.ProfileSetOverridesRequest{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if err := p.setProfileOverrides(request.GetProfileOverridesJson()); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	return adapterResponse(&pb.ProfileSetOverridesResponse{}, nil)
}

func (c *Context) freezeProfile(invocation *pb.InvocationContext) {
	c.profileID = invocation.GetProfileId()
	c.config = c.plugin.config
	// 只在建立 Context 时读取一次；后续更新不会改变本次调用。
	if views := c.plugin.profileViews.Load(); views != nil {
		if values, ok := views.values[c.profileID]; ok {
			c.config = values
		}
	}
}
