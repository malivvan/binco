package varnum

import (
	"math"

	"github.com/malivvan/binco"
)

// PositiveInt8 serializes a non-negative int8 without ZigZag mapping.
//
// Unmarshalling a value above math.MaxInt8 returns ErrOverflow instead of
// wrapping to a negative number. Marshalling a negative value panics with
// ErrOverflow, because a positive codec cannot represent it.
var PositiveInt8 = posI8Codec{}

type posI8Codec struct{}

func (s posI8Codec) Marshal(v int8, bs []byte) (n int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return putUnsigned(uint8(v), bs)
}

func (s posI8Codec) Unmarshal(bs []byte) (v int8, n int, err error) {
	uv, n, err := Uint8.Unmarshal(bs)
	if err != nil {
		return
	}
	if uv > math.MaxInt8 {
		return 0, n, binco.ErrOverflow
	}
	return int8(uv), n, nil
}

func (s posI8Codec) Size(v int8) (size int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return unsignedLen(uint8(v))
}

func (s posI8Codec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}

// PositiveInt16 serializes a non-negative int16 without ZigZag mapping.
var PositiveInt16 = posI16Codec{}

type posI16Codec struct{}

func (s posI16Codec) Marshal(v int16, bs []byte) (n int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return putUnsigned(uint16(v), bs)
}

func (s posI16Codec) Unmarshal(bs []byte) (v int16, n int, err error) {
	uv, n, err := Uint16.Unmarshal(bs)
	if err != nil {
		return
	}
	if uv > math.MaxInt16 {
		return 0, n, binco.ErrOverflow
	}
	return int16(uv), n, nil
}

func (s posI16Codec) Size(v int16) (size int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return unsignedLen(uint16(v))
}

func (s posI16Codec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}

// PositiveInt32 serializes a non-negative int32 without ZigZag mapping.
var PositiveInt32 = posI32Codec{}

type posI32Codec struct{}

func (s posI32Codec) Marshal(v int32, bs []byte) (n int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return putUnsigned(uint32(v), bs)
}

func (s posI32Codec) Unmarshal(bs []byte) (v int32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(bs)
	if err != nil {
		return
	}
	if uv > math.MaxInt32 {
		return 0, n, binco.ErrOverflow
	}
	return int32(uv), n, nil
}

func (s posI32Codec) Size(v int32) (size int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return unsignedLen(uint32(v))
}

func (s posI32Codec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}

// PositiveInt64 serializes a non-negative int64 without ZigZag mapping.
var PositiveInt64 = posI64Codec{}

type posI64Codec struct{}

func (s posI64Codec) Marshal(v int64, bs []byte) (n int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return putUnsigned(uint64(v), bs)
}

func (s posI64Codec) Unmarshal(bs []byte) (v int64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(bs)
	if err != nil {
		return
	}
	if uv > math.MaxInt64 {
		return 0, n, binco.ErrOverflow
	}
	return int64(uv), n, nil
}

func (s posI64Codec) Size(v int64) (size int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return unsignedLen(uint64(v))
}

func (s posI64Codec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}

// PositiveInt serializes a non-negative int without ZigZag mapping.
//
// The wire width is always 64 bits. A value that does not fit the host int, or
// a negative value, is reported as ErrOverflow.
var PositiveInt = posWordCodec{}

type posWordCodec struct{}

func (s posWordCodec) Marshal(v int, bs []byte) (n int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return Uint64.Marshal(uint64(v), bs)
}

func (s posWordCodec) Unmarshal(bs []byte) (v int, n int, err error) {
	uv, n, err := Uint64.Unmarshal(bs)
	if err != nil {
		return
	}
	iv := int64(uv)
	if uv > math.MaxInt64 || int64(int(iv)) != iv {
		return 0, n, binco.ErrOverflow
	}
	return int(iv), n, nil
}

func (s posWordCodec) Size(v int) (size int) {
	if v < 0 {
		panic(binco.ErrOverflow)
	}
	return Uint64.Size(uint64(v))
}

func (s posWordCodec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}
