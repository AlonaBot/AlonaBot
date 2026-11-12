package generate

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
)

// handlerParamNames 复用已检查的构建视图，仅记录插件及其本模块依赖中的 Handler。
func (s *scan) handlerParamNames(info *types.Info, paths []string) {
	if s.adapterProject {
		return
	}
	for _, path := range paths {
		if path == sdkPath {
			continue
		}
		for _, file := range s.syntax[path] {
			ast.Inspect(file, func(node ast.Node) bool {
				var signature *types.Signature
				var key string
				switch fn := node.(type) {
				case *ast.FuncDecl:
					object := info.Defs[fn.Name].(*types.Func)
					signature = object.Type().(*types.Signature)
					prefix := path
					if object.Pkg().Name() == "main" {
						prefix = "main"
					}
					key = prefix + "." + object.Name()
					if receiver := signature.Recv(); receiver != nil {
						name := types.TypeString(receiver.Type(), func(*types.Package) string { return "" })
						if _, pointer := types.Unalias(receiver.Type()).(*types.Pointer); pointer {
							name = "(" + name + ")"
						}
						key = prefix + "." + name + "." + object.Name()
					}
				case *ast.FuncLit:
					signature = info.TypeOf(fn).(*types.Signature)
					prefix := path
					if file.Name.Name == "main" {
						prefix = "main"
					}
					pos := s.files.PositionFor(fn.Type.Func, false)
					key = fmt.Sprintf("%s:%s:%d", prefix, filepath.Base(pos.Filename), pos.Line)
				default:
					return true
				}
				params := signature.Params()
				if params.Len() == 0 {
					return true
				}
				pointer, ok := types.Unalias(params.At(0).Type()).(*types.Pointer)
				if !ok {
					return true
				}
				named, ok := types.Unalias(pointer.Elem()).(*types.Named)
				if !ok || named.Obj().Pkg() == nil || named.Obj().Pkg().Path() != sdkPath || named.Obj().Name() != "Context" {
					return true
				}
				names := make([]string, params.Len()-1)
				for i := range names {
					names[i] = params.At(i + 1).Name()
				}
				if s.paramNames == nil {
					s.paramNames = make(map[string][]string)
				}
				s.paramNames[key] = names
				return true
			})
		}
	}
}
