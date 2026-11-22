package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

type profileViews struct {
	values    map[string]config.Values
	overrides map[string]map[string]json.RawMessage
}

func (p *Plugin) setProfileOverrides(raw string) error {
	var overrides map[string]map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &overrides); err != nil {
		return fmt.Errorf("decode profile overrides: %w", err)
	}
	if overrides == nil {
		return fmt.Errorf("profile overrides must be a JSON object")
	}
	views := &profileViews{overrides: overrides, values: make(map[string]config.Values, len(overrides))}
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

// ProfileOverrides 返回调用插件各配置文件的有效覆盖副本；后续通知只影响下一次读取。
func ProfileOverrides(ctx context.Context) (map[string]map[string]json.RawMessage, error) {
	session := componentFor(ctx)
	if session == nil || session.plugin == nil {
		return nil, errcode.FailedPrecondition.New("plugin session is not ready")
	}
	result := make(map[string]map[string]json.RawMessage)
	if views := session.plugin.profileViews.Load(); views != nil {
		for id, fields := range views.overrides {
			result[id] = make(map[string]json.RawMessage, len(fields))
			for key, value := range fields {
				result[id][key] = append(json.RawMessage(nil), value...)
			}
		}
	}
	return result, nil
}
