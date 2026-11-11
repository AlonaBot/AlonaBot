package adapter

import pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"

// ExtensionChecker 按实例所属 Adapter 的声明校验片段。
// 未声明返回 UNSUPPORTED，JSON 或 Schema 校验失败返回 INVALID_ARGUMENT。
type ExtensionChecker interface {
	CheckExtension(name string, version uint32, payloadJSON string) *pb.ErrorInfo
}
