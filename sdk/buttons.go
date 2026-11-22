package sdk

const buttonRoutePrefix = "button."

// buttonBuilder 将私有事件路由交给类型化按钮登记器。
func (r *pluginScope) buttonBuilder(name string, version uint32) (any, func(func(string) (any, error), func(*Context, any) error)) {
	b := r.on(buttonRoutePrefix+name, ButtonClicked)
	b.route.actionID, b.route.actionVersion = name, version
	b.button = true
	return b, func(decode func(string) (any, error), handle func(*Context, any) error) {
		b.plugin.checkRegistration()
		b.route.decodeButton = decode
		if handle != nil {
			b.route.handleButton = handle
		}
	}
}
