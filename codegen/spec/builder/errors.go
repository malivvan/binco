package builder

import (
	"fmt"
	"reflect"

	"github.com/malivvan/binco/codegen/typename"
)

// NewUnexpectedDefinedTypeError tells the caller to use AddDefinedType.
func NewUnexpectedDefinedTypeError(t reflect.Type) error {
	return fmt.Errorf("use the Generator.AddDefinedType method for the %v type",
		t)
}

// NewUnexpectedStructTypeError tells the caller to use AddStruct.
func NewUnexpectedStructTypeError(t reflect.Type) error {
	return fmt.Errorf("use the Generator.AddStruct method for the %v type", t)
}

// NewUnexpectedInterfaceTypeError tells the caller to use AddInterface.
func NewUnexpectedInterfaceTypeError(t reflect.Type) error {
	return fmt.Errorf("use the Generator.AddInterface method for the %v type",
		t)
}

// NewUnsupportedTypeError reports a type the generator cannot handle.
func NewUnsupportedTypeError(t reflect.Type) error {
	return fmt.Errorf("unsupported %v type", t)
}

// NewNotStructError reports a type that is not a defined struct.
func NewNotStructError(t reflect.Type) error {
	return fmt.Errorf("requires a defined struct, but %v is not", t)
}

// NewTwoPathsSameAliasError reports two import paths that share a package name.
func NewTwoPathsSameAliasError(pkgPath1, pkgPath2 typename.PkgPath,
	alias typename.Package) error {
	return fmt.Errorf("two pkgPath '%v' and '%v' have the same alias '%v'",
		pkgPath1, pkgPath2, alias)
}

// NewWrongFieldsCountError reports a WithField count that does not match the struct.
func NewWrongFieldsCountError(want int) error {
	return fmt.Errorf("not enough structops.WithField() calls, want %v", want)
}

// NewAnyWithoutImplsError reports an interface{} serializer without implementations.
func NewAnyWithoutImplsError() error {
	return fmt.Errorf("an interface{} serializer needs at least one implementation; add one with intropts.WithStructImpl or intropts.WithDefinedTypeImpl")
}

// NewAnyMapKeyError reports an interface{} map key.
//
// A decoded key would hold a value of one of the registered types, and a type
// that is not comparable would panic inside the map assignment, so the shape is
// rejected while the type is generated instead of on hostile input.
func NewAnyMapKeyError() error {
	return fmt.Errorf("an interface{} cannot be a map key; a decoded key may hold a dynamic type that is not comparable, so use a concrete key type")
}
