package tspec

import (
	"go/ast"
	"go/doc"
	"strings"

	"github.com/pkg/errors"
)

// ParsePkgWithDecorator parses package and returns all types with the gaven decorator
func ParsePkgWithDecorator(pkg *ast.Package, decorator string) (objs map[string]*ast.Object, err error) {
	docPackage := doc.New(pkg, "", doc.AllDecls)
	objs = make(map[string]*ast.Object)
	for _, t := range docPackage.Types {
		hasDecorator := false
		docs := strings.Split(t.Doc, "\n")
		for _, line := range docs {
			line = strings.TrimSpace(line)
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == decorator {
				hasDecorator = true
				break
			}
		}
		if !hasDecorator {
			continue
		}

		var spec *ast.TypeSpec
		for _, si := range t.Decl.Specs {
			s, ok := si.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if s.Name.Name == t.Name {
				spec = s
			}
		}
		if spec == nil {
			err = errors.Errorf("invalid type %s", t.Name)
			return
		}

		objs[t.Name] = spec.Name.Obj
	}
	return
}
