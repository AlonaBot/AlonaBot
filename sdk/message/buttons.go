package message

import (
	"encoding/json"

	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
)

type encodingFailure = messagecodec.Failure

// Button 保存动作或 URL、展示标签、JSON 值与 Core 签发的回调令牌。
type Button = messagecodec.Button

// ButtonsSegment 按行保存按钮及构造时的编码错误。
type ButtonsSegment struct {
	encodingFailure
	Rows [][]Button
}

// Buttons 构造按钮键盘，编码失败随片段保留到发送入口。
func Buttons(rows ...[]Button) Segment {
	var err error
	for _, row := range rows {
		for _, button := range row {
			if err == nil {
				err = messagecodec.ButtonError(button)
			}
		}
	}
	return ButtonsSegment{encodingFailure: messagecodec.NewFailure(err), Rows: rows}
}

// ExtensionSegment 保存 Adapter 声明的版本化 JSON 片段及编码错误。
type ExtensionSegment struct {
	encodingFailure
	AdapterID, Name string
	Version         uint32
	PayloadJSON     string
}

// Extension 按 Adapter 的声明编码 payload，保留原始编码错误。
func Extension(adapterID, name string, version uint32, payload any) Segment {
	encoded, err := json.Marshal(payload)
	return ExtensionSegment{encodingFailure: messagecodec.NewFailure(err), AdapterID: adapterID, Name: name, Version: version, PayloadJSON: string(encoded)}
}

// URLButton 构造无需回调令牌的链接按钮。
func URLButton(label, url string) Button { return Button{Label: label, URL: url} }
