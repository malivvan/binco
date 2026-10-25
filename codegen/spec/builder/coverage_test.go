package builder_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/codegen/spec"
	bldr "github.com/malivvan/binco/codegen/spec/builder"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
	"github.com/malivvan/binco/testdata"
)

// errorConverter delegates to the real converter but fails for names that
// contain one of the configured markers.
type errorConverter struct {
	inner bldr.TypeNameConverter
	full  string
	rel   string
	err   error
}

func (c errorConverter) ConvertToFullName(name typename.CompleteName) (typename.FullName, error) {
	if c.full != "" && strings.Contains(string(name), c.full) {
		return "", c.err
	}
	return c.inner.ConvertToFullName(name)
}

func (c errorConverter) ConvertToRelativeName(name typename.FullName) (typename.RelativeName, error) {
	if c.rel != "" && strings.Contains(string(name), c.rel) {
		return "", c.err
	}
	return c.inner.ConvertToRelativeName(name)
}

// newErrorBuilder returns a builder whose conversions fail for the markers.
func newErrorBuilder(t *testing.T, full, rel string) bldr.TypeBuilder {
	t.Helper()
	gops := genopts.New()
	if err := genopts.Apply([]genopts.SetOption{genopts.WithPath(pkgPath)}, &gops); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	inner := cnvtr.NewTypeNameConverter(gops)
	return bldr.NewTypeBuilder(errorConverter{
		inner: inner, full: full, rel: rel, err: errors.New("conversion failed"),
	}, gops)
}

// TestBuildDefinedTypeErrors covers the three early returns of BuildDefinedType
// that a valid type never reaches.
func TestBuildDefinedTypeConversionErrors(t *testing.T) {
	b := newBuilder(t)
	if _, err := b.BuildDefinedType(reflect.TypeFor[testdata.Int](),
		tpopts.Options{Validator: "not an expr"}); err == nil {
		t.Fatal("a validator that is not an expression was accepted")
	}
	if _, err := newErrorBuilder(t, "testdata.Int", "").
		BuildDefinedType(reflect.TypeFor[testdata.Int](), tpopts.Options{}); err == nil {
		t.Fatal("a failing name conversion was accepted")
	}
	if _, err := newErrorBuilder(t, "int", "").
		BuildDefinedType(reflect.TypeFor[testdata.Int](), tpopts.Options{}); err == nil {
		t.Fatal("a failing underlying-type conversion was accepted")
	}
}

// TestBuildStructTypeErrors covers the struct checks a valid type never reaches.
func TestBuildStructTypeErrors(t *testing.T) {
	b := newBuilder(t)

	// The field count has to match the struct, or the count check fires first.
	var sops stopts.Options
	stopts.Apply([]stopts.SetOption{
		stopts.WithField(fldopts.WithType(tpopts.WithValidator("not an expr"))),
		stopts.WithField(),
	}, &sops)
	if _, err := b.BuildStructType(reflect.TypeFor[testdata.SimpleStruct](), sops); err == nil {
		t.Fatal("a field validator that is not an expression was accepted")
	}

	if _, err := b.BuildStructType(reflect.TypeFor[testdata.SimpleStruct](),
		stopts.Options{Validator: "not an expr"}); err == nil {
		t.Fatal("a struct validator that is not an expression was accepted")
	}

	// time.Time is a defined struct from another package with unexported fields.
	if _, err := b.BuildStructType(reflect.TypeFor[time.Time](), stopts.Options{}); err == nil {
		t.Fatal("a foreign struct with unexported fields was accepted")
	}

	if _, err := newErrorBuilder(t, "testdata.SimpleStruct", "").
		BuildStructType(reflect.TypeFor[testdata.SimpleStruct](), stopts.Options{}); err == nil {
		t.Fatal("a failing struct-name conversion was accepted")
	}
	if _, err := newErrorBuilder(t, "int", "").
		BuildStructType(reflect.TypeFor[testdata.SimpleStruct](), stopts.Options{}); err == nil {
		t.Fatal("a failing field-name conversion was accepted")
	}

	// A field whose type has no name (a channel) fails while the field names
	// are converted.
	if _, err := b.BuildStructType(reflect.TypeFor[testdata.UnsupportedFieldStruct](),
		stopts.Options{}); err == nil {
		t.Fatal("a struct with an unsupported field type was accepted")
	}
}

// TestBuildVersionedTypeErrors covers the version and migration checks.
func TestBuildVersionedTypeErrors(t *testing.T) {
	b := newBuilder(t)
	if _, err := newErrorBuilder(t, "testdata.VerFoo", "").
		BuildVersionedType(reflect.TypeFor[testdata.VerFoo](), veropts.Options{}); err == nil {
		t.Fatal("a failing name conversion was accepted")
	}

	var vops veropts.Options
	if err := veropts.Apply(&vops,
		veropts.WithVersion(reflect.TypeFor[testdata.VerFooVer1](), "")); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := newErrorBuilder(t, "testdata.VerFooVer1", "").
		BuildVersionedType(reflect.TypeFor[testdata.VerFoo](), vops); err == nil {
		t.Fatal("a failing version conversion was accepted")
	}

	var badMigration veropts.Options
	if err := veropts.Apply(&badMigration,
		veropts.WithVersion(reflect.TypeFor[testdata.VerFooVer1](), "not an expr")); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := b.BuildVersionedType(reflect.TypeFor[testdata.VerFoo](), badMigration); err == nil {
		t.Fatal("a migration that is not an expression was accepted")
	}
}

