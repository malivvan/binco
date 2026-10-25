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
	if _, _, err := binco.DecodeExact(composite.String, bs); !errors.Is(err, binco.ErrTrailingBytes) {
		t.Fatalf("DecodeExact returned %v, want ErrTrailingBytes", err)
	}
	exact := binco.Encode(varnum.Int, 300)
	if v, n, err := binco.DecodeExact(varnum.Int, exact); err != nil || v != 300 || n != len(exact) {
		t.Fatalf("DecodeExact of an exact payload = (%d, %d, %v)", v, n, err)
	}
}
