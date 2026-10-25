package codegen

import (
	"reflect"
	"testing"
	"text/template"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/spec"
	bldr "github.com/malivvan/binco/codegen/spec/builder"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
	"github.com/malivvan/binco/testdata"
)

const targetPkg = "example.com/target"

func newTestSerializerFns(t *testing.T, opts ...genopts.SetOption) SerializerFns {
	t.Helper()
	opts = append([]genopts.SetOption{genopts.WithPath(targetPkg)}, opts...)
	gops := genopts.New()
	if err := genopts.Apply(opts, &gops); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	conv := cnvtr.NewTypeNameConverter(gops)
	fns, err := NewSerializerFns(conv, bldr.NewTypeBuilder(conv, gops),
		map[typename.FullName]struct{}{}, TmplFns{}, gops)
	if err != nil {
		t.Fatalf("NewSerializerFns: %v", err)
	}
	return fns
}

// mustName returns a helper that fails the test when a naming helper returned
// an error. It takes the (value, error) pair as its only argument, which is the
// one place Go allows a multi-value call to be expanded.
func mustName(t *testing.T) func(string, error) string {
	return func(value string, err error) string {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return value
	}
}

func TestSerializerFnsNames(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t)
	name := typename.FullName("target.Foo")

	if got := must(f.Var(name)); got != "FooBinco" {
		t.Fatalf("Var = %q", got)
	}
	if got := must(f.TypedVar(name)); got != "FooTypedBinco" {
		t.Fatalf("TypedVar = %q", got)
	}
	if got := must(f.DTMVar(name)); got != "FooDTM" {
		t.Fatalf("DTMVar = %q", got)
	}
	if got := must(f.Type(name)); got != "fooBinco" {
		t.Fatalf("Type = %q", got)
	}
	if got := must(f.SchemaVar(name)); got != "FooSchemaFingerprint" {
		t.Fatalf("SchemaVar = %q", got)
	}
	if got := f.Val(struct{}{}); got != "v" {
		t.Fatalf("Val = %q", got)
	}
}

func TestSerializerFnsOf(t *testing.T) {
	safe := newTestSerializerFns(t)
	tests := []struct {
		name string
		tops tpopts.Options
		want string
	}{
		{"int", tpopts.Options{}, "varnum.Int"},
		{"int64", tpopts.Options{}, "varnum.Int64"},
		{"int32", tpopts.Options{}, "varnum.Int32"},
		{"int16", tpopts.Options{}, "varnum.Int16"},
		{"int", tpopts.Options{NumEncoding: tpopts.NumEncodingRaw}, "scalar.Int"},
		{"int", tpopts.Options{NumEncoding: tpopts.NumEncodingVarintPositive},
			"varnum.PositiveInt"},
		{"int8", tpopts.Options{}, "scalar.Int8"},
		{"uint8", tpopts.Options{}, "scalar.Uint8"},
		{"uint", tpopts.Options{}, "varnum.Uint"},
		{"byte", tpopts.Options{}, "scalar.Byte"},
		{"float64", tpopts.Options{}, "scalar.Float64"},
		{"float32", tpopts.Options{}, "scalar.Float32"},
		{"bool", tpopts.Options{}, "composite.Bool"},
		{"string", tpopts.Options{}, "composite.String"},
		{"[]byte", tpopts.Options{}, "composite.ByteSlice"},
		{"[]uint8", tpopts.Options{}, "composite.ByteSlice"},
		{"time.Time", tpopts.Options{}, "scalar.TimeUnixUTC"},
		{"time.Time", tpopts.Options{TimeUnit: tpopts.TimeUnitNano}, "scalar.TimeUnixNano"},
		{"target.Foo", tpopts.Options{}, "FooBinco"},
	}
	for _, tc := range tests {
		value, err := safe.Of(typename.FullName(tc.name), tc.tops, safe.gopts)
		if err != nil {
			t.Fatalf("Of(%q, %+v): %v", tc.name, tc.tops, err)
		}
		if value != tc.want {
			t.Fatalf("Of(%q, %+v) = %q, want %q", tc.name, tc.tops, value, tc.want)
		}
	}

	unsafeFns := newTestSerializerFns(t, genopts.WithUnsafe())
	for _, tc := range []struct{ name, want string }{
		{"int", "zerocpy.Int"},
		{"int8", "zerocpy.Int8"},
		{"uint64", "zerocpy.Uint64"},
		{"float64", "zerocpy.Float64"},
		{"bool", "zerocpy.Bool"},
		{"string", "zerocpy.String"},
		{"[]byte", "zerocpy.ByteSlice"},
		{"time.Time", "zerocpy.TimeUnixUTC"},
	} {
		value, err := unsafeFns.Of(typename.FullName(tc.name), tpopts.Options{}, unsafeFns.gopts)
		if err != nil {
			t.Fatalf("unsafe Of(%q): %v", tc.name, err)
		}
		if value != tc.want {
			t.Fatalf("unsafe Of(%q) = %q, want %q", tc.name, value, tc.want)
		}
	}

	notUnsafe := newTestSerializerFns(t, genopts.WithNotUnsafe())
	if value, err := notUnsafe.Of("int", tpopts.Options{}, notUnsafe.gopts); err != nil || value != "zerocpy.Int" {
		t.Fatalf("not-unsafe Of(int) = (%q, %v)", value, err)
	}
	if value, err := notUnsafe.Of("string", tpopts.Options{}, notUnsafe.gopts); err != nil || value != "composite.String" {
		t.Fatalf("not-unsafe Of(string) = (%q, %v)", value, err)
	}
}

