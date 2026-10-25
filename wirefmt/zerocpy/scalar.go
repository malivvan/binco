package zerocpy

import (
	"math"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	"github.com/malivvan/binco/wirefmt/scalar"
)

// Bool serializes a bool.
var Bool = boolView{}

type boolView struct{}

func (s boolView) Marshal(v bool, w wirefmt.Writer) (n int, err error) {
	err = w.WriteByte(*(*byte)(unsafe_mod.Pointer(&v)))
	if err != nil {
		return
	}
	n = 1
	return
}

func (s boolView) Unmarshal(r wirefmt.Reader) (v bool, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n = 1
	if b > 1 {
		err = binco.ErrWrongFormat
		return
	}
	v = *(*bool)(unsafe_mod.Pointer(&b))
	return
}

func (s boolView) Size(v bool) (n int) {
	return composite.SizeBool(v)
}

func (s boolView) Skip(r wirefmt.Reader) (n int, err error) {
	return composite.SkipBool(r)
}

// Byte serializes a byte.
var Byte = byteView{}

type byteView struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s byteView) FixedWidth() int { return binco.Num8RawSize }

func (s byteView) Marshal(v byte, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s byteView) Unmarshal(r wirefmt.Reader) (v byte, n int, err error) {
	return loadLE8[byte](r)
}

func (s byteView) Size(v byte) (n int) {
	return scalar.Byte.Size(v)
}

func (s byteView) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Byte.Skip(r)
}

// Uint8 serializes a uint8.
var Uint8 = u8View{}

type u8View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u8View) FixedWidth() int { return binco.Num8RawSize }

func (s u8View) Marshal(v uint8, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s u8View) Unmarshal(r wirefmt.Reader) (v uint8, n int, err error) {
	return loadLE8[uint8](r)
}

func (s u8View) Size(v uint8) (n int) {
	return scalar.Uint8.Size(v)
}

func (s u8View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Uint8.Skip(r)
}

// Uint16 serializes a uint16.
var Uint16 = u16View{}

type u16View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u16View) FixedWidth() int { return binco.Num16RawSize }

func (s u16View) Marshal(v uint16, w wirefmt.Writer) (n int, err error) {
	return writeLE16(v, w)
}

func (s u16View) Unmarshal(r wirefmt.Reader) (v uint16, n int, err error) {
	return loadLE16[uint16](r)
}

func (s u16View) Size(v uint16) (n int) {
	return scalar.Uint16.Size(v)
}

func (s u16View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Uint16.Skip(r)
}

// Uint32 serializes a uint32.
var Uint32 = u32View{}

type u32View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u32View) FixedWidth() int { return binco.Num32RawSize }

func (s u32View) Marshal(v uint32, w wirefmt.Writer) (n int, err error) {
	return writeLE32(v, w)
}

func (s u32View) Unmarshal(r wirefmt.Reader) (v uint32, n int, err error) {
	return loadLE32[uint32](r)
}

func (s u32View) Size(v uint32) (n int) {
	return scalar.Uint32.Size(v)
}

func (s u32View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Uint32.Skip(r)
}

// Uint64 serializes a uint64.
var Uint64 = u64View{}

type u64View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s u64View) FixedWidth() int { return binco.Num64RawSize }

func (s u64View) Marshal(v uint64, w wirefmt.Writer) (n int, err error) {
	return writeLE64(v, w)
}

func (s u64View) Unmarshal(r wirefmt.Reader) (v uint64, n int, err error) {
	return loadLE64[uint64](r)
}

func (s u64View) Size(v uint64) (n int) {
	return scalar.Uint64.Size(v)
}

func (s u64View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Uint64.Skip(r)
}

// Uint serializes a uint of the host word size.
var Uint = wordView{}

type wordView struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordView) FixedWidth() int { return binco.Num64RawSize }

func (s wordView) Marshal(v uint, w wirefmt.Writer) (n int, err error) {
	return writeWordUnsigned(v, w)
}

func (s wordView) Unmarshal(r wirefmt.Reader) (v uint, n int, err error) {
	return loadWordUnsigned(r)
}

func (s wordView) Size(v uint) (n int) {
	return wordUnsignedSize(v)
}

