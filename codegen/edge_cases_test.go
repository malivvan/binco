package codegen_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen"
	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	"github.com/malivvan/binco/testdata"
	"github.com/malivvan/binco/testdata/other"
)

// TestAllFieldsIgnored covers a struct whose every field is left out: the
// generated serializer must still be valid Go and encode nothing.
func TestAllFieldsIgnored(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[testdata.IgnoredFieldStruct](),
		stopts.WithField(fldopts.WithIgnore()),
		stopts.WithField(fldopts.WithIgnore()),
	); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, g)
	mustContain(t, src,
		"func (s ignoredFieldStructBinco) Marshal(v IgnoredFieldStruct, bs []byte) (n int) {",
		"func (s ignoredFieldStructBinco) Unmarshal(bs []byte) (v IgnoredFieldStruct, n int, err error) {",
	)
}

// TestSameTypeNameInTwoPackages covers D7: two types with the same bare name
// used to collapse into one generated serializer. They are rejected unless one
// of them is given an explicit name.
func TestSameTypeNameInTwoPackages(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[testdata.SameName]()); err != nil {
		t.Fatalf("AddStruct(testdata.SameName): %v", err)
	}
	if err := g.AddStruct(reflect.TypeFor[other.SameName]()); err != nil {
		t.Fatalf("AddStruct(other.SameName): %v", err)
	}
	if _, err := g.Generate(); !errors.Is(err, codegen.ErrNameCollision) {
		t.Fatalf("Generate returned %v, want ErrNameCollision", err)
	}

	// An explicit name resolves the collision.
	fixed := newGenerator(t,
		genopts.WithSerializer(reflect.TypeFor[other.SameName](), "OtherSameName"))
	if err := fixed.AddStruct(reflect.TypeFor[testdata.SameName]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	if err := fixed.AddStruct(reflect.TypeFor[other.SameName]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, fixed)
	mustContain(t, src, "OtherSameNameBinco", "SameNameBinco")
}
