package dbhook

// ConfigView 随 Values 按值保存解析函数；准备完成前由插件拒绝读取。
type ConfigView[T any] struct{ resolve func() T }

func (s *ConfigView[T]) configView() *ConfigView[T] { return s }

type configOwner[T any] interface{ configView() *ConfigView[T] }

// BindConfig 为视图保存准备后执行的解析函数。
func BindConfig[T any](v configOwner[T], resolve func() T) { v.configView().resolve = resolve }

// ResolveConfig 返回延迟视图的解析结果及其是否已绑定。
func ResolveConfig[T any](v configOwner[T]) (T, bool) {
	if resolve := v.configView().resolve; resolve != nil {
		return resolve(), true
	}
	var zero T
	return zero, false
}
