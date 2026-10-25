package fixedwidth_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/fixedwidth"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

// TestOfAcceptsRawScalars checks the codecs that write a raw little-endian
// value: their width is the size of the Go type.
func TestOfAcceptsRawScalars(t *testing.T) {
	for _, tc := range []struct {
		name  string
		width int
		ok    func() (int, bool)
	}{
		{"int32", 4, func() (int, bool) { return fixedwidth.Of[int32](scalar.Int32) }},
		{"uint64", 8, func() (int, bool) { return fixedwidth.Of[uint64](scalar.Uint64) }},
		{"float64", 8, func() (int, bool) { return fixedwidth.Of[float64](scalar.Float64) }},
		{"byte", 1, func() (int, bool) { return fixedwidth.Of[byte](scalar.Byte) }},
	} {
		width, ok := tc.ok()
		if !ok || width != tc.width {
			t.Fatalf("%s: Of = (%d, %v), want (%d, true)", tc.name, width, ok, tc.width)
		}
	}
}

// TestOfRejectsMismatches checks the three reasons to fall back to the
// element-by-element path.
func TestOfRejectsMismatches(t *testing.T) {
	// Not a fixed-width codec.
	if _, ok := fixedwidth.Of[int](varnum.Int); ok {
		t.Fatal("Of accepted a varint codec")
	}
	// Fixed width, but not the size of the Go type.
	if _, ok := fixedwidth.Of[int16](scalar.Int32); ok {
		t.Fatal("Of accepted a codec whose width differs from the type")
	}
	// A codec that reports a width but is not implemented as raw bytes cannot
	// occur here; the interface documents that requirement.
	if w := scalar.Int32.FixedWidth(); w != 4 {
		t.Fatalf("Int32.FixedWidth = %d, want 4", w)
	}
	if !binco.HostLittleEndian {
		t.Skip("the bulk path is disabled on a big-endian host")
	}
}
