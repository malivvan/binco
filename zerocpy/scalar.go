package zerocpy

import (
	"math"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
)

// Bool serializes a bool.
var Bool = boolView{}

type boolView struct{}

func (s boolView) Marshal(v bool, bs []byte) (n int) {
	*(*bool)(unsafe_mod.Pointer(&bs[0])) = v
	return 1
}

func (s boolView) Unmarshal(bs []byte) (v bool, n int, err error) {
	if len(bs) < 1 {
		return false, 0, binco.ErrTooSmallByteSlice
	}
	if bs[0] > 1 {
		err = binco.ErrWrongFormat
		return
	}
	return *(*bool)(unsafe_mod.Pointer(&bs[0])), 1, nil
}

func (s boolView) Size(_ bool) (size int) {
	return 1
}

func (s boolView) Skip(bs []byte) (n int, err error) {
	return composite.SkipBool(bs)
}

// Byte serializes a byte.
var Byte = byteView{}

type byteView struct{}

func (s byteView) Marshal(v byte, bs []byte) (n int) {
	return writeLE8(v, bs)
}

func (s byteView) Unmarshal(bs []byte) (v byte, n int, err error) {
	return loadLE8[byte](bs)
}

func (s byteView) Size(v byte) (size int) {
	return binco.Num8RawSize
}

func (s byteView) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger8(bs)
}

// Uint8 serializes a uint8.
var Uint8 = u8View{}

type u8View struct{}

func (s u8View) Marshal(v uint8, bs []byte) (n int) {
	return writeLE8(v, bs)
}

func (s u8View) Unmarshal(bs []byte) (v uint8, n int, err error) {
	return loadLE8[uint8](bs)
}

func (s u8View) Size(v uint8) (size int) {
	return binco.Num8RawSize
}

func (s u8View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger8(bs)
}

// Uint16 serializes a uint16.
var Uint16 = u16View{}

type u16View struct{}

func (s u16View) Marshal(v uint16, bs []byte) (n int) {
	return writeLE16(v, bs)
}

func (s u16View) Unmarshal(bs []byte) (v uint16, n int, err error) {
	return loadLE16[uint16](bs)
}

func (s u16View) Size(v uint16) (size int) {
	return binco.Num16RawSize
}

func (s u16View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger16(bs)
}

// Uint32 serializes a uint32.
var Uint32 = u32View{}

type u32View struct{}

func (s u32View) Marshal(v uint32, bs []byte) (n int) {
	return writeLE32(v, bs)
}

func (s u32View) Unmarshal(bs []byte) (v uint32, n int, err error) {
	return loadLE32[uint32](bs)
}

func (s u32View) Size(v uint32) (size int) {
	return binco.Num32RawSize
}

func (s u32View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger32(bs)
}

// Uint64 serializes a uint64.
var Uint64 = u64View{}

type u64View struct{}

func (s u64View) Marshal(v uint64, bs []byte) (n int) {
	return writeLE64(v, bs)
}

func (s u64View) Unmarshal(bs []byte) (v uint64, n int, err error) {
	return loadLE64[uint64](bs)
}

func (s u64View) Size(v uint64) (size int) {
	return binco.Num64RawSize
}

func (s u64View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger64(bs)
}

// Uint serializes a uint of the host word size.
var Uint = wordView{}

type wordView struct{}

func (s wordView) Marshal(v uint, bs []byte) (n int) {
	return writeWordUnsigned(v, bs)
}

func (s wordView) Unmarshal(bs []byte) (v uint, n int, err error) {
	return loadWordUnsigned(bs)
}

func (s wordView) Size(v uint) (size int) {
	return wordUnsignedSize
}

func (s wordView) Skip(bs []byte) (n int, err error) {
	return skipWordUnsigned(bs)
}

// Int8 serializes an int8.
var Int8 = i8View{}

type i8View struct{}

func (s i8View) Marshal(v int8, bs []byte) (n int) {
	return writeLE8(v, bs)
}

func (s i8View) Unmarshal(bs []byte) (v int8, n int, err error) {
	return loadLE8[int8](bs)
}

func (s i8View) Size(v int8) (size int) {
	return binco.Num8RawSize
}

func (s i8View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger8(bs)
}

// Int16 serializes an int16.
var Int16 = i16View{}

type i16View struct{}

func (s i16View) Marshal(v int16, bs []byte) (n int) {
	return writeLE16(v, bs)
}

func (s i16View) Unmarshal(bs []byte) (v int16, n int, err error) {
	return loadLE16[int16](bs)
}

func (s i16View) Size(v int16) (size int) {
	return binco.Num16RawSize
}

func (s i16View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger16(bs)
}

// Int32 serializes an int32.
var Int32 = i32View{}

type i32View struct{}

func (s i32View) Marshal(v int32, bs []byte) (n int) {
	return writeLE32(v, bs)
}

func (s i32View) Unmarshal(bs []byte) (v int32, n int, err error) {
	return loadLE32[int32](bs)
}

func (s i32View) Size(v int32) (size int) {
	return binco.Num32RawSize
}

func (s i32View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger32(bs)
}

// Int64 serializes an int64.
var Int64 = i64View{}

type i64View struct{}

func (s i64View) Marshal(v int64, bs []byte) (n int) {
	return writeLE64(v, bs)
}

func (s i64View) Unmarshal(bs []byte) (v int64, n int, err error) {
	return loadLE64[int64](bs)
}

func (s i64View) Size(v int64) (size int) {
	return binco.Num64RawSize
}

func (s i64View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger64(bs)
}

// Int serializes an int of the host word size.
var Int = wordIntView{}

type wordIntView struct{}

func (s wordIntView) Marshal(v int, bs []byte) (n int) {
	return writeWord(v, bs)
}

func (s wordIntView) Unmarshal(bs []byte) (v int, n int, err error) {
	return loadWord(bs)
}

func (s wordIntView) Size(v int) (size int) {
	return wordSize
}

func (s wordIntView) Skip(bs []byte) (n int, err error) {
	return skipWord(bs)
}

// Float64 serializes a float64.
var Float64 = f64View{}

type f64View struct{}

func (s f64View) Marshal(v float64, bs []byte) (n int) {
	return writeLE64(math.Float64bits(v), bs)
}

func (s f64View) Unmarshal(bs []byte) (v float64, n int, err error) {
	uv, n, err := loadLE64[uint64](bs)
	if err != nil {
		return
	}
	return math.Float64frombits(uv), n, nil
}

func (s f64View) Size(v float64) (size int) {
	return binco.Num64RawSize
}

func (s f64View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger64(bs)
}

// Float32 serializes a float32.
var Float32 = f32View{}

type f32View struct{}

func (s f32View) Marshal(v float32, bs []byte) (n int) {
	return writeLE32(math.Float32bits(v), bs)
}

func (s f32View) Unmarshal(bs []byte) (v float32, n int, err error) {
	uv, n, err := loadLE32[uint32](bs)
	if err != nil {
		return
	}
	return math.Float32frombits(uv), n, nil
}

func (s f32View) Size(v float32) (size int) {
	return binco.Num32RawSize
}

func (s f32View) Skip(bs []byte) (n int, err error) {
	return scalar.SkipInteger32(bs)
}
