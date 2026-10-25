package scanner

import (
	"github.com/malivvan/binco/codegen/typename"
)

// TypeInfo describes one type found in a name: its kind, package, parameters and position.
type TypeInfo[T QualifiedName] struct {
	PkgPath typename.PkgPath // always in the format "PkgPath/Pkg.Type", thanks
	// to the scanner
	Stars     string
	Package   typename.Package
	Name      typename.TypeName
	Params    []T
	ArrLength string

	KeyType  T
	ElemType T

	Kind     Kind
	Position Position
}

// Kind tells how a type is written: a defined name, a container or a primitive.
type Kind int

// The kinds a scanned type can have.
const (
	KindUndefined Kind = iota
	KindDefined
	KindArray
	KindSlice
	KindMap
	KindPrim
)

// Position tells where in the name a type appears.
type Position int

// The positions a type can appear in.
const (
	PositionUndefined Position = iota
	PositionKey
	PositionElem
	PositionParam
)
