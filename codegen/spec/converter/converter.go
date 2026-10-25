// Package converter translates between the three ways this module names a type.
//
// A complete name is what reflect produces: the package path followed by the
// package name and the type ("example.com/p/types/types.Foo"). A full name is
// what the templates parse ("types.Foo"). A relative name is what the generated
// file uses ("Foo" inside the target package, "types.Foo" outside it).
package converter

import (
	"fmt"
	"strings"
	"sync"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

// TypeNameConverter converts complete type names to full names, and full names to the names used in generated code.
//
// The converter records the package aliases it has seen, so it is stateful and
// serialises that state with a mutex. It returns errors instead of panicking: a
// name that cannot be parsed is a caller mistake that Generate can report.
type TypeNameConverter struct {
	knownPkgs map[typename.Package]typename.PkgPath
	mu        *sync.Mutex
	gopts     genopts.Options
}

// ConvertToRelativeName returns the name the generated file uses for fname.
func (c TypeNameConverter) ConvertToRelativeName(fname typename.FullName) (
	rname typename.RelativeName, err error) {
	op := newRelativeNameOp(c.gopts)
	if err = scanner.Scan(scanner.Config{}, fname, op, tpopts.Options{}); err != nil {
		return "", fmt.Errorf("cannot convert %v to a relative name: %w", fname, err)
	}
	return op.RelativeName(), nil
}

// ConvertToFullName returns the full name the templates parse for cname.
func (c TypeNameConverter) ConvertToFullName(cname typename.CompleteName) (
	fname typename.FullName, err error) {
	op := newFullNameOp(c.knownPkgs, c.mu, c.gopts)
	if err = scanner.Scan(scanner.Config{}, cname, op, tpopts.Options{}); err != nil {
		return "", fmt.Errorf("cannot convert %v to a full name: %w", cname, err)
	}
	return op.FullName(), nil
}

// NewTypeNameConverter returns a converter that knows the package of the generated file.
func NewTypeNameConverter(gops genopts.Options) TypeNameConverter {
	return TypeNameConverter{
		knownPkgs: map[typename.Package]typename.PkgPath{gops.Package: gops.PkgPath},
		mu:        new(sync.Mutex),
		gopts:     gops,
	}
}

// fullNameOp rewrites the packages of a complete name into their aliases.
type fullNameOp struct {
	b         *strings.Builder
	knownPkgs map[typename.Package]typename.PkgPath
	mu        *sync.Mutex
	gopts     genopts.Options
}

func (o *fullNameOp) checkPkg(pkg typename.Package,
	t scanner.TypeInfo[typename.CompleteName]) (err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if pkgPath, pst := o.knownPkgs[pkg]; pst && pkgPath != t.PkgPath {
		err = NewTwoPathsSameAliasError(pkgPath, t.PkgPath, pkg)
	}
	return
}

func (o *fullNameOp) choosePkg(t scanner.TypeInfo[typename.CompleteName]) (
	pkg typename.Package, err error,
) {
	importPath := genopts.ImportPath(t.PkgPath)

	if alias, pst := o.gopts.ImportAliases()[importPath]; pst {
		pkg = typename.Package(alias)
	} else {
		pkg = t.Package
		if t.Position == scanner.PositionParam {

			if o.gopts.Logger != nil {
				o.gopts.Logger.Printf("WARNING: no alias for '%v' in the generator options\n",
					t.PkgPath)
			}
		}
	}
	if err = o.checkPkg(pkg, t); err != nil {
		return
	}
	o.mu.Lock()
	o.knownPkgs[pkg] = t.PkgPath
	o.mu.Unlock()
	return
}

func (o *fullNameOp) FullName() typename.FullName {
	return typename.FullName(o.b.String())
}

func (o *fullNameOp) ProcessRightSquare() {
	o.b.WriteString("]")
}

func (o *fullNameOp) ProcessComma() {
	o.b.WriteString(",")
}

func (o *fullNameOp) ProcessLeftSquare() {
	o.b.WriteString("[")
}

func (o *fullNameOp) ProcessType(t scanner.TypeInfo[typename.CompleteName],
	tops tpopts.Options,
) (err error) {
	var pkg typename.Package
	o.b.WriteString(t.Stars)
	switch t.Kind {
	case scanner.KindDefined:
		if pkg, err = o.choosePkg(t); err != nil {
			return
		}
		o.b.WriteString(string(pkg))
		o.b.WriteString(".")
		o.b.WriteString(string(t.Name))
	case scanner.KindArray:
		o.b.WriteString("[")
		o.b.WriteString(t.ArrLength)
		o.b.WriteString("]")
	case scanner.KindSlice:
		o.b.WriteString("[]")
	case scanner.KindMap:
		o.b.WriteString("map")
	case scanner.KindPrim:
		o.b.WriteString(string(t.Name))
	default:
		return fmt.Errorf("unexpected %v kind", t.Kind)
	}
	return
}

// newFullNameOp returns a full-name rewriter.
func newFullNameOp(knownPkgs map[typename.Package]typename.PkgPath,
	mu *sync.Mutex, gops genopts.Options) *fullNameOp {
	return &fullNameOp{&strings.Builder{}, knownPkgs, mu, gops}
}

// relativeNameOp drops the package of the generated file from a full name.
type relativeNameOp struct {
	b    *strings.Builder
	gops genopts.Options
}

func (o *relativeNameOp) choosePkg(t scanner.TypeInfo[typename.FullName]) (
	pkg typename.Package, err error) {
	if t.Package == o.gops.Package {
		pkg = ""
	} else {
		pkg = t.Package
	}
	return
}

func (o *relativeNameOp) RelativeName() typename.RelativeName {
	return typename.RelativeName(o.b.String())
}

func (o *relativeNameOp) ProcessRightSquare() {
	o.b.WriteString("]")
}

func (o *relativeNameOp) ProcessComma() {
	o.b.WriteString(",")
}

func (o *relativeNameOp) ProcessLeftSquare() {
	o.b.WriteString("[")
}

func (o *relativeNameOp) ProcessType(t scanner.TypeInfo[typename.FullName],
	tops tpopts.Options,
) (err error) {
	var pkg typename.Package
	o.b.WriteString(t.Stars)
	switch t.Kind {
	case scanner.KindDefined:
		if pkg, err = o.choosePkg(t); err != nil {
			return
		}
		if pkg != "" {
			o.b.WriteString(string(pkg))
			o.b.WriteString(".")
		}
		o.b.WriteString(string(t.Name))
	case scanner.KindArray:
		o.b.WriteString("[")
		o.b.WriteString(t.ArrLength)
		o.b.WriteString("]")
	case scanner.KindSlice:
		o.b.WriteString("[]")
	case scanner.KindMap:
		o.b.WriteString("map")
	case scanner.KindPrim:
		o.b.WriteString(string(t.Name))
	default:
		return fmt.Errorf("unexpected %v kind", t.Kind)
	}
	return
}

// newRelativeNameOp returns a relative-name rewriter.
func newRelativeNameOp(gops genopts.Options) *relativeNameOp {
	return &relativeNameOp{&strings.Builder{}, gops}
}
