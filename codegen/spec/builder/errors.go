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
