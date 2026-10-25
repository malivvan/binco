package converter_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

func TestErrorConstructors(t *testing.T) {
	intType := reflect.TypeFor[int]()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unexpected defined type", converter.NewUnexpectedDefinedTypeError(intType),
			"use the Generator.AddDefinedType method for the int type"},
		{"unexpected struct type", converter.NewUnexpectedStructTypeError(intType),
			"use the Generator.AddStruct method for the int type"},
		{"unexpected interface type", converter.NewUnexpectedInterfaceTypeError(intType),
			"use the Generator.AddInterface method for the int type"},
		{"unsupported type", converter.NewUnsupportedTypeError(intType),
			"unsupported int type"},
		{"not defined type", converter.NewNotDefinedTypeError(intType),
			"requires a defined type, but int is not"},
		{"not struct", converter.NewNotStructError(intType),
			"requires a defined struct, but int is not"},
		{"not interface", converter.NewNotInterfaceError(intType),
			"requires a defined interface, but int is not"},
		{"two paths same alias",
			converter.NewTwoPathsSameAliasError(
				typename.PkgPath("example.com/a"), typename.PkgPath("example.com/b"),
				typename.Package("pkg")),
			"two pkgPath 'example.com/a' and 'example.com/b' have the same alias 'pkg'"},
		{"wrong fields count", converter.NewWrongFieldsCountError(3),
			"not enough structops.WithField() calls, want 3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("constructor returned nil")
			}
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}