func TestSerializerFnsOfReturnsError(t *testing.T) {
	f := newTestSerializerFns(t)
	if _, err := f.Of("struct{}", tpopts.Options{}, f.gopts); err == nil {
		t.Fatal("Of of an unsupported type returned no error")
	}
}

func TestSerializerFnsKeyElem(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t)
	anon := spec.AnonType{KeyType: "string", ElemType: "int"}
	if got := must(f.Key(anon, f.gopts)); got != "composite.String" {
		t.Fatalf("Key = %q", got)
	}
	if got := must(f.Elem(anon, f.gopts)); got != "varnum.Int" {
		t.Fatalf("Elem = %q", got)
	}
}

func TestSerializerFnsRelName(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t)
	if got := must(f.RelName("target.Foo")); got != "Foo" {
		t.Fatalf("RelName = %q", got)
	}
	if got := must(f.RelName("other.Foo")); got != "other.Foo" {
		t.Fatalf("RelName = %q", got)
	}
}

func TestSerializerFnsOpts(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t)
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"string empty", must(f.StringOpts(spec.AnonType{})), ""},
		{"string len serializer", must(f.StringOpts(spec.AnonType{LenSerializer: "scalar.Int"})),
			"stropts.WithLenSerializer(scalar.Int)"},
		{"string all", must(f.StringOpts(spec.AnonType{LenSerializer: "scalar.Int", LenVl: "vl"})),
			"stropts.WithLenSerializer(scalar.Int), stropts.WithLenValidator(vl)"},
		{"string nil", must(f.StringOpts(spec.AnonType{LenSerializer: "nil", LenVl: "nil"})), ""},
		{"byte slice all", must(f.ByteSliceOpts(spec.AnonType{LenSerializer: "scalar.Int", LenVl: "vl"})),
			"bslopts.WithLenSerializer(scalar.Int), bslopts.WithLenValidator(vl)"},
		{"array all",
			must(f.ArrayOpts(spec.AnonType{ElemType: "int", LenSerializer: "scalar.Int", ElemVl: "ev"})),
			"arropts.WithLenSerializer[int](scalar.Int), arropts.WithElemValidator[int](ev)"},
		{"slice all", must(f.SliceOpts(spec.AnonType{ElemType: "int", LenSerializer: "scalar.Int",
			LenVl: "lv", ElemVl: "ev"})),
			"slopts.WithLenSerializer[int](scalar.Int), slopts.WithLenValidator[int](lv), " +
				"slopts.WithElemValidator[int](ev)"},
		{"map all", must(f.MapOpts(spec.AnonType{KeyType: "int", ElemType: "string",
			LenSerializer: "scalar.Int", LenVl: "lv", KeyVl: "kv", ElemVl: "ev"})),
			"mapopts.WithLenSerializer[int, string](scalar.Int), " +
				"mapopts.WithLenValidator[int, string](lv), " +
				"mapopts.WithKeyValidator[int, string](kv), " +
				"mapopts.WithValueValidator[int, string](ev)"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Fatalf("%v: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestSerializerFnsFieldTmplPipe(t *testing.T) {
	f := newTestSerializerFns(t)
	structType := spec.StructType{
		FullName: "target.Foo",
		Fields: []spec.FieldType{
			{FullName: "int", FieldName: "A", Index: 0, Gops: f.gopts},
			{FullName: "string", FieldName: "B", Index: 1, Gops: f.gopts},
		},
		Sops: stopts.Options{Fields: []fldopts.Options{
			{},
			{Type: tpopts.Options{Validator: "ValidateB"}},
		}},
		Gops: f.gopts,
	}

	pipe := f.FieldTmplPipe(structType, structType.Fields[1], 1, f.gopts)
	if pipe.Val != "v" || pipe.FieldsCount != 2 || pipe.Index != 1 {
		t.Fatalf("unexpected pipe %+v", pipe)
	}
	if pipe.Field.FieldName != "B" {
		t.Fatalf("Field = %+v", pipe.Field)
	}
	if pipe.Tops.Validator != "ValidateB" {
		t.Fatalf("Tops were not copied from the field options: %+v", pipe.Tops)
	}

	noOpts := structType
	noOpts.Sops = stopts.Options{}
	if got := f.FieldTmplPipe(noOpts, noOpts.Fields[0], 0, f.gopts); got.Tops != (tpopts.Options{}) {
		t.Fatalf("Tops = %+v, want zero", got.Tops)
	}
}

func TestSerializerFnsTypeName(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t)
	if got := must(f.codecName("target.Foo")); got != "Foo" {
		t.Fatalf("codecName = %q", got)
	}
	if got := must(f.codecName("other.Bar")); got != "other.Bar" {
		t.Fatalf("codecName = %q", got)
	}

	f.foreignTypes[typename.FullName("other.Bar")] = struct{}{}
	if got := must(f.codecName("other.Bar")); got != "Bar" {
		t.Fatalf("codecName = %q", got)
	}
	if got := must(f.Var("other.Bar")); got != "BarBinco" {
		t.Fatalf("Var = %q", got)
	}
	if got := f.upperFirst("foo"); got != "Foo" {
		t.Fatalf("upperFirst = %q", got)
	}
	if got := f.lowerFirst("Foo"); got != "foo" {
		t.Fatalf("lowerFirst = %q", got)
	}
	if got := f.lowerFirst(""); got != "" {
		t.Fatalf("lowerFirst of an empty name = %q", got)
	}
}

