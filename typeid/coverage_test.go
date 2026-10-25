package typeid_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

// TestDTMSerializerRejectsNonPositiveIdentifiers checks the write-side rule: a
// zero or negative identifier would be written as a huge unsigned number, so it
// is refused instead.
func TestDTMSerializerRejectsNonPositiveIdentifiers(t *testing.T) {
	calls := []struct {
		name string
		call func()
	}{
		{"Marshal zero", func() { typeid.DTMSerializer.Marshal(0, make([]byte, 8)) }},
		{"Marshal negative", func() { typeid.DTMSerializer.Marshal(-1, make([]byte, 8)) }},
		{"Size zero", func() { typeid.DTMSerializer.Size(0) }},
		{"Size negative", func() { typeid.DTMSerializer.Size(-1) }},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, binco.ErrOverflow) {
					t.Fatalf("panic value %v, want ErrOverflow", err)
				}
			}()
			tc.call()
		})
	}
}

// TestDTMSerializerRejectsBadInput checks the decode-side rule: missing input is
// a truncation, and a zero identifier is out of range.
func TestDTMSerializerRejectsBadInput(t *testing.T) {
	if _, _, err := typeid.DTMSerializer.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Unmarshal(nil) = %v, want ErrTruncated", err)
	}
	if _, _, err := typeid.DTMSerializer.Unmarshal([]byte{0}); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Unmarshal of identifier 0 = %v, want ErrOverflow", err)
	}
}

// TestTypedSkipReportsATruncatedIdentifier checks the typed serializer's skip of
// an input that ends before the identifier.
func TestTypedSkipReportsATruncatedIdentifier(t *testing.T) {
	typed := typeid.NewSerializer(1, varnum.Int)
	if _, err := typed.Skip(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Skip(nil) = %v, want ErrTruncated", err)
	}
}
