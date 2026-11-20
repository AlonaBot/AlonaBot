package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path"
	"slices"
	"strconv"
)

// nativeHarness 只保留场景入口实际引用的声明，排除 Unix 测试及其平台 API。
// 原有断言按完整 AST 声明复用，native.go 中的同名声明拥有平台资源的替换权。
func nativeHarness(sources map[string]string, roots []string) ([]byte, error) {
	files := token.NewFileSet()
	type declaration struct {
		node    ast.Decl
		imports map[string]*ast.ImportSpec
	}
	var declarations []*declaration
	byName := make(map[string][]*declaration)
	globals := make(map[*ast.Object]bool)
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	slices.Sort(names)
	// native.go 最后登记，让原生实现替换同名的 Unix 所有者。
	names = append(slices.DeleteFunc(names, func(name string) bool { return name == "native.go" }), "native.go")
	for _, name := range names {
		file, err := parser.ParseFile(files, name, sources[name], 0)
		if err != nil {
			return nil, err
		}
		for _, object := range file.Scope.Objects {
			globals[object] = true
		}
		imports := make(map[string]*ast.ImportSpec)
		for _, spec := range file.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return nil, err
			}
			alias := path.Base(value)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			imports[alias] = spec
		}
		for _, node := range file.Decls {
			if gen, ok := node.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				continue
			}
			decl := &declaration{node: node, imports: imports}
			declarations = append(declarations, decl)
			var identifiers []string
			switch node := node.(type) {
			case *ast.FuncDecl:
				identifiers = append(identifiers, node.Name.Name)
			case *ast.GenDecl:
				for _, spec := range node.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						identifiers = append(identifiers, spec.Name.Name)
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							identifiers = append(identifiers, name.Name)
						}
					}
				}
			}
			for _, identifier := range identifiers {
				if name == "native.go" {
					byName[identifier] = nil
				}
				byName[identifier] = append(byName[identifier], decl)
			}
		}
	}
	// 依赖闭包保留声明的完整函数体，导入按实际使用的包标识符选择。
	selected := make(map[*declaration]bool)
	imports := make(map[string]*ast.ImportSpec)
	var selectName func(string)
	selectName = func(name string) {
		for _, decl := range byName[name] {
			if selected[decl] {
				continue
			}
			selected[decl] = true
			ast.Inspect(decl.node, func(node ast.Node) bool {
				if selector, ok := node.(*ast.SelectorExpr); ok {
					if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Obj == nil {
						if spec := decl.imports[identifier.Name]; spec != nil {
							imports[spec.Path.Value] = spec
						}
					}
				}
				if identifier, ok := node.(*ast.Ident); ok && (identifier.Obj == nil || globals[identifier.Obj]) {
					selectName(identifier.Name)
				}
				return true
			})
		}
	}
	for _, root := range roots {
		if len(byName[root]) == 0 {
			return nil, fmt.Errorf("missing native scenario declaration %s", root)
		}
		selectName(root)
	}
	var output bytes.Buffer
	output.WriteString("package main\nimport (\n")
	paths := make([]string, 0, len(imports))
	for name := range imports {
		paths = append(paths, name)
	}
	slices.Sort(paths)
	for _, name := range paths {
		if err := format.Node(&output, files, imports[name]); err != nil {
			return nil, err
		}
		output.WriteByte('\n')
	}
	output.WriteString(")\n")
	for _, decl := range declarations {
		if !selected[decl] {
			continue
		}
		if err := format.Node(&output, files, decl.node); err != nil {
			return nil, err
		}
		output.WriteByte('\n')
	}
	return format.Source(output.Bytes())
}
