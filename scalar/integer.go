package scalar

import (
	"github.com/malivvan/binco"
)

// putLE64 writes t as eight little-endian bytes.
func putLE64[T binco.Integer64](t T, bs []byte) int {
	binco.CheckWrite(bs, 0, binco.Num64RawSize)
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
	binco.CheckWrite(bs, 0, binco.Num32RawSize)
	bs[0] = byte(t)
	bs[1] = byte(t >> 8)
	bs[2] = byte(t >> 16)
	bs[3] = byte(t >> 24)
	return binco.Num32RawSize
}

// putLE16 writes t as two little-endian bytes.
func putLE16[T binco.Integer16](t T, bs []byte) int {
	binco.CheckWrite(bs, 0, binco.Num16RawSize)
	bs[0] = byte(t)
	bs[1] = byte(t >> 8)
	return binco.Num16RawSize
}

// putLE8 writes t as a single byte.
func putLE8[T binco.Integer8](t T, bs []byte) int {
	binco.CheckWrite(bs, 0, binco.Num8RawSize)
	bs[0] = byte(t)
	return binco.Num8RawSize
}

// readLE64 reads eight little-endian bytes.
func readLE64[T binco.Integer64](bs []byte) (T, int, error) {
	var t T
	if len(bs) < binco.Num64RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
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

// wordSize is the encoded width of an int, and of a uint.
//
// The width is a constant that is resolved at compile time. It used to be a
// package variable chosen in init, which forced every Int and Uint call through
// an indirect call the compiler could not inline or bounds-check-eliminate.
const wordSize = binco.IntWordBytes

// putWord writes an int in the wire word width.
func putWord(v int, bs []byte) int {
	return putLE64(int64(v), bs)
}

// readWord reads an int of the wire word width and reports ErrOverflow when the
// host int is narrower than the value, as it is on 32-bit builds.
func readWord(bs []byte) (int, int, error) {
	uv, n, err := readLE64[int64](bs)
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
func skipWord(bs []byte) (int, error) {
	_, n, err := readWord(bs)
	return n, err
}

// putWordUnsigned writes a uint in the wire word width.
func putWordUnsigned(v uint, bs []byte) int {
	return putLE64(uint64(v), bs)
}

// readWordUnsigned reads a uint of the wire word width and reports ErrOverflow
// when the host uint is narrower than the value.
func readWordUnsigned(bs []byte) (uint, int, error) {
	uv, n, err := readLE64[uint64](bs)
	if err != nil {
		return 0, n, err
	}
	v := uint(uv)
	if uint64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// skipWordUnsigned advances over a uint of the wire word width.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWordUnsigned(bs []byte) (int, error) {
	_, n, err := readWordUnsigned(bs)
	return n, err
}
