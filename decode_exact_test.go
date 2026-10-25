package binco_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

// TestDecodeExact covers the D9 note that Decode ignores trailing bytes: a
// protocol that carries exactly one value per message needs the strict variant.
func TestDecodeExact(t *testing.T) {
	bs := binco.Encode(composite.String, "hi")
	bs = append(bs, 0xff)

	// Decode keeps ignoring trailing bytes, which is the documented behaviour.
	if v, err := binco.Decode(composite.String, bs); err != nil || v != "hi" {
		t.Fatalf("Decode = (%q, %v)", v, err)
	}
	if _, err := binco.DecodeExact(composite.String, bs); !errors.Is(err, binco.ErrTrailingBytes) {
		t.Fatalf("DecodeExact returned %v, want ErrTrailingBytes", err)
	}
	exact := binco.Encode(varnum.Int, 300)
	if v, err := binco.DecodeExact(varnum.Int, exact); err != nil || v != 300 {
		t.Fatalf("DecodeExact of an exact payload = (%d, %v)", v, err)
	}
}

// TestDecodeExactReportsDecodeErrors checks that a decode failure wins over the
// trailing-bytes check: the caller needs the real cause.
func TestDecodeExactReportsDecodeErrors(t *testing.T) {
	// A string that declares five bytes and carries one.
	if _, err := binco.DecodeExact(composite.String, []byte{5, 'a'}); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("DecodeExact of a truncated value = %v, want ErrTruncated", err)
	}
}
