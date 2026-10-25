package fldopts_test

import (
	"testing"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func TestOptions(t *testing.T) {
	o := fldopts.Options{}
	fldopts.Apply(&o, fldopts.WithIgnore(), nil)
	if !o.Ignore {
		t.Fatal("WithIgnore was not applied")
	}
	if o.Type != (tpopts.Options{}) {
		t.Fatalf("WithIgnore changed the type options: %+v", o.Type)
	}
}

func TestWithType(t *testing.T) {
	o := fldopts.Options{}
	fldopts.Apply(&o,
		fldopts.WithType(
			tpopts.WithNumEncoding(tpopts.NumEncodingRaw),
			tpopts.WithLenValidator("ValidateLength"),
			tpopts.WithElemValidator("ValidateElem"),
		),
		nil,
	)
	if o.Type.NumEncoding != tpopts.NumEncodingRaw {
		t.Fatalf("NumEncoding = %v", o.Type.NumEncoding)
	}
	if o.Type.LenValidator != "ValidateLength" || o.Type.ElemValidator != "ValidateElem" {
		t.Fatalf("unexpected type options %+v", o.Type)
	}

	if o.Ignore {
		t.Fatal("WithType set Ignore")
	}
}

func TestApplyEmpty(t *testing.T) {
	o := fldopts.Options{}
	fldopts.Apply(&o)
	if o.Ignore || o.Type != (tpopts.Options{}) {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
