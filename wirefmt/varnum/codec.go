package varnum

import (
	"math/bits"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// putUnsigned writes t to w in LEB128 form.
func putUnsigned[T binco.Unsigned](t T, w wirefmt.Writer) (n int,
	err error,
) {
	for t >= 0x80 {
		err = w.WriteByte(byte(t) | 0x80)
		if err != nil {
			return
		}
		t >>= 7
		n++
	}
	err = w.WriteByte(byte(t))
	if err != nil {
		return
	}
	n++
	return
}

// getUnsigned reads an unsigned varint of at most maxLen bytes.
func getUnsigned[T binco.Unsigned](maxVarintLen int, maxLastByte byte,
	r wirefmt.Reader,
) (t T, n int, err error) {
	var b byte

	var shift int

	for {
		b, err = r.ReadByte()
		if err != nil {
			return
		}
		n++
		if n == maxVarintLen && b > maxLastByte {
			return 0, n, binco.ErrOverflow
		}
		if b < 0x80 {
			t = t | T(b)<<shift
			return
		}
		t = t | T(b&0x7F)<<shift
		shift += 7
	}
}

// unsignedLen reports how many bytes t occupies once encoded.
func unsignedLen[T binco.Unsigned](t T) (size int) {
	return (bits.Len64(uint64(t)|1) + 6) / 7
}

// skipUnsigned advances over one encoded unsigned varnum.
func skipUnsigned(maxVarintLen int, maxLastByte byte, r wirefmt.Reader) (n int,
	err error,
) {
	var b byte
	for {
		b, err = r.ReadByte()
		if err != nil {
			return
		}
		n++

		if n == maxVarintLen && b > maxLastByte {
			err = binco.ErrOverflow
			return
		}
		if b < 0x80 {
			return
		}
	}
}

// Byte serializes a byte.
var Byte = byteCodec{}

type byteCodec struct{}

func (byteCodec) Marshal(v byte, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (byteCodec) Unmarshal(r wirefmt.Reader) (v byte, n int, err error) {
	return getUnsigned[byte](binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte,
		r)
}

func (byteCodec) Size(v byte) (size int) {
	return unsignedLen(v)
}

func (byteCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, r)
}

// Uint8 serializes a uint8.
var Uint8 = u8Codec{}

type u8Codec struct{}

func (u8Codec) Marshal(v uint8, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (u8Codec) Unmarshal(r wirefmt.Reader) (v uint8, n int, err error) {
	return getUnsigned[uint8](binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, r)
}

func (u8Codec) Size(v uint8) int {
	return unsignedLen(v)
}

func (u8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.Uint8MaxVarintLen, binco.Uint8MaxLastByte, r)
}

// Uint16 serializes a uint16.
var Uint16 = u16Codec{}

type u16Codec struct{}

func (u16Codec) Marshal(v uint16, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (u16Codec) Unmarshal(r wirefmt.Reader) (v uint16, n int, err error) {
	return getUnsigned[uint16](binco.Uint16MaxVarintLen, binco.Uint16MaxLastByte, r)
}

func (u16Codec) Size(v uint16) int {
	return unsignedLen(v)
}

func (u16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.Uint16MaxVarintLen, binco.Uint16MaxLastByte, r)
}

// Uint32 serializes a uint32.
var Uint32 = u32Codec{}

type u32Codec struct{}

func (u32Codec) Marshal(v uint32, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (u32Codec) Unmarshal(r wirefmt.Reader) (v uint32, n int, err error) {
	return getUnsigned[uint32](binco.Uint32MaxVarintLen,
		binco.Uint32MaxLastByte,
		r)
}

func (u32Codec) Size(v uint32) int {
	return unsignedLen(v)
}

func (u32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.Uint32MaxVarintLen, binco.Uint32MaxLastByte, r)
}

// Uint64 serializes a uint64.
var Uint64 = u64Codec{}

type u64Codec struct{}

func (u64Codec) Marshal(v uint64, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (u64Codec) Unmarshal(r wirefmt.Reader) (v uint64, n int, err error) {
	return getUnsigned[uint64](binco.Uint64MaxVarintLen,
		binco.Uint64MaxLastByte,
		r)
}

func (u64Codec) Size(v uint64) int {
	return unsignedLen(v)
}

func (u64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.Uint64MaxVarintLen, binco.Uint64MaxLastByte, r)
}

// Uint serializes a uint, following the word size of the host.
var Uint = wordCodec{}

type wordCodec struct{}

func (wordCodec) Marshal(v uint, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(v, w)
}

func (wordCodec) Unmarshal(r wirefmt.Reader) (v uint, n int, err error) {
	return getUnsigned[uint](binco.UintMaxVarintLen(), binco.UintMaxLastByte(), r)
}

func (wordCodec) Size(v uint) int {
	return unsignedLen(v)
}

func (wordCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipUnsigned(binco.UintMaxVarintLen(), binco.UintMaxLastByte(), r)
}
