package scalar

import (
	"strconv"

	"github.com/malivvan/binco"
)

// putLE64 writes t as eight little-endian bytes.
func putLE64[T binco.Integer64](t T, bs []byte) int {
	_ = bs[7]
	bs[0] = byte(t)
	bs[1] = byte(t >> 8)
	bs[2] = byte(t >> 16)
	bs[3] = byte(t >> 24)
	bs[4] = byte(t >> 32)
	bs[5] = byte(t >> 40)
	bs[6] = byte(t >> 48)
	bs[7] = byte(t >> 56)
	return binco.Num64RawSize
}

// putLE32 writes t as four little-endian bytes.
func putLE32[T binco.Integer32](t T, bs []byte) int {
	_ = bs[3]
	bs[0] = byte(t)
	bs[1] = byte(t >> 8)
	bs[2] = byte(t >> 16)
	bs[3] = byte(t >> 24)
	return binco.Num32RawSize
}

// putLE16 writes t as two little-endian bytes.
func putLE16[T binco.Integer16](t T, bs []byte) int {
	_ = bs[1]
	bs[0] = byte(t)
	bs[1] = byte(t >> 8)
	return binco.Num16RawSize
}

// putLE8 writes t as a single byte.
func putLE8[T binco.Integer8](t T, bs []byte) int {
	bs[0] = byte(t)
	return binco.Num8RawSize
}

// readLE64 reads eight little-endian bytes.
func readLE64[T binco.Integer64](bs []byte) (T, int, error) {
	var t T
	if len(bs) < binco.Num64RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	_ = bs[7]
	t = T(bs[0])
	t |= T(bs[1]) << 8
	t |= T(bs[2]) << 16
	t |= T(bs[3]) << 24
	t |= T(bs[4]) << 32
	t |= T(bs[5]) << 40
	t |= T(bs[6]) << 48
	t |= T(bs[7]) << 56
	return t, binco.Num64RawSize, nil
}

// readLE32 reads four little-endian bytes.
func readLE32[T binco.Integer32](bs []byte) (T, int, error) {
	var t T
	if len(bs) < binco.Num32RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	_ = bs[3]
	t = T(bs[0])
	t |= T(bs[1]) << 8
	t |= T(bs[2]) << 16
	t |= T(bs[3]) << 24
	return t, binco.Num32RawSize, nil
}

// readLE16 reads two little-endian bytes.
func readLE16[T binco.Integer16](bs []byte) (T, int, error) {
	var t T
	if len(bs) < binco.Num16RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	_ = bs[1]
	t = T(bs[0])
	t |= T(bs[1]) << 8
	return t, binco.Num16RawSize, nil
}

// readLE8 reads a single byte.
func readLE8[T binco.Integer8](bs []byte) (T, int, error) {
	var t T
	if len(bs) < binco.Num8RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return T(bs[0]), binco.Num8RawSize, nil
}

// SkipInteger64 advances over an eight-byte integer.
func SkipInteger64(bs []byte) (int, error) {
	if len(bs) < binco.Num64RawSize {
		return 0, binco.ErrTooSmallByteSlice
	}
	return binco.Num64RawSize, nil
}

// SkipInteger32 advances over a four-byte integer.
func SkipInteger32(bs []byte) (int, error) {
	if len(bs) < binco.Num32RawSize {
		return 0, binco.ErrTooSmallByteSlice
	}
	return binco.Num32RawSize, nil
}

// SkipInteger16 advances over a two-byte integer.
func SkipInteger16(bs []byte) (int, error) {
	if len(bs) < binco.Num16RawSize {
		return 0, binco.ErrTooSmallByteSlice
	}
	return binco.Num16RawSize, nil
}

// SkipInteger8 advances over a one-byte integer.
func SkipInteger8(bs []byte) (int, error) {
	if len(bs) < binco.Num8RawSize {
		return 0, binco.ErrTooSmallByteSlice
	}
	return binco.Num8RawSize, nil
}

// putWord points at the integer writer for the host word size.
var putWord func(v int, bs []byte) int

// readWord points at the integer reader for the host word size.
var readWord func(bs []byte) (int, int, error)

// wordSize is the number of bytes in an int.
var wordSize int

// skipWord points at the integer skipper for the host word size.
var skipWord func(bs []byte) (int, error)

// putWordUnsigned points at the unsigned writer for the host word size.
var putWordUnsigned func(v uint, bs []byte) int

// readWordUnsigned points at the unsigned reader for the host word size.
var readWordUnsigned func(bs []byte) (uint, int, error)

// wordUnsignedSize is the number of bytes in a uint.
var wordUnsignedSize int

// skipWordUnsigned points at the unsigned skipper for the host word size.
var skipWordUnsigned func(bs []byte) (int, error)

// bindWordInt selects the integer helpers for the host word size.
func bindWordInt(intSize int) {
	switch intSize {
	case 64:
		putWord = putLE64[int]
		readWord = readLE64[int]
		wordSize = binco.Num64RawSize
		skipWord = SkipInteger64
	case 32:
		putWord = putLE32[int]
		readWord = readLE32[int]
		wordSize = binco.Num32RawSize
		skipWord = SkipInteger32
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
}

// bindWordUnsigned selects the unsigned helpers for the host word size.
func bindWordUnsigned(intSize int) {
	switch intSize {
	case 64:
		putWordUnsigned = putLE64[uint]
		readWordUnsigned = readLE64[uint]
		wordUnsignedSize = binco.Num64RawSize
		skipWordUnsigned = SkipInteger64
	case 32:
		putWordUnsigned = putLE32[uint]
		readWordUnsigned = readLE32[uint]
		wordUnsignedSize = binco.Num32RawSize
		skipWordUnsigned = SkipInteger32
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
}

func init() {
	bindWordInt(strconv.IntSize)
}

func init() {
	bindWordUnsigned(strconv.IntSize)
}
