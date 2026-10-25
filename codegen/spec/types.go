package spec

import (
	"fmt"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	structops "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/codegen/typename"
)

// Type is the common part of every specification: the full name of the Go type and the generator options.
type Type struct {
	FullName typename.FullName
	Gops     genopts.Options
}

// DefinedType describes a type whose underlying type is a number, string, container or pointer.
type DefinedType struct {
	FullName           typename.FullName
	UnderlyingTypeName typename.FullName // e.g., 'int' for 'type Int int'

	Tops tpopts.Options
	Gops genopts.Options
}

// StructType describes a struct, its fields and its field options.
type StructType struct {
	FullName typename.FullName
	Fields   []FieldType
	Sops     structops.Options
	Gops     genopts.Options
}

func (d StructType) SerializedFields() (sl []FieldType) {
	if len(d.Sops.Fields) == 0 {
		return d.Fields
	}
	sl = make([]FieldType, 0, len(d.Fields))
	for i, f := range d.Fields {
		if !d.Sops.Fields[i].Ignore {
			sl = append(sl, f)
		}
	}
	return
}

// InterfaceType describes an interface and the implementations it dispatches to.
type InterfaceType struct {
	FullName typename.FullName
	Impls    []typename.FullName
	Iops     intropts.Options
	Gops     genopts.Options
}

// VersionedType describes a type with several versions.
type VersionedType struct {
	FullName typename.FullName
	Versions []typename.FullName
	Vops     veropts.Options
	Gops     genopts.Options
}

// FieldType describes one struct field.
type FieldType struct {
	FullName  typename.FullName
	FieldName string
	Gops      genopts.Options
}

// AnonType describes an anonymous serializer that the generator has to declare for a container or pointer type.
type AnonType struct {
	SerName AnonSerName
	Kind    AnonKind

	ArrType   typename.FullName
	ArrLength string

	LenSer string
	LenVl  string

	KeyType typename.FullName
	KeyVl   string

	ElemType typename.FullName
	ElemVl   string
}

// AnonSerName is the generated name of an anonymous serializer.
type AnonSerName string

// AnonKind says which container an anonymous serializer belongs to.
type AnonKind int

func (k AnonKind) String() string {
	switch k {
	case AnonKindString:
		return "string"
	case AnonKindArray:
		return "array"
	case AnonKindByteSlice:
		return "byteSlice"
	case AnonKindSlice:
		return "slice"
	case AnonKindMap:
		return "map"
	case AnonKindPtr:
		return "ptr"
	default:
		panic(fmt.Sprintf("unexpected %v AnonSerKind", int(k)))
	}
}

// The kinds an anonymous serializer can have.
const (
	AnonKindUndefined AnonKind = iota
	AnonKindString
	AnonKindArray
	AnonKindByteSlice
	AnonKindSlice
	AnonKindMap
	AnonKindPtr
)
