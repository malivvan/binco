package builder_test

import (
	"reflect"
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	"github.com/malivvan/binco/codegen/spec/builder"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

type field struct{ A, B int }

// TestValidateExpr covers S11: validator and migration expressions are copied
// into generated source, so a malformed one must be rejected where it was
// supplied instead of surfacing as a generic code-generation failure.
func TestValidateExpr(t *testing.T) {
	valid := []string{"", "ValidateFoo", "pkg.Validate", "func(v Foo) error { return nil }"}
	for _, expr := range valid {
		if err := builder.ValidateExpr("validator", expr); err != nil {
			t.Fatalf("ValidateExpr(%q) = %v, want nil", expr, err)
		}
	}
	invalid := []string{"Validate(", "Validate Foo", "func(", ")"}
	for _, expr := range invalid {
		if err := builder.ValidateExpr("validator", expr); err == nil {
			t.Fatalf("ValidateExpr(%q) returned nil, want an error", expr)
		}
	}
}

func TestValidateIdent(t *testing.T) {
	if err := builder.ValidateIdent("alias", "varnum"); err != nil {
		t.Fatalf("ValidateIdent returned %v", err)
	}
	if err := builder.ValidateIdent("alias", "1abc"); err == nil {
		t.Fatal("ValidateIdent accepted a non-identifier")
	}
}

// TestInvalidValidatorIsRejectedByTheBuilder checks that the check runs for the
// options a caller actually passes to the generator.
func TestInvalidValidatorIsRejectedByTheBuilder(t *testing.T) {
	gops := genopts.New()
	gops.PkgPath = typename.PkgPath("example.com/p")
	gops.Package = typename.Package("p")
	b := builder.NewTypeBuilder(cnvtr.NewTypeNameConverter(gops), gops)

	_, err := b.BuildStructType(reflect.TypeFor[field](), stopts.Options{
		Validator: "NotAValidator(",
	})
	if err == nil {
		t.Fatal("BuildStructType accepted a malformed struct validator")
	}

	_, err = b.BuildStructType(reflect.TypeFor[field](), stopts.Options{})
	if err != nil {
		t.Fatalf("BuildStructType with no field options returned %v", err)
	}

	_, err = b.BuildStructType(reflect.TypeFor[field](), stopts.Options{
		Validator: "func(",
	})
	if err == nil {
		t.Fatal("BuildStructType accepted a malformed validator")
	}
}

// TestValidateIdentAllowsEmpty checks that an unset option is not an error.
func TestValidateIdentAllowsEmpty(t *testing.T) {
	if err := builder.ValidateIdent("validator", ""); err != nil {
		t.Fatalf("ValidateIdent(%q) = %v, want nil", "", err)
	}
}
