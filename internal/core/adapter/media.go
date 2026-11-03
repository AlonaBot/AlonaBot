package adapter

import "context"

// MediaReader 提供实例可选的平台文件读取能力，length 为 1～1 MiB。
type MediaReader interface {
	ReadMedia(ctx context.Context, fileID string, offset uint64, length uint32) (data []byte, next uint64, eof bool, err error)
}

// MediaSession 提供实例所属 Adapter Session 的身份，供 Core 绑定出站只读令牌。
type MediaSession interface {
	MediaSessionKey() string
}
