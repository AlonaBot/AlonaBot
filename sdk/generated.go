package sdk

import (
	"fmt"
	"reflect"
)

// Generated 提供构建期生成的配置与 Service Schema；nil 字段表示没有对应项。
type Generated struct {
	ServiceSchemas map[string][]byte // schema_id 对应的规范 JSON Service Schema。
	ConfigSchema   []byte            // config_schema.json 的规范 JSON；没有该文件时为 nil。
}

// generatedRegistrations 仅在生成代码的 init 中写入，初始化完成后由 Collect 和准备阶段读取。
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
