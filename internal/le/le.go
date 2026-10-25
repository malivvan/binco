// Package le reads and writes little-endian fixed-width values in a byte slice.
//
// The buffer codecs of package scalar and package zerocopy need the same
// byte-level work; this package is the one implementation they share. The
// encoding is always little-endian, whatever the host byte order is, and every
// function checks the input or output length before touching it.
package le

import (
	"encoding/binary"
	"strconv"

	"github.com/malivvan/binco"
)

// Put64 writes t as eight little-endian bytes and reports the size.
func Put64[T binco.Integer64](bs []byte, t T) int {
	binco.CheckWrite(bs, 0, binco.Num64RawSize)
	binary.LittleEndian.PutUint64(bs, uint64(t))
	return binco.Num64RawSize
}

// Put32 writes t as four little-endian bytes and reports the size.
func Put32[T binco.Integer32](bs []byte, t T) int {
	binco.CheckWrite(bs, 0, binco.Num32RawSize)
	binary.LittleEndian.PutUint32(bs, uint32(t))
	return binco.Num32RawSize
}

// Put16 writes t as two little-endian bytes and reports the size.
func Put16[T binco.Integer16](bs []byte, t T) int {
	binco.CheckWrite(bs, 0, binco.Num16RawSize)
	binary.LittleEndian.PutUint16(bs, uint16(t))
	return binco.Num16RawSize
}

// Put8 writes t as one byte and reports the size.
func Put8[T binco.Integer8](bs []byte, t T) int {
	binco.CheckWrite(bs, 0, binco.Num8RawSize)
	bs[0] = byte(t)
	return binco.Num8RawSize
}

// Get64 reads eight little-endian bytes.
func Get64[T binco.Integer64](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num64RawSize {
		return t, 0, binco.ErrTruncated
	}
	return T(binary.LittleEndian.Uint64(bs)), binco.Num64RawSize, nil
}

// Get32 reads four little-endian bytes.
func Get32[T binco.Integer32](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num32RawSize {
		return t, 0, binco.ErrTruncated
	}
	return T(binary.LittleEndian.Uint32(bs)), binco.Num32RawSize, nil
}

// Get16 reads two little-endian bytes.
func Get16[T binco.Integer16](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num16RawSize {
		return t, 0, binco.ErrTruncated
	}
	return T(binary.LittleEndian.Uint16(bs)), binco.Num16RawSize, nil
}

// Get8 reads one byte.
func Get8[T binco.Integer8](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num8RawSize {
		return t, 0, binco.ErrTruncated
	}
	return T(bs[0]), binco.Num8RawSize, nil
}

// Skip64 advances over eight bytes.
func Skip64(bs []byte) (int, error) { return skip(bs, binco.Num64RawSize) }

// Skip32 advances over four bytes.
func Skip32(bs []byte) (int, error) { return skip(bs, binco.Num32RawSize) }

// Skip16 advances over two bytes.
func Skip16(bs []byte) (int, error) { return skip(bs, binco.Num16RawSize) }

// Skip8 advances over one byte.
func Skip8(bs []byte) (int, error) { return skip(bs, binco.Num8RawSize) }

func skip(bs []byte, size int) (int, error) {
	if len(bs) < size {
		return 0, binco.ErrTruncated
	}
	return size, nil
}

// PutWord writes an int in the wire word width, which is always 64 bits.
func PutWord(bs []byte, v int) int {
	return Put64(bs, int64(v))
}

// GetWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func GetWord(bs []byte) (int, int, error) {
	raw, n, err := Get64[int64](bs)
	if err != nil {
		return 0, n, err
	}
	v, err := binco.WordInt(raw, strconv.IntSize)
	return v, n, err
}

// SkipWord advances over an int of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func SkipWord(bs []byte) (int, error) {
	_, n, err := GetWord(bs)
	return n, err
}

// PutWordUnsigned writes a uint in the wire word width.
func PutWordUnsigned(bs []byte, v uint) int {
	return Put64(bs, uint64(v))
}

// GetWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func GetWordUnsigned(bs []byte) (uint, int, error) {
	raw, n, err := Get64[uint64](bs)
	if err != nil {
		return 0, n, err
	}
	v, err := binco.WordUint(raw, strconv.IntSize)
	return v, n, err
}

// SkipWordUnsigned advances over a uint of the wire word width.
func SkipWordUnsigned(bs []byte) (int, error) {
	_, n, err := GetWordUnsigned(bs)
	return n, err
}
