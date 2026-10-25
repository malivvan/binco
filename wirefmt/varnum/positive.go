package varnum

import (
	"github.com/malivvan/binco/wirefmt"
)

// PositiveInt8 serializes a non-negative int8 without ZigZag mapping.
var PositiveInt8 = posI8Codec{}

type posI8Codec struct{}

func (posI8Codec) Marshal(v int8, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint8(v), w)
}

func (posI8Codec) Unmarshal(r wirefmt.Reader) (v int8, n int, err error) {
	uv, n, err := Uint8.Unmarshal(r)
	v = int8(uv)
	return
}

func (posI8Codec) Size(v int8) (size int) {
	return Uint8.Size(uint8(v))
}

func (posI8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint8.Skip(r)
}

// PositiveInt16 serializes a non-negative int16 without ZigZag mapping.
var PositiveInt16 = posI16Codec{}

type posI16Codec struct{}

func (posI16Codec) Marshal(v int16, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint16(v), w)
}

func (posI16Codec) Unmarshal(r wirefmt.Reader) (v int16, n int, err error) {
	uv, n, err := Uint16.Unmarshal(r)
	v = int16(uv)
	return
}

func (posI16Codec) Size(v int16) (size int) {
	return Uint16.Size(uint16(v))
}

func (posI16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint16.Skip(r)
}

// PositiveInt32 serializes a non-negative int32 without ZigZag mapping.
var PositiveInt32 = posI32Codec{}

type posI32Codec struct{}

func (posI32Codec) Marshal(v int32, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint32(v), w)
}

func (posI32Codec) Unmarshal(r wirefmt.Reader) (v int32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(r)
	v = int32(uv)
	return
}

func (posI32Codec) Size(v int32) int {
	return Uint32.Size(uint32(v))
}

func (posI32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint32.Skip(r)
}

// PositiveInt64 serializes a non-negative int64 without ZigZag mapping.
var PositiveInt64 = posI64Codec{}

type posI64Codec struct{}

func (posI64Codec) Marshal(v int64, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint64(v), w)
}

func (posI64Codec) Unmarshal(r wirefmt.Reader) (v int64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(r)
	v = int64(uv)
	return
}

func (posI64Codec) Size(v int64) int {
	return unsignedLen(uint64(v))
}

func (posI64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint64.Skip(r)
}

// PositiveInt serializes a non-negative int without ZigZag mapping.
var PositiveInt = posWordCodec{}

type posWordCodec struct{}

func (posWordCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint(v), w)
}

func (posWordCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	uv, n, err := Uint.Unmarshal(r)
	v = int(uv)
	return
}

func (posWordCodec) Size(v int) (size int) {
	return Uint.Size(uint(v))
}

func (posWordCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint.Skip(r)
}
