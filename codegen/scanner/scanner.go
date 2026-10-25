// Package scanner decomposes Go type names.
//
// The generator never holds reflect values while rendering: it works on names
// such as "[]pkg.Foo" or "map[string][]*pkg.Bar" and asks the scanner to report
// the pieces. The operations of an Op are called in the order the pieces appear,
// which is how the templates learn about nesting.
package scanner

import (
	"path/filepath"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/typename"
)

// Config tunes the scanner; WithoutParams skips the type parameters of a name.
type Config struct {
	WithoutParams bool
}

// Scan walks name and reports every type it contains to op.
func Scan[T QualifiedName](cfg Config, name T, op Op[T], tops tpopts.Options) (
	err error) {
	return New(cfg, op, tops).Scan(name)
}

// Scanner decomposes type names, outermost type first.
type Scanner[T QualifiedName] struct {
	cfg  Config
	op   Op[T]
	tops tpopts.Options
}

func (s Scanner[T]) scanPrimitive(t TypeInfo[T], position Position) (err error) {
	t.Kind = KindPrim
	t.Position = position
	return s.op.ProcessType(t, s.tops)
}

func (s Scanner[T]) scanContainer(t TypeInfo[T], keyType, elemType T,
	kind Kind, position Position,
) (err error) {
	t.Kind = kind
	t.Position = position
	if err = s.op.ProcessType(t, s.tops); err != nil {
		return
	}
	if keyType != "" {
		s.op.ProcessLeftSquare()
		keyScanner := s
		keyScanner.tops = tpopts.Options{}
		if err = keyScanner.scan(keyType, PositionKey); err != nil {
			return
		}
		s.op.ProcessRightSquare()
	}
	elemScanner := s
	elemScanner.tops = tpopts.Options{}
	return elemScanner.scan(elemType, PositionElem)
}

func (s Scanner[T]) scanDefined(t TypeInfo[T], position Position) (err error) {
	t.Kind = KindDefined
	t.Position = position

	if t.Position == PositionParam {
		t = joinParamPath(t)
	}

	if err = s.op.ProcessType(t, s.tops); err != nil {
		return
	}
	pS := s
	if s.cfg.WithoutParams {
		pS.op = noopOp[T]{}
	}
	pS.tops = tpopts.Options{}
	for i := range t.Params {
		if i == 0 {
			pS.op.ProcessLeftSquare()
		}
		if err = pS.scan(t.Params[i], PositionParam); err != nil {
			return
		}
		if i < len(t.Params)-1 {
			pS.op.ProcessComma()
		}
		if i == len(t.Params)-1 {
			pS.op.ProcessRightSquare()
		}
	}
	return
}

func (s Scanner[T]) scan(name T, position Position) (err error) {
	if t, ok := ParseDefinedType(name); ok {
		return s.scanDefined(t, position)
	}
	if t, keyType, elemType, kind, ok := ParseContainerType(name); ok {
		return s.scanContainer(t, keyType, elemType, kind, position)
	}
	if t, ok := ParsePrimitiveType(name); ok {
		return s.scanPrimitive(t, position)
	}
	return NewUnsupportedQualifiedNameError(name)
}

func (s Scanner[T]) Scan(name T) error {
	return s.scan(name, 0)
}

// New returns a Scanner for one call.
func New[T QualifiedName](cfg Config, op Op[T], opts tpopts.Options) Scanner[T] {
	return Scanner[T]{cfg, op, opts}
}

// joinParamPath repairs the package path of a type parameter, which reflect reports without the package name.
func joinParamPath[T QualifiedName](t TypeInfo[T]) TypeInfo[T] {
	t.PkgPath = typename.PkgPath(filepath.Join(string(t.PkgPath), string(t.Package)))
	return t
}

// Op receives the pieces of a scanned name: a type, a left square, a comma or a right square.
type Op[T QualifiedName] interface {
	ProcessType(t TypeInfo[T], tops tpopts.Options) error
	ProcessLeftSquare()
	ProcessComma()
	ProcessRightSquare()
}

// QualifiedName is the set of name types the scanner accepts: complete and full names.
type QualifiedName interface {
	typename.CompleteName | typename.FullName
}

// noopOp ignores everything, which is how parameters are skipped.
type noopOp[T QualifiedName] struct{}

func (o noopOp[T]) ProcessRightSquare() {}

func (o noopOp[T]) ProcessComma() {}

func (o noopOp[T]) ProcessLeftSquare() {}

func (o noopOp[T]) ProcessType(t TypeInfo[T], tops tpopts.Options) (err error) {
	return
}
