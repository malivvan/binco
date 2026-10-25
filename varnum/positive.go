package varnum

// PositiveInt8 serializes a non-negative int8 without ZigZag mapping.
var PositiveInt8 = posI8Codec{}

type posI8Codec struct{}

func (s posI8Codec) Marshal(v int8, bs []byte) (n int) {
	return putUnsigned(uint8(v), bs)
}

func (s posI8Codec) Unmarshal(bs []byte) (v int8, n int, err error) {
	var uv uint8
	uv, n, err = Uint8.Unmarshal(bs)
	v = int8(uv)
	return
}

func (s posI8Codec) Size(v int8) (size int) {
	return unsignedLen(uint8(v))
}

func (s posI8Codec) Skip(bs []byte) (n int, err error) {
	return Uint8.Skip(bs)
}

// PositiveInt16 serializes a non-negative int16 without ZigZag mapping.
var PositiveInt16 = posI16Codec{}

type posI16Codec struct{}

func (s posI16Codec) Marshal(v int16, bs []byte) (n int) {
	return putUnsigned(uint16(v), bs)
}

func (s posI16Codec) Unmarshal(bs []byte) (v int16, n int, err error) {
	var uv uint16
	uv, n, err = Uint16.Unmarshal(bs)
	v = int16(uv)
	return
}

func (s posI16Codec) Size(v int16) (size int) {
	return unsignedLen(uint16(v))
}

func (s posI16Codec) Skip(bs []byte) (n int, err error) {
	return Uint16.Skip(bs)
}

// PositiveInt32 serializes a non-negative int32 without ZigZag mapping.
var PositiveInt32 = posI32Codec{}

type posI32Codec struct{}

func (s posI32Codec) Marshal(v int32, bs []byte) (n int) {
	return putUnsigned(uint32(v), bs)
}

func (s posI32Codec) Unmarshal(bs []byte) (v int32, n int, err error) {
	var uv uint32
	uv, n, err = Uint32.Unmarshal(bs)
	v = int32(uv)
	return
}

func (s posI32Codec) Size(v int32) (size int) {
	return unsignedLen(uint32(v))
}

func (s posI32Codec) Skip(bs []byte) (n int, err error) {
	return Uint32.Skip(bs)
}

// PositiveInt64 serializes a non-negative int64 without ZigZag mapping.
var PositiveInt64 = posI64Codec{}

type posI64Codec struct{}

func (s posI64Codec) Marshal(v int64, bs []byte) (n int) {
	return putUnsigned(uint64(v), bs)
}

func (s posI64Codec) Unmarshal(bs []byte) (v int64, n int, err error) {
	var uv uint64
	uv, n, err = Uint64.Unmarshal(bs)
	v = int64(uv)
	return
}

func (s posI64Codec) Size(v int64) (size int) {
	return unsignedLen(uint64(v))
}

func (s posI64Codec) Skip(bs []byte) (n int, err error) {
	return Uint64.Skip(bs)
}

// PositiveInt serializes a non-negative int without ZigZag mapping.
var PositiveInt = posWordCodec{}

type posWordCodec struct{}

func (s posWordCodec) Marshal(v int, bs []byte) (n int) {
	return putUnsigned(uint(v), bs)
}

func (s posWordCodec) Unmarshal(bs []byte) (v int, n int, err error) {
	uv, n, err := Uint.Unmarshal(bs)
	if err != nil {
		return
	}
	return int(uv), n, nil
}

func (s posWordCodec) Size(v int) (size int) {
	return unsignedLen(uint(v))
}

func (s posWordCodec) Skip(bs []byte) (n int, err error) {
	return Uint.Skip(bs)
}
