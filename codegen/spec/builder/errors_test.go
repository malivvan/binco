package builder_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen/spec/builder"
	"github.com/malivvan/binco/codegen/typename"
)

// TestErrorConstructors pins the messages the generator shows a caller. They are
// the main diagnostic when a type is registered through the wrong method.
func TestErrorConstructors(t *testing.T) {
	intType := reflect.TypeFor[int]()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unexpected defined type", builder.NewUnexpectedDefinedTypeError(intType),
			"use the Generator.AddDefinedType method for the int type"},
		{"unexpected struct type", builder.NewUnexpectedStructTypeError(intType),
			"use the Generator.AddStruct method for the int type"},
		{"unexpected interface type", builder.NewUnexpectedInterfaceTypeError(intType),
			"use the Generator.AddInterface method for the int type"},
		{"unsupported type", builder.NewUnsupportedTypeError(intType),
			"unsupported int type"},
		{"not struct", builder.NewNotStructError(intType),
			"requires a defined struct, but int is not"},
		{"two paths same alias",
			builder.NewTwoPathsSameAliasError(
				typename.PkgPath("example.com/a"), typename.PkgPath("example.com/b"),
				typename.Package("pkg")),
			"two pkgPath 'example.com/a' and 'example.com/b' have the same alias 'pkg'"},
		{"wrong fields count", builder.NewWrongFieldsCountError(3),
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
