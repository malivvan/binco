package varnum

import (
	"math"

	"github.com/malivvan/binco/wirefmt"
)

// Float64 serializes a float64 through its IEEE 754 bit pattern.
var Float64 = f64Codec{}

type f64Codec struct{}

func (f64Codec) Marshal(v float64, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(math.Float64bits(v), w)
}

func (f64Codec) Unmarshal(r wirefmt.Reader) (v float64, n int, err error) {
	uv, n, err := Uint64.Unmarshal(r)
	if err != nil {
		return
	}
	return math.Float64frombits(uv), n, nil
}

func (f64Codec) Size(v float64) int {
	return unsignedLen(math.Float64bits(v))
}

func (f64Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint64.Skip(r)
}

// Float32 serializes a float32 through its IEEE 754 bit pattern.
var Float32 = f32Codec{}

type f32Codec struct{}

func (f32Codec) Marshal(v float32, w wirefmt.Writer) (n int, err error) {
	return putUnsigned(math.Float32bits(v), w)
}

func (f32Codec) Unmarshal(r wirefmt.Reader) (v float32, n int, err error) {
	uv, n, err := Uint32.Unmarshal(r)
	if err != nil {
		return
	}
	return math.Float32frombits(uv), n, nil
}

func (f32Codec) Size(v float32) int {
	return unsignedLen(math.Float32bits(v))
}

func (f32Codec) Skip(r wirefmt.Reader) (n int, err error) {
	return Uint32.Skip(r)
}
