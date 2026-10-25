package converter_test

import (
	"testing"

	"github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

// TestTwoPathsSameAliasError pins the message for the one error the converter
// raises itself; the generator's other type errors live in package builder.
func TestTwoPathsSameAliasError(t *testing.T) {
	err := converter.NewTwoPathsSameAliasError(
		typename.PkgPath("example.com/a"), typename.PkgPath("example.com/b"),
		typename.Package("pkg"))
	if got, want := err.Error(),
		"two pkgPath 'example.com/a' and 'example.com/b' have the same alias 'pkg'"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
