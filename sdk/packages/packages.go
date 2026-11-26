// Package packages 查询已安装包并从 GitHub 安装组件。
//
// List 与 Install 接收活跃 Handler 的 Context，Core 按父事件检查管理员身份。
package packages

import (
	"context"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

// Info 保存包的安装身份与公开状态。
type Info struct {
	ID, Name, Version, Role, Source string
	Enabled                         bool
}

// List 按 ID 字典序返回已安装包，要求父事件的 Actor 为管理员。
func List(ctx context.Context) ([]Info, error) {
	if _, active := registry.Parent(ctx); !active {
		return nil, errcode.FailedPrecondition.New("package list requires an active handler context")
	}
	result, err := registry.Proto[*pb.PackageListRequest, *pb.PackageListResponse](wire.ServicePackageList).Call(ctx, &pb.PackageListRequest{})
	if err != nil {
		return nil, err
	}
	entries := make([]Info, 0, len(result.GetPackages()))
	for _, info := range result.GetPackages() {
		entries = append(entries, fromProto(info))
	}
	return entries, nil
}

// Install 从 owner/repo 安装组件，管理员调用同意首次来源信任与同类替换。
// 返回安装信息、受影响对象与完整错误；部分结果随错误保留。
func Install(ctx context.Context, source string) (Info, []string, error) {
	if _, active := registry.Parent(ctx); !active {
		return Info{}, nil, errcode.FailedPrecondition.New("package install requires an active handler context")
	}
	result, err := registry.Proto[*pb.PackageInstallRequest, *pb.PackageInstallResponse](wire.ServicePackageInstall).Call(ctx, &pb.PackageInstallRequest{Source: source})
	return fromProto(result.GetPackage()), result.GetAffected(), err
}

func fromProto(info *pb.PackageInfo) Info {
	return Info{ID: info.GetId(), Name: info.GetName(), Version: info.GetVersion(), Role: info.GetRole(), Enabled: info.GetEnabled(), Source: info.GetSource()}
}
