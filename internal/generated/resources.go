// Package generated 保存组件入口包的只读构建资源与仓库测试的对象绑定。
//
// CLI 注入和 plugintest 扫描共用登记表；所有登记与查询均并发安全。
// Resources 中的切片与映射由登记方发布后保持只读。
package generated

import (
	"fmt"
	"reflect"
	"sync"
)

// Resources 保存组件声明及扫描得到的资源。
type Resources struct {
	Manifest            []byte              // 入口目录 plugin.yaml 的原文。
	ConfigSchema        []byte              // 规范 JSON；没有声明时为 nil。
	AdapterConfigSchema []byte              // Adapter 实例的规范 JSON Schema；插件为 nil。
	ServiceSchemas      map[string][]byte   // schema_id 对应的规范 JSON Service Schema。
	I18n                map[string][]byte   // i18n 文件名对应的原文。
	ParamNames          map[string][]string // Handler 全名或包路径:文件名:行号对应的参数名。
}

var registry = struct {
	sync.RWMutex
	packages map[string]Resources
	objects  map[any]Resources
	entries  map[any]string
}{packages: make(map[string]Resources), objects: make(map[any]Resources), entries: make(map[any]string)}

// Register 按入口包登记资源；重复登记同一导入路径会 panic。
func Register(pkgPath string, r Resources) {
	registry.Lock()
	defer registry.Unlock()
	if _, ok := registry.packages[pkgPath]; ok {
		panic(fmt.Sprintf("resources for entry package %q already registered", pkgPath))
	}
	registry.packages[pkgPath] = r
}

// Bind 按对象绑定仓库测试夹具与 plugintest 测试 Adapter 的资源。
func Bind(component any, r Resources) {
	registry.Lock()
	defer registry.Unlock()
	registry.objects[component] = r
}

// Lookup 先查对象绑定，再查入口包的资源。
func Lookup(component any, pkgPath string) (Resources, bool) {
	registry.RLock()
	defer registry.RUnlock()
	if component == nil || reflect.TypeOf(component).Comparable() {
		if r, ok := registry.objects[component]; ok {
			return r, true
		}
	}
	r, ok := registry.packages[pkgPath]
	return r, ok
}

// SetEntry 记录插件对象的入口包。
func SetEntry(p any, pkgPath string) {
	registry.Lock()
	defer registry.Unlock()
	registry.entries[p] = pkgPath
}

// Entry 返回插件对象在构造时记录的入口包。
func Entry(p any) (pkgPath string, ok bool) {
	registry.RLock()
	defer registry.RUnlock()
	if p != nil && !reflect.TypeOf(p).Comparable() {
		return "", false
	}
	pkgPath, ok = registry.entries[p]
	return
}
