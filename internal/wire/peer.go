// 本文件负责核对握手认证之后对端帧的会话身份与协商版本；Core 准备阶段、SDK 与 Session 共用。
package wire

import (
	"fmt"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// CheckPeer 核对认证后收到的帧（WIRE §1）：session_id 等于本 Session、protocol_minor 等于协商值、
// source_plugin_id 在插件发出的帧上等于其认证 ID，在 Core 发出的帧上为空。peerPluginID 为空表示对端是 Core。
// 不符时返回 PROTOCOL_ERROR，调用方应使当前 Session 失败。
func CheckPeer(env *pb.Envelope, sid []byte, minor uint32, peerPluginID string) error {
	switch {
	case string(env.GetSessionId()) != string(sid):
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR, Msg: "帧的 session_id 与本 Session 不符"}
	case env.GetProtocolMinor() != minor:
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR,
			Msg: fmt.Sprintf("帧的 protocol_minor=%d，与协商值 %d 不符", env.GetProtocolMinor(), minor)}
	case env.GetSourcePluginId() != peerPluginID:
		return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_PROTOCOL_ERROR,
			Msg: fmt.Sprintf("帧的 source_plugin_id=%q，对端认证身份为 %q", env.GetSourcePluginId(), peerPluginID)}
	}
	return nil
}
