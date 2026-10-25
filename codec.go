package binco

import (
	"fmt"
	"slices"
)

// Serializer is the contract every codec in this module implements: encode into a caller-owned buffer, decode from one, report the encoded size, and skip a value.
//
// Size and Marshal are separate calls, not one atomic operation: a value that is
// mutated between them (by another goroutine, or by a map that is written while
// it is being encoded) can make Marshal write past the buffer Size reported.
// Encode one immutable value, or hold the lock that protects it across both.
type Serializer[T any] interface {
	Marshal(t T, bs []byte) (n int)
	Unmarshal(bs []byte) (t T, n int, err error)
	Size(t T) (size int)
	Skip(bs []byte) (n int, err error)
}

// Marshaller lets a value encode itself, which is what binco.Marshal and generated code rely on.
type Marshaller interface {
	MarshalBinco(bs []byte) (n int)
	SizeBinco() (size int)
}

// MarshallerTyped is Marshaller with a type identifier in front of the payload.
type MarshallerTyped interface {
	MarshalTypedBinco(bs []byte) (n int)
	SizeTypedBinco() (size int)
}

// Marshal allocates a buffer of SizeBinco bytes and lets v fill it.
func Marshal(v Marshaller) (bs []byte) {
	bs = make([]byte, v.SizeBinco())
	v.MarshalBinco(bs)
	return
}

// MarshalTyped is Marshal with the type identifier included.
func MarshalTyped(v MarshallerTyped) (bs []byte) {
	bs = make([]byte, v.SizeTypedBinco())
	v.MarshalTypedBinco(bs)
	return
}

// Encode returns a fresh buffer holding the encoded v.
func Encode[T any](s Serializer[T], v T) (bs []byte) {
	bs = make([]byte, s.Size(v))
	s.Marshal(v, bs)
	return
}

// Decode unmarshals v and returns only the value and the error.
func Decode[T any](s Serializer[T], bs []byte) (v T, err error) {
	v, _, err = s.Unmarshal(bs)
	return
}

// DecodeExact is Decode, but refuses input that continues after the value.
//
// Unmarshal deliberately ignores trailing bytes, which lets a caller decode the
// first field of a longer record. A protocol that carries exactly one value per
// message wants the opposite, and this is the variant that provides it: it
// returns ErrTrailingBytes instead of silently dropping the rest.
func DecodeExact[T any](s Serializer[T], bs []byte) (v T, err error) {
	var n int
	v, n, err = s.Unmarshal(bs)
	if err != nil {
		return
	}
	if n != len(bs) {
		err = ErrTrailingBytes
	}
	return
}

// AppendEncode appends the encoded form of v to dst and returns the grown
// slice.
//
// dst comes first, like in append. It is the form to use with a reusable
// buffer: Size is exact, so dst is grown once and no temporary allocation is
// needed.
func AppendEncode[T any](dst []byte, s Serializer[T], v T) []byte {
	size := s.Size(v)
	start := len(dst)
	dst = slices.Grow(dst, size)
	dst = dst[:start+size]
	if n := s.Marshal(v, dst[start:]); n != size {
		panic(fmt.Errorf("%w: serializer wrote %d of %d bytes", ErrWrongSize, n, size))
	}
	return dst
}

// AppendMarshal is AppendEncode for a value that encodes itself.
func AppendMarshal(dst []byte, v Marshaller) []byte {
	size := v.SizeBinco()
	start := len(dst)
	dst = slices.Grow(dst, size)
	dst = dst[:start+size]
	if n := v.MarshalBinco(dst[start:]); n != size {
		panic(fmt.Errorf("%w: marshaller wrote %d of %d bytes", ErrWrongSize, n, size))
	}
	return dst
}

// AppendMarshalTyped is AppendMarshal with the type identifier included.
func AppendMarshalTyped(dst []byte, v MarshallerTyped) []byte {
	size := v.SizeTypedBinco()
	start := len(dst)
	dst = slices.Grow(dst, size)
	dst = dst[:start+size]
	if n := v.MarshalTypedBinco(dst[start:]); n != size {
		panic(fmt.Errorf("%w: typed marshaller wrote %d of %d bytes", ErrWrongSize, n, size))
	}
	return dst
}
