package typeid_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
)

// TestOptionalDTMSerializerNilMarker checks the reserved identifier an interface
// payload writes for a nil value: one zero byte, on the way in and on the way
// out.
func TestOptionalDTMSerializerNilMarker(t *testing.T) {
	bs := binco.Encode(typeid.OptionalDTMSerializer, binco.DTM(0))
	if len(bs) != 1 || bs[0] != 0 {
		t.Fatalf("Marshal(0) = %v, want [0]", bs)
	}
	if got := typeid.OptionalDTMSerializer.Size(0); got != 1 {
		t.Fatalf("Size(0) = %d, want 1", got)
	}

	dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(bs)
	if err != nil || n != 1 || dtm != 0 {
		t.Fatalf("Unmarshal([0]) = (%d, %d, %v)", dtm, n, err)
	}
	if n, err = typeid.OptionalDTMSerializer.Skip(bs); err != nil || n != 1 {
		t.Fatalf("Skip([0]) = (%d, %v)", n, err)
	}
}

// TestOptionalDTMSerializerMatchesDTMSerializer checks that every non-zero
// identifier keeps the bytes, size and skip of the plain DTM codec.
func TestOptionalDTMSerializerMatchesDTMSerializer(t *testing.T) {
	for _, v := range []binco.DTM{1, 7, 127, 128, 300, 1 << 20} {
		want := binco.Encode(typeid.DTMSerializer, v)
		got := binco.Encode(typeid.OptionalDTMSerializer, v)
		if len(got) != len(want) {
			t.Fatalf("identifier %d: %v, want %v", v, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("identifier %d: %v, want %v", v, got, want)
			}
		}
		if typeid.OptionalDTMSerializer.Size(v) != typeid.DTMSerializer.Size(v) {
			t.Fatalf("identifier %d: Size differs", v)
		}
		dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(got)
		if err != nil || n != len(got) || dtm != v {
			t.Fatalf("identifier %d: Unmarshal = (%d, %d, %v)", v, dtm, n, err)
		}
		if n, err = typeid.OptionalDTMSerializer.Skip(got); err != nil || n != len(want) {
			t.Fatalf("identifier %d: Skip = (%d, %v)", v, n, err)
		}
	}
}

// TestOptionalDTMSerializerRejectsBadInput checks truncation, non-canonical
// encodings and the identifier width: an identifier above the host's int cannot
// be held by binco.DTM, and Skip must agree with Unmarshal about it.
func TestOptionalDTMSerializerRejectsBadInput(t *testing.T) {
	if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Unmarshal(nil) = %v, want ErrTruncated", err)
	}
	if _, _, err := typeid.OptionalDTMSerializer.Unmarshal([]byte{0x80, 0x00}); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Unmarshal of a non-canonical identifier = %v, want ErrNonCanonical", err)
	}

	if strconv.IntSize == 32 {
		// MaxInt64 fits PositiveInt64 but not binco.DTM on a 32-bit host.
		big := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}
		if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(big); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("Unmarshal of MaxInt64 = %v, want ErrOverflow", err)
		}
		if _, err := typeid.OptionalDTMSerializer.Skip(big); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("Skip of MaxInt64 = %v, want ErrOverflow", err)
		}
	}
}

// TestOptionalDTMSerializerRejectsNegative checks the write side: a negative
// identifier is not a reserved marker and cannot be encoded.
func TestOptionalDTMSerializerRejectsNegative(t *testing.T) {
	calls := []struct {
		name string
		call func()
	}{
		{"Marshal", func() { typeid.OptionalDTMSerializer.Marshal(-1, make([]byte, 8)) }},
		{"Size", func() { typeid.OptionalDTMSerializer.Size(-1) }},
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
