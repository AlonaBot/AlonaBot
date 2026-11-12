package sdk

import (
	"fmt"
	"reflect"

	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

// Generated 提供构建期嵌入的工程身份、Schema 与文案原文；nil 字段表示没有对应项。
type Generated struct {
	Manifest            []byte            // 工程入口包内 plugin.yaml 的原文。
	AdapterConfigSchema []byte            // Adapter 实例的规范 JSON Schema；插件为 nil。
	ServiceSchemas      map[string][]byte // schema_id 对应的规范 JSON Service Schema。
	I18n                map[string][]byte // i18n 文件名对应的原文；无文案资源时为 nil。
	ConfigSchema        []byte            // config_schema.json 的规范 JSON；没有该文件时为 nil。
}

// generatedRegistrations 仅在生成代码的 init 中写入，初始化完成后由 Collect 和准备阶段读取。
var generatedRegistrations = make(map[reflect.Type]Generated)

// UseGenerated 按根类型登记生成项，只供生成代码在 init 中调用。
// root 传入根类型的 nil 指针；同一根类型重复登记会 panic。
func UseGenerated(root Root, g Generated) {
	registerGenerated(reflect.TypeOf(root), g)
}

// UseGeneratedAdapter 按 Adapter 的具名类型登记生成项，只供生成代码在 init 中调用。
// a 传入具体类型的 nil 指针；同一类型的值和指针共用登记，重复登记会 panic。
func UseGeneratedAdapter(a adapter.Adapter, g Generated) {
	registerGenerated(adapterType(reflect.TypeOf(a)), g)
}

func registerGenerated(component reflect.Type, g Generated) {
	if _, exists := generatedRegistrations[component]; exists {
		panic(fmt.Sprintf("generated registration for %s already exists", component))
	}
	generatedRegistrations[component] = g
}

func adapterType(component reflect.Type) reflect.Type {
	if component != nil && component.Kind() == reflect.Pointer {
		return component.Elem()
	}
	return component
}

// GeneratedFor 按组件类型读取生成资源；bool 表示该类型是否已登记。
// 插件沿用根类型键；Adapter 的值与指针读取同一登记。
func GeneratedFor(component any) (Generated, bool) {
	typ := reflect.TypeOf(component)
	generated, ok := generatedRegistrations[typ]
	if !ok {
		generated, ok = generatedRegistrations[adapterType(typ)]
	}
	return generated, ok
}

func generatedFor(root Root) Generated {
	generated, _ := GeneratedFor(root)
	return generated
}
