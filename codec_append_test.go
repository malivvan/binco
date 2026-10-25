package binco_test

import (
	"bytes"
	"errors"
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
	dst := binco.AppendMarshal([]byte{0xAA}, selfEncoded{v: 300})
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendMarshal = %v, want %v", dst, want)
	}
	dst = binco.AppendMarshalTyped([]byte{0xAA}, selfEncodedTyped{v: 300})
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendMarshalTyped = %v, want %v", dst, want)
	}
}

// shortSerializer reports two bytes for every value but writes one. It is a
// broken codec, and the append helpers must refuse to trust it.
type shortSerializer struct{}

func (shortSerializer) Marshal(v int, bs []byte) (n int) {
	bs[0] = byte(v)
	return 1
}
func (shortSerializer) Unmarshal(bs []byte) (v int, n int, err error) { return int(bs[0]), 1, nil }
func (shortSerializer) Size(v int) (size int)                         { return 2 }
func (shortSerializer) Skip(bs []byte) (n int, err error)             { return 2, nil }

// shortMarshaller is the self-encoding counterpart of shortSerializer.
type shortMarshaller struct{}

func (shortMarshaller) MarshalBinco(bs []byte) (n int) { return 1 }
func (shortMarshaller) SizeBinco() (size int)          { return 2 }

type shortTypedMarshaller struct{}

func (shortTypedMarshaller) MarshalTypedBinco(bs []byte) (n int) { return 1 }
func (shortTypedMarshaller) SizeTypedBinco() (size int)          { return 2 }

// TestAppendEncode checks the serializer flavour of the append helper.
func TestAppendEncode(t *testing.T) {
	dst := binco.AppendEncode([]byte{0xAA}, varnum.Int, 300)
	if want := append([]byte{0xAA}, binco.Encode(varnum.Int, 300)...); !bytes.Equal(dst, want) {
		t.Fatalf("AppendEncode = %v, want %v", dst, want)
	}
}

// TestAppendRejectsInconsistentSize checks that a codec which writes a different
// number of bytes than it reported is reported instead of leaving zero padding.
func TestAppendRejectsInconsistentSize(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"AppendEncode", func() { binco.AppendEncode(nil, shortSerializer{}, 1) }},
		{"AppendMarshal", func() { binco.AppendMarshal(nil, shortMarshaller{}) }},
		{"AppendMarshalTyped", func() { binco.AppendMarshalTyped(nil, shortTypedMarshaller{}) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, binco.ErrWrongSize) {
					t.Fatalf("panic value %v, want ErrWrongSize", err)
				}
			}()
			tc.call()
		})
	}
}
