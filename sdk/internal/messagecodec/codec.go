// Package messagecodec 在 SDK 根包与消息构造包之间传递编码失败，保持消息公开接口精简。
//
// Button 保存按钮的 JSON 编码结果；Failure 随片段保留错误，发起请求前由 Error 读取。
package messagecodec

// Button 保存动作、标签、JSON 值与编码错误。
type Button struct {
	ActionID, Label, Value string
	err                    error
}

// NewButton 保留按钮编码的原始错误。
func NewButton(action, label, value string, err error) Button {
	return Button{ActionID: action, Label: label, Value: value, err: err}
}

// ButtonError 返回按钮的编码错误。
func ButtonError(b Button) error { return b.err }

// Failure 通过嵌入将错误随具体片段传递，方法只在本内部包可见。
type Failure struct{ err error }

// NewFailure 创建携带原始编码错误的片段状态。
func NewFailure(err error) Failure { return Failure{err: err} }

func (f Failure) encodingError() error { return f.err }

// Error 读取消息构造包嵌入的失败状态；普通片段返回 nil。
func Error(segment any) error {
	if s, ok := segment.(interface{ encodingError() error }); ok {
		return s.encodingError()
	}
	return nil
}
