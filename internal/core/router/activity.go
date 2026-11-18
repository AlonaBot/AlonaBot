package router

import (
	"context"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type activityPlugin struct {
	Plugin
	acquire func() (func(), error)
}

// WithActivity 为实际调用和事件写出登记活动，身份、声明和业务结果由原 Plugin 提供。
func WithActivity(p Plugin, acquire func() (release func(), err error)) Plugin {
	return &activityPlugin{Plugin: p, acquire: acquire}
}

func (p *activityPlugin) Call(ctx context.Context, request wire.Request) (wire.Response, error) {
	release, err := p.acquire()
	if err != nil {
		return wire.Response{}, err
	}
	var once sync.Once // 原 Late 与正常返回可能竞争，只归还本次持有的一份引用。
	finish := func() { once.Do(release) }
	late := request.Late
	request.Late = func(response wire.Response) {
		defer finish()
		if late != nil {
			late(response)
		}
	}
	response, err := p.Plugin.Call(ctx, request)
	code := wire.CodeOf(err)
	// RawErrorCode 只由真实远端 RESPONSE 填写；本地结束仍等待 Late 或旧树退出。
	if err == nil || response.RawErrorCode != pb.ErrorCode_ERROR_CODE_UNSPECIFIED ||
		code != pb.ErrorCode_ERROR_CODE_CANCELLED && code != pb.ErrorCode_ERROR_CODE_DEADLINE_EXCEEDED && code != pb.ErrorCode_ERROR_CODE_UNAVAILABLE {
		finish()
	}
	return response, err
}

func (p *activityPlugin) SendEvent(ctx context.Context, target string, event *pb.BusEvent) error {
	release, err := p.acquire()
	if err != nil {
		return err
	}
	ctx, returned := wire.TrackEvent(ctx, release)
	defer returned()
	return p.Plugin.SendEvent(ctx, target, event)
}
