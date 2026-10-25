package codegen

import (
	"reflect"
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/spec"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
	"github.com/malivvan/binco/testdata"
	other "github.com/malivvan/binco/testdata/other"
)

// badName is a name the scanner cannot parse, which is what a conversion error
// looks like from the naming helpers.
const badName = typename.FullName("not a name")

// TestUpperFirstEmpty checks the empty case of the capitalization helper.
func TestUpperFirstEmpty(t *testing.T) {
	f := SerializerFns{}
	if got := f.upperFirst(""); got != "" {
		t.Fatalf("upperFirst(%q) = %q, want %q", "", got, "")
	}
}

// TestNamingHelpersReportConversionErrors checks that every name helper returns
// the conversion error instead of a half-built name.
func TestNamingHelpersReportConversionErrors(t *testing.T) {
	f := newTestSerializerFns(t)
	calls := []struct {
		name string
		call func() (string, error)
	}{
		{"Var", func() (string, error) { return f.Var(badName) }},
		{"TypedVar", func() (string, error) { return f.TypedVar(badName) }},
		{"DTMVar", func() (string, error) { return f.DTMVar(badName) }},
		{"SchemaVar", func() (string, error) { return f.SchemaVar(badName) }},
		{"Type", func() (string, error) { return f.Type(badName) }},
		{"RelName", func() (string, error) { return f.RelName(badName) }},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.call(); err == nil {
				t.Fatalf("%s of an unparseable name returned nil", tc.name)
			}
		})
	}
}

// TestOptionRenderersReportConversionErrors checks the option-list renderers.
func TestOptionRenderersReportConversionErrors(t *testing.T) {
	f := newTestSerializerFns(t)
	calls := []struct {
		name string
		call func() (string, error)
	}{
		{"MapOpts key", func() (string, error) {
			return f.MapOpts(spec.AnonType{KeyType: badName})
		}},
		{"MapOpts elem", func() (string, error) {
			return f.MapOpts(spec.AnonType{KeyType: "int", ElemType: badName})
		}},
		{"SliceOpts", func() (string, error) {
			return f.SliceOpts(spec.AnonType{ElemType: badName})
		}},
		{"ArrayOpts", func() (string, error) {
			return f.ArrayOpts(spec.AnonType{ElemType: badName})
		}},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.call(); err == nil {
				t.Fatalf("%s of an unparseable name returned nil", tc.name)
			}
		})
	}
}

// TestOfNamesTheDocumentedTypes checks the codec names the templates rely on.
func TestOfNamesTheDocumentedTypes(t *testing.T) {
	f := newTestSerializerFns(t)
	for _, tc := range []struct {
		name string
		tops tpopts.Options
		want string
	}{
		{"int64", tpopts.Options{}, "varnum.Int64"},
		{"int64", tpopts.Options{NumEncoding: tpopts.NumEncodingVarintPositive}, "varnum.PositiveInt64"},
		{"uint32", tpopts.Options{}, "varnum.Uint32"},
		{"string", tpopts.Options{}, "composite.String"},
		{"[]byte", tpopts.Options{}, "composite.ByteSlice"},
		{"time.Time", tpopts.Options{}, "scalar.TimeUnixUTC"},
	} {
		got, err := f.Of(typename.FullName(tc.name), tc.tops, genopts.Options{})
		if err != nil || got != tc.want {
			t.Fatalf("Of(%q) = (%q, %v), want %q", tc.name, got, err, tc.want)
		}
	}
}

// TestNewSerializerFnsErrors checks the two conversions that prepare the
// per-type serializer names.
func TestNewSerializerFnsErrors(t *testing.T) {
	// A type that has no name at all.
	_, err := NewSerializerFns(cnvtr.NewTypeNameConverter(genopts.New()), nil, nil, TmplFns{},
		genopts.Options{SerializerNames: map[reflect.Type]string{
			reflect.TypeFor[struct{ X int }](): "X",
		}})
	if err == nil {
		t.Fatal("NewSerializerFns accepted an unnamed type")
	}

	// A conversion that fails: two packages mapped to one alias cannot both be
	// imported, so whichever is converted second is reported.
	opts := genopts.New()
	if err = genopts.WithImportAlias("github.com/malivvan/binco/testdata", "dup")(&opts); err != nil {
		t.Fatalf("WithImportAlias: %v", err)
	}
	if err = genopts.WithImportAlias("github.com/malivvan/binco/testdata/other", "dup")(&opts); err != nil {
		t.Fatalf("WithImportAlias: %v", err)
	}
	opts.SerializerNames = map[reflect.Type]string{
		reflect.TypeFor[testdata.SameName](): "SameName",
		reflect.TypeFor[other.SameName]():    "SameName2",
	}
	_, err = NewSerializerFns(cnvtr.NewTypeNameConverter(opts), nil, nil, TmplFns{}, opts)
	if err == nil {
		t.Fatal("NewSerializerFns ignored a failing conversion")
	}
}
