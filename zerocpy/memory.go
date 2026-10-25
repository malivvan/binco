package zerocpy

import (
	"encoding/binary"

	"github.com/malivvan/binco"
)

// The helpers below read and write fixed-width values through
// encoding/binary rather than through a direct `*(*T)(unsafe.Pointer(...))`.
//
// A direct pointer dereference is only correct for an aligned address; a value
// that starts at an odd offset inside the input buffer is undefined behaviour
// on architectures that fault on unaligned access (some ARM, MIPS and RISC-V
// builds). The compiler turns binary.LittleEndian.Uint64 and PutUint64 into a
// single load or store on amd64 and arm64 anyway, so nothing is lost there,
// and every other architecture gets a correct result.

// writeLE64 stores a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, binco.Num64RawSize)
	binary.LittleEndian.PutUint64(bs, uint64(t))
	return binco.Num64RawSize
}

// writeLE32 stores a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, binco.Num32RawSize)
	binary.LittleEndian.PutUint32(bs, uint32(t))
	return binco.Num32RawSize
}

// writeLE16 stores a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, binco.Num16RawSize)
	binary.LittleEndian.PutUint16(bs, uint16(t))
	return binco.Num16RawSize
}

// writeLE8 stores an 8-bit value.
func writeLE8[T binco.Integer8](t T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, binco.Num8RawSize)
	bs[0] = byte(t)
	return binco.Num8RawSize
}

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num64RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return T(binary.LittleEndian.Uint64(bs)), binco.Num64RawSize, nil
}

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num32RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return T(binary.LittleEndian.Uint32(bs)), binco.Num32RawSize, nil
}

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num16RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return T(binary.LittleEndian.Uint16(bs)), binco.Num16RawSize, nil
}

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](bs []byte) (t T, n int, err error) {
	if len(bs) < binco.Num8RawSize {
		return t, 0, binco.ErrTooSmallByteSlice
	}
	return T(bs[0]), binco.Num8RawSize, nil
}

// wordSize is the encoded width of an int, and of a uint.
//
// It is a constant rather than a variable chosen in init: a package-level
// function variable cannot be inlined or bounds-check-eliminated, so every
// Int and Uint call used to pay for an indirect call.
const wordSize = binco.IntWordBytes

// writeWord stores an int of the wire word size.
func writeWord(v int, bs []byte) int {
	return writeLE64(int64(v), bs)
}

// loadWord reads an int of the wire word size, reporting ErrOverflow when the
// host int is narrower than the 64-bit wire value.
func loadWord(bs []byte) (int, int, error) {
	uv, n, err := loadLE64[int64](bs)
	if err != nil {
		return 0, n, err
	}
	v := int(uv)
	if int64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// skipWord advances over an int of the wire word size.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWord(bs []byte) (int, error) {
	_, n, err := loadWord(bs)
	return n, err
}

// writeWordUnsigned stores a uint of the wire word size.
func writeWordUnsigned(v uint, bs []byte) int {
	return writeLE64(uint64(v), bs)
}

// loadWordUnsigned reads a uint of the wire word size, reporting ErrOverflow
// when the host uint is narrower than the 64-bit wire value.
func loadWordUnsigned(bs []byte) (uint, int, error) {
	uv, n, err := loadLE64[uint64](bs)
	if err != nil {
		return 0, n, err
	}
	v := uint(uv)
	if uint64(v) != uv {
		return 0, n, binco.ErrOverflow
	}
	return v, n, nil
}

// skipWordUnsigned advances over a uint of the wire word size.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWordUnsigned(bs []byte) (int, error) {
	_, n, err := loadWordUnsigned(bs)
	return n, err
}
