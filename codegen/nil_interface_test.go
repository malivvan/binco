package codegen_test

import (
	"reflect"
	"testing"

	intropts "github.com/malivvan/binco/codegen/config/interface"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/testdata"
)

// TestNilInterfaceIsRepresentable covers D5: a nil interface value used to panic
// in the generated Marshal and Size. It is now written as the reserved zero type
// identifier and decoded back to nil.
func TestNilInterfaceIsRepresentable(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	mustContain(t, generate(t, g),
		"if v == nil {\n\t\tbs[0] = 0\n\t\treturn 1\n\t}",
		"case 0:\n\t\t// The writer stored a nil interface value.\n\t\treturn nil, n, nil",
		"if v == nil {\n\t\treturn 1\n\t}",
		"case 0:\n\t\treturn n, nil",
	)
}

// TestVersionedCountsFailedValueBytes covers the rest of D5: on an inner error
// the versioned serializer returned fewer bytes than the interface serializer
// for the same input.
func TestVersionedCountsFailedValueBytes(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterVersioned(reflect.TypeFor[testdata.VerFoo](),
		veropts.WithVersion(reflect.TypeFor[testdata.VerFooVer1](), "MigrateVerFooVer1"),
		veropts.WithCurrentVersion(reflect.TypeFor[testdata.VerFooVer2]())); err != nil {
		t.Fatalf("RegisterVersioned: %v", err)
	}
	mustContain(t, generate(t, g), "n += n1\n\t\tif err != nil {")
}
