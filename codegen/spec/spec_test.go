package spec_test

import (
	"testing"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	structopts "github.com/malivvan/binco/codegen/config/struct"
	"github.com/malivvan/binco/codegen/spec"
)

func TestStructTypeSerializedFields(t *testing.T) {
	fields := []spec.FieldType{{FieldName: "A"}, {FieldName: "B"}, {FieldName: "C"}}

	t.Run("no options", func(t *testing.T) {
		got := spec.StructType{Fields: fields}.SerializedFields()
		if len(got) != len(fields) {
			t.Fatalf("len = %d, want %d", len(got), len(fields))
		}
		for i := range fields {
			if got[i].FieldName != fields[i].FieldName {
				t.Fatalf("field %d = %q, want %q", i, got[i].FieldName, fields[i].FieldName)
			}
		}
	})

	t.Run("one ignored", func(t *testing.T) {
		st := spec.StructType{
			Fields: fields,
			Sops: structopts.Options{Fields: []fldopts.Options{
				{}, {Ignore: true}, {},
			}},
		}
		got := st.SerializedFields()
		if len(got) != 2 || got[0].FieldName != "A" || got[1].FieldName != "C" {
			t.Fatalf("SerializedFields() = %+v, want A and C", got)
		}
	})

	t.Run("all ignored", func(t *testing.T) {
		st := spec.StructType{
			Fields: fields,
			Sops: structopts.Options{Fields: []fldopts.Options{
				{Ignore: true}, {Ignore: true}, {Ignore: true},
			}},
		}
		if got := st.SerializedFields(); len(got) != 0 {
			t.Fatalf("SerializedFields() = %+v, want none", got)
		}
	})
}

func TestAnonKindString(t *testing.T) {
	tests := []struct {
		kind spec.AnonKind
		want string
	}{
		{spec.AnonKindString, "string"},
		{spec.AnonKindArray, "array"},
		{spec.AnonKindByteSlice, "byteSlice"},
		{spec.AnonKindSlice, "slice"},
		{spec.AnonKindMap, "map"},
		{spec.AnonKindPtr, "ptr"},
	}
	for _, tc := range tests {
		if got := tc.kind.String(); got != tc.want {
			t.Fatalf("AnonKind(%d).String() = %q, want %q", int(tc.kind), got, tc.want)
		}
	}

	// An unknown kind used to panic inside a template execution, where the panic
	// surfaced as an opaque failure; it now renders a name that says so.
	if got, want := spec.AnonKindUndefined.String(), "unknown(0)"; got != want {
		t.Fatalf("AnonKindUndefined.String() = %q, want %q", got, want)
	}
}
