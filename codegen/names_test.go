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

func newTestSerFns(t *testing.T, opts ...genopts.SetOption) SerFns {
	t.Helper()
	opts = append([]genopts.SetOption{genopts.WithPkgPath(targetPkg)}, opts...)
	gops := genopts.New()
	if err := genopts.Apply(opts, &gops); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	conv := cnvtr.NewTypeNameConverter(gops)
	return NewSerFns(conv, bldr.NewTypeBuilder(conv, gops),
		map[typename.FullName]struct{}{}, TmplFns{}, gops)
}

func TestSerFnsNames(t *testing.T) {
	f := newTestSerFns(t)
	name := typename.FullName("target.Foo")

	if got := f.Var(name); got != "FooBinco" {
		t.Fatalf("Var = %q", got)
	}
	if got := f.TypedVar(name); got != "FooTypedBinco" {
		t.Fatalf("TypedVar = %q", got)
	}
	if got := f.DTMVar(name); got != "FooDTM" {
		t.Fatalf("DTMVar = %q", got)
	}
	if got := f.Type(name); got != "fooBinco" {
		t.Fatalf("Type = %q", got)
	}
	if got := f.Val(struct{}{}); got != "v" {
		t.Fatalf("Val = %q", got)
	}
}

func TestSerFnsOf(t *testing.T) {
	safe := newTestSerFns(t)
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
		if got := safe.Of(typename.FullName(tc.name), tc.tops, safe.gopts); got != tc.want {
			t.Fatalf("Of(%q, %+v) = %q, want %q", tc.name, tc.tops, got, tc.want)
		}
	}

	unsafeFns := newTestSerFns(t, genopts.WithUnsafe())
	for _, tc := range []struct{ name, want string }{
		{"int", "zerocpy.Int"},
		{"int8", "zerocpy.Int8"},
		{"uint64", "zerocpy.Uint64"},
		{"float64", "zerocpy.Float64"},
		{"bool", "zerocpy.Bool"},
		{"string", "zerocpy.String"},
		{"[]byte", "zerocpy.ByteSlice"},
		{"time.Time", "zerocpy.TimeUnixUTC"},
		{"time.Time", "zerocpy.TimeUnixUTC"},
	} {
		if got := unsafeFns.Of(typename.FullName(tc.name), tpopts.Options{},
			unsafeFns.gopts); got != tc.want {
			t.Fatalf("unsafe Of(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}

	notUnsafe := newTestSerFns(t, genopts.WithNotUnsafe())
	if got := notUnsafe.Of("int", tpopts.Options{}, notUnsafe.gopts); got != "zerocpy.Int" {
		t.Fatalf("not-unsafe Of(int) = %q", got)
	}
	if got := notUnsafe.Of("string", tpopts.Options{}, notUnsafe.gopts); got != "composite.String" {
		t.Fatalf("not-unsafe Of(string) = %q", got)
	}
}

func TestSerFnsOfPanics(t *testing.T) {
	f := newTestSerFns(t)
	defer func() {
		if recover() == nil {
			t.Fatal("Of did not panic for an unsupported type")
		}
	}()
	f.Of("struct{}", tpopts.Options{}, f.gopts)
}

func TestSerFnsKeyElem(t *testing.T) {
	f := newTestSerFns(t)
	anon := spec.AnonType{KeyType: "string", ElemType: "int"}
	if got := f.Key(anon, f.gopts); got != "composite.String" {
		t.Fatalf("Key = %q", got)
	}
	if got := f.Elem(anon, f.gopts); got != "varnum.Int" {
		t.Fatalf("Elem = %q", got)
	}
}

func TestSerFnsRelName(t *testing.T) {
	f := newTestSerFns(t)
	if got := f.RelName("target.Foo"); got != "Foo" {
		t.Fatalf("RelName = %q", got)
	}
	if got := f.RelName("other.Foo"); got != "other.Foo" {
		t.Fatalf("RelName = %q", got)
	}
}

func TestSerFnsOpts(t *testing.T) {
	f := newTestSerFns(t)
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"string empty", f.StringOpts(spec.AnonType{}), ""},
		{"string len ser", f.StringOpts(spec.AnonType{LenSer: "scalar.Int"}),
			"stropts.WithLenSer(scalar.Int)"},
		{"string all", f.StringOpts(spec.AnonType{LenSer: "scalar.Int", LenVl: "vl"}),
			"stropts.WithLenSer(scalar.Int), stropts.WithLenValidator(vl)"},
		{"string nil", f.StringOpts(spec.AnonType{LenSer: "nil", LenVl: "nil"}), ""},
		{"byte slice all", f.ByteSliceOpts(spec.AnonType{LenSer: "scalar.Int", LenVl: "vl"}),
			"bslopts.WithLenSer(scalar.Int), bslopts.WithLenValidator(vl)"},
		{"array all",
			f.ArrayOpts(spec.AnonType{ElemType: "int", LenSer: "scalar.Int", ElemVl: "ev"}),
			"arropts.WithLenSer[int](scalar.Int), arropts.WithElemValidator[int](ev)"},
		{"slice all", f.SliceOpts(spec.AnonType{ElemType: "int", LenSer: "scalar.Int",
			LenVl: "lv", ElemVl: "ev"}),
			"slopts.WithLenSer[int](scalar.Int), slopts.WithLenValidator[int](lv), " +
				"slopts.WithElemValidator[int](ev)"},
		{"map all", f.MapOpts(spec.AnonType{KeyType: "int", ElemType: "string",
			LenSer: "scalar.Int", LenVl: "lv", KeyVl: "kv", ElemVl: "ev"}),
			"mapopts.WithLenSer[int, string](scalar.Int), " +
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

func TestSerFnsFieldTmplPipe(t *testing.T) {
	f := newTestSerFns(t)
	structType := spec.StructType{
		FullName: "target.Foo",
		Fields: []spec.FieldType{
			{FullName: "int", FieldName: "A", Gops: f.gopts},
			{FullName: "string", FieldName: "B", Gops: f.gopts},
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

func TestSerFnsTypeName(t *testing.T) {
	f := newTestSerFns(t)
	if got := f.codecName("target.Foo"); got != "Foo" {
		t.Fatalf("codecName = %q", got)
	}
	if got := f.codecName("other.Bar"); got != "other.Bar" {
		t.Fatalf("codecName = %q", got)
	}

	f.foreignTypes[typename.FullName("other.Bar")] = struct{}{}
	if got := f.codecName("other.Bar"); got != "Bar" {
		t.Fatalf("codecName = %q", got)
	}
	if got := f.Var("other.Bar"); got != "BarBinco" {
		t.Fatalf("Var = %q", got)
	}
	if got := f.upperFirst("foo"); got != "Foo" {
		t.Fatalf("upperFirst = %q", got)
	}
	if got := f.lowerFirst("Foo"); got != "foo" {
		t.Fatalf("lowerFirst = %q", got)
	}
}

func TestSerFnsCustomNames(t *testing.T) {
	f := newTestSerFns(t,
		genopts.WithSerName(reflect.TypeFor[testdata.Int](), "CustomInt"))
	if got := f.Var("testdata.Int"); got != "CustomIntBinco" {
		t.Fatalf("Var = %q", got)
	}
}

func TestTmplFns(t *testing.T) {
	f := TmplFns{}
	stream := genopts.Options{Stream: true}

	if !f.PtrType("*int") || f.PtrType("int") {
		t.Fatal("PtrType misclassified a name")
	}
	if !f.ArrayType("[3]int") || f.ArrayType("int") {
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
	if got := f.TimeSer(tpopts.Options{TimeUnit: tpopts.TimeUnitNano}); got != "TimeUnixNano" {
		t.Fatalf("TimeSer = %q", got)
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
		g, err := NewGenerator(genopts.WithPkgPath(pkgPath))
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
