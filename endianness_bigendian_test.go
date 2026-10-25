//go:build ppc64 || s390x || mips || mips64

package binco_test

import (
	"bytes"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/internal/fixedwidth"
	"github.com/malivvan/binco/scalar"
)

// TestBigEndianHost runs only on a big-endian target (ppc64, s390x, mips,
// mips64). It proves the three things the wire format promises there: the host
// is detected as big-endian, the bulk memory copy is disabled, and every codec
// still writes little-endian bytes.
func TestBigEndianHost(t *testing.T) {
	if binco.HostLittleEndian {
		t.Fatal("a big-endian target reports HostLittleEndian")
	}
	if width, ok := fixedwidth.Of[int32](scalar.Int32); ok {
		t.Fatalf("the bulk path is enabled on a big-endian host (width %d)", width)
	}
	if got, want := binco.Encode(scalar.Uint32, 0x01020304), []byte{0x04, 0x03, 0x02, 0x01}; !bytes.Equal(got, want) {
		t.Fatalf("scalar.Uint32 = %v, want %v", got, want)
	}

	values := []int32{1, -2, 3}
	want := []byte{3}
	for _, v := range values {
		want = append(want, littleEndianInt32(v)...)
	}
	got := binco.Encode(composite.NewSliceSerializer[int32](scalar.Int32), values)
	if !bytes.Equal(got, want) {
		t.Fatalf("slice = %v, want %v", got, want)
	}
	back, err := binco.Decode(composite.NewSliceSerializer[int32](scalar.Int32), got)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for i := range values {
		if back[i] != values[i] {
			t.Fatalf("Unmarshal[%d] = %d, want %d", i, back[i], values[i])
		}
	}
}
