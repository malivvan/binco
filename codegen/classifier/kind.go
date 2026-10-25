package classifier

import (
	"reflect"
)

// DefinedType reports whether t was declared with a name in a package.
func DefinedType(t reflect.Type) (ok bool) {
	return t.PkgPath() != ""
}

// DefinedBasicType reports whether t is a defined type with a number, string, container or pointer underlying type.
func DefinedBasicType(t reflect.Type) (ok bool) {
	return DefinedType(t) && BasicType(t)
}

// DefinedStruct reports whether t is a defined struct.
func DefinedStruct(t reflect.Type) (ok bool) {
	return DefinedType(t) && t.Kind() == reflect.Struct
}

// DefinedInterface reports whether t is a defined interface.
func DefinedInterface(t reflect.Type) (ok bool) {
	return DefinedType(t) && InterfaceType(t)
}

// DefinedNonEmptyInterface reports whether t is a defined interface with at least one method.
func DefinedNonEmptyInterface(t reflect.Type) (ok bool) {
	return DefinedType(t) && InterfaceType(t) && t.NumMethod() > 0
}

// BasicType reports whether t is a primitive, container or pointer type.
func BasicType(t reflect.Type) (ok bool) {
	return PrimitiveType(t) || ContainerType(t) || PtrType(t)
}

// PrimitiveType reports whether t is a built-in scalar type.
func PrimitiveType(t reflect.Type) (ok bool) {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return
	}
}

// InterfaceType reports whether t is an interface.
func InterfaceType(t reflect.Type) (ok bool) {
	return t.Kind() == reflect.Interface
}

// PtrType reports whether t is a pointer.
func PtrType(t reflect.Type) bool {
	return t.Kind() == reflect.Pointer
}

// ContainerType reports whether t is an array, a slice or a map.
func ContainerType(t reflect.Type) (ok bool) {
	return ArrayType(t) || SliceType(t) || MapType(t)
}

// ArrayType reports whether t is an array.
func ArrayType(t reflect.Type) (ok bool) {
	return t.Kind() == reflect.Array
}

// SliceType reports whether t is a slice.
func SliceType(t reflect.Type) (ok bool) {
	return t.Kind() == reflect.Slice
}

// MapType reports whether t is a map.
func MapType(t reflect.Type) (ok bool) {
	return t.Kind() == reflect.Map
}
