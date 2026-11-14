package generate

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

const projectFileName = "plugin.yaml"
const adapterRole = "adapter"
const adapterPath = sdkPath + "/adapter"

func (s *scan) readProject() error {
	raw, err := s.resourceFile(projectFileName)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read project manifest: %w", err)
	}
	project, err := manifest.ParseProject(raw)
	if err != nil {
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{projectFileName + ":1:1", err.Error()})
		return nil
	}
	s.projectManifest = raw
	s.adapterProject = project.Manifest.Role == adapterRole
	return nil
}

// checkAdapterType 只查入口包的具体具名类型，值与指针的方法集共享同一项候选。
func (s *scan) checkAdapterType(importer *schemaImporter) error {
	contractPackage, err := importer.Import(adapterPath)
	if err != nil {
		return fmt.Errorf("import Adapter contract: %w", err)
	}
	contract := contractPackage.Scope().Lookup("Adapter").Type().Underlying().(*types.Interface)
	entry, err := importer.Import(s.entry.ImportPath)
	if err != nil {
		return err
	}
	if s.requireAdapterConstructor {
		constructor := entry.Scope().Lookup("New")
		fn, valid := constructor.(*types.Func)
		if valid {
			sig := fn.Signature()
			valid = sig.Recv() == nil && sig.Params().Len() == 0 && !sig.Variadic() && sig.TypeParams().Len() == 0 && sig.Results().Len() == 1
			if valid {
				valid = types.Identical(sig.Results().At(0).Type(), contractPackage.Scope().Lookup("Adapter").Type())
			}
		}
		if !valid {
			position := s.syntax[s.entry.ImportPath][0].Name.Pos()
			if constructor != nil {
				position = constructor.Pos()
			}
			s.diagnostic(position, "entry constructor must have signature func New() adapter.Adapter")
		}
	}
	var candidates []declaration
	for _, name := range entry.Scope().Names() {
		object, ok := entry.Scope().Lookup(name).(*types.TypeName)
		if !ok || object.IsAlias() {
			continue
		}
		named, ok := object.Type().(*types.Named)
		if !ok || named.TypeParams().Len() != 0 {
			continue
		}
		if _, abstract := named.Underlying().(*types.Interface); abstract {
			continue
		}
		if types.Implements(named, contract) || types.Implements(types.NewPointer(named), contract) {
			candidates = append(candidates, declaration{name, object.Pos(), s.entry.ImportPath})
		}
	}
	if len(candidates) != 1 {
		var positions []string
		for _, candidate := range candidates {
			positions = append(positions, s.position(candidate.pos))
		}
		files := s.entry.GoFiles
		if len(files) == 0 {
			files = s.entry.CgoFiles
		}
		position := filepath.ToSlash(files[0]) + ":1:1"
		if len(positions) > 0 {
			position = positions[0]
		}
		message := fmt.Sprintf("expected exactly one concrete Adapter type, found %d", len(candidates))
		if len(positions) > 0 {
			message += " at " + strings.Join(positions, ", ")
		}
		s.result.Diagnostics = append(s.result.Diagnostics, Diagnostic{position, message})
		return nil
	}
	s.adapterRoot = candidates[0]
	s.result.AdapterType = s.adapterRoot.pkg + "." + s.adapterRoot.name
	// 实现 AdapterHTTP 必然引用 SDK 类型；没有这条依赖的 Adapter 无需加载 SDK。
	sdk := importedPackage(entry, sdkPath)
	if sdk == nil {
		return nil
	}
	http := sdk.Scope().Lookup("AdapterHTTP").Type().Underlying().(*types.Interface)
	typ := entry.Scope().Lookup(s.adapterRoot.name).Type()
	s.result.HTTP = types.Implements(typ, http) || types.Implements(types.NewPointer(typ), http)
	return nil
}

func importedPackage(pkg *types.Package, path string) *types.Package {
	seen := make(map[*types.Package]bool)
	for queue := []*types.Package{pkg}; len(queue) > 0; {
		current := queue[0]
		queue = queue[1:]
		if current.Path() == path {
			return current
		}
		if !seen[current] {
			seen[current] = true
			queue = append(queue, current.Imports()...)
		}
	}
	return nil
}
