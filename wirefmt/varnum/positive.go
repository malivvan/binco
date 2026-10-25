package varnum

import (
	"math"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// PositiveInt8 serializes a non-negative int8 without ZigZag mapping.
//
// A decoded value above math.MaxInt8 is reported as ErrOverflow instead of
// wrapping to a negative number; marshalling a negative value panics with
// ErrOverflow.
var PositiveInt8 = posI8Codec{}

type posI8Codec struct{}

func (posI8Codec) Marshal(v int8, w wirefmt.Writer) (n int, err error) {
	if v < 0 {
		return 0, binco.ErrOverflow
	}
	return putUnsigned(uint8(v), w)
}

func (posI8Codec) Unmarshal(r wirefmt.Reader) (v int8, n int, err error) {
	uv, n, err := Uint8.Unmarshal(r)
	if err != nil {
		return
	}
	if uv > math.MaxInt8 {
		return 0, n, binco.ErrOverflow
	}
	return int8(uv), n, nil
}

func (posI8Codec) Size(v int8) (size int) {
	return Uint8.Size(uint8(v))
}

func (posI8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (posI8Codec{}).Unmarshal(r)
	return
}

// PositiveInt16 serializes a non-negative int16 without ZigZag mapping.
var PositiveInt16 = posI16Codec{}

type posI16Codec struct{}

func (posI16Codec) Marshal(v int16, w wirefmt.Writer) (n int, err error) {
	if v < 0 {
		return 0, binco.ErrOverflow
	}
	return putUnsigned(uint16(v), w)
}

func (posI16Codec) Unmarshal(r wirefmt.Reader) (v int16, n int, err error) {
	uv, n, err := Uint16.Unmarshal(r)
	if err != nil {
		return
	}
	if uv > math.MaxInt16 {
		return 0, n, binco.ErrOverflow
	}
	return int16(uv), n, nil
}

func (posI16Codec) Size(v int16) (size int) {
	return Uint16.Size(uint16(v))
}

func (posI16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (posI16Codec{}).Unmarshal(r)
	return
}

// PositiveInt32 serializes a non-negative int32 without ZigZag mapping.
var PositiveInt32 = posI32Codec{}

type posI32Codec struct{}

func (posI32Codec) Marshal(v int32, w wirefmt.Writer) (n int, err error) {
	if v < 0 {
		return 0, binco.ErrOverflow
	}
	return putUnsigned(uint32(v), w)
}

func (posI32Codec) Unmarshal(r wirefmt.Reader) (v int32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(r)
	if err != nil {
		return
	}
	if uv > math.MaxInt32 {
		return 0, n, binco.ErrOverflow
	}
	return int32(uv), n, nil
}

func (posI32Codec) Size(v int32) int {
	return Uint32.Size(uint32(v))
}

func (posI32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (posI32Codec{}).Unmarshal(r)
	return
}

// PositiveInt64 serializes a non-negative int64 without ZigZag mapping.
var PositiveInt64 = posI64Codec{}

type posI64Codec struct{}

func (posI64Codec) Marshal(v int64, w wirefmt.Writer) (n int, err error) {
	if v < 0 {
		return 0, binco.ErrOverflow
	}
	return putUnsigned(uint64(v), w)
}

func (posI64Codec) Unmarshal(r wirefmt.Reader) (v int64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(r)
	if err != nil {
		return
	}
	if uv > math.MaxInt64 {
		return 0, n, binco.ErrOverflow
	}
	return int64(uv), n, nil
}

func (posI64Codec) Size(v int64) int {
	return unsignedLen(uint64(v))
}

func (posI64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (posI64Codec{}).Unmarshal(r)
	return
}

// PositiveInt serializes a non-negative int without ZigZag mapping.
//
// The wire width is always 64 bits; values that do not fit the host int are
// reported as ErrOverflow.
var PositiveInt = posWordCodec{}

type posWordCodec struct{}

func (posWordCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	if v < 0 {
		return 0, binco.ErrOverflow
	}
	return Uint64.Marshal(uint64(v), w)
}

func (posWordCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	uv, n, err := Uint64.Unmarshal(r)
	if err != nil {
		return
	}
	iv := int64(uv)
	if uv > math.MaxInt64 || int64(int(iv)) != iv {
		return 0, n, binco.ErrOverflow
	}
	return int(iv), n, nil
}

func (posWordCodec) Size(v int) (size int) {
	return Uint64.Size(uint64(v))
}

func (posWordCodec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (posWordCodec{}).Unmarshal(r)
	return
}
