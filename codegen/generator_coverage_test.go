package codegen_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/testdata"
	other "github.com/malivvan/binco/testdata/other"
)

// TestDuplicateDTMNamesAreRejected checks that two registered types with the
// same bare name cannot both declare an identifier constant.
func TestDuplicateDTMNamesAreRejected(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.SameName]()),
		intropts.WithStructImpl(reflect.TypeFor[other.SameName]()),
	); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	if _, err := g.Generate(); !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("Generate = %v, want ErrDuplicateDTM", err)
	}
}

// TestDTMRegistryReportsDuplicateIDs checks the pinned-identifier validation.
func TestDTMRegistryReportsDuplicateIDs(t *testing.T) {
	g := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 5),
		genopts.WithDTM(reflect.TypeFor[testdata.DefinedImpl](), 5),
	)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]()),
	); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	if _, err := g.DTMRegistry(); !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("DTMRegistry = %v, want ErrDuplicateDTM", err)
	}
}

// TestAddStructUnderlyingTimeRejectsNonStructs checks the time-wrapper path.
func TestAddStructUnderlyingTimeRejectsNonStructs(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[testdata.Int](), stopts.WithUnderlyingTime()); err == nil {
		t.Fatal("AddStruct with WithUnderlyingTime accepted a non-struct")
	}
}

// TestGenerateReportsInvalidGeneratedSource checks the error path of the import
// processor: a serializer name that is not an identifier produces source that
// does not parse.
func TestGenerateReportsInvalidGeneratedSource(t *testing.T) {
	g := newGenerator(t,
		genopts.WithSerializer(reflect.TypeFor[testdata.Int](), "not an identifier"))
	if err := g.AddDefinedType(reflect.TypeFor[testdata.Int]()); err != nil {
		t.Fatalf("AddDefinedType: %v", err)
	}
	if _, err := g.Generate(); !errors.Is(err, codegen.ErrCodeGenFailed) {
		t.Fatalf("Generate = %v, want ErrCodeGenFailed", err)
	}
}

// TestGeneratedImportsUseTheConfiguredAlias checks the aliased-import branch of
// the import block builder.
func TestGeneratedImportsUseTheConfiguredAlias(t *testing.T) {
	g := newGenerator(t,
		genopts.WithImportAlias("github.com/malivvan/binco/testdata/other", "oth"))
	if err := g.AddStruct(reflect.TypeFor[other.SameName]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	mustContain(t, generate(t, g), `oth "github.com/malivvan/binco/testdata/other"`)
}

// TestRegisterVersionedErrors covers the two ways registering versions fails
// before any type is added.
func TestRegisterVersionedErrors(t *testing.T) {
	// Two current versions are ambiguous.
	g := newGenerator(t)
	err := g.RegisterVersioned(reflect.TypeFor[testdata.VerFoo](),
		veropts.WithCurrentVersion(reflect.TypeFor[testdata.VerFooVer2]()),
		veropts.WithCurrentVersion(reflect.TypeFor[testdata.VerFooVer1]()),
	)
	if err == nil {
		t.Fatal("RegisterVersioned accepted two current versions")
	}

	// A version type that cannot be generated stops the registration.
	g = newGenerator(t)
	err = g.RegisterVersioned(reflect.TypeFor[testdata.VerFoo](),
		veropts.WithCurrentVersion(reflect.TypeFor[time.Time]()))
	if err == nil {
		t.Fatal("RegisterVersioned accepted a version from another package")
	}
}

// TestRegisterInterfaceErrors covers the two ways registering implementations
// fails before any type is added.
func TestRegisterInterfaceErrors(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[time.Time]())); err == nil {
		t.Fatal("RegisterInterface accepted a struct from another package")
	}

	g = newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.SimpleStruct]())); err == nil {
		t.Fatal("RegisterInterface accepted a struct as a defined-type implementation")
	}
}

// TestNewGeneratorRejectsUnnamedTypes checks that a pinned identifier or an
// explicit serializer name needs a type that has a name.
func TestNewGeneratorRejectsUnnamedTypes(t *testing.T) {
	if _, err := codegen.NewGenerator(genopts.WithPath(modelPkgPath),
		genopts.WithDTM(reflect.TypeFor[struct{ X int }](), 1)); err == nil {
		t.Fatal("NewGenerator accepted a pinned unnamed type")
	}
	if _, err := codegen.NewGenerator(genopts.WithPath(modelPkgPath),
		genopts.WithSerializer(reflect.TypeFor[struct{ X int }](), "X")); err == nil {
		t.Fatal("NewGenerator accepted a serializer name for an unnamed type")
	}
}

// TestNewGeneratorRejectsConflictingAliases checks the conversion error when an
// import alias shadows the generated package's own name.
func TestNewGeneratorRejectsConflictingAliases(t *testing.T) {
	_, err := codegen.NewGenerator(
		genopts.WithPath(modelPkgPath),
		genopts.WithImportAlias("github.com/malivvan/binco/testdata/other", "testdata"),
		genopts.WithDTM(reflect.TypeFor[other.SameName](), 1),
	)
	if err == nil {
		t.Fatal("NewGenerator accepted an alias that shadows the package name")
	}
}
