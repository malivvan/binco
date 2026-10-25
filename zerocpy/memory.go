package zerocpy

import (
	"strconv"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
)

// writeLE64 stores a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, bs []byte) (n int) {
	*(*T)(unsafe_mod.Pointer(&bs[0])) = t
	return binco.Num64RawSize
}

// writeLE32 stores a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, bs []byte) (n int) {
	*(*T)(unsafe_mod.Pointer(&bs[0])) = t
	return binco.Num32RawSize
}

// writeLE16 stores a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, bs []byte) (n int) {
	*(*T)(unsafe_mod.Pointer(&bs[0])) = t
	return binco.Num16RawSize
}

// writeLE8 stores an 8-bit value.
func writeLE8[T binco.Integer8](t T, bs []byte) (n int) {
	*(*T)(unsafe_mod.Pointer(&bs[0])) = t
	return binco.Num8RawSize
}

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num64RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return *(*T)(unsafe_mod.Pointer(&bs[0])), binco.Num64RawSize, nil
}

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num32RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return *(*T)(unsafe_mod.Pointer(&bs[0])), binco.Num32RawSize, nil
}

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num16RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return *(*T)(unsafe_mod.Pointer(&bs[0])), binco.Num16RawSize, nil
}

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num8RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return *(*T)(unsafe_mod.Pointer(&bs[0])), binco.Num8RawSize, nil
}

// writeWord stores an int of the host word size.
var writeWord func(v int, bs []byte) int

// loadWord reads an int of the host word size.
var loadWord func(bs []byte) (int, int, error)

// wordSize is the width of an int on this host.
var wordSize int

// skipWord advances over an int of the host word size.
var skipWord func(bs []byte) (int, error)

// writeWordUnsigned stores a uint of the host word size.
var writeWordUnsigned func(v uint, bs []byte) int

// loadWordUnsigned reads a uint of the host word size.
var loadWordUnsigned func(bs []byte) (uint, int, error)

// wordUnsignedSize is the width of a uint on this host.
var wordUnsignedSize int

// skipWordUnsigned advances over a uint of the host word size.
var skipWordUnsigned func(bs []byte) (int, error)

// bindWordInt selects the int helpers for the host word size.
func bindWordInt(intSize int) {
	switch intSize {
	case 64:
		writeWord = writeLE64[int]
		loadWord = loadLE64[int]
		wordSize = binco.Num64RawSize
		skipWord = scalar.SkipInteger64
	case 32:
		writeWord = writeLE32[int]
		loadWord = loadLE32[int]
		wordSize = binco.Num32RawSize
		skipWord = scalar.SkipInteger32
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
}

// bindWordUnsigned selects the uint helpers for the host word size.
func bindWordUnsigned(intSize int) {
	switch intSize {
	case 64:
		writeWordUnsigned = writeLE64[uint]
		loadWordUnsigned = loadLE64[uint]
		wordUnsignedSize = binco.Num64RawSize
		skipWordUnsigned = scalar.SkipInteger64
	case 32:
		writeWordUnsigned = writeLE32[uint]
		loadWordUnsigned = loadLE32[uint]
		wordUnsignedSize = binco.Num32RawSize
		skipWordUnsigned = scalar.SkipInteger32
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