func (s wordView) Skip(r wirefmt.Reader) (n int, err error) {
	return skipWordUnsigned(r)
}

// Int8 serializes an int8.
var Int8 = i8View{}

type i8View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i8View) FixedWidth() int { return binco.Num8RawSize }

func (s i8View) Marshal(v int8, w wirefmt.Writer) (n int, err error) {
	return writeLE8(v, w)
}

func (s i8View) Unmarshal(r wirefmt.Reader) (v int8, n int, err error) {
	return loadLE8[int8](r)
}

func (s i8View) Size(v int8) (n int) {
	return scalar.Int8.Size(v)
}

func (s i8View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Int8.Skip(r)
}

// Int16 serializes an int16.
var Int16 = i16View{}

type i16View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i16View) FixedWidth() int { return binco.Num16RawSize }

func (s i16View) Marshal(v int16, w wirefmt.Writer) (n int, err error) {
	return writeLE16(v, w)
}

func (s i16View) Unmarshal(r wirefmt.Reader) (v int16, n int, err error) {
	return loadLE16[int16](r)
}

func (s i16View) Size(v int16) (n int) {
	return scalar.Int16.Size(v)
}

func (s i16View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Int16.Skip(r)
}

// Int32 serializes an int32.
var Int32 = i32View{}

type i32View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i32View) FixedWidth() int { return binco.Num32RawSize }

func (s i32View) Marshal(v int32, w wirefmt.Writer) (n int, err error) {
	return writeLE32(v, w)
}

func (s i32View) Unmarshal(r wirefmt.Reader) (v int32, n int, err error) {
	return loadLE32[int32](r)
}

func (s i32View) Size(v int32) (n int) {
	return scalar.Int32.Size(v)
}

func (s i32View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Int32.Skip(r)
}

// Int64 serializes an int64.
var Int64 = i64View{}

type i64View struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s i64View) FixedWidth() int { return binco.Num64RawSize }

func (s i64View) Marshal(v int64, w wirefmt.Writer) (n int, err error) {
	return writeLE64(v, w)
}

func (s i64View) Unmarshal(r wirefmt.Reader) (v int64, n int, err error) {
	return loadLE64[int64](r)
}

func (s i64View) Size(v int64) (n int) {
	return scalar.Int64.Size(v)
}

func (s i64View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Int64.Skip(r)
}

// Int serializes an int of the host word size.
var Int = wordIntView{}

type wordIntView struct{}

// FixedWidth reports the encoded size of one value, which is the width of
// the little-endian value itself.
func (s wordIntView) FixedWidth() int { return binco.Num64RawSize }

func (s wordIntView) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	return writeWord(v, w)
}

func (s wordIntView) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	return loadWord(r)
}

func (s wordIntView) Size(v int) (n int) {
	return wordSize(v)
}

func (s wordIntView) Skip(r wirefmt.Reader) (n int, err error) {
	return skipWord(r)
}

// Float64 serializes a float64.
var Float64 = f64View{}

type f64View struct{}

func (s f64View) Marshal(v float64, w wirefmt.Writer) (n int, err error) {
	return writeLE64(math.Float64bits(v), w)
}

func (s f64View) Unmarshal(r wirefmt.Reader) (v float64, n int, err error) {
	uv, n, err := loadLE64[uint64](r)
	if err != nil {
		return
	}
	return math.Float64frombits(uv), n, nil
}

func (s f64View) Size(v float64) (n int) {
	return scalar.Float64.Size(v)
}

func (s f64View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Float64.Skip(r)
}

// Float32 serializes a float32.
var Float32 = f32View{}

type f32View struct{}

func (s f32View) Marshal(v float32, w wirefmt.Writer) (n int, err error) {
	return writeLE32(math.Float32bits(v), w)
}

func (s f32View) Unmarshal(r wirefmt.Reader) (v float32, n int, err error) {
	uv, n, err := loadLE32[uint32](r)
	if err != nil {
		return
	}
	return math.Float32frombits(uv), n, nil
}

func (s f32View) Size(v float32) (n int) {
	return scalar.Float32.Size(v)
}

func (s f32View) Skip(r wirefmt.Reader) (n int, err error) {
	return scalar.Float32.Skip(r)
}
