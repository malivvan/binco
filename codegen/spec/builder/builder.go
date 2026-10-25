// Package builder turns reflect types into generator specifications.
//
// It validates that a type can be declared at all (a defined type is not a
// struct, a struct is not an interface, and so on), derives the names the
// templates need, and collects the anonymous serializers that a struct or a
// defined type requires because one of its fields is a container or a pointer.
package builder

import (
	"reflect"

	"github.com/malivvan/binco/codegen/classifier"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/spec"
	"github.com/malivvan/binco/codegen/typename"
)

// TypeBuilder converts reflect types into the specifications of package spec.
type TypeBuilder struct {
	converter TypeNameConverter
	gops      genopts.Options
}

func (b TypeBuilder) notInterfaceError(t reflect.Type) error {
	switch {
	case classifier.DefinedBasicType(t):
		return NewUnexpectedDefinedTypeError(t)
	case classifier.DefinedStruct(t):
		return NewUnexpectedStructTypeError(t)
	default:
		return NewUnsupportedTypeError(t)
	}
}

func (b TypeBuilder) notStructError(t reflect.Type) error {
	switch {
	case classifier.DefinedBasicType(t):
		return NewUnexpectedDefinedTypeError(t)
	case classifier.DefinedNonEmptyInterface(t):
		return NewUnexpectedInterfaceTypeError(t)
	default:
		return NewUnsupportedTypeError(t)
	}
}

func (b TypeBuilder) notDefinedTypeError(t reflect.Type) error {
	switch {
	case classifier.DefinedStruct(t):
		return NewUnexpectedStructTypeError(t)
	case classifier.DefinedNonEmptyInterface(t):
		return NewUnexpectedInterfaceTypeError(t)
	default:
		return NewUnsupportedTypeError(t)
	}
}

func (b TypeBuilder) parseVersions(vops veropts.Options) []typename.FullName {
	versions := make([]typename.FullName, len(vops.Versions))
	for i, version := range vops.Versions {
		versions[i] = b.parseTypeName(version.Type)
	}
	return versions
}

func (b TypeBuilder) parseImpls(iops intropts.Options) []typename.FullName {
	impls := make([]typename.FullName, len(iops.Impls))
	for i, impl := range iops.Impls {
		impls[i] = b.parseTypeName(impl)
	}
	return impls
}

func (b TypeBuilder) parseFieldNames(t reflect.Type) []typename.FullName {
	fieldNames := make([]typename.FullName, t.NumField())
	for i := range t.NumField() {
		fieldNames[i] = b.parseTypeName(t.Field(i).Type)
	}
	return fieldNames
}

func (b TypeBuilder) parseUnderlyingTypeName(t reflect.Type) (typeName typename.FullName) {
	cname, err := typename.BaseTypeCompleteName(t)
	if err != nil {
		return
	}
	return b.converter.ConvertToFullName(cname)
}

func (b TypeBuilder) parseTypeName(t reflect.Type) (typeName typename.FullName) {
	cname, err := typename.TypeCompleteName(t)
	if err != nil {
		return
	}
	return b.converter.ConvertToFullName(cname)
}

func (b TypeBuilder) makeFields(t reflect.Type, fieldNames []typename.FullName) []spec.FieldType {
	fields := make([]spec.FieldType, len(fieldNames))
	for i := range fieldNames {
		fields[i] = spec.FieldType{
			FullName:  fieldNames[i],
			FieldName: t.Field(i).Name,
			Gops:      b.gops,
		}
	}
	return fields
}

func (b TypeBuilder) checkFields(t reflect.Type, sops stopts.Options) (
	err error,
) {
	if len(sops.Fields) == 0 {
		return
	}
	var (
		want   = t.NumField()
		actual = len(sops.Fields)
	)
	if actual != want {
		err = NewWrongFieldsCountError(want)
	}
	return
}

func (b TypeBuilder) CollectAnonTypes(name typename.FullName,
	m map[spec.AnonSerializerName]spec.AnonType,
	tops tpopts.Options,
) (err error) {
	op := NewCollectAnonOp(m, b.converter)
	return scanner.Scan(scanner.Config{}, name, op, tops)
}

