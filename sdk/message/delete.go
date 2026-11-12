package message

import "github.com/AlonaBot/AlonaBot/sdk/errcode"

// DeleteItemResult 保存一个输入消息 ID 的撤回状态及安全错误。
type DeleteItemResult struct {
	PlatformMessageID string
	Status            PartStatus
	Error             *errcode.Error
}

// DeleteResult 按输入顺序保存撤回结果，重复 ID 也各自占一项。
type DeleteResult struct {
	Session           Session
	AdapterInstanceID string
	Items             []DeleteItemResult
}
