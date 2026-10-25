// Package builder turns reflect types into generator specifications.
//
// It validates that a type can be declared at all (a defined type is not a
// struct, a struct is not an interface, and so on), derives the names the
// templates need, and collects the anonymous serializers that a struct or a
// defined type requires because one of its fields is a container or a pointer.
package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

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

func (b TypeBuilder) parseVersions(vops veropts.Options) ([]typename.FullName, error) {
	versions := make([]typename.FullName, len(vops.Versions))
	for i, version := range vops.Versions {
		name, err := b.parseTypeName(version.Type)
		if err != nil {
			return nil, err
		}
		versions[i] = name
	}
	return versions, nil
}

func (b TypeBuilder) parseImpls(iops intropts.Options) ([]typename.FullName, error) {
	impls := make([]typename.FullName, len(iops.Impls))
	for i, impl := range iops.Impls {
		name, err := b.parseTypeName(impl)
		if err != nil {
			return nil, err
		}
		impls[i] = name
	}
	return impls, nil
}

func (b TypeBuilder) parseFieldNames(t reflect.Type) ([]typename.FullName, error) {
	fieldNames := make([]typename.FullName, t.NumField())
	for i := range t.NumField() {
		name, err := b.parseTypeName(t.Field(i).Type)
		if err != nil {
			return nil, err
		}
		fieldNames[i] = name
	}
	return fieldNames, nil
}

func (b TypeBuilder) parseUnderlyingTypeName(t reflect.Type) (typename.FullName, error) {
	cname, err := typename.BaseTypeCompleteName(t)
	if err != nil {
		return "", err
	}
	return b.converter.ConvertToFullName(cname)
}

func (b TypeBuilder) parseTypeName(t reflect.Type) (typename.FullName, error) {
	cname, err := typename.TypeCompleteName(t)
	if err != nil {
		return "", err
	}
	return b.converter.ConvertToFullName(cname)
}

func (b TypeBuilder) makeFields(t reflect.Type, fieldNames []typename.FullName) []spec.FieldType {
	fields := make([]spec.FieldType, len(fieldNames))
	for i := range fieldNames {
		fields[i] = spec.FieldType{
			FullName:  fieldNames[i],
			FieldName: t.Field(i).Name,
			Index:     i,
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
	if err = validateTypeOptions(tops); err != nil {
		return
	}
	op := NewCollectAnonOp(m, b.converter)
	return scanner.Scan(scanner.Config{}, name, op, tops)
}

func (b TypeBuilder) BuildAnonType(name typename.FullName, tops tpopts.Options) (
	anonType spec.AnonType, ok bool, err error) {
	if err = validateTypeOptions(tops); err != nil {
		return
	}
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
	if err = validateTypeOptions(tops); err != nil {
		return
	}

	name, err := b.parseTypeName(t)
	if err != nil {
		return definedType, err
	}
	definedType.FullName = name
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
	typeName, err := b.parseTypeName(t)
	if err != nil {
		return versionedType, err
	}
	versions, err := b.parseVersions(vops)
	if err != nil {
		return versionedType, err
	}
	for _, version := range vops.Versions {
		if err = ValidateExpr("migration", version.Migration); err != nil {
			return
		}
	}
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
	typeName, err := b.parseTypeName(t)
	if err != nil {
		return interfaceType, err
	}
	impls, err := b.parseImpls(iops)
	if err != nil {
		return interfaceType, err
	}
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
	tp.FullName, err = b.parseTypeName(t)
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
	if err = ValidateExpr("struct validator", sops.Validator); err != nil {
		return
	}
	for i := range sops.Fields {
		if err = validateTypeOptions(sops.Fields[i].Type); err != nil {
			return
		}
	}
	if err = b.checkUnexported(t); err != nil {
		return
	}
	typeName, err := b.parseTypeName(t)
	if err != nil {
		return structType, err
	}
	fieldNames, err := b.parseFieldNames(t)
	if err != nil {
		return structType, err
	}
	structType.FullName = typeName
	structType.Fields = b.makeFields(t, fieldNames)
	structType.Sops = sops
	structType.Gops = b.gops
	structType.Fingerprint = schemaFingerprint(structType.Fields)
	return
}

// checkUnexported rejects a foreign struct with unexported fields.
//
// Generated code lives in the target package, so it cannot read or write an
// unexported field of a struct declared elsewhere. Catching it here reports the
// type that cannot be generated instead of failing later with a compile error
// inside the generated file.
func (b TypeBuilder) checkUnexported(t reflect.Type) error {
	if t.PkgPath() == string(b.gops.PkgPath) {
		return nil
	}
	for i := range t.NumField() {
		if name := t.Field(i).Name; !t.Field(i).IsExported() {
			return fmt.Errorf("%v has the unexported field %q; a struct from another package can only be generated when every field is exported", t, name)
		}
	}
	return nil
}

// schemaFingerprint hashes the field names and wire types in declaration order.
func schemaFingerprint(fields []spec.FieldType) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(f.FieldName)
		b.WriteByte(':')
		b.WriteString(string(f.FullName))
		b.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func (b TypeBuilder) BuildDefinedType(t reflect.Type, tops tpopts.Options) (
	definedTime spec.DefinedType, err error) {
	if !classifier.DefinedBasicType(t) {
		err = b.notDefinedTypeError(t)
		return
	}
	if err = validateTypeOptions(tops); err != nil {
		return
	}
	typeName, err := b.parseTypeName(t)
	if err != nil {
		return definedTime, err
	}
	underlyingTypeName, err := b.parseUnderlyingTypeName(t)
	if err != nil {
		return definedTime, err
	}
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
	ConvertToFullName(cname typename.CompleteName) (fname typename.FullName, err error)
	ConvertToRelativeName(fname typename.FullName) (rname typename.RelativeName, err error)
}

// validateTypeOptions rejects validator expressions that are not Go expressions.
//
// The expressions are copied into generated source, so checking them here keeps
// the failure next to the option that caused it.
func validateTypeOptions(tops tpopts.Options) (err error) {
	for _, item := range []struct {
		option string
		expr   string
	}{
		{"validator", tops.Validator},
		{"length validator", tops.LenValidator},
		{"element validator", tops.ElemValidator},
		{"key validator", tops.KeyValidator},
	} {
		if err = ValidateExpr(item.option, item.expr); err != nil {
			return
		}
	}
	return
}
