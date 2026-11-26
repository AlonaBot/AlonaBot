package manage

import (
	"context"
	"errors"
	"slices"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"google.golang.org/protobuf/proto"
)

// RegisterPackages 将管理员包查询与安装服务接入固定服务分流器。
func (s *Service) RegisterPackages(mux *router.Mux) {
	mux.Handle(wire.ServicePackageList, s.callPackage)
	mux.Handle(wire.ServicePackageInstall, s.callPackage)
}

func (s *Service) callPackage(ctx context.Context, _ router.Caller, in *wire.Inbound) wire.Response {
	// 1. Session 校验父调用仍在途，管理员身份只取自 Core 保存的父事件。
	parent, ok := in.ParentAttach.(*router.Invocation)
	if in.ParentCallID == 0 || !ok || parent == nil {
		return packageFailure(precondition("package service requires an active invocation"))
	}
	admin, err := s.opt.Policy.IsAdmin(ctx, parent.Event.GetAdapterId(), parent.Event.GetActor())
	if err != nil {
		return packageFailure(err)
	}
	if !admin {
		return packageFailure(denied("package service requires Core Admin"))
	}
	var request proto.Message
	switch in.Target {
	case wire.ServicePackageList:
		request = &pb.PackageListRequest{}
	case wire.ServicePackageInstall:
		request = &pb.PackageInstallRequest{}
	}
	if in.Payload.Encoding != pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF || in.Payload.SchemaID != validate.SchemaID(request) || in.Payload.SchemaVersion != 1 {
		return packageFailure(invalid("invalid package service payload metadata"))
	}
	if err := proto.Unmarshal(in.Payload.Bytes, request); err != nil {
		return packageFailure(invalid("decode package service request: " + err.Error()))
	}

	// 2. 查询保留安装记录与 manifest 原值；安装复用管理流程及其完整错误。
	switch request := request.(type) {
	case *pb.PackageListRequest:
		installs, err := s.opt.Store.Installs(ctx)
		if err != nil {
			return packageFailure(err)
		}
		result := &pb.PackageListResponse{}
		for _, in := range installs {
			if in.Uninstalled {
				continue
			}
			mutex := s.packageLock(in.PluginID)
			if !mutex.TryLock() {
				return packageFailure(precondition("component management is in progress: " + in.PluginID))
			}
			m, err := s.manifest(ctx, in.PluginID)
			if err != nil {
				mutex.Unlock()
				return packageFailure(err)
			}
			enabled, err := s.opt.Store.Enabled(ctx, in.PluginID)
			mutex.Unlock()
			if err != nil {
				return packageFailure(err)
			}
			result.Packages = append(result.Packages, &pb.PackageInfo{Id: in.PluginID, Name: m.Name, Version: in.Version, Role: string(in.Role), Enabled: enabled, Source: in.Source})
		}
		slices.SortFunc(result.Packages, func(a, b *pb.PackageInfo) int { return strings.Compare(a.GetId(), b.GetId()) })
		return packageResponse(result)
	case *pb.PackageInstallRequest:
		if err := checkRepository(request.GetSource()); err != nil {
			return packageFailure(err)
		}
		installed, err := s.Install(ctx, InstallRequest{Source: request.GetSource(), Confirm: true, Replace: true, invocation: true})
		result := &pb.PackageInstallResponse{Affected: installed.Affected}
		if installed.ID != "" {
			result.Package = &pb.PackageInfo{Id: installed.ID, Name: installed.Name, Version: installed.Version, Role: installed.Role, Enabled: installed.Enabled, Source: installed.Source}
		}
		var confirmation *ErrNeedsConfirm
		if errors.As(err, &confirmation) {
			result.Affected = confirmation.Affected
		}
		response := packageResponse(result)
		if err != nil && response.Err == nil {
			response.Err = packageFailure(err).Err
		}
		return response
	}
	panic("unregistered package service")
}

func packageFailure(err error) wire.Response {
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	var confirmation *ErrNeedsConfirm
	var detail *validate.Error
	if errors.As(err, &confirmation) {
		code = pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION
	} else if errors.As(err, &detail) {
		code = detail.Code
	}
	return wire.Response{Err: &pb.ErrorInfo{Code: code, SafeMessage: err.Error()}}
}

func packageResponse(message proto.Message) wire.Response {
	encoded, err := proto.Marshal(message)
	if err != nil {
		return packageFailure(err)
	}
	return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_PROTOBUF, SchemaID: validate.SchemaID(message), SchemaVersion: 1, Bytes: encoded}}
}
