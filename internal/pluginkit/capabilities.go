package pluginkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// requiredCapabilities 读取实际链接 SDK 的声明，避免 CLI 自身版本替目标 SDK 声明能力。
// 这里只检查声明和常量，不运行目标程序，也不为读取私有 sdkOffer 增加 SDK API。
func requiredCapabilities(ctx context.Context, options generate.Options, overlay, modFile string, development, http bool) ([]manifest.Capability, error) {
	listed, err := targetGoCommand(ctx, options, development, "list", "-modfile", modFile, "-mod=mod", "-overlay", overlay, "-deps", "-export", "-json", sdkPath)
	if err != nil {
		return nil, err
	}
	exports := make(map[string]string)
	var files []*ast.File
	fset := token.NewFileSet()
	decoder := json.NewDecoder(bytes.NewReader(listed))
	for {
		var pkg struct {
			Dir, ImportPath, Export string
			GoFiles                 []string
		}
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode SDK capability imports: %w", err)
		}
		exports[pkg.ImportPath] = pkg.Export
		if pkg.ImportPath != sdkPath {
			continue
		}
		for _, name := range pkg.GoFiles {
			file, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, 0)
			if err != nil {
				return nil, err
			}
			files = append(files, file)
		}
	}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	config := types.Config{IgnoreFuncBodies: true, Sizes: types.SizesFor("gc", options.GOARCH), Importer: importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		return os.Open(exports[path])
	})}
	sdk, err := config.Check(sdkPath, fset, files, info)
	if err != nil {
		return nil, fmt.Errorf("check SDK capability declarations: %w", err)
	}
	var capabilities []manifest.Capability
	for _, file := range files {
		for _, declaration := range file.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.VAR {
				continue
			}
			for _, spec := range group.Specs {
				value := spec.(*ast.ValueSpec)
				for index, name := range value.Names {
					if name.Name != "sdkOffer" {
						continue
					}
					capabilities, err = offerCapabilities(value.Values[index], info)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if capabilities == nil {
		return nil, fmt.Errorf("SDK sdkOffer declaration was not found")
	}
	if http {
		version := sdk.Scope().Lookup("httpCapabilityVersion").(*types.Const).Val()
		for _, dependency := range sdk.Imports() {
			if dependency.Path() == modulePath+"/internal/wire" {
				name := dependency.Scope().Lookup("CapabilityHTTPListener").(*types.Const).Val()
				number, _ := constant.Uint64Val(version)
				capabilities = append(capabilities, manifest.Capability{Name: constant.StringVal(name), Version: uint32(number)})
				return capabilities, nil
			}
		}
		return nil, fmt.Errorf("SDK HTTP capability declaration was not found")
	}
	return capabilities, nil
}

func offerCapabilities(offer ast.Expr, info *types.Info) ([]manifest.Capability, error) {
	literal, ok := offer.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("SDK sdkOffer must be a static offer literal")
	}
	capabilities := []manifest.Capability{}
	for _, element := range literal.Elts {
		field := element.(*ast.KeyValueExpr)
		if field.Key.(*ast.Ident).Name != "Required" {
			continue
		}
		list, ok := field.Value.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("SDK sdkOffer.Required must be a static capability list")
		}
		for _, item := range list.Elts {
			if pointer, ok := item.(*ast.UnaryExpr); ok {
				item = pointer.X
			}
			entry, ok := item.(*ast.CompositeLit)
			if !ok {
				return nil, fmt.Errorf("SDK required capability must be a static literal")
			}
			var capability manifest.Capability
			for _, element := range entry.Elts {
				field := element.(*ast.KeyValueExpr)
				value := info.Types[field.Value].Value
				if value == nil {
					return nil, fmt.Errorf("SDK required capability %s must be constant", field.Key)
				}
				switch field.Key.(*ast.Ident).Name {
				case "Name":
					capability.Name = constant.StringVal(value)
				case "Version":
					number, _ := constant.Uint64Val(value)
					capability.Version = uint32(number)
				}
			}
			capabilities = append(capabilities, capability)
		}
	}
	return capabilities, nil
}
