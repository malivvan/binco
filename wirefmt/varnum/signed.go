package varnum

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// EncodeZigZag maps a signed value onto the unsigned range, so that -1, 1, -2, 2 become 1, 2, 3, 4.
func EncodeZigZag[T binco.Signed](t T) T {
	if t < 0 {
		return ^(t << 1)
	} else {
		return t << 1
	}
}

// DecodeZigZag is the inverse of EncodeZigZag.
func DecodeZigZag[T binco.Unsigned](t T) T {
	if t&1 == 1 {
		return ^(t >> 1)
	} else {
		return t >> 1
	}
}

// Int8 serializes an int8.
var Int8 = i8Codec{}

type i8Codec struct{}

func (i8Codec) Marshal(v int8, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint8(EncodeZigZag(v)), w)
}

func (i8Codec) Unmarshal(r wirefmt.Reader) (v int8, n int, err error) {
	uv, n, err := Uint8.Unmarshal(r)
	if err != nil {
		return
	}
	return int8(DecodeZigZag(uv)), n, nil
}

func (i8Codec) Size(v int8) (size int) {
	return Uint8.Size(uint8(EncodeZigZag(v)))
}

func (i8Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint8.Skip(r)
}

// Int16 serializes an int16.
var Int16 = i16Codec{}

type i16Codec struct{}

func (i16Codec) Marshal(v int16, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint16(EncodeZigZag(v)), w)
}

func (i16Codec) Unmarshal(r wirefmt.Reader) (v int16, n int, err error) {
	uv, n, err := Uint16.Unmarshal(r)
	if err != nil {
		return
	}
	return int16(DecodeZigZag(uv)), n, nil
}

func (i16Codec) Size(v int16) (size int) {
	return Uint16.Size(uint16(EncodeZigZag(v)))
}

func (i16Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint16.Skip(r)
}

// Int32 serializes an int32.
var Int32 = i32Codec{}

type i32Codec struct{}

func (i32Codec) Marshal(v int32, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint32(EncodeZigZag(v)), w)
}

func (i32Codec) Unmarshal(r wirefmt.Reader) (v int32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(r)
	if err != nil {
		return
	}
	return int32(DecodeZigZag(uv)), n, nil
}

func (i32Codec) Size(v int32) int {
	return Uint32.Size(uint32(EncodeZigZag(v)))
}

func (i32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint32.Skip(r)
}

// Int64 serializes an int64.
var Int64 = i64Codec{}

type i64Codec struct{}

func (i64Codec) Marshal(v int64, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint64(EncodeZigZag(v)), w)
}

func (i64Codec) Unmarshal(r wirefmt.Reader) (v int64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(r)
	if err != nil {
		return
	}
	return int64(DecodeZigZag(uv)), n, nil
}

func (i64Codec) Size(v int64) int {
	return unsignedLen(uint64(EncodeZigZag(v)))
}

func (i64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint64.Skip(r)
}

// Int serializes an int, following the word size of the host.
var Int = wordIntCodec{}

type wordIntCodec struct{}

func (wordIntCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(uint(EncodeZigZag(v)), w)
}

func (wordIntCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	uv, n, err := Uint.Unmarshal(r)
	if err != nil {
		return
	}
	return int(DecodeZigZag(uv)), n, nil
}

func (wordIntCodec) Size(v int) (size int) {
	return Uint.Size(uint(EncodeZigZag(v)))
}

func (wordIntCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint.Skip(r)
}