// TestBuildInterfaceTypeErrors covers the interface and implementation checks.
func TestBuildInterfaceTypeErrors(t *testing.T) {
	if _, err := newErrorBuilder(t, "testdata.Universal", "").
		BuildInterfaceType(reflect.TypeFor[testdata.Universal](), intropts.Options{}); err == nil {
		t.Fatal("a failing interface-name conversion was accepted")
	}

	var iops intropts.Options
	intropts.Apply(&iops, intropts.WithImpl(reflect.TypeFor[testdata.StructImpl]()))
	if _, err := newErrorBuilder(t, "testdata.StructImpl", "").
		BuildInterfaceType(reflect.TypeFor[testdata.Universal](), iops); err == nil {
		t.Fatal("a failing implementation conversion was accepted")
	}
}

// TestBuildTimeTypeErrors covers the time-wrapper checks.
func TestBuildTimeTypeErrors(t *testing.T) {
	b := newBuilder(t)
	if _, err := b.BuildTimeType(reflect.TypeFor[testdata.Time](),
		stopts.UnderlyingTimeOptions{}, "not an expr"); err == nil {
		t.Fatal("a validator that is not an expression was accepted")
	}
	if _, err := newErrorBuilder(t, "testdata.Time", "").
		BuildTimeType(reflect.TypeFor[testdata.Time](), stopts.UnderlyingTimeOptions{}, ""); err == nil {
		t.Fatal("a failing name conversion was accepted")
	}
}

// TestCollectAnonTypesDrivesTheScanner checks the container callbacks: every
// shape records its anonymous serializer.
func TestCollectAnonTypesDrivesTheScanner(t *testing.T) {
	b := newBuilder(t)
	m := map[spec.AnonSerializerName]spec.AnonType{}
	names := []typename.FullName{
		"[]int", "map[string]int", "[2]int", "[]byte", "string", "int",
	}
	for _, name := range names {
		if err := b.CollectAnonTypes(name, m, tpopts.Options{}); err != nil {
			t.Fatalf("CollectAnonTypes(%q): %v", name, err)
		}
	}
	if err := b.CollectAnonTypes("[]byte", m,
		tpopts.Options{LenEnc: tpopts.NumEncodingRaw}); err != nil {
		t.Fatalf("CollectAnonTypes with a length encoding: %v", err)
	}
	if len(m) == 0 {
		t.Fatal("no anonymous serializer was recorded")
	}

	// A validator that is not an expression is rejected before the scan.
	if err := b.CollectAnonTypes("[]int", m,
		tpopts.Options{Validator: "not an expr"}); err == nil {
		t.Fatal("CollectAnonTypes accepted an invalid validator")
	}
	if _, _, err := b.BuildAnonType("[]int",
		tpopts.Options{Validator: "not an expr"}); err == nil {
		t.Fatal("BuildAnonType accepted an invalid validator")
	}
}

// TestAnonTypeValidatorErrors checks that a validator whose type cannot be
// converted is reported for every container shape.
func TestAnonTypeValidatorErrors(t *testing.T) {
	cases := []struct {
		name string
		rel  string
		tops tpopts.Options
	}{
		{"map[string]int", "int", tpopts.Options{LenValidator: "Validate"}},
		{"map[string]int", "string", tpopts.Options{KeyValidator: "Validate"}},
		{"map[string]int", "int", tpopts.Options{ElemValidator: "Validate"}},
		{"[]int", "int", tpopts.Options{LenValidator: "Validate"}},
		{"[]int", "int", tpopts.Options{ElemValidator: "Validate"}},
		{"[]byte", "int", tpopts.Options{LenValidator: "Validate"}},
		{"[2]int", "int", tpopts.Options{ElemValidator: "Validate"}},
		{"string", "int", tpopts.Options{LenValidator: "Validate"}},
	}
	for _, tc := range cases {
		b := newErrorBuilder(t, "", tc.rel)
		if err := b.CollectAnonTypes(typename.FullName(tc.name), map[spec.AnonSerializerName]spec.AnonType{}, tc.tops); err == nil {
			t.Fatalf("CollectAnonTypes(%q, %+v) returned nil", tc.name, tc.tops)
		}
	}

	// The probe used for a single field reports the same error.
	b := newErrorBuilder(t, "", "int")
	if _, _, err := b.BuildAnonType("[]int",
		tpopts.Options{LenValidator: "Validate"}); err == nil {
		t.Fatal("BuildAnonType accepted an unconvertible validator type")
	}
}
