package zerocpy

import (
	"encoding/binary"
	"io"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// The helpers below assemble fixed-width values in a local buffer and issue one
// Write, instead of casting the value with unsafe and issuing one WriteByte per
// byte. Besides being portable across byte orders, that turns eight virtual
// calls per uint64 into one.

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

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num64RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint64(buf[:])), n, nil
}

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num32RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint32(buf[:])), n, nil
}

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num16RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(binary.LittleEndian.Uint16(buf[:])), n, nil
}

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](r wirefmt.Reader) (t T, n int, err error) {
	var buf [binco.Num8RawSize]byte
	n, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}
	return T(buf[0]), n, nil
}

// wordSize is the encoded width of an int, and of a uint.
//
// It is a constant, so the compiler can inline and constant-fold the Size
// methods; the earlier package-level function variables prevented that.
func wordSize(v int) int {
	return binco.IntWordBytes
}

// writeWord writes an int in the wire word width.
func writeWord(v int, w wirefmt.Writer) (int, error) {
	return writeLE64(int64(v), w)
}

// loadWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func loadWord(r wirefmt.Reader) (int, int, error) {
	uv, n, err := loadLE64[int64](r)
	if err != nil {
		return 0, n, err
	}
	v := int(uv)
	if int64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// skipWord advances over an int of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWord(r wirefmt.Reader) (int, error) {
	_, n, err := loadWord(r)
	return n, err
}

// writeWordUnsigned writes a uint in the wire word width.
func writeWordUnsigned(v uint, w wirefmt.Writer) (int, error) {
	return writeLE64(uint64(v), w)
}

// loadWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func loadWordUnsigned(r wirefmt.Reader) (uint, int, error) {
	uv, n, err := loadLE64[uint64](r)
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
	_, n, err := loadWordUnsigned(r)
	return n, err
}

// skipLE64 advances over eight bytes.
func skipLE64(r wirefmt.Reader) (int, error) {
	var buf [binco.Num64RawSize]byte
	return io.ReadFull(r, buf[:])
}
