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

// adapterGeneratedRegistrations 仅由 Adapter 生成代码的 init 写入，初始化后只读。
var adapterGeneratedRegistrations = make(map[reflect.Type]Generated)

// UseGenerated 把只读构建资源显式绑定到插件对象；重复绑定在 Collect 时报错。
func UseGenerated(p *Plugin, g Generated) {
	p.checkRegistration()
	p.generatedCount++
	if p.generatedCount > 1 {
		p.errs = append(p.errs, fmt.Errorf("generated resources bound more than once"))
	}
	p.generated = g
}

// Generated 返回插件对象绑定的只读构建资源。
func (p *Plugin) Generated() Generated { return p.generated }

// UseGeneratedAdapter 按 Adapter 的具名类型登记生成项，只供生成代码在 init 中调用。
// a 传入具体类型的 nil 指针；同一类型的值和指针共用登记，重复登记会 panic。
func UseGeneratedAdapter(a adapter.Adapter, g Generated) {
	registerGenerated(adapterType(reflect.TypeOf(a)), g)
}

func registerGenerated(component reflect.Type, g Generated) {
	if _, exists := adapterGeneratedRegistrations[component]; exists {
		panic(fmt.Sprintf("generated registration for %s already exists", component))
	}
	adapterGeneratedRegistrations[component] = g
}

func adapterType(component reflect.Type) reflect.Type {
	if component != nil && component.Kind() == reflect.Pointer {
		return component.Elem()
	}
	return component
}

// GeneratedFor 按 Adapter 类型读取生成资源；值与指针读取同一登记。
// 插件资源通过 Plugin.Generated 读取。
func GeneratedFor(component any) (Generated, bool) {
	typ := reflect.TypeOf(component)
	generated, ok := adapterGeneratedRegistrations[typ]
	if !ok {
		generated, ok = adapterGeneratedRegistrations[adapterType(typ)]
	}
	return generated, ok
}

func generatedFor(p *Plugin) Generated { return p.generated }
