package scalar

import (
	"encoding/binary"
	"io"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// The helpers below assemble a fixed-width value in a local buffer and issue one
// Write, or read it with one io.ReadFull.
//
// Writing one byte per WriteByte call made eight interface calls out of every
// uint64; reading did the same. A single call per value is both faster and
// shorter, and the local buffer keeps the code free of unsafe pointer casts.

// writeLE64 writes a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, w wirefmt.Writer) (n int, err error) {
	var buf [binco.Num64RawSize]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(t))
	return w.Write(buf[:])
}

// writeLE32 writes a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, w wirefmt.Writer) (n int, err error) {
	var buf [binco.Num32RawSize]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(t))
	return w.Write(buf[:])
}

// writeLE16 writes a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, w wirefmt.Writer) (n int, err error) {
	var buf [binco.Num16RawSize]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(t))
	return w.Write(buf[:])
}

// writeLE8 writes an 8-bit value.
func writeLE8[T binco.Integer8](t T, w wirefmt.Writer) (n int, err error) {
	var buf [binco.Num8RawSize]byte
	buf[0] = byte(t)
	return w.Write(buf[:])
}

// readLE64 reads a 64-bit value from little-endian bytes.
func readLE64[T binco.Integer64](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num64RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint64(buf[:])), n, nil
}

// readLE32 reads a 32-bit value from little-endian bytes.
func readLE32[T binco.Integer32](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num32RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint32(buf[:])), n, nil
}

// readLE16 reads a 16-bit value from little-endian bytes.
func readLE16[T binco.Integer16](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num16RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint16(buf[:])), n, nil
}

// readLE8 reads an 8-bit value.
func readLE8[T binco.Integer8](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num8RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(buf[0]), n, nil
}

// skipLE64 advances over eight bytes.
func skipLE64(r wirefmt.Reader) (int, error) {
	var buf [binco.Num64RawSize]byte
	return io.ReadFull(r, buf[:])
}

// skipLE32 advances over four bytes.
func skipLE32(r wirefmt.Reader) (int, error) {
	var buf [binco.Num32RawSize]byte
	return io.ReadFull(r, buf[:])
}

// skipLE16 advances over two bytes.
func skipLE16(r wirefmt.Reader) (int, error) {
	var buf [binco.Num16RawSize]byte
	return io.ReadFull(r, buf[:])
}

// skipLE8 advances over one byte.
func skipLE8(r wirefmt.Reader) (int, error) {
	var buf [binco.Num8RawSize]byte
	return io.ReadFull(r, buf[:])
}

// widthOf64 is the encoded width of a 64-bit value.
func widthOf64[T binco.Num64](t T) int {
	return binco.Num64RawSize
}

// widthOf32 is the encoded width of a 32-bit value.
func widthOf32[T binco.Num32](t T) int {
	return binco.Num32RawSize
}

// writeWord writes an int in the wire word width.
//
// The wire width is always 64 bits, whatever the host word size is, so a
// payload written on amd64 reads back on a 32-bit build. These used to be
// package-level function variables chosen in init, which stopped the compiler
// from inlining every Int and Uint call.
func writeWord(v int, w wirefmt.Writer) (int, error) {
	return writeLE64(int64(v), w)
}

// readWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func readWord(r wirefmt.Reader) (int, int, error) {
	uv, n, err := readLE64[int64](r)
	if err != nil {
		return 0, n, err
	}
	v := int(uv)
	if int64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// wordSize reports the encoded width of an int.
func wordSize(v int) int {
	return binco.IntWordBytes
}

// skipWord advances over an int of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWord(r wirefmt.Reader) (int, error) {
	_, n, err := readWord(r)
	return n, err
}

// writeWordUnsigned writes a uint in the wire word width.
func writeWordUnsigned(v uint, w wirefmt.Writer) (int, error) {
	return writeLE64(uint64(v), w)
}

// readWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func readWordUnsigned(r wirefmt.Reader) (uint, int, error) {
	uv, n, err := readLE64[uint64](r)
	if err != nil {
		return 0, n, err
	}
	v := uint(uv)
	if uint64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// wordUnsignedSize reports the encoded width of a uint.
func wordUnsignedSize(v uint) int {
	return binco.IntWordBytes
}

// skipWordUnsigned advances over a uint of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWordUnsigned(r wirefmt.Reader) (int, error) {
	_, n, err := readWordUnsigned(r)
	return n, err
}
