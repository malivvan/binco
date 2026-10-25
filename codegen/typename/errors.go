package typename

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrTypeMismatch reports a type that is not a defined type where one is required.
var ErrTypeMismatch = errors.New("type mismatch")

// NewInvalidPkgPathError reports a string that is not a valid import path.
func NewInvalidPkgPathError(str string) error {
	return fmt.Errorf("invalid '%v' pkg path format", str)
}

// NewInvalidPackageError reports a string that is not a valid package name.
func NewInvalidPackageError(str string) error {
	return fmt.Errorf("invalid '%v' package format", str)
}

// NewMultiPointerError reports a type behind more than one pointer.
func NewMultiPointerError(t reflect.Type) error {
	return fmt.Errorf("do not support multi-pointer types, like %v", t)
}

// NewUnsupportedTypeError reports a type the generator cannot name.
func NewUnsupportedTypeError(t reflect.Type) error {
	return fmt.Errorf("can't get a name for the %v type", t)
}
