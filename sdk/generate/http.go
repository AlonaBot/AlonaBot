package generate

import (
	"go/ast"
	"go/types"
)

// httpCalls 保守收集工程包内的 SDK HTTP 登记调用，不做可达性分析。
// shortcut: 经函数值或反射间接登记的路由不识别，需要覆盖这些写法时扩展静态分析。
func (s *scan) httpCalls(info *types.Info, paths []string) {
	for _, path := range paths {
		for _, file := range s.syntax[path] {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
				if !ok {
					return true
				}
				fn, ok := info.Uses[selector.Sel].(*types.Func)
				if !ok || fn.Pkg() == nil || fn.Pkg().Path() != sdkPath || fn.Name() != "Route" {
					return true
				}
				receiver := fn.Type().(*types.Signature).Recv().Type().(*types.Pointer).Elem().(*types.Named)
				if receiver.Obj().Name() == "Plugin" || receiver.Obj().Name() == "SubPlugin" {
					s.result.HTTP = true
				}
				return true
			})
		}
	}
}
