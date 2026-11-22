package codegen_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/codegen"
	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	"github.com/malivvan/binco/testdata"
)

// registerAny declares the union the empty interface dispatches over.
func registerAny(t *testing.T, g *codegen.Generator) {
	t.Helper()
	if err := g.RegisterInterface(reflect.TypeFor[any](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyRecord]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.AnyNumber]()),
	); err != nil {
		t.Fatalf("RegisterInterface(any): %v", err)
	}
}

func TestGenerateAnyInterface(t *testing.T) {
	g := newGenerator(t)
	registerAny(t, g)
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, g)

	mustContain(t, src,
		"var anyBinco = anyBincoImpl{}",
		"func (s anyBincoImpl) Marshal(v any, bs []byte) (n int)",
		"func (s anyBincoImpl) Unmarshal(bs []byte) (v any, n int, err error)",
		"return typeid.OptionalDTMSerializer.Marshal(0, bs)",
		"dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(bs)",
		"case 0:\n\t\t// The writer stored a nil interface value.\n\t\treturn nil, n, nil",
		"case AnyRecordDTM:",
		"case AnyNumberDTM:",
		"var AnyRecordTypedBinco = typeid.NewSerializer(AnyRecordDTM, AnyRecordBinco)",
		"anyBinco.Marshal(v.Payload, bs[n:])",
		"composite.NewMapSerializer[string, any](composite.String, anyBinco",
		"composite.NewSliceSerializer[any](anyBinco)",
	)
}

// TestGenerateAnyStructFieldOrder checks that the generated struct codec walks
// the fields in declaration order.
func TestGenerateAnyStructFieldOrder(t *testing.T) {
	g := newGenerator(t)
	registerAny(t, g)
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, g)
	want := []string{
		"v.Kind",
		"anyBinco.Marshal(v.Payload",
		"v.Meta",
		"v.Items",
	}
	last := -1
	for _, fragment := range want {
		i := strings.Index(src, fragment)
		if i < 0 {
			t.Fatalf("generated code does not contain %q", fragment)
		}
		if i <= last {
			t.Fatalf("%q is out of field order", fragment)
		}
		last = i
	}
}

// TestAddStructAnyNeedsRegistration checks that a type holding any cannot be
// added before the union exists.
func TestAddStructAnyNeedsRegistration(t *testing.T) {
	g := newGenerator(t)
	err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent]())
	if !errors.Is(err, codegen.ErrNoAnySerializer) {
		t.Fatalf("AddStruct = %v, want ErrNoAnySerializer", err)
	}
	if !strings.Contains(err.Error(), "Payload") {
		t.Fatalf("error %q does not name the field", err)
	}
}

// TestAddDefinedTypeAnyNeedsRegistration checks the same rule for a defined type
// whose underlying type is a container of any.
func TestAddDefinedTypeAnyNeedsRegistration(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddDefinedType(reflect.TypeFor[testdata.AnyList]()); !errors.Is(err, codegen.ErrNoAnySerializer) {
		t.Fatalf("AddDefinedType without the union = %v, want ErrNoAnySerializer", err)
	}

	g = newGenerator(t)
	registerAny(t, g)
	if err := g.AddDefinedType(reflect.TypeFor[testdata.AnyList]()); err != nil {
		t.Fatalf("AddDefinedType: %v", err)
	}
	mustContain(t, generate(t, g),
		"composite.NewSliceSerializer[any](anyBinco)",
		"[]any(v)",
	)
}

// TestRegisterAnyBeforeItsImpls checks the order rule inside RegisterInterface:
// an implementation may itself hold the empty interface.
func TestRegisterAnyBeforeItsImpls(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[any](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyNode]()),
	); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	mustContain(t, generate(t, g), "AnyNodeBinco", "anyBinco")
}

