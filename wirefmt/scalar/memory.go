package scalar

import (
	"strconv"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// writeLE64 writes a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, w wirefmt.Writer) (n int,
	err error,
) {
	err = w.WriteByte(byte(t))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 8))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 16))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 24))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 32))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 40))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 48))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 56))
	if err != nil {
		return
	}
	n++
	return
}

// writeLE32 writes a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, w wirefmt.Writer) (n int,
	err error,
) {
	err = w.WriteByte(byte(t))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 8))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 16))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 24))
	if err != nil {
		return
	}
	n++
	return
}

// writeLE16 writes a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, w wirefmt.Writer) (n int,
	err error,
) {
	err = w.WriteByte(byte(t))
	if err != nil {
		return
	}
	n++
	err = w.WriteByte(byte(t >> 8))
	if err != nil {
		return
	}
	n++
	return
}

// writeLE8 writes an 8-bit value.
func writeLE8[T binco.Integer8](t T, w wirefmt.Writer) (n int,
	err error,
) {
	err = w.WriteByte(byte(t))
	if err != nil {
		return
	}
	n++
	return
}

// readLE64 reads a 64-bit value from little-endian bytes.
func readLE64[T binco.Integer64](r wirefmt.Reader) (t T, n int,
	err error,
) {
	var b byte
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t = T(b)
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 8
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 16
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 24
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 32
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 40
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 48
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 56
	return
}

// readLE32 reads a 32-bit value from little-endian bytes.
func readLE32[T binco.Integer32](r wirefmt.Reader) (t T, n int,
	err error,
) {
	var b byte
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t = T(b)
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 8
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 16
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 24
	return
}

// readLE16 reads a 16-bit value from little-endian bytes.
func readLE16[T binco.Integer16](r wirefmt.Reader) (t T, n int,
	err error,
) {
	var b byte
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t = T(b)
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t |= T(b) << 8
	return
}

// readLE8 reads an 8-bit value.
func readLE8[T binco.Integer8](r wirefmt.Reader) (t T, n int,
	err error,
) {
	var b byte
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	n++
	t = T(b)
	return
}

// skipLE64 advances over eight bytes.
func skipLE64(r wirefmt.Reader) (int, error) {
	return skipBytes(binco.Num64RawSize, r)
}

// skipLE32 advances over four bytes.
func skipLE32(r wirefmt.Reader) (int, error) {
	return skipBytes(binco.Num32RawSize, r)
}

// skipLE16 advances over two bytes.
func skipLE16(r wirefmt.Reader) (int, error) {
	return skipBytes(binco.Num16RawSize, r)
}

// skipLE8 advances over one byte.
func skipLE8(r wirefmt.Reader) (int, error) {
	return skipBytes(binco.Num8RawSize, r)
}

// skipBytes advances over size bytes.
func skipBytes(integerSize int, r wirefmt.Reader) (n int, err error) {
	for range integerSize {
		_, err = r.ReadByte()
		if err != nil {
			return
		}
		n++
	}
	return
}

// widthOf64 is the encoded width of a 64-bit value.
func widthOf64[T binco.Num64](t T) int {
	return binco.Num64RawSize
}

// widthOf32 is the encoded width of a 32-bit value.
func widthOf32[T binco.Num32](t T) int {
	return binco.Num32RawSize
}

// writeWord writes an int of the host word size.
var writeWord func(v int, w wirefmt.Writer) (int, error)

// readWord reads an int of the host word size.
var readWord func(r wirefmt.Reader) (int, int, error)

// wordSize is the width of an int on this host.
var wordSize func(v int) int

// skipWord advances over an int of the host word size.
var skipWord func(r wirefmt.Reader) (int, error)

// writeWordUnsigned writes a uint of the host word size.
var writeWordUnsigned func(v uint, w wirefmt.Writer) (int, error)

// readWordUnsigned reads a uint of the host word size.
var readWordUnsigned func(r wirefmt.Reader) (uint, int, error)

// wordUnsignedSize is the width of a uint on this host.
var wordUnsignedSize func(v uint) int

// skipWordUnsigned advances over a uint of the host word size.
var skipWordUnsigned func(r wirefmt.Reader) (int, error)

// bindWordInt selects the int helpers for the host word size.
func bindWordInt(intSize int) {
	switch intSize {
	case 64:
		writeWord = writeLE64[int]
		readWord = readLE64[int]
		wordSize = widthOf64[int]
		skipWord = skipLE64
	case 32:
		writeWord = writeLE32[int]
		readWord = readLE32[int]
		wordSize = widthOf32[int]
		skipWord = skipLE32
	default:
		panic(binco.ErrUnsupportedIntSize)
	}
}

// bindWordUnsigned selects the uint helpers for the host word size.
func bindWordUnsigned(intSize int) {
	switch intSize {
	case 64:
		writeWordUnsigned = writeLE64[uint]
		readWordUnsigned = readLE64[uint]
		wordUnsignedSize = widthOf64[uint]
		skipWordUnsigned = skipLE64
	case 32:
		writeWordUnsigned = writeLE32[uint]
		readWordUnsigned = readLE32[uint]
		wordUnsignedSize = widthOf32[uint]
		skipWordUnsigned = skipLE32
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
