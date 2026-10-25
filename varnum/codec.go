package varnum

import (
	"math/bits"

	"github.com/malivvan/binco"
)

// putUnsigned stores t in LEB128 form and returns the number of bytes written.
func putUnsigned[T binco.Unsigned](t T, bs []byte) (n int) {
	if t < 0x80 {
		if len(bs) < 1 {
			panic(binco.ErrTooSmallByteSlice)
		}
		bs[0] = byte(t)
		return 1
	}
	binco.CheckWrite(bs, 0, unsignedLen(t))
	for t >= 0x80 {
		bs[n] = byte(t) | 0x80
		t >>= 7
		n++
	}
	bs[n] = byte(t)
	return n + 1
}

// getUnsigned reads an unsigned varint of at most maxLen bytes.
//
// The encoding must be canonical: a multi-byte value whose final byte is zero
// is rejected with ErrNonCanonical, so every value has exactly one byte
// representation. Without that rule "80 80 80 00" and "00" would both decode
// to zero, which breaks hashing, signing and de-duplication by bytes.
func getUnsigned[T binco.Unsigned](maxVarintLen int, maxLastByte byte,
	bs []byte,
) (t T, n int, err error) {
	if len(bs) == 0 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if bs[0] < 0x80 {
		return T(bs[0]), 1, nil
	}
	var (
		b     byte
		shift int
	)
	for n, b = range bs {
		n++
		if n == maxVarintLen && b > maxLastByte {
			return 0, n, binco.ErrOverflow
		}
		if b < 0x80 {
			if b == 0 {
				return 0, n, binco.ErrNonCanonical
			}
			t = t | T(b)<<shift
			return
		}
		t = t | T(b&0x7F)<<shift
		shift += 7
	}
	return 0, n, binco.ErrTooSmallByteSlice
}

// unsignedLen reports how many bytes t occupies once encoded.
func unsignedLen[T binco.Unsigned](t T) (size int) {
	return (bits.Len64(uint64(t)|1) + 6) / 7
}

// skipUnsigned advances over one encoded unsigned varnum, applying the same
// canonical and overflow rules as getUnsigned.
func skipUnsigned(maxVarintLen int, maxLastByte byte, bs []byte) (n int,
	err error,
) {
	if len(bs) == 0 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	for i := range bs {
		if i+1 == maxVarintLen && bs[i] > maxLastByte {
			return i + 1, binco.ErrOverflow
		}
		if bs[i] < 0x80 {
			if bs[i] == 0 && i > 0 {
				return i + 1, binco.ErrNonCanonical
			}
			return i + 1, nil
		}
	}
	return len(bs), binco.ErrTooSmallByteSlice
}

// Byte serializes a byte.
var Byte = byteCodec{}

type byteCodec struct{}

func (s byteCodec) Marshal(v byte, bs []byte) (n int) {
	return putUnsigned(v, bs)
}

func (s byteCodec) Unmarshal(bs []byte) (v byte, n int, err error) {
	return getUnsigned[byte](binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, bs)
}

func (s byteCodec) Size(v byte) (size int) {
	return unsignedLen(v)
}

func (s byteCodec) Skip(bs []byte) (n int, err error) {
	return skipUnsigned(binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, bs)
}

// Uint8 serializes a uint8.
var Uint8 = u8Codec{}

type u8Codec struct{}

func (s u8Codec) Marshal(v uint8, bs []byte) (n int) {
	return putUnsigned(v, bs)
}

func (s u8Codec) Unmarshal(bs []byte) (v uint8, n int, err error) {
	return getUnsigned[uint8](binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte,
		bs)
}

func (s u8Codec) Size(v uint8) (size int) {
	return unsignedLen(v)
}

func (s u8Codec) Skip(bs []byte) (n int, err error) {
	return skipUnsigned(binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, bs)
}

// Uint16 serializes a uint16.
var Uint16 = u16Codec{}

type u16Codec struct{}

func (s u16Codec) Marshal(v uint16, bs []byte) (n int) {
	return putUnsigned(v, bs)
}

func (s u16Codec) Unmarshal(bs []byte) (v uint16, n int, err error) {
	return getUnsigned[uint16](binco.Uint16MaxVarintLen, binco.Uint16MaxLastByte,
		bs)
}

func (s u16Codec) Size(v uint16) (size int) {
	return unsignedLen(v)
}

func (s u16Codec) Skip(bs []byte) (n int, err error) {
	return skipUnsigned(binco.Uint16MaxVarintLen, binco.Uint16MaxLastByte, bs)
}

// Uint32 serializes a uint32.
var Uint32 = u32Codec{}

type u32Codec struct{}

func (s u32Codec) Marshal(v uint32, bs []byte) (n int) {
	return putUnsigned(v, bs)
}

func (s u32Codec) Unmarshal(bs []byte) (v uint32, n int, err error) {
	return getUnsigned[uint32](binco.Uint32MaxVarintLen, binco.Uint32MaxLastByte,
		bs)
}

func (s u32Codec) Size(v uint32) (size int) {
	return unsignedLen(v)
}

func (s u32Codec) Skip(bs []byte) (n int, err error) {
	return skipUnsigned(binco.Uint32MaxVarintLen, binco.Uint32MaxLastByte, bs)
}

// Uint64 serializes a uint64.
var Uint64 = u64Codec{}

type u64Codec struct{}

func (s u64Codec) Marshal(v uint64, bs []byte) (n int) {
	return putUnsigned(v, bs)
}

func (s u64Codec) Unmarshal(bs []byte) (v uint64, n int, err error) {
	return getUnsigned[uint64](binco.Uint64MaxVarintLen, binco.Uint64MaxLastByte,
		bs)
}

func (s u64Codec) Size(v uint64) (size int) {
	return unsignedLen(v)
}

func (s u64Codec) Skip(bs []byte) (n int, err error) {
	return skipUnsigned(binco.Uint64MaxVarintLen, binco.Uint64MaxLastByte, bs)
}

// Uint serializes a uint. The wire width is always 64 bits, whatever the host
// word size is, so a payload written on amd64 reads back on a 32-bit build.
// Values that do not fit the host uint are reported as ErrOverflow.
var Uint = wordCodec{}

type wordCodec struct{}

func (s wordCodec) Marshal(v uint, bs []byte) (n int) {
	return Uint64.Marshal(uint64(v), bs)
}

func (s wordCodec) Unmarshal(bs []byte) (v uint, n int, err error) {
	uv, n, err := Uint64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = uint(uv)
	if uint64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return
}

func (s wordCodec) Size(v uint) (size int) {
	return Uint64.Size(uint64(v))
}

func (s wordCodec) Skip(bs []byte) (n int, err error) {
	// Decode so that a value the host uint cannot hold is an error here just as
	// it is in Unmarshal.
	_, n, err = s.Unmarshal(bs)
	return
}
