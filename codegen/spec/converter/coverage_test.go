package converter_test

import (
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	"github.com/malivvan/binco/codegen/typename"
)

// typenameComplete widens a literal to a complete name.
func typenameComplete(s string) typename.CompleteName { return typename.CompleteName(s) }

// TestImports checks the packages the converter reports: foreign ones only, in
// path order, with their aliases.
func TestImports(t *testing.T) {
	c := newConverter(t, genopts.WithPath("example.com/target"))
	// A defined type is "<import path>/<package>.<Type>", the convention Go
	// itself uses.
	for _, name := range []string{
		"example.com/zeta/zeta.Zeta",
		"example.com/target/target.Local",
		"example.com/alpha/alpha.Alpha",
	} {
		if _, err := c.ConvertToFullName(typenameComplete(name)); err != nil {
			t.Fatalf("ConvertToFullName(%q): %v", name, err)
		}
	}

	imports := c.Imports()
	if len(imports) != 2 {
		t.Fatalf("Imports = %v, want two foreign packages", imports)
	}
	if imports[0].Path != "example.com/alpha" || imports[1].Path != "example.com/zeta" {
		t.Fatalf("Imports are not sorted: %v", imports)
	}
	if imports[0].Alias != "alpha" {
		t.Fatalf("Imports[0].Alias = %q", imports[0].Alias)
	}
}
