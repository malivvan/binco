package varnum_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// TestNonCanonicalVarintsAreRejected covers S8: "80 80 80 00" used to decode to
// zero, so a value had more than one byte representation.
func TestNonCanonicalVarintsAreRejected(t *testing.T) {
	overlong := []byte{0x80, 0x80, 0x80, 0x00}
	if _, _, err := varnum.Uint64.Unmarshal(overlong); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Unmarshal returned %v, want ErrNonCanonical", err)
	}
	if _, err := varnum.Uint64.Skip(overlong); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Skip returned %v, want ErrNonCanonical", err)
	}
	if _, _, err := varnum.Int.Unmarshal(overlong); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Int.Unmarshal returned %v, want ErrNonCanonical", err)
	}

	// The canonical form of zero is a single byte and still decodes.
	if got, n, err := varnum.Uint64.Unmarshal([]byte{0x00}); err != nil || n != 1 || got != 0 {
		t.Fatalf("Unmarshal of canonical zero = (%d, %d, %v)", got, n, err)
	}
}

// TestPositiveIntDoesNotWrap covers S8: PositiveInt8 used to return -56 for the
// byte string c8 01 (200).
func TestPositiveIntDoesNotWrap(t *testing.T) {
	if _, _, err := varnum.PositiveInt8.Unmarshal([]byte{0xc8, 0x01}); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("PositiveInt8.Unmarshal returned %v, want ErrOverflow", err)
	}
	if _, _, err := varnum.PositiveInt16.Unmarshal([]byte{0xff, 0xff, 0x04}); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("PositiveInt16.Unmarshal returned %v, want ErrOverflow", err)
	}
	if _, _, err := varnum.PositiveInt32.Unmarshal(
		[]byte{0xff, 0xff, 0xff, 0xff, 0x10}); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("PositiveInt32.Unmarshal returned %v, want ErrOverflow", err)
	}

	for name, fn := range map[string]func(){
		"PositiveInt8":  func() { varnum.PositiveInt8.Marshal(-1, make([]byte, 2)) },
		"PositiveInt16": func() { varnum.PositiveInt16.Marshal(-1, make([]byte, 3)) },
		"PositiveInt32": func() { varnum.PositiveInt32.Marshal(-1, make([]byte, 5)) },
		"PositiveInt64": func() { varnum.PositiveInt64.Marshal(-1, make([]byte, 10)) },
		"PositiveInt":   func() { varnum.PositiveInt.Marshal(-1, make([]byte, 10)) },
	} {
		t.Run(name+" marshal negative", func(t *testing.T) {
			defer func() {
				if r := recover(); !errors.Is(toError(r), binco.ErrOverflow) {
					t.Fatalf("panic value %v, want ErrOverflow", r)
				}
			}()
			fn()
		})
	}
}

func toError(v any) error {
	err, _ := v.(error)
	return err
}

// TestWordWidthIsHostIndependent covers D6: int and uint are encoded as 64 bits
// on every architecture, so a payload is portable.
//
// A 32-bit build cannot hold the value, so there the check is that the host
// width is not silently truncated: decoding reports ErrOverflow instead.
func TestWordWidthIsHostIndependent(t *testing.T) {
	big := uint64(1) << 40
	bs := binco.Encode(varnum.Uint64, big)
	if got, err := binco.Decode(varnum.Uint64, bs); err != nil || got != big {
		t.Fatalf("uint64 payload = (%d, %v)", got, err)
	}
	if strconv.IntSize == 64 {
		if got, err := binco.Decode(varnum.Uint, bs); err != nil || uint64(got) != big {
			t.Fatalf("uint payload = (%d, %v)", got, err)
		}
		if got := varnum.Uint.Size(uint(big)); got != 6 {
			t.Fatalf("Uint.Size = %d, want 6 (a 64-bit value)", got)
		}
	} else if _, _, err := varnum.Uint.Unmarshal(bs); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Uint.Unmarshal of a 64-bit value on a 32-bit host returned %v, want ErrOverflow", err)
	}
	if got, err := binco.Decode(varnum.Int64, binco.Encode(varnum.Int64, int64(big))); err != nil || got != int64(big) {
		t.Fatalf("int64 payload = (%d, %v)", got, err)
	}
}
