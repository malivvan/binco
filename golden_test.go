package binco_test

import (
	"bytes"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

// TestGoldenVectors pins the bytes of the primitive codecs. The format is a
// compatibility surface, so a change to any of these vectors is a wire-format
// change and must be deliberate.
func TestGoldenVectors(t *testing.T) {
	vectors := []struct {
		name string
		fn   func() []byte
		want []byte
	}{
		{"varint 0", func() []byte { return binco.Encode(varnum.Uint64, 0) }, []byte{0x00}},
		{"varint 1", func() []byte { return binco.Encode(varnum.Uint64, 1) }, []byte{0x01}},
		{"varint 127", func() []byte { return binco.Encode(varnum.Uint64, 127) }, []byte{0x7f}},
		{"varint 128", func() []byte { return binco.Encode(varnum.Uint64, 128) }, []byte{0x80, 0x01}},
		{"varint 300", func() []byte { return binco.Encode(varnum.Uint64, 300) }, []byte{0xac, 0x02}},
		{"varint max", func() []byte {
			return binco.Encode(varnum.Uint64, ^uint64(0))
		}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
		{"zigzag -1", func() []byte { return binco.Encode(varnum.Int64, -1) }, []byte{0x01}},
		{"zigzag -2", func() []byte { return binco.Encode(varnum.Int64, -2) }, []byte{0x03}},
		{"zigzag 1", func() []byte { return binco.Encode(varnum.Int64, 1) }, []byte{0x02}},
		{"scalar uint32", func() []byte { return binco.Encode(scalar.Uint32, 0x01020304) },
			[]byte{0x04, 0x03, 0x02, 0x01}},
		{"scalar int32 -1", func() []byte { return binco.Encode(scalar.Int32, -1) },
			[]byte{0xff, 0xff, 0xff, 0xff}},
		{"scalar uint16", func() []byte { return binco.Encode(scalar.Uint16, 0x0102) },
			[]byte{0x02, 0x01}},
	}
	for _, tc := range vectors {
		if got := tc.fn(); !bytes.Equal(got, tc.want) {
			t.Fatalf("%s = %x, want %x", tc.name, got, tc.want)
		}
	}
}
