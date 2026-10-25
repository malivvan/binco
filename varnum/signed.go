package varnum

import "github.com/malivvan/binco"

// EncodeZigZag maps a signed value onto the unsigned range, so that -1, 1, -2, 2 become 1, 2, 3, 4.
func EncodeZigZag[T binco.Signed](t T) T {
	if t < 0 {
		return ^(t << 1)
	}
	return t << 1
}

// DecodeZigZag is the inverse of EncodeZigZag.
func DecodeZigZag[T binco.Unsigned](t T) T {
	if t&1 == 1 {
		return ^(t >> 1)
	}
	return t >> 1
}

// Int8 serializes an int8.
var Int8 = i8Codec{}

type i8Codec struct{}

func (s i8Codec) Marshal(v int8, bs []byte) (n int) {
	return putUnsigned(uint8(EncodeZigZag(v)), bs)
}

func (s i8Codec) Unmarshal(bs []byte) (v int8, n int, err error) {
	uv, n, err := Uint8.Unmarshal(bs)
	if err != nil {
		return
	}
	return int8(DecodeZigZag(uv)), n, nil
}

func (s i8Codec) Size(v int8) (size int) {
	return unsignedLen(uint8(EncodeZigZag(v)))
}

func (s i8Codec) Skip(bs []byte) (n int, err error) {
	return Uint8.Skip(bs)
}

// Int16 serializes an int16.
var Int16 = i16Codec{}

type i16Codec struct{}

func (s i16Codec) Marshal(v int16, bs []byte) (n int) {
	return putUnsigned(uint16(EncodeZigZag(v)), bs)
}

func (s i16Codec) Unmarshal(bs []byte) (v int16, n int, err error) {
	uv, n, err := Uint16.Unmarshal(bs)
	if err != nil {
		return
	}
	return int16(DecodeZigZag(uv)), n, nil
}

func (s i16Codec) Size(v int16) (size int) {
	return unsignedLen(uint16(EncodeZigZag(v)))
}

func (s i16Codec) Skip(bs []byte) (n int, err error) {
	return Uint16.Skip(bs)
}

// Int32 serializes an int32.
var Int32 = i32Codec{}

type i32Codec struct{}

func (s i32Codec) Marshal(v int32, bs []byte) (n int) {
	return putUnsigned(uint32(EncodeZigZag(v)), bs)
}

func (s i32Codec) Unmarshal(bs []byte) (v int32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(bs)
	if err != nil {
		return
	}
	return int32(DecodeZigZag(uv)), n, nil
}

func (s i32Codec) Size(v int32) (size int) {
	return unsignedLen(uint32(EncodeZigZag(v)))
}

func (s i32Codec) Skip(bs []byte) (n int, err error) {
	return Uint32.Skip(bs)
}

// Int64 serializes an int64.
var Int64 = i64Codec{}

type i64Codec struct{}

func (s i64Codec) Marshal(v int64, bs []byte) (n int) {
	return putUnsigned(uint64(EncodeZigZag(v)), bs)
}

func (s i64Codec) Unmarshal(bs []byte) (v int64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(bs)
	if err != nil {
		return
	}
	return int64(DecodeZigZag(uv)), n, nil
}

func (s i64Codec) Size(v int64) (size int) {
	return unsignedLen(uint64(EncodeZigZag(v)))
}

func (s i64Codec) Skip(bs []byte) (n int, err error) {
	return Uint64.Skip(bs)
}

// Int serializes an int, following the word size of the host.
var Int = IntSerializer{}

type IntSerializer struct{}

func (s IntSerializer) Marshal(v int, bs []byte) (n int) {
	return putUnsigned(uint(EncodeZigZag(v)), bs)
}

func (s IntSerializer) Unmarshal(bs []byte) (v int, n int, err error) {
	uv, n, err := Uint.Unmarshal(bs)
	if err != nil {
		return
	}
	return int(DecodeZigZag(uv)), n, nil
}

func (s IntSerializer) Size(v int) (size int) {
	return unsignedLen(uint(EncodeZigZag(v)))
}

func (s IntSerializer) Skip(bs []byte) (n int, err error) {
	return Uint.Skip(bs)
}
