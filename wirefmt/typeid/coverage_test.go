package typeid_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt/typeid"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// TestDTMSerializerRejectsNonPositiveIdentifiers checks the write-side rule: a
// zero or negative identifier is refused instead of being written as a huge
// unsigned number.
func TestDTMSerializerRejectsNonPositiveIdentifiers(t *testing.T) {
	calls := []struct {
		name string
		call func() (int, error)
	}{
		{"Marshal zero", func() (int, error) { return typeid.DTMSerializer.Marshal(0, &bytes.Buffer{}) }},
		{"Marshal negative", func() (int, error) { return typeid.DTMSerializer.Marshal(-1, &bytes.Buffer{}) }},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.call(); !errors.Is(err, binco.ErrOverflow) {
				t.Fatalf("%s = %v, want ErrOverflow", tc.name, err)
			}
		})
	}

	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("panic value %v, want ErrOverflow", err)
		}
	}()
	typeid.DTMSerializer.Size(0)
}

// TestDTMSerializerRejectsBadInput checks the decode-side rule: a missing
// identifier is a clean end of stream, and identifier zero is out of range.
func TestDTMSerializerRejectsBadInput(t *testing.T) {
	if _, _, err := typeid.DTMSerializer.Unmarshal(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
	}
	if _, _, err := typeid.DTMSerializer.Unmarshal(bytes.NewReader([]byte{0})); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Unmarshal of identifier 0 = %v, want ErrOverflow", err)
	}
}

// TestTypedSkipReportsAnEmptyStream checks the typed serializer's skip of a
// stream that ends before the identifier.
func TestTypedSkipReportsAnEmptyStream(t *testing.T) {
	typed := typeid.NewSerializer(1, varnum.Int)
	if _, err := typed.Skip(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream = %v, want io.EOF", err)
	}
}
