package scalar

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Byte serializes a byte.
var Byte = byteCodec{}

type byteCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s byteCodec) FixedWidth() int { return binco.Num8RawSize }

func (s byteCodec) Marshal(v byte, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s byteCodec) Unmarshal(r wirefmt.Reader) (v byte, n int, err error) {
	return readLE8[byte](r)
}

func (s byteCodec) Size(v byte) (n int) {
	return binco.Num8RawSize
}

func (s byteCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE8(r)
}

// Uint8 serializes a uint8.
var Uint8 = u8Codec{}

type u8Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u8Codec) FixedWidth() int { return binco.Num8RawSize }

func (s u8Codec) Marshal(v uint8, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s u8Codec) Unmarshal(r wirefmt.Reader) (v uint8, n int, err error) {
	return readLE8[uint8](r)
}

func (s u8Codec) Size(v uint8) (n int) {
	return binco.Num8RawSize
}

func (s u8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE8(r)
}

// Uint16 serializes a uint16.
var Uint16 = u16Codec{}

type u16Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u16Codec) FixedWidth() int { return binco.Num16RawSize }

func (s u16Codec) Marshal(v uint16, w wirefmt.Writer) (n int, err error) {
	return writeLE16(v, w)
}

func (s u16Codec) Unmarshal(r wirefmt.Reader) (v uint16, n int, err error) {
	return readLE16[uint16](r)
}

func (s u16Codec) Size(v uint16) (n int) {
	return binco.Num16RawSize
}

func (s u16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE16(r)
}

// Uint32 serializes a uint32.
var Uint32 = u32Codec{}

type u32Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u32Codec) FixedWidth() int { return binco.Num32RawSize }

func (s u32Codec) Marshal(v uint32, w wirefmt.Writer) (n int, err error) {
	return writeLE32(v, w)
}

func (s u32Codec) Unmarshal(r wirefmt.Reader) (v uint32, n int, err error) {
	return readLE32[uint32](r)
}

func (s u32Codec) Size(v uint32) (n int) {
	return widthOf32(v)
}

func (s u32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE32(r)
}

// Uint64 serializes a uint64.
var Uint64 = u64Codec{}

type u64Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u64Codec) FixedWidth() int { return binco.Num64RawSize }

func (s u64Codec) Marshal(v uint64, w wirefmt.Writer) (n int, err error) {
	return writeLE64(v, w)
}

func (s u64Codec) Unmarshal(r wirefmt.Reader) (v uint64, n int, err error) {
	return readLE64[uint64](r)
}

func (s u64Codec) Size(v uint64) (n int) {
	return widthOf64(v)
}

func (s u64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE64(r)
}

// Uint serializes a uint, following the word size of the host.
var Uint = wordUnsignedCodec{}

type wordUnsignedCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordUnsignedCodec) FixedWidth() int { return binco.Num64RawSize }

func (s wordUnsignedCodec) Marshal(v uint, w wirefmt.Writer) (n int, err error) {
	return writeWordUnsigned(v, w)
}

func (s wordUnsignedCodec) Unmarshal(r wirefmt.Reader) (v uint, n int, err error) {
	return readWordUnsigned(r)
}

func (s wordUnsignedCodec) Size(v uint) (n int) {
	return wordUnsignedSize(v)
}

func (s wordUnsignedCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipWordUnsigned(r)
}

// Int8 serializes an int8.
var Int8 = i8Codec{}

type i8Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i8Codec) FixedWidth() int { return binco.Num8RawSize }

func (s i8Codec) Marshal(v int8, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s i8Codec) Unmarshal(r wirefmt.Reader) (v int8, n int, err error) {
	return readLE8[int8](r)
}

func (s i8Codec) Size(v int8) (n int) {
	return binco.Num8RawSize
}

func (s i8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE8(r)
}

// Int16 serializes an int16.
var Int16 = i16Codec{}

type i16Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i16Codec) FixedWidth() int { return binco.Num16RawSize }

func (s i16Codec) Marshal(v int16, w wirefmt.Writer) (n int, err error) {
	return writeLE16(v, w)
}

func (s i16Codec) Unmarshal(r wirefmt.Reader) (v int16, n int, err error) {
	return readLE16[int16](r)
}

func (s i16Codec) Size(v int16) (n int) {
	return binco.Num16RawSize
}

func (s i16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE16(r)
}

// Int32 serializes an int32.
var Int32 = i32Codec{}

type i32Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i32Codec) FixedWidth() int { return binco.Num32RawSize }

func (s i32Codec) Marshal(v int32, w wirefmt.Writer) (n int, err error) {
	return writeLE32(v, w)
}

func (s i32Codec) Unmarshal(r wirefmt.Reader) (v int32, n int, err error) {
	return readLE32[int32](r)
}

func (s i32Codec) Size(v int32) (n int) {
	return widthOf32(v)
}

func (s i32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE32(r)
}

// Int64 serializes an int64.
var Int64 = i64Codec{}

type i64Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i64Codec) FixedWidth() int { return binco.Num64RawSize }

func (s i64Codec) Marshal(v int64, w wirefmt.Writer) (n int, err error) {
	return writeLE64(v, w)
}

func (s i64Codec) Unmarshal(r wirefmt.Reader) (v int64, n int, err error) {
	return readLE64[int64](r)
}

func (s i64Codec) Size(v int64) (n int) {
	return widthOf64(v)
}

func (s i64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipLE64(r)
}

// Int serializes an int, following the word size of the host.
var Int = wordIntCodec{}

type wordIntCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordIntCodec) FixedWidth() int { return binco.Num64RawSize }

func (s wordIntCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	return writeWord(v, w)
}

func (s wordIntCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	return readWord(r)
}

func (s wordIntCodec) Size(v int) (n int) {
	return wordSize(v)
}

func (s wordIntCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return skipWord(r)
}
