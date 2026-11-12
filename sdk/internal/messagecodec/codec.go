// Package messagecodec 在 SDK 根包与消息构造包之间传递消息的私有编码状态。
//
// Button 保存按钮的 JSON 编码结果；Failure 保留编码错误；ReferenceOrigin 保留引用目标的省略信息。
package messagecodec

// Button 保存动作、标签、JSON 值与编码错误。
type Button struct {
	ActionID, Label, Value, CoreToken string
	err                               error
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

// ReferenceOrigin 保留入站引用省略目标时补全的句柄，以恢复原始可选字段。
type ReferenceOrigin struct{ ref string }

func (r *ReferenceOrigin) setImplicitReference(ref string) { r.ref = ref }
func (r ReferenceOrigin) implicitReference() string        { return r.ref }

// SetImplicitReference 将补全的句柄随引用片段保存。
func SetImplicitReference(segment interface{ setImplicitReference(string) }, ref string) {
	segment.setImplicitReference(ref)
}

// ImplicitReference 返回入站省略目标时补全的句柄，普通引用返回空串。
func ImplicitReference(segment interface{ implicitReference() string }) string {
	return segment.implicitReference()
}
