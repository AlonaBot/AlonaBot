package message

import "time"

// Readability 指定入站媒体是否可读取。
type Readability int

const (
	Readable   Readability = iota + 1 // 可读取。
	Unreadable                        // 平台不支持读取。
)

// IncomingMedia 保存尚未取件的入站对象授权，Size 为零表示未知。
type IncomingMedia struct {
	Token       string
	Size        uint64
	MIME        string
	Readability Readability
	ExpiresAt   time.Time
}

// MediaSource 指定恰好一个媒体来源。
type MediaSource struct {
	MediaRef     string
	Incoming     *IncomingMedia
	HTTPSURL     string
	InlineBase64 string
	PlatformFile string // 只供 Adapter 读取 Core 准备的来源，插件发送时拒绝
}

// MediaRef 保存已校验的媒体引用及消费范围。
type MediaRef struct {
	Token             string
	Size              uint64
	MIME              string
	SHA256            []byte
	ExpiresAt         time.Time
	ConsumerPluginIDs []string
	Sessions          []string
}

// Source 返回此已上传媒体的来源。
func (m *MediaRef) Source() MediaSource { return MediaSource{MediaRef: m.Token} }

// Source 返回此入站对象的来源。
func (m *IncomingMedia) Source() MediaSource { return MediaSource{Incoming: m} }
