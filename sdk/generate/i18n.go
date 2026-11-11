package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/i18n"
)

const i18nDirectory = "i18n"
const i18nExtension = ".json"

func (s *scan) readI18n() error {
	directory := filepath.Join(s.dir, i18nDirectory)
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read i18n directory: %w", err)
	}
	files := make(map[string][]byte)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), i18nExtension) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return fmt.Errorf("read i18n file: %w", err)
		}
		files[entry.Name()] = data
	}
	// 存在 i18n 目录时必须提供源语言；Parse 的空输入留给无资源的组件。
	if len(files) == 0 {
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{
			Pos: i18nDirectory + "/en-US.json", Message: "source catalog is missing",
		})
		return nil
	}
	s.i18nFiles = files
	_, problems := i18n.Parse(files)
	for _, problem := range problems {
		position := i18nDirectory + "/" + problem.File
		if problem.Key != "" {
			position += ":" + problem.Key
		}
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{Pos: position, Message: problem.Message})
	}
	return nil
}

// translationCalls 仅检查本 module 扫描闭包中的真实 SDK 调用，复用类型检查结果。
func (s *scan) translationCalls(info *types.Info, paths []string) {
	var source struct {
		Messages map[string]string `json:"messages"`
	}
	if data := s.i18nFiles["en-US.json"]; data != nil {
		if err := json.Unmarshal(data, &source); err != nil {
			s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{Pos: "i18n/en-US.json", Message: err.Error()})
			return
		}
	}
	for _, path := range paths {
		for _, file := range s.syntax[path] {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				keyIndex := 0
				var ident *ast.Ident
				switch function := ast.Unparen(call.Fun).(type) {
				case *ast.Ident:
					ident = function
				case *ast.SelectorExpr:
					ident = function.Sel
					if selection := info.Selections[function]; selection != nil && selection.Kind() == types.MethodExpr {
						keyIndex++ // 方法表达式的第一个实参是接收者。
					}
				default:
					return true
				}
				function, ok := info.Uses[ident].(*types.Func)
				if !ok || function.Pkg() == nil || function.Pkg().Path() != sdkPath || function.Name() != "T" {
					return true
				}
				if receiver := function.Signature().Recv(); receiver == nil {
					keyIndex++ // sdk.T 的第一个实参是 context.Context。
				} else {
					pointer, ok := receiver.Type().(*types.Pointer)
					if !ok {
						return true
					}
					named, ok := pointer.Elem().(*types.Named)
					if !ok || named.Obj().Name() != "Context" {
						return true
					}
				}
				s.translationCall(info, call, keyIndex, source.Messages)
				return true
			})
		}
	}
}

func (s *scan) translationCall(info *types.Info, call *ast.CallExpr, keyIndex int, messages map[string]string) {
	if len(call.Args) == 1 {
		if _, tuple := info.Types[call.Args[0]].Type.(*types.Tuple); tuple {
			return // 多返回值表达式中的键来自运行时，没有独立的常量 AST 实参。
		}
	}
	key, ok := stringConstant(info, call.Args[keyIndex])
	if !ok {
		return
	}
	message, exists := messages[key]
	if !exists {
		s.diagnostic(call.Args[keyIndex].Pos(), fmt.Sprintf("unknown message key %q in en-US", key))
		return
	}
	args := call.Args[keyIndex+1:]
	if call.Ellipsis.IsValid() || len(args)%2 != 0 {
		s.diagnostic(call.Pos(), "translation arguments must contain explicit name/value pairs")
		return
	}
	names := make(map[string]bool)
	for i := 0; i < len(args); i += 2 {
		name, ok := stringConstant(info, args[i])
		if !ok {
			s.diagnostic(args[i].Pos(), "translation argument name must be a string constant")
			return
		}
		typ := info.Types[args[i]].Type
		if !types.Identical(typ, types.Typ[types.String]) && !types.Identical(typ, types.Typ[types.UntypedString]) {
			s.diagnostic(args[i].Pos(), "translation argument name must have type string")
			return
		}
		names[name] = true
	}
	placeholders, err := i18n.Placeholders(message)
	if err != nil {
		s.diagnostic(call.Args[keyIndex].Pos(), err.Error())
		return
	}
	if !slices.Equal(slices.Sorted(maps.Keys(names)), placeholders) {
		s.diagnostic(call.Args[keyIndex].Pos(), fmt.Sprintf("translation argument names must match placeholders %q for message %q", placeholders, key))
	}
}

func stringConstant(info *types.Info, expression ast.Expr) (string, bool) {
	value := info.Types[expression].Value
	if value == nil || value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(value), true
}
