package scalar

import "github.com/malivvan/binco"

// Byte serializes a byte.
var Byte = byteCodec{}

type byteCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s byteCodec) FixedWidth() int { return binco.Num8RawSize }

func (s byteCodec) Marshal(v byte, bs []byte) (n int) {
	return putLE8(v, bs)
}

func (s byteCodec) Unmarshal(bs []byte) (v byte, n int, err error) {
	return readLE8[byte](bs)
}

func (s byteCodec) Size(v byte) (size int) {
	return binco.Num8RawSize
}

func (s byteCodec) Skip(bs []byte) (n int, err error) {
	return SkipInteger8(bs)
}

// Uint8 serializes a uint8.
var Uint8 = u8Codec{}

type u8Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u8Codec) FixedWidth() int { return binco.Num8RawSize }

func (s u8Codec) Marshal(v uint8, bs []byte) (n int) {
	return putLE8(v, bs)
}

func (s u8Codec) Unmarshal(bs []byte) (v uint8, n int, err error) {
	return readLE8[uint8](bs)
}

func (s u8Codec) Size(v uint8) (size int) {
	return binco.Num8RawSize
}

func (s u8Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger8(bs)
}

// Uint16 serializes a uint16.
var Uint16 = u16Codec{}

type u16Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u16Codec) FixedWidth() int { return binco.Num16RawSize }

func (s u16Codec) Marshal(v uint16, bs []byte) (n int) {
	return putLE16(v, bs)
}

func (s u16Codec) Unmarshal(bs []byte) (v uint16, n int, err error) {
	return readLE16[uint16](bs)
}

func (s u16Codec) Size(v uint16) (size int) {
	return binco.Num16RawSize
}

func (s u16Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger16(bs)
}

// Uint32 serializes a uint32.
var Uint32 = u32Codec{}

type u32Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u32Codec) FixedWidth() int { return binco.Num32RawSize }

func (s u32Codec) Marshal(v uint32, bs []byte) (n int) {
	return putLE32(v, bs)
}

func (s u32Codec) Unmarshal(bs []byte) (v uint32, n int, err error) {
	return readLE32[uint32](bs)
}

func (s u32Codec) Size(v uint32) (size int) {
	return binco.Num32RawSize
}

func (s u32Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger32(bs)
}

// Uint64 serializes a uint64.
var Uint64 = u64Codec{}

type u64Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u64Codec) FixedWidth() int { return binco.Num64RawSize }

func (s u64Codec) Marshal(v uint64, bs []byte) (n int) {
	return putLE64(v, bs)
}

func (s u64Codec) Unmarshal(bs []byte) (v uint64, n int, err error) {
	return readLE64[uint64](bs)
}

func (s u64Codec) Size(v uint64) (size int) {
	return binco.Num64RawSize
}

func (s u64Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger64(bs)
}

// Uint serializes a uint, following the word size of the host.
var Uint = wordUnsignedCodec{}

type wordUnsignedCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordUnsignedCodec) FixedWidth() int { return binco.Num64RawSize }

func (s wordUnsignedCodec) Marshal(v uint, bs []byte) (n int) {
	return putWordUnsigned(v, bs)
}

func (s wordUnsignedCodec) Unmarshal(bs []byte) (v uint, n int, err error) {
	return readWordUnsigned(bs)
}

func (s wordUnsignedCodec) Size(v uint) (size int) {
	return wordSize
}

func (s wordUnsignedCodec) Skip(bs []byte) (n int, err error) {
	return skipWordUnsigned(bs)
}

// Int8 serializes an int8.
var Int8 = i8Codec{}

type i8Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i8Codec) FixedWidth() int { return binco.Num8RawSize }

func (s i8Codec) Marshal(v int8, bs []byte) (n int) {
	return putLE8[int8](v, bs)
}

func (s i8Codec) Unmarshal(bs []byte) (v int8, n int, err error) {
	return readLE8[int8](bs)
}

func (s i8Codec) Size(v int8) (size int) {
	return binco.Num8RawSize
}

func (s i8Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger8(bs)
}

// Int16 serializes an int16.
var Int16 = i16Codec{}

type i16Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i16Codec) FixedWidth() int { return binco.Num16RawSize }

func (s i16Codec) Marshal(v int16, bs []byte) (n int) {
	return putLE16[int16](v, bs)
}

func (s i16Codec) Unmarshal(bs []byte) (v int16, n int, err error) {
	return readLE16[int16](bs)
}

func (s i16Codec) Size(v int16) (size int) {
	return binco.Num16RawSize
}

func (s i16Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger16(bs)
}

// Int32 serializes an int32.
var Int32 = i32Codec{}

type i32Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i32Codec) FixedWidth() int { return binco.Num32RawSize }

func (s i32Codec) Marshal(v int32, bs []byte) (n int) {
	return putLE32[int32](v, bs)
}

func (s i32Codec) Unmarshal(bs []byte) (v int32, n int, err error) {
	return readLE32[int32](bs)
}

func (s i32Codec) Size(v int32) (size int) {
	return binco.Num32RawSize
}

func (s i32Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger32(bs)
}

// Int64 serializes an int64.
var Int64 = i64Codec{}

type i64Codec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i64Codec) FixedWidth() int { return binco.Num64RawSize }

func (s i64Codec) Marshal(v int64, bs []byte) (n int) {
	return putLE64(v, bs)
}

func (s i64Codec) Unmarshal(bs []byte) (v int64, n int, err error) {
	return readLE64[int64](bs)
}

func (s i64Codec) Size(v int64) (size int) {
	return binco.Num64RawSize
}

func (s i64Codec) Skip(bs []byte) (n int, err error) {
	return SkipInteger64(bs)
}

// Int serializes an int, following the word size of the host.
var Int = wordIntCodec{}

type wordIntCodec struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordIntCodec) FixedWidth() int { return binco.Num64RawSize }

func (s wordIntCodec) Marshal(v int, bs []byte) (n int) {
	return putWord(v, bs)
}

func (s wordIntCodec) Unmarshal(bs []byte) (v int, n int, err error) {
	return readWord(bs)
}

func (s wordIntCodec) Size(v int) (size int) {
	return wordSize
}

func (s wordIntCodec) Skip(bs []byte) (n int, err error) {
	return skipWord(bs)
}
