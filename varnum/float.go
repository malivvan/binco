package varnum

import "math"

// Float64 serializes a float64 through its IEEE 754 bit pattern.
var Float64 = f64Codec{}

type f64Codec struct{}

func (s f64Codec) Marshal(v float64, bs []byte) (n int) {
	return putUnsigned(math.Float64bits(v), bs)
}

func (s f64Codec) Unmarshal(bs []byte) (v float64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(bs)
	if err != nil {
		return
	}
	v = math.Float64frombits(uv)
	return
}

func (s f64Codec) Size(v float64) (size int) {
	return unsignedLen(math.Float64bits(v))
}

func (s f64Codec) Skip(bs []byte) (n int, err error) {
	return Uint64.Skip(bs)
}

// Float32 serializes a float32 through its IEEE 754 bit pattern.
var Float32 = f32Codec{}

type f32Codec struct{}

func (s f32Codec) Marshal(v float32, bs []byte) (n int) {
	return putUnsigned(math.Float32bits(v), bs)
}

func (s f32Codec) Unmarshal(bs []byte) (v float32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(bs)
	if err != nil {
		return
	}
	v = math.Float32frombits(uv)
	return
}

func (s f32Codec) Size(v float32) (size int) {
	return unsignedLen(math.Float32bits(v))
}

func (s f32Codec) Skip(bs []byte) (n int, err error) {
	return Uint32.Skip(bs)
}
