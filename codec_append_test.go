package binco_test

import (
	"bytes"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

type selfEncoded struct{ v int }

func (m selfEncoded) MarshalBinco(bs []byte) (n int) { return varnum.Int.Marshal(m.v, bs) }
func (m selfEncoded) SizeBinco() (size int)          { return varnum.Int.Size(m.v) }

type selfEncodedTyped struct{ v int }

func (m selfEncodedTyped) MarshalTypedBinco(bs []byte) (n int) {
	n = varnum.Int.Marshal(m.v, bs)
	return
}
func (m selfEncodedTyped) SizeTypedBinco() (size int) { return varnum.Int.Size(m.v) }

// TestAppendMarshal covers the P4 append helpers for self-encoding values.
func TestAppendMarshal(t *testing.T) {
	dst := binco.AppendMarshal(selfEncoded{v: 300}, []byte{0xAA})
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendMarshal = %v, want %v", dst, want)
	}
	dst = binco.AppendMarshalTyped(selfEncodedTyped{v: 300}, []byte{0xAA})
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendMarshalTyped = %v, want %v", dst, want)
	}
}
