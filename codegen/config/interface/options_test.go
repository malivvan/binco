package intropts_test

import (
	"reflect"
	"testing"

	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func TestInterfaceOptions(t *testing.T) {
	a, b := reflect.TypeFor[int](), reflect.TypeFor[string]()
	o := intropts.Options{}
	intropts.Apply(&o,
		intropts.WithImpl(a),
		intropts.WithImpls(b),
		intropts.WithMarshaller(),
		nil,
	)
	if len(o.Impls) != 2 || o.Impls[0] != a || o.Impls[1] != b {
		t.Fatalf("Impls = %v", o.Impls)
	}
	if !o.Marshaller {
		t.Fatal("WithMarshaller was not applied")
	}
}

func TestRegisterOptions(t *testing.T) {
	structImpl := reflect.TypeFor[struct{ A int }]()
	definedImpl := reflect.TypeFor[int]()
	o := intropts.NewRegisterOptions()
	intropts.ApplyRegister(&o,
		intropts.WithStructImpl(structImpl, stopts.WithValidator("Validate")),
		intropts.WithDefinedTypeImpl(definedImpl, tpopts.WithNumEncoding(tpopts.NumEncodingRaw)),
		intropts.WithRegisterMarshaller(),
		nil,
	)
	if len(o.StructImpls) != 1 || o.StructImpls[0].Type != structImpl {
		t.Fatalf("StructImpls = %v", o.StructImpls)
	}
	if len(o.StructImpls[0].Opts) != 1 {
		t.Fatalf("struct impl options = %v", o.StructImpls[0].Opts)
	}
	if len(o.DefinedTypeImpls) != 1 || o.DefinedTypeImpls[0].Type != definedImpl {
		t.Fatalf("DefinedTypeImpls = %v", o.DefinedTypeImpls)
	}
	if len(o.DefinedTypeImpls[0].Opts) != 1 {
		t.Fatalf("defined type impl options = %v", o.DefinedTypeImpls[0].Opts)
	}
	if !o.Marshaller {
		t.Fatal("WithRegisterMarshaller was not applied")
	}
}
