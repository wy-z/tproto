package tspec

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

const (
	// DefaultRefPrefix defines the default value of ref prefix
	DefaultRefPrefix = "#/definitions/"
)

// ParserOptions defines tspec parser options
type ParserOptions struct {
	IgnoreJSONTag bool
	RefPrefix     string
}

// DefaultParserOptions defines default tspec parser options
var DefaultParserOptions = ParserOptions{
	IgnoreJSONTag: false,
	RefPrefix:     DefaultRefPrefix,
}

// Parser defines tspec parser
type Parser struct {
	fset       *token.FileSet
	dirPkgMap  map[string]*ast.Package
	pkgObjsMap map[*ast.Package]map[string]*ast.Object
	typeMap    map[string]*spec.Schema
	opts       ParserOptions
	lock       sync.Mutex
}

// NewParser returns a new tspec parser
func NewParser() (parser *Parser) {
	parser = new(Parser)
	parser.fset = token.NewFileSet()
	parser.dirPkgMap = make(map[string]*ast.Package)
	parser.pkgObjsMap = make(map[*ast.Package]map[string]*ast.Object)
	parser.typeMap = make(map[string]*spec.Schema)
	parser.opts = DefaultParserOptions
	return
}

// Options gets or sets parser options
func (t *Parser) Options(opts ...ParserOptions) ParserOptions {
	if len(opts) != 0 {
		t.opts = opts[0]
	}
	return t.opts
}

// ParseDir parses the dir and cache it
func (t *Parser) ParseDir(dirPath string, pkgName string) (pkg *ast.Package, err error) {
	if tmpPkg, ok := t.dirPkgMap[dirPath]; ok {
		pkg = tmpPkg
		return
	}

	pkgs, err := parser.ParseDir(t.fset, dirPath, nil, parser.ParseComments)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for k := range pkgs {
		if k == pkgName {
			pkg = pkgs[k]
			break
		}
	}
	if pkg == nil {
		err = errors.Errorf("%s not found in %s", pkgName, dirPath)
		return
	}

	t.dirPkgMap[dirPath] = pkg
	return
}

