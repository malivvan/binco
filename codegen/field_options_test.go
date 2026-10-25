package codegen

import (
	"reflect"
	"testing"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	bldr "github.com/malivvan/binco/codegen/spec/builder"
)

type ignoredMiddleField struct {
	A int
	B string
	C int
}

// TestIgnoredFieldKeepsOptionsAligned covers D1: SerializedFields drops ignored
// fields, so indexing the options with the position in the filtered list gave
// field C the options of field B.
func TestIgnoredFieldKeepsOptionsAligned(t *testing.T) {
	f := newTestSerializerFns(t)

	sops := stopts.Options{}
	stopts.Apply([]stopts.SetOption{
		stopts.WithField(fldopts.WithType(tpopts.WithValidator("ValidateA"))),
		stopts.WithField(fldopts.WithIgnore(), fldopts.WithType(tpopts.WithValidator("ValidateB"))),
		stopts.WithField(fldopts.WithType(tpopts.WithValidator("ValidateC"))),
	}, &sops)

	builder := bldr.NewTypeBuilder(f.converter, f.gopts)
	structType, err := builder.BuildStructType(reflect.TypeFor[ignoredMiddleField](), sops)
	if err != nil {
		t.Fatalf("BuildStructType: %v", err)
	}

	serialized := structType.SerializedFields()
	if len(serialized) != 2 {
		t.Fatalf("SerializedFields returned %d fields, want 2", len(serialized))
	}
	if serialized[0].FieldName != "A" || serialized[1].FieldName != "C" {
		t.Fatalf("SerializedFields = %v, %v", serialized[0].FieldName, serialized[1].FieldName)
	}

	pipe := f.FieldTmplPipe(structType, serialized[1], 1, f.gopts)
	if pipe.Tops.Validator != "ValidateC" {
		t.Fatalf("field C received validator %q, want ValidateC", pipe.Tops.Validator)
	}
	pipe = f.FieldTmplPipe(structType, serialized[0], 0, f.gopts)
	if pipe.Tops.Validator != "ValidateA" {
		t.Fatalf("field A received validator %q, want ValidateA", pipe.Tops.Validator)
	}
}