// TestRegisterAnyWithoutImpls checks that a failing registration does not make a
// later AddStruct succeed.
func TestRegisterAnyWithoutImpls(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[any]()); err == nil {
		t.Fatal("RegisterInterface(any) accepted an empty implementation list")
	}
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent]()); !errors.Is(err, codegen.ErrNoAnySerializer) {
		t.Fatalf("AddStruct = %v, want ErrNoAnySerializer", err)
	}
}

// TestGenerateAnyIgnoredField checks that an ignored any field needs no union.
func TestGenerateAnyIgnoredField(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent](),
		stopts.WithField(fldopts.WithIgnore()),
		stopts.WithField(fldopts.WithIgnore()),
		stopts.WithField(fldopts.WithIgnore()),
		stopts.WithField(fldopts.WithIgnore()),
	); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	mustContain(t, generate(t, g), "func (s anyEventBinco) Marshal(v AnyEvent, bs []byte) (n int) {\n\treturn\n}")
}

// TestGenerateStreamAny checks the streaming shape of the dispatcher.
func TestGenerateStreamAny(t *testing.T) {
	g := newGenerator(t, genopts.WithStream())
	registerAny(t, g)
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyEvent]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, g)
	mustContain(t, src,
		"func (s anyBincoImpl) Marshal(v any, w wirefmt.Writer) (n int, err error)",
		"func (s anyBincoImpl) Unmarshal(r wirefmt.Reader) (v any, n int, err error)",
		"return typeid.OptionalDTMSerializer.Marshal(0, w)",
		"anyBinco.Marshal(v.Payload, w)",
		"\"github.com/malivvan/binco/wirefmt/typeid\"",
	)
}

// TestAnyMapKeyIsRejectedByAddStruct checks the map-key rule where a caller
// meets it: while the struct is added.
func TestAnyMapKeyIsRejectedByAddStruct(t *testing.T) {
	g := newGenerator(t)
	registerAny(t, g)
	err := g.AddStruct(reflect.TypeFor[testdata.AnyKeyedMap]())
	if err == nil {
		t.Fatal("AddStruct accepted a map with an interface key")
	}
	if !strings.Contains(err.Error(), "map key") {
		t.Fatalf("error %q does not explain the map-key rule", err)
	}
}

// TestAnyDTMNumberingAndRegistry checks the identifiers of the implementations.
func TestAnyDTMNumberingAndRegistry(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	registerAny(t, g)
	mustContain(t, generate(t, g),
		"StructImplDTM  binco.DTM = 1",
		"DefinedImplDTM binco.DTM = 2",
		"AnyRecordDTM binco.DTM = 3",
		"AnyNumberDTM binco.DTM = 4",
	)
	registry, err := g.DTMRegistry()
	if err != nil {
		t.Fatalf("DTMRegistry: %v", err)
	}
	if _, ok := registry["any"]; ok {
		t.Fatal("the empty interface has an identifier of its own")
	}
	for _, name := range []string{
		"testdata.StructImpl", "testdata.DefinedImpl",
		"testdata.AnyRecord", "testdata.AnyNumber",
	} {
		if _, ok := registry[name]; !ok {
			t.Fatalf("registry has no identifier for %s: %v", name, registry)
		}
	}
}

// TestAnySharesAnImplementationIsRejected documents the existing rule: an
// implementation belongs to one union because its identifier constant is named
// after the type.
func TestAnySharesAnImplementationIsRejected(t *testing.T) {
	g := newGenerator(t)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.AnyRecord]())); err != nil {
		t.Fatalf("RegisterInterface(Universal): %v", err)
	}
	registerAny(t, g)
	if _, err := g.Generate(); !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("Generate = %v, want ErrDuplicateDTM", err)
	}
}

// TestGenerateAnyPointerField checks the last container shape: a pointer whose
// element is the empty interface.
func TestGenerateAnyPointerField(t *testing.T) {
	g := newGenerator(t)
	registerAny(t, g)
	if err := g.AddStruct(reflect.TypeFor[testdata.AnyPointer]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	mustContain(t, generate(t, g), "composite.NewPtrSerializer[any](anyBinco)")
}
