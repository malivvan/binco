package builder_test

import (
	"reflect"
	"strings"
	"testing"

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

const pkgPath = "github.com/malivvan/binco/testdata"

func typenameFullName(s string) typename.FullName { return typename.FullName(s) }

func newBuilder(t *testing.T) bldr.TypeBuilder {
	t.Helper()
	gops := genopts.New()
	if err := genopts.Apply([]genopts.SetOption{genopts.WithPath(pkgPath)}, &gops); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return bldr.NewTypeBuilder(cnvtr.NewTypeNameConverter(gops), gops)
}

func TestBuildDefinedType(t *testing.T) {
	b := newBuilder(t)

	got, err := b.BuildDefinedType(reflect.TypeFor[testdata.Int](),
		tpopts.Options{NumEncoding: tpopts.NumEncodingRaw})
	if err != nil {
		t.Fatalf("BuildDefinedType: %v", err)
	}
	if got.FullName != "testdata.Int" {
		t.Fatalf("FullName = %q", got.FullName)
	}
	if got.UnderlyingTypeName != "int" {
		t.Fatalf("UnderlyingTypeName = %q", got.UnderlyingTypeName)
	}
	if got.Tops.NumEncoding != tpopts.NumEncodingRaw {
		t.Fatalf("Tops were not copied: %+v", got.Tops)
	}
	if got.Gops.PkgPath != pkgPath {
		t.Fatalf("Gops were not copied: %+v", got.Gops)
	}

	for _, tc := range []struct {
		typ  reflect.Type
		want string
	}{
		{reflect.TypeFor[testdata.IntSlice](), "[]int"},
		{reflect.TypeFor[testdata.Str](), "string"},
	} {
		got, err = b.BuildDefinedType(tc.typ, tpopts.Options{})
		if err != nil {
			t.Fatalf("BuildDefinedType(%v): %v", tc.typ, err)
		}
		if string(got.UnderlyingTypeName) != tc.want {
			t.Fatalf("UnderlyingTypeName = %q, want %q", got.UnderlyingTypeName, tc.want)
		}
	}
}

func TestBuildDefinedTypeErrors(t *testing.T) {
	b := newBuilder(t)
	if _, err := b.BuildDefinedType(reflect.TypeFor[testdata.SimpleStruct](),
		tpopts.Options{}); err == nil {
		t.Fatal("a struct was accepted as a defined type")
	}
	if _, err := b.BuildDefinedType(reflect.TypeFor[int](),
		tpopts.Options{}); err == nil {
		t.Fatal("a primitive was accepted as a defined type")
	}
	if _, err := b.BuildDefinedType(reflect.TypeFor[testdata.Universal](),
		tpopts.Options{}); err == nil {
		t.Fatal("an interface was accepted as a defined type")
	}
}

func TestBuildStructType(t *testing.T) {
	b := newBuilder(t)
	got, err := b.BuildStructType(reflect.TypeFor[testdata.SimpleStruct](), stopts.Options{})
	if err != nil {
		t.Fatalf("BuildStructType: %v", err)
	}
	if got.FullName != "testdata.SimpleStruct" {
		t.Fatalf("FullName = %q", got.FullName)
	}
	if len(got.Fields) != 2 {
		t.Fatalf("Fields = %d, want 2", len(got.Fields))
	}
	if got.Fields[0].FieldName != "Num" || got.Fields[0].FullName != "int" {
		t.Fatalf("unexpected first field %+v", got.Fields[0])
	}
	if got.Fields[1].FieldName != "Str" || got.Fields[1].FullName != "string" {
		t.Fatalf("unexpected second field %+v", got.Fields[1])
	}

	sops := stopts.Options{}
	stopts.Apply([]stopts.SetOption{
		stopts.WithField(),
		stopts.WithField(fldopts.WithIgnore()),
	}, &sops)
	if _, err = b.BuildStructType(reflect.TypeFor[testdata.SimpleStruct](), sops); err != nil {
		t.Fatalf("BuildStructType with field options: %v", err)
	}

	if _, err = b.BuildStructType(reflect.TypeFor[testdata.SimpleStruct](),
		stopts.Options{Fields: []fldopts.Options{{}}}); err == nil {
		t.Fatal("a wrong field count was accepted")
	} else if !strings.Contains(err.Error(), "want 2") {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err = b.BuildStructType(reflect.TypeFor[testdata.Int](), stopts.Options{}); err == nil {
		t.Fatal("a defined int was accepted as a struct")
	}
}

func TestBuildTyped(t *testing.T) {
	b := newBuilder(t)
	for _, typ := range []reflect.Type{
		reflect.TypeFor[testdata.Int](),
		reflect.TypeFor[testdata.IntSlice](),
		reflect.TypeFor[testdata.SimpleStruct](),
		reflect.TypeFor[testdata.Universal](),
	} {
		if _, err := b.BuildTyped(typ); err != nil {
			t.Fatalf("BuildTyped(%v): %v", typ, err)
		}
	}
	if _, err := b.BuildTyped(reflect.TypeFor[any]()); err == nil {
		t.Fatal("an empty interface was accepted as a typed type")
	}
}

func TestBuildInterfaceType(t *testing.T) {
	b := newBuilder(t)
	iops := intropts.Options{}
	intropts.Apply(&iops,
		intropts.WithImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithImpls(reflect.TypeFor[testdata.DefinedImpl]()),
	)
	got, err := b.BuildInterfaceType(reflect.TypeFor[testdata.Universal](), iops)
	if err != nil {
		t.Fatalf("BuildInterfaceType: %v", err)
	}
	if got.FullName != "testdata.Universal" {
		t.Fatalf("FullName = %q", got.FullName)
	}
	if len(got.Impls) != 2 || got.Impls[0] != "testdata.StructImpl" ||
		got.Impls[1] != "testdata.DefinedImpl" {
		t.Fatalf("Impls = %v", got.Impls)
	}
	if _, err = b.BuildInterfaceType(reflect.TypeFor[testdata.Int](), iops); err == nil {
		t.Fatal("a defined int was accepted as an interface")
	}
}

func TestBuildVersionedType(t *testing.T) {
	b := newBuilder(t)
	vops := veropts.Options{}
	if err := veropts.Apply(&vops,
		veropts.WithVersion(reflect.TypeFor[testdata.VerFooVer1](), "MigrateVerFooVer1"),
		veropts.WithCurrentVersion(reflect.TypeFor[testdata.VerFooVer2]()),
	); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := b.BuildVersionedType(reflect.TypeFor[testdata.VerFoo](), vops)
	if err != nil {
		t.Fatalf("BuildVersionedType: %v", err)
	}
	if got.FullName != "testdata.VerFoo" {
		t.Fatalf("FullName = %q", got.FullName)
	}
	if len(got.Versions) != 2 || got.Versions[0] != "testdata.VerFooVer1" ||
		got.Versions[1] != "testdata.VerFooVer2" {
		t.Fatalf("Versions = %v", got.Versions)
	}
	if _, err = b.BuildVersionedType(reflect.TypeFor[int](), vops); err == nil {
		t.Fatal("a primitive was accepted as a versioned type")
	}
}

func TestBuildTimeType(t *testing.T) {
	b := newBuilder(t)
	got, err := b.BuildTimeType(reflect.TypeFor[testdata.Time](),
		stopts.UnderlyingTimeOptions{TimeUnit: tpopts.TimeUnitMilli}, "ValidateInt")
	if err != nil {
		t.Fatalf("BuildTimeType: %v", err)
	}
	if got.UnderlyingTypeName != "time.Time" {
		t.Fatalf("UnderlyingTypeName = %q", got.UnderlyingTypeName)
	}
	if got.Tops.TimeUnit != tpopts.TimeUnitMilli || got.Tops.Validator != "ValidateInt" {
		t.Fatalf("unexpected tops %+v", got.Tops)
	}
	if _, err = b.BuildTimeType(reflect.TypeFor[testdata.Int](),
		stopts.UnderlyingTimeOptions{}, ""); err == nil {
		t.Fatal("a defined int was accepted as a time type")
	}
}

func TestBuildAnonType(t *testing.T) {
	b := newBuilder(t)

	got, ok, err := b.BuildAnonType("[]int", tpopts.Options{})
	if err != nil || !ok {
		t.Fatalf("BuildAnonType([]int) = (%v, %v)", ok, err)
	}
	if got.Kind != spec.AnonKindSlice || got.ElemType != "int" {
		t.Fatalf("unexpected anon type %+v", got)
	}

	got, ok, err = b.BuildAnonType("*int", tpopts.Options{})
	if err != nil || !ok || got.Kind != spec.AnonKindPtr {
		t.Fatalf("BuildAnonType(*int) = (%+v, %v, %v)", got, ok, err)
	}

	if _, ok, err = b.BuildAnonType("int", tpopts.Options{}); err != nil || ok {
		t.Fatalf("BuildAnonType(int) = (%v, %v)", ok, err)
	}

	if _, ok, err = b.BuildAnonType("string",
		tpopts.Options{LenValidator: "Validate"}); err != nil || !ok {
		t.Fatalf("BuildAnonType(string with validator) = (%v, %v)", ok, err)
	}

	if _, _, err = b.BuildAnonType("struct{}", tpopts.Options{}); err == nil {
		t.Fatal("an unsupported type was accepted")
	}
}

func TestCollectAnonTypes(t *testing.T) {
	b := newBuilder(t)
	m := map[spec.AnonSerializerName]spec.AnonType{}
	if err := b.CollectAnonTypes("map[string][]int", m, tpopts.Options{}); err != nil {
		t.Fatalf("CollectAnonTypes: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("collected %d anonymous types, want 2 (map and slice): %v", len(m), m)
	}
	kinds := map[spec.AnonKind]bool{}
	for _, anonType := range m {
		kinds[anonType.Kind] = true
	}
	if !kinds[spec.AnonKindMap] || !kinds[spec.AnonKindSlice] {
		t.Fatalf("unexpected kinds %v", kinds)
	}

	if err := b.CollectAnonTypes("chan int", map[spec.AnonSerializerName]spec.AnonType{},
		tpopts.Options{}); err == nil {
		t.Fatal("an unsupported type was accepted")
	}
}

func TestBase64KeywordEncoding(t *testing.T) {

	got := bldr.Base64KeywordEncoding.EncodeToString([]byte{0xFB, 0xFF})
	if strings.ContainsAny(got, "+/=") {
		t.Fatalf("encoded %q still contains a base64 keyword", got)
	}
	if got != "ΣΔ8Ξ" {
		t.Fatalf("encoded = %q, want ΣΔ8Ξ", got)
	}
}

func TestCollectContainerAnonTypes(t *testing.T) {
	b := newBuilder(t)
	tests := []struct {
		name  string
		scan  string
		tops  tpopts.Options
		kinds []spec.AnonKind
	}{
		{"array", "[3]int", tpopts.Options{}, []spec.AnonKind{spec.AnonKindArray}},
		{"array with options", "[3]int", tpopts.Options{
			LenEnc: tpopts.NumEncodingRaw, ElemValidator: "ValidateElem",
		}, []spec.AnonKind{spec.AnonKindArray}},
		{"byte slice with validator", "[]byte", tpopts.Options{
			LenValidator: "ValidateLength",
		}, []spec.AnonKind{spec.AnonKindByteSlice}},
		{"byte slice without options", "[]byte", tpopts.Options{}, nil},
		{"string with length encoding", "string", tpopts.Options{
			LenEnc: tpopts.NumEncodingRaw,
		}, []spec.AnonKind{spec.AnonKindString}},
		{"string with validator", "string", tpopts.Options{
			LenValidator: "ValidateLength",
		}, []spec.AnonKind{spec.AnonKindString}},
		{"string without options", "string", tpopts.Options{}, nil},
		{"slice with all options", "[]int", tpopts.Options{
			LenEnc:        tpopts.NumEncodingRaw,
			LenValidator:  "ValidateLength",
			ElemValidator: "ValidateElem",
		}, []spec.AnonKind{spec.AnonKindSlice}},
		{"map with all options", "map[string]int", tpopts.Options{
			LenEnc:        tpopts.NumEncodingRaw,
			LenValidator:  "ValidateLength",
			KeyValidator:  "ValidateKey",
			ElemValidator: "ValidateElem",
		}, []spec.AnonKind{spec.AnonKindMap}},
		{"generic type", "pkg.Type[pkg.A, pkg.B]", tpopts.Options{}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := map[spec.AnonSerializerName]spec.AnonType{}
			if err := b.CollectAnonTypes(typenameFullName(tc.scan), m, tc.tops); err != nil {
				t.Fatalf("CollectAnonTypes: %v", err)
			}
			got := map[spec.AnonKind]bool{}
			for _, anonType := range m {
				got[anonType.Kind] = true
			}
			if len(got) != len(tc.kinds) {
				t.Fatalf("collected %d kinds, want %d: %v", len(got), len(tc.kinds), got)
			}
			for _, kind := range tc.kinds {
				if !got[kind] {
					t.Fatalf("kind %v was not collected: %v", kind, got)
				}
			}
		})
	}
}

func TestBuildStructTypeForInterface(t *testing.T) {
	b := newBuilder(t)
	if _, err := b.BuildStructType(reflect.TypeFor[testdata.Universal](),
		stopts.Options{}); err == nil {
		t.Fatal("an interface was accepted as a struct")
	}

	if _, err := b.BuildStructType(reflect.TypeFor[chan int](),
		stopts.Options{}); err == nil {
		t.Fatal("a channel was accepted as a struct")
	}
}

func TestBuildInterfaceTypeForStruct(t *testing.T) {
	b := newBuilder(t)
	if _, err := b.BuildInterfaceType(reflect.TypeFor[testdata.SimpleStruct](),
		intropts.Options{}); err == nil {
		t.Fatal("a struct was accepted as an interface")
	}
	if _, err := b.BuildInterfaceType(reflect.TypeFor[chan int](),
		intropts.Options{}); err == nil {
		t.Fatal("a channel was accepted as an interface")
	}
}

// TestBuildInterfaceTypeForAny checks that the predeclared empty interface is
// accepted with its implementations.
func TestBuildInterfaceTypeForAny(t *testing.T) {
	b := newBuilder(t)
	iops := intropts.Options{}
	intropts.Apply(&iops,
		intropts.WithImpl(reflect.TypeFor[testdata.AnyRecord]()),
		intropts.WithImpls(reflect.TypeFor[testdata.AnyNumber]()),
	)
	got, err := b.BuildInterfaceType(reflect.TypeFor[any](), iops)
	if err != nil {
		t.Fatalf("BuildInterfaceType(any): %v", err)
	}
	if got.FullName != "any" {
		t.Fatalf("FullName = %q, want any", got.FullName)
	}
	if len(got.Impls) != 2 || got.Impls[0] != "testdata.AnyRecord" ||
		got.Impls[1] != "testdata.AnyNumber" {
		t.Fatalf("Impls = %v", got.Impls)
	}
}

// TestBuildInterfaceTypeAnyNeedsImpls checks that a union with no
// implementations is refused: it could only encode nil.
func TestBuildInterfaceTypeAnyNeedsImpls(t *testing.T) {
	b := newBuilder(t)
	_, err := b.BuildInterfaceType(reflect.TypeFor[any](), intropts.Options{})
	if err == nil {
		t.Fatal("BuildInterfaceType(any) accepted an empty implementation list")
	}
	if !strings.Contains(err.Error(), "WithStructImpl") {
		t.Fatalf("error %q does not name the option that fixes it", err)
	}
}

// TestBuildStructTypeWithAnyFields checks the names of the fields that hold the
// empty interface.
func TestBuildStructTypeWithAnyFields(t *testing.T) {
	b := newBuilder(t)
	got, err := b.BuildStructType(reflect.TypeFor[testdata.AnyEvent](), stopts.Options{})
	if err != nil {
		t.Fatalf("BuildStructType: %v", err)
	}
	want := []typename.FullName{"string", "any", "map[string]any", "[]any"}
	if len(got.Fields) != len(want) {
		t.Fatalf("fields = %v", got.Fields)
	}
	for i, name := range want {
		if got.Fields[i].FullName != name {
			t.Fatalf("field %d = %q, want %q", i, got.Fields[i].FullName, name)
		}
	}
}

// TestUsesAny checks the probe the generator uses to require the union before a
// type that holds it.
func TestUsesAny(t *testing.T) {
	b := newBuilder(t)
	for _, name := range []typename.FullName{
		"any", "*any", "[]any", "map[string]any", "pkg.Pair[any, int]",
	} {
		if !b.UsesAny(name) {
			t.Fatalf("UsesAny(%q) = false", name)
		}
	}
	for _, name := range []typename.FullName{
		"int", "[]string", "pkg.Type", "chan int",
	} {
		if b.UsesAny(name) {
			t.Fatalf("UsesAny(%q) = true", name)
		}
	}
}

// TestCollectAnyAnonTypes checks that any adds no serializer of its own but that
// the containers around it still get one.
func TestCollectAnyAnonTypes(t *testing.T) {
	b := newBuilder(t)
	tests := []struct {
		name  typename.FullName
		count int
		kind  spec.AnonKind
		elem  typename.FullName
	}{
		{"any", 0, spec.AnonKindUndefined, ""},
		{"[]any", 1, spec.AnonKindSlice, "any"},
		{"map[string]any", 1, spec.AnonKindMap, "any"},
		{"*any", 1, spec.AnonKindPtr, "any"},
	}
	for _, tc := range tests {
		t.Run(string(tc.name), func(t *testing.T) {
			m := map[spec.AnonSerializerName]spec.AnonType{}
			if err := b.CollectAnonTypes(tc.name, m, tpopts.Options{}); err != nil {
				t.Fatalf("CollectAnonTypes: %v", err)
			}
			if len(m) != tc.count {
				t.Fatalf("collected %d anonymous types, want %d: %v", len(m), tc.count, m)
			}
			for _, anonType := range m {
				if anonType.Kind != tc.kind || anonType.ElemType != tc.elem {
					t.Fatalf("unexpected anonymous type %+v", anonType)
				}
			}
		})
	}
}

// TestAnyMapKeyIsRejected checks the map-key rule, including a key that nests
// the empty interface in an array.
func TestAnyMapKeyIsRejected(t *testing.T) {
	b := newBuilder(t)
	for _, name := range []typename.FullName{"map[any]int", "map[[2]any]int"} {
		if _, _, err := b.BuildAnonType(name, tpopts.Options{}); err == nil {
			t.Fatalf("BuildAnonType(%q) accepted an interface key", name)
		}
		if err := b.CollectAnonTypes(name,
			map[spec.AnonSerializerName]spec.AnonType{}, tpopts.Options{}); err == nil {
			t.Fatalf("CollectAnonTypes(%q) accepted an interface key", name)
		}
	}
}

// TestAnonNamesDependOnEveryDimension checks that two anonymous serializers that
// differ in something the generated code reads get different names.
//
// The name used to include the package of the length codec rather than the codec
// itself, and an array's length was missing entirely. [3]int and [4]int then
// shared one serializer (the generated code did not compile), and a slice with
// the default length shared one with a slice whose length is written by
// varnum.Int (silently changing the bytes).
func TestAnonNamesDependOnEveryDimension(t *testing.T) {
	b := newBuilder(t)
	collect := func(name typename.FullName, opts tpopts.Options) map[spec.AnonSerializerName]spec.AnonType {
		t.Helper()
		m := map[spec.AnonSerializerName]spec.AnonType{}
		if err := b.CollectAnonTypes(name, m, opts); err != nil {
			t.Fatalf("CollectAnonTypes(%q): %v", name, err)
		}
		return m
	}
	onlyName := func(m map[spec.AnonSerializerName]spec.AnonType) spec.AnonSerializerName {
		t.Helper()
		if len(m) != 1 {
			t.Fatalf("collected %d anonymous types, want 1: %v", len(m), m)
		}
		for name := range m {
			return name
		}
		return ""
	}

	positive := tpopts.Options{LenEnc: tpopts.NumEncodingVarintPositive}
	signed := tpopts.Options{LenEnc: tpopts.NumEncodingVarint}
	for _, tc := range []struct {
		left, right typename.FullName
		leftOpts    tpopts.Options
		rightOpts   tpopts.Options
	}{
		// Different array lengths.
		{"[3]int", "[4]int", tpopts.Options{}, tpopts.Options{}},
		// The same container with two length codecs that live in one package:
		// varnum.PositiveInt and varnum.Int.
		{"[]int", "[]int", positive, signed},
		{"[]byte", "[]byte", positive, signed},
		{"map[string]int", "map[string]int", positive, signed},
		// A different validator.
		{"[]int", "[]int", tpopts.Options{},
			tpopts.Options{ElemValidator: "ValidateElem"}},
	} {
		left := onlyName(collect(tc.left, tc.leftOpts))
		right := onlyName(collect(tc.right, tc.rightOpts))
		if left == right {
			t.Fatalf("%q with %+v and %q with %+v share the name %q",
				tc.left, tc.leftOpts, tc.right, tc.rightOpts, left)
		}
	}
}