func (b TypeBuilder) BuildAnonType(name typename.FullName, tops tpopts.Options) (
	anonType spec.AnonType, ok bool, err error) {
	op := NewMakeAnonOp(b.converter)

	if err = scanner.Scan(scanner.Config{}, name, op, tops); err != nil {
		if err != ErrStop {
			return
		}
		err = nil
	}
	anonType, ok = op.AnonType()
	return
}

func (b TypeBuilder) BuildTimeType(t reflect.Type, uops stopts.UnderlyingTimeOptions,
	validator string,
) (definedType spec.DefinedType, err error) {
	if !classifier.DefinedStruct(t) {
		err = NewNotStructError(t)
		return
	}
	var tops tpopts.Options
	tpopts.Apply(&tops,
		tpopts.WithTimeUnit(uops.TimeUnit),
		tpopts.WithValidator(validator),
	)

	definedType.FullName = b.parseTypeName(t)
	definedType.UnderlyingTypeName = "time.Time"
	definedType.Tops = tops
	definedType.Gops = b.gops
	return
}

func (b TypeBuilder) BuildVersionedType(t reflect.Type, vops veropts.Options) (
	versionedType spec.VersionedType, err error,
) {
	if !classifier.DefinedType(t) {
		err = b.notInterfaceError(t)
		return
	}
	var (
		typeName = b.parseTypeName(t)
		versions = b.parseVersions(vops)
	)
	versionedType.FullName = typeName
	versionedType.Versions = versions
	versionedType.Vops = vops
	versionedType.Gops = b.gops
	return
}

func (b TypeBuilder) BuildInterfaceType(t reflect.Type, iops intropts.Options) (
	interfaceType spec.InterfaceType, err error,
) {
	if !classifier.DefinedInterface(t) {
		err = b.notInterfaceError(t)
		return
	}
	var (
		typeName = b.parseTypeName(t)
		impls    = b.parseImpls(iops)
	)
	interfaceType.FullName = typeName
	interfaceType.Impls = impls
	interfaceType.Iops = iops
	interfaceType.Gops = b.gops
	return
}

func (b TypeBuilder) BuildTyped(t reflect.Type) (
	tp spec.Type, err error,
) {
	if !classifier.DefinedBasicType(t) && !classifier.DefinedStruct(t) &&
		!classifier.DefinedNonEmptyInterface(t) {
		err = NewUnsupportedTypeError(t)
		return
	}
	tp.FullName = b.parseTypeName(t)
	tp.Gops = b.gops
	return
}

func (b TypeBuilder) BuildStructType(t reflect.Type, sops stopts.Options) (
	structType spec.StructType, err error,
) {
	if !classifier.DefinedStruct(t) {
		err = b.notStructError(t)
		return
	}
	if err = b.checkFields(t, sops); err != nil {
		return
	}
	var (
		typeName   = b.parseTypeName(t)
		fieldNames = b.parseFieldNames(t)
	)
	structType.FullName = typeName
	structType.Fields = b.makeFields(t, fieldNames)
	structType.Sops = sops
	structType.Gops = b.gops
	return
}

func (b TypeBuilder) BuildDefinedType(t reflect.Type, tops tpopts.Options) (
	definedTime spec.DefinedType, err error) {
	if !classifier.DefinedBasicType(t) {
		err = b.notDefinedTypeError(t)
		return
	}
	var (
		typeName           = b.parseTypeName(t)
		underlyingTypeName = b.parseUnderlyingTypeName(t)
	)
	definedTime.FullName = typeName
	definedTime.UnderlyingTypeName = underlyingTypeName
	definedTime.Tops = tops
	definedTime.Gops = b.gops
	return
}

// NewTypeBuilder returns a builder that names types through converter.
func NewTypeBuilder(converter TypeNameConverter, gops genopts.Options) TypeBuilder {
	return TypeBuilder{converter, gops}
}

// TypeNameConverter rewrites Go type names so that the generated code refers to them correctly.
type TypeNameConverter interface {
	ConvertToFullName(cname typename.CompleteName) (fname typename.FullName)
	ConvertToRelativeName(fname typename.FullName) (rname typename.RelativeName)
}
