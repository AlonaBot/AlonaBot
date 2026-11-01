package sdk

import (
	"fmt"
	"reflect"
)

// Generated 提供构建期生成的命令与事件注册函数；nil 字段表示没有对应注册项。
type Generated struct {
	Commands func(*CommandRegistry)
	Events   func(*EventRegistry)
}

// generatedRegistrations 仅在生成代码的 init 中写入，初始化完成后由 Collect 读取。
var generatedRegistrations = make(map[reflect.Type]Generated)

// UseGenerated 按根类型登记生成项，只供生成代码在 init 中调用。
// root 传入根类型的 nil 指针；同一根类型重复登记会 panic。
func UseGenerated(root Root, g Generated) {
	rootType := reflect.TypeOf(root)
	if _, exists := generatedRegistrations[rootType]; exists {
		panic(fmt.Sprintf("generated registration for %s already exists", rootType))
	}
	generatedRegistrations[rootType] = g
}

func generatedFor(root Root) Generated {
	return generatedRegistrations[reflect.TypeOf(root)]
}

func (g Generated) registerCommands(r *CommandRegistry) {
	if g.Commands != nil {
		g.Commands(r)
	}
}

func (g Generated) registerEvents(r *EventRegistry) {
	if g.Events != nil {
		g.Events(r)
	}
}