// TestSerializerFnsNameCollision covers D7: two types with the same bare name
// used to collapse into one generated serializer.
func TestSerializerFnsNameCollision(t *testing.T) {
	f := newTestSerializerFns(t)
	f.foreignTypes[typename.FullName("a.Foo")] = struct{}{}
	f.foreignTypes[typename.FullName("b.Foo")] = struct{}{}
	if _, err := f.codecName("a.Foo"); err != nil {
		t.Fatalf("first codecName: %v", err)
	}
	if _, err := f.codecName("b.Foo"); err == nil {
		t.Fatal("a name collision was not reported")
	}
}

func TestSerializerFnsCustomNames(t *testing.T) {
	must := mustName(t)
	f := newTestSerializerFns(t,
		genopts.WithSerializer(reflect.TypeFor[testdata.Int](), "CustomInt"))
	if got := must(f.Var("testdata.Int")); got != "CustomIntBinco" {
		t.Fatalf("Var = %q", got)
	}
}

func TestTmplFns(t *testing.T) {
	f := TmplFns{}
	stream := genopts.Options{Stream: true}

	if !f.PtrType("*int") || f.PtrType("int") {
		t.Fatal("PtrType misclassified a name")
	}
	if !f.ArrayType("[3]int") || f.ArrayType("int") || f.ArrayType("int32") || f.ArrayType("Foo1Bar") {
		t.Fatal("ArrayType misclassified a name")
	}
	if f.WithComma("") != "" || f.WithComma("x") != ", x" {
		t.Fatal("WithComma is wrong")
	}
	if f.Minus(5, 2) != 3 {
		t.Fatal("Minus is wrong")
	}
	if !f.ByteSliceStream("[]byte", stream) || !f.ByteSliceStream("[]uint8", stream) {
		t.Fatal("ByteSliceStream should be true for byte slices in stream mode")
	}
	if f.ByteSliceStream("[]int", stream) || f.ByteSliceStream("[]byte", genopts.Options{}) {
		t.Fatal("ByteSliceStream should be false otherwise")
	}
	if got := f.TimeSerializer(tpopts.Options{TimeUnit: tpopts.TimeUnitNano}); got != "TimeUnixNano" {
		t.Fatalf("TimeSerializer = %q", got)
	}

	if got := f.MarshalSignatureLastParam(genopts.Options{}); got != "bs []byte" {
		t.Fatalf("MarshalSignatureLastParam = %q", got)
	}
	if got := f.MarshalSignatureLastParam(stream); got != "w wirefmt.Writer" {
		t.Fatalf("MarshalSignatureLastParam = %q", got)
	}
	if got := f.UnmarshalSignatureLastParam(genopts.Options{}); got != "bs []byte" {
		t.Fatalf("UnmarshalSignatureLastParam = %q", got)
	}
	if got := f.UnmarshalSignatureLastParam(stream); got != "r wirefmt.Reader" {
		t.Fatalf("UnmarshalSignatureLastParam = %q", got)
	}
	if got := f.SkipSignatureLastParam(stream); got != "r wirefmt.Reader" {
		t.Fatalf("SkipSignatureLastParam = %q", got)
	}
	if got := f.MarshalLastParam(true, genopts.Options{}); got != "bs" {
		t.Fatalf("MarshalLastParam = %q", got)
	}
	if got := f.MarshalLastParam(false, genopts.Options{}); got != "bs[n:]" {
		t.Fatalf("MarshalLastParam = %q", got)
	}
	if got := f.MarshalLastParam(true, stream); got != "w" {
		t.Fatalf("MarshalLastParam = %q", got)
	}
	if got := f.UnmarshalLastParam(true, genopts.Options{}); got != "bs" {
		t.Fatalf("UnmarshalLastParam = %q", got)
	}
	if got := f.UnmarshalLastParam(false, genopts.Options{}); got != "bs[n:]" {
		t.Fatalf("UnmarshalLastParam = %q", got)
	}
	if got := f.UnmarshalLastParam(false, stream); got != "r" {
		t.Fatalf("UnmarshalLastParam = %q", got)
	}
	if got := f.SkipLastParam(genopts.Options{}); got != "bs[n:]" {
		t.Fatalf("SkipLastParam = %q", got)
	}
	if got := f.SkipLastParam(stream); got != "r" {
		t.Fatalf("SkipLastParam = %q", got)
	}
	if got := f.ModImportName(stream); got != "wirefmt" {
		t.Fatalf("ModImportName = %q", got)
	}
	if got := f.ModImportName(genopts.Options{}); got != "binco" {
		t.Fatalf("ModImportName = %q", got)
	}
}

