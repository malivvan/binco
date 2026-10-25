package zerocpy

import (
	"io"
	"strconv"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/scalar"
)

// writeLE64 writes a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, w wirefmt.Writer) (n int,
	err error,
) {
	arr := *(*[binco.Num64RawSize]byte)(unsafe_mod.Pointer(&t))
	return w.Write(arr[:])
}

// writeLE32 writes a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, w wirefmt.Writer) (n int,
	err error,
) {
	arr := *(*[binco.Num32RawSize]byte)(unsafe_mod.Pointer(&t))
	return w.Write(arr[:])
}

// writeLE16 writes a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, w wirefmt.Writer) (n int,
	err error,
) {
	arr := *(*[binco.Num16RawSize]byte)(unsafe_mod.Pointer(&t))
	return w.Write(arr[:])
}

// writeLE8 writes an 8-bit value.
func writeLE8[T binco.Integer8](t T, w wirefmt.Writer) (n int,
	err error,
) {
	arr := *(*[binco.Num8RawSize]byte)(unsafe_mod.Pointer(&t))
	return w.Write(arr[:])
}

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](r wirefmt.Reader) (t T, n int,
	err error,
) {
	b := (*byte)(unsafe_mod.Pointer(&t))
	n, err = io.ReadFull(r, unsafe_mod.Slice(b, binco.Num64RawSize))
	return
}

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](r wirefmt.Reader) (t T, n int,
	err error,
) {
	b := (*byte)(unsafe_mod.Pointer(&t))
	n, err = io.ReadFull(r, unsafe_mod.Slice(b, binco.Num32RawSize))
	return
}

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](r wirefmt.Reader) (t T, n int,
	err error,
) {
	b := (*byte)(unsafe_mod.Pointer(&t))
	n, err = io.ReadFull(r, unsafe_mod.Slice(b, binco.Num16RawSize))
	return
}

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](r wirefmt.Reader) (t T, n int,
	err error,
) {
	b := (*byte)(unsafe_mod.Pointer(&t))
	n, err = io.ReadFull(r, unsafe_mod.Slice(b, binco.Num8RawSize))
	return
}

// writeWord writes an int of the host word size.
var writeWord func(v int, w wirefmt.Writer) (int, error)

// loadWord reads an int of the host word size.
var loadWord func(r wirefmt.Reader) (int, int, error)

// wordSize is the width of an int on this host.
var wordSize func(v int) int

// skipWord advances over an int of the host word size.
var skipWord func(r wirefmt.Reader) (int, error)

// writeWordUnsigned writes a uint of the host word size.
var writeWordUnsigned func(v uint, w wirefmt.Writer) (int, error)

// loadWordUnsigned reads a uint of the host word size.
var loadWordUnsigned func(r wirefmt.Reader) (uint, int, error)

// wordUnsignedSize is the width of a uint on this host.
var wordUnsignedSize func(v uint) int

// skipWordUnsigned advances over a uint of the host word size.
var skipWordUnsigned func(r wirefmt.Reader) (int, error)

// bindWordInt selects the int helpers for the host word size.
func bindWordInt(intSize int) {
	switch intSize {
	case 64:
		writeWord = writeLE64[int]
		loadWord = loadLE64[int]
	case 32:
		writeWord = writeLE32[int]
		loadWord = loadLE32[int]
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
	wordSize = scalar.Int.Size
	skipWord = scalar.Int.Skip
}

// bindWordUnsigned selects the uint helpers for the host word size.
func bindWordUnsigned(intSize int) {
	switch intSize {
	case 64:
		writeWordUnsigned = writeLE64[uint]
		loadWordUnsigned = loadLE64[uint]
	case 32:
		writeWordUnsigned = writeLE32[uint]
		loadWordUnsigned = loadLE32[uint]
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
	wordUnsignedSize = scalar.Uint.Size
	skipWordUnsigned = scalar.Uint.Skip
}

func init() {
	bindWordInt(strconv.IntSize)
}

func init() {
	bindWordUnsigned(strconv.IntSize)
}
