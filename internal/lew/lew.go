// Package lew reads and writes little-endian fixed-width values on a stream.
//
// It is the streaming counterpart of internal/le: the streaming codecs of
// package scalar and package zerocpy share this one implementation. Each value
// is assembled in a local buffer and written with a single Write, or read with a
// single io.ReadFull, instead of one call per byte.
package lew

import (
	"encoding/binary"
	"io"

	"github.com/malivvan/binco"
)

// Write64 writes t as eight little-endian bytes.
func Write64[T binco.Integer64](w io.Writer, t T) (int, error) {
	var buf [binco.Num64RawSize]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(t))
	return w.Write(buf[:])
}

// Write32 writes t as four little-endian bytes.
func Write32[T binco.Integer32](w io.Writer, t T) (int, error) {
	var buf [binco.Num32RawSize]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(t))
	return w.Write(buf[:])
}

// Write16 writes t as two little-endian bytes.
func Write16[T binco.Integer16](w io.Writer, t T) (int, error) {
	var buf [binco.Num16RawSize]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(t))
	return w.Write(buf[:])
}

// Write8 writes t as one byte.
func Write8[T binco.Integer8](w io.Writer, t T) (int, error) {
	var buf [binco.Num8RawSize]byte
	buf[0] = byte(t)
	return w.Write(buf[:])
}

// Read64 reads eight little-endian bytes.
func Read64[T binco.Integer64](r io.Reader) (t T, n int, err error) {
	var buf [binco.Num64RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint64(buf[:])), n, nil
}

// Read32 reads four little-endian bytes.
func Read32[T binco.Integer32](r io.Reader) (t T, n int, err error) {
	var buf [binco.Num32RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint32(buf[:])), n, nil
}

// Read16 reads two little-endian bytes.
func Read16[T binco.Integer16](r io.Reader) (t T, n int, err error) {
	var buf [binco.Num16RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint16(buf[:])), n, nil
}

// Read8 reads one byte.
func Read8[T binco.Integer8](r io.Reader) (t T, n int, err error) {
	var buf [binco.Num8RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(buf[0]), n, nil
}

// Skip64 advances over eight bytes.
func Skip64(r io.Reader) (int, error) { return skip(r, binco.Num64RawSize) }

// Skip32 advances over four bytes.
func Skip32(r io.Reader) (int, error) { return skip(r, binco.Num32RawSize) }

// Skip16 advances over two bytes.
func Skip16(r io.Reader) (int, error) { return skip(r, binco.Num16RawSize) }

// Skip8 advances over one byte.
func Skip8(r io.Reader) (int, error) { return skip(r, binco.Num8RawSize) }

func skip(r io.Reader, size int) (int, error) {
	var buf [binco.Num64RawSize]byte
	return io.ReadFull(r, buf[:size])
}

// WriteWord writes an int in the wire word width, which is always 64 bits.
func WriteWord(w io.Writer, v int) (int, error) {
	return Write64(w, int64(v))
}

// ReadWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func ReadWord(r io.Reader) (int, int, error) {
	raw, n, err := Read64[int64](r)
	if err != nil {
		return 0, n, err
	}
	v := int(raw)
	if int64(v) != raw {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// SkipWord advances over an int of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func SkipWord(r io.Reader) (int, error) {
	_, n, err := ReadWord(r)
	return n, err
}

// WriteWordUnsigned writes a uint in the wire word width.
func WriteWordUnsigned(w io.Writer, v uint) (int, error) {
	return Write64(w, uint64(v))
}

// ReadWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func ReadWordUnsigned(r io.Reader) (uint, int, error) {
	raw, n, err := Read64[uint64](r)
	if err != nil {
		return 0, n, err
	}
	v := uint(raw)
	if uint64(v) != raw {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// SkipWordUnsigned advances over a uint of the wire word width.
func SkipWordUnsigned(r io.Reader) (int, error) {
	_, n, err := ReadWordUnsigned(r)
	return n, err
}