func TestTmplFnsInclude(t *testing.T) {
	tmpl := template.Must(template.New("base").New("greet").Parse("hello {{.}}"))
	f := TmplFns{tmpl: tmpl}

	got, err := f.Include("greet", "world")
	if err != nil || got != "hello world" {
		t.Fatalf("Include = (%q, %v)", got, err)
	}
	if _, err = f.Include("missing", nil); err == nil {
		t.Fatal("Include of a missing template returned no error")
	}
}

func TestFillCrossgen(t *testing.T) {
	newGen := func(t *testing.T, pkgPath string) *Generator {
		t.Helper()
		g, err := NewGenerator(genopts.WithPath(pkgPath))
		if err != nil {
			t.Fatalf("NewGenerator: %v", err)
		}
		return g
	}

	other := newGen(t, targetPkg)
	other.markForeign(reflect.TypeFor[testdata.SimpleStruct](), "testdata.SimpleStruct")
	if _, ok := other.foreignTypes["testdata.SimpleStruct"]; !ok {
		t.Fatal("the external type was not marked as cross-generated")
	}

	same := newGen(t, "github.com/malivvan/binco/testdata")
	same.markForeign(reflect.TypeFor[testdata.SimpleStruct](), "testdata.SimpleStruct")
	if len(same.foreignTypes) != 0 {
		t.Fatalf("a local type was marked as cross-generated: %v", same.foreignTypes)
	}
}