// Import imports package dir and returns related package
func (t *Parser) Import(pkgPath string) (pkg *ast.Package, err error) {
	wd, err := os.Getwd()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	importPkg, err := build.Import(pkgPath, wd, build.ImportComment)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	pkg, err = t.ParseDir(importPkg.Dir, importPkg.Name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

// ParsePkg collects all objs and cache them
func (t *Parser) ParsePkg(pkg *ast.Package) (objs map[string]*ast.Object, err error) {
	if tmpObjs, ok := t.pkgObjsMap[pkg]; ok {
		objs = tmpObjs
		return
	}

	objs = make(map[string]*ast.Object)
	for _, f := range pkg.Files {
		for key, obj := range f.Scope.Objects {
			if obj.Kind == ast.Typ {
				objs[key] = obj
			}
		}
	}

	t.pkgObjsMap[pkg] = objs
	return
}

func (t *Parser) parseTypeStr(oPkg *ast.Package, typeStr string) (pkg *ast.Package,
	obj *ast.Object, err error) {
	var objs map[string]*ast.Object
	var pkgName, typeTitle string
	var ok bool

	strs := strings.Split(typeStr, ".")
	l := len(strs)
	if l == 0 || l > 2 {
		err = errors.Errorf("invalid type str %s", typeStr)
		return
	}
	if l == 1 {
		pkgName = oPkg.Name
		typeTitle = strs[0]
	} else {
		pkgName = strs[0]
		typeTitle = strs[1]
	}
	if pkgName == oPkg.Name {
		pkg = oPkg
		objs, err = t.ParsePkg(pkg)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
		if obj, ok = objs[typeTitle]; !ok {
			err = errors.Errorf("%s not found in package %s", typeTitle, pkg.Name)
			return
		}
		return
	}
	var p *ast.Package
	for _, file := range oPkg.Files {
		for _, ispec := range file.Imports {
			pkgPath := strings.Trim(ispec.Path.Value, "\"")
			p, err = t.Import(pkgPath)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
			if !(pkgName == p.Name || (ispec.Name != nil && pkgName == ispec.Name.Name)) {
				continue
			}
			objs, err = t.ParsePkg(p)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
			if _, ok = objs[typeTitle]; !ok {
				continue
			} else {
				pkg = p
				obj = objs[typeTitle]
				break
			}
		}
	}
	if pkg == nil || obj == nil {
		err = errors.Errorf("%s.%s not found", pkgName, typeTitle)
		return
	}

	return
}

func (t *Parser) parseIdentExpr(oExpr ast.Expr, pkg *ast.Package) (expr ast.Expr, err error) {
	expr = starExprX(oExpr)
	if ident, ok := expr.(*ast.Ident); ok {
		if ident.Obj != nil {
			ts, e := objDeclTypeSpec(ident.Obj)
			if e != nil {
				err = errors.WithStack(e)
				return
			}
			expr = starExprX(ts.Type)
		} else {
			// try to find obj in pkg
			objs, e := t.ParsePkg(pkg)
			if e != nil {
				err = errors.WithStack(e)
				return
			}
			if obj, ok := objs[ident.Name]; ok {
				ts, e := objDeclTypeSpec(obj)
				if e != nil {
					err = errors.WithStack(e)
					return
				}
				expr = starExprX(ts.Type)
			}
		}
	}
	return
}

func (t *Parser) parseTypeRef(pkg *ast.Package, expr ast.Expr, typeTitle string) (
	schema *spec.Schema, err error) {
	ident, isIdent := starExprX(expr).(*ast.Ident)
	typeExpr, err := t.parseIdentExpr(expr, pkg)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	switch typ := typeExpr.(type) {
	case *ast.StructType:
		if isIdent {
			typeTitle = ident.Name
		}
		schema = spec.RefProperty(t.opts.RefPrefix + typeTitle)
		_, err = t.parseType(pkg, typ, typeTitle)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
		return
	case *ast.SelectorExpr:
		typeStr, e := selectorExprTypeStr(typ)
		if e != nil {
			err = errors.WithStack(err)
			return
		}
		if typeStr != "time.Time" {
			typeTitle := typ.Sel.Name
			schema = spec.RefProperty(t.opts.RefPrefix + typeTitle)
			_, err = t.Parse(pkg, typeStr)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
			return
		}
	case *ast.ArrayType:
		var eltTitle string
		if _, isAnonymousStruct := starExprX(typ.Elt).(*ast.StructType); isAnonymousStruct {
			eltTitle = typeTitle + "_Elt"
		}
		itemsSchema, e := t.parseTypeRef(pkg, typ.Elt, eltTitle)
		if e != nil {
			err = errors.WithStack(e)
			return
		}
		schema = spec.ArrayProperty(itemsSchema)
		return
	case *ast.MapType:
		if ident, isIdent := typ.Key.(*ast.Ident); !isIdent || ident.Name != "string" {
			err = errors.Errorf("the type of map key must be string, got %s", ident.Name)
			return
		}
		var eltTitle string
		if _, isAnonymousStruct := starExprX(typ.Value).(*ast.StructType); isAnonymousStruct {
			eltTitle = typeTitle + "_Elt"
		}
		valueSchema, e := t.parseTypeRef(pkg, typ.Value, eltTitle)
		if e != nil {
			err = errors.WithStack(e)
			return
		}
		schema = spec.MapProperty(valueSchema)
		return
	}
	return t.parseType(pkg, typeExpr, typeTitle)
}

func (t *Parser) parseType(pkg *ast.Package, expr ast.Expr, typeTitle string) (schema *spec.Schema,
	err error) {
	t.lock.Lock()
	if tmpSchema, ok := t.typeMap[typeTitle]; ok {
		schema = tmpSchema
		t.lock.Unlock()
		return
	}
	if typeTitle != "" {
		t.typeMap[typeTitle] = nil
	}
	t.lock.Unlock()

	// parse ident expr
	expr, err = t.parseIdentExpr(expr, pkg)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	schema = new(spec.Schema)
	schema.WithTitle(typeTitle)
	switch typ := expr.(type) {
	case *ast.StructType:
		schema.Typed("object", "")
		if typ.Fields.List == nil {
			break
		}
		for _, field := range typ.Fields.List {
			if len(field.Names) != 0 {
				fieldName := field.Names[0].Name
				if !ast.IsExported(fieldName) {
					continue
				}
				tags := parseFieldTag(field)
				if !t.opts.IgnoreJSONTag && tags["json"] == "-" {
					continue
				}

				var fTypeTitle string
				switch starExprX(field.Type).(type) {
				case *ast.StructType, *ast.ArrayType, *ast.MapType:
					fTypeTitle = typeTitle + "_" + fieldName
				}
				prop, e := t.parseTypeRef(pkg, field.Type, fTypeTitle)
				if e != nil {
					err = errors.WithStack(e)
					return
				}

				jName := fieldName
				if !t.opts.IgnoreJSONTag && len(tags["json"]) > 0 {
					jName = strings.TrimSpace(strings.Split(tags["json"], ",")[0])
				}
				if tags["required"] == "true" {
					schema.AddRequired(jName)
				}
				prop.WithDescription(tags["description"])
				schema.SetProperty(jName, *prop)
			} else {
				tags := parseFieldTag(field)
				if !t.opts.IgnoreJSONTag && tags["json"] == "-" {
					continue
				}

				var fieldTypeTitle string
				ident, isIdent := starExprX(field.Type).(*ast.Ident)
				fieldExpr, e := t.parseIdentExpr(field.Type, pkg)
				if e != nil {
					err = errors.WithStack(e)
					return
				}
				switch fieldTyp := fieldExpr.(type) {
				case *ast.StructType:
					if isIdent {
						fieldTypeTitle = ident.Name
					}
				case *ast.SelectorExpr:
					typeStr, e := selectorExprTypeStr(fieldTyp)
					if e != nil {
						err = errors.WithStack(err)
						return
					}
					if typeStr != "time.Time" {
						fieldTypeTitle = fieldTyp.Sel.Name
					}
				}
				prop, e := t.parseTypeRef(pkg, field.Type, fieldTypeTitle)
				if e != nil {
					err = errors.WithStack(e)
					return
				}

				if !t.opts.IgnoreJSONTag && len(tags["json"]) > 0 {
					jName := strings.TrimSpace(strings.Split(tags["json"], ",")[0])
					if tags["required"] == "true" {
						schema.AddRequired(jName)
						prop.WithDescription(tags["description"])
					}
					schema.SetProperty(jName, *prop)
				} else {
					// inheritance
					schema.AddToAllOf(*prop)
				}
			}
		}
	case *ast.ArrayType, *ast.MapType:
		schema, err = t.parseTypeRef(pkg, typ, typeTitle)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	case *ast.SelectorExpr:
		typeStr, e := selectorExprTypeStr(typ)
		if e != nil {
			err = errors.WithStack(err)
			return
		}
		if typeStr == "time.Time" {
			typeType, typeFormat, e := parseBasicType("time")
			if e != nil {
				err = errors.WithStack(e)
				return
			}
			schema.Typed(typeType, typeFormat)
		} else {
			schema, err = t.Parse(pkg, typeStr)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
		}
	case *ast.Ident: // basic type only
		typeType, typeFormat, e := parseBasicType(typ.Name)
		if e != nil {
			err = errors.WithStack(e)
			return
		}
		schema.Typed(typeType, typeFormat)
	case *ast.InterfaceType:
	default:
		err = errors.Errorf("invalid expr type %T", typ)
		return
	}
	// combine schemas
	if len(schema.AllOf) != 0 && len(schema.Properties) != 0 {
		newSchema := spec.Schema{}
		newSchema.Properties = schema.Properties
		schema.AddToAllOf(newSchema)
		schema.Properties = nil
	}

	if typeTitle != "" {
		t.typeMap[typeTitle] = schema
	}
	return
}

// Parse parses type expr and returns related json schema
func (t *Parser) Parse(oPkg *ast.Package, typeStr string) (
	schema *spec.Schema, err error) {
	pkg, obj, err := t.parseTypeStr(oPkg, typeStr)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	ts, err := objDeclTypeSpec(obj)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	schema, err = t.parseType(pkg, ts.Type, obj.Name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

// Definitions returns all related definitions
func (t *Parser) Definitions() (defs spec.Definitions) {
	defs = make(spec.Definitions)
	for k, v := range t.typeMap {
		defs[k] = *v
	}
	return
}

// Reset cleans all definitions
func (t *Parser) Reset() {
	t.typeMap = make(map[string]*spec.Schema)
	return
}

func starExprX(expr ast.Expr) ast.Expr {
	if star, ok := expr.(*ast.StarExpr); ok {
		return star.X
	}
	return expr
}

func objDeclTypeSpec(obj *ast.Object) (ts *ast.TypeSpec, err error) {
	ts, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		err = errors.Errorf("invalid object decl, want *ast.TypeSpec, got %T", ts)
		return
	}
	return
}

func selectorExprTypeStr(expr *ast.SelectorExpr) (typeStr string, err error) {
	xIdent, ok := expr.X.(*ast.Ident)
	if !ok {
		err = errors.Errorf("invalid selector expr %#v", expr)
		return
	}
	typeStr = xIdent.Name + "." + expr.Sel.Name
	return
}

var fieldTagList = []string{"json", "required", "description"}

func parseFieldTag(field *ast.Field) (tags map[string]string) {
	if field.Tag == nil {
		return
	}
	tags = make(map[string]string)
	stag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
	for _, k := range fieldTagList {
		tags[k] = stag.Get(k)
	}
	return
}

var basicTypes = map[string]string{
	"bool": "boolean:",
	"uint": "integer:int64", "uint8": "integer:int32", "uint16": "integer:int32",
	"uint32": "integer:int32", "uint64": "integer:int64",
	"int": "integer:int64", "int8": "integer:int32", "int16": "integer:int32",
	"int32": "integer:int32", "int64": "integer:int64",
	"uintptr": "integer:int64",
	"float32": "number:float", "float64": "number:double",
	"string":    "string",
	"complex64": "number:float", "complex128": "number:double",
	"byte": "string:byte", "rune": "string:byte", "time": "string:date-time",
}

func parseBasicType(typeTitle string) (typ, format string, err error) {
	typeStr, ok := basicTypes[typeTitle]
	if !ok {
		err = errors.Errorf("invalid ident %s", typeTitle)
		return
	}
	exprs := strings.Split(typeStr, ":")
	typ = exprs[0]
	if len(exprs) > 1 {
		format = exprs[1]
	}
	return
}
