package generate

import (
	"fmt"
	"go/ast"
	"go/types"
	"log/slog"
	"path/filepath"
)

// handlerParamNames 复用已检查的构建视图，仅记录插件及其本模块依赖中的 Handler。
func (s *scan) handlerParamNames(info *types.Info, paths []string) {
	if s.adapterProject {
		return
	}
	conflicts := make(map[string]bool)
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
						typ := types.Unalias(receiver.Type())
						pointer, isPointer := typ.(*types.Pointer)
						if isPointer {
							typ = types.Unalias(pointer.Elem())
						}
						name := typ.(*types.Named).Obj().Name()
						if isPointer {
							name = "(*" + name + ")"
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
					if name := params.At(i + 1).Name(); name != "_" {
						names[i] = name
					}
				}
				if s.paramNames == nil {
					s.paramNames = make(map[string][]string)
				}
				if conflicts[key] {
					return true
				}
				if _, exists := s.paramNames[key]; exists {
					// 文件与行号无法区分这些 Handler；整组省略，避免取错名称。
					delete(s.paramNames, key)
					conflicts[key] = true
					slog.Warn("ambiguous handler parameter names; using argN, specify Args for Optional names", "position", s.position(node.Pos()), "key", key)
					return true
				}
				s.paramNames[key] = names
				return true
			})
		}
	}
}
