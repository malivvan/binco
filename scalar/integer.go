package scalar

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/le"
)

// The fixed-width codecs delegate the byte-level work to internal/le, which
// package zerocpy shares. The encoding is little-endian on every host.

// putLE64 writes t as eight little-endian bytes.
func putLE64[T binco.Integer64](t T, bs []byte) int { return le.Put64(bs, t) }

// putLE32 writes t as four little-endian bytes.
func putLE32[T binco.Integer32](t T, bs []byte) int { return le.Put32(bs, t) }

// putLE16 writes t as two little-endian bytes.
func putLE16[T binco.Integer16](t T, bs []byte) int { return le.Put16(bs, t) }

// putLE8 writes t as a single byte.
func putLE8[T binco.Integer8](t T, bs []byte) int { return le.Put8(bs, t) }

// readLE64 reads eight little-endian bytes.
func readLE64[T binco.Integer64](bs []byte) (T, int, error) { return le.Get64[T](bs) }

// readLE32 reads four little-endian bytes.
func readLE32[T binco.Integer32](bs []byte) (T, int, error) { return le.Get32[T](bs) }

// readLE16 reads two little-endian bytes.
func readLE16[T binco.Integer16](bs []byte) (T, int, error) { return le.Get16[T](bs) }

// readLE8 reads a single byte.
func readLE8[T binco.Integer8](bs []byte) (T, int, error) { return le.Get8[T](bs) }

// SkipInteger64 advances over an eight-byte integer.
func SkipInteger64(bs []byte) (int, error) { return le.Skip64(bs) }

// SkipInteger32 advances over a four-byte integer.
func SkipInteger32(bs []byte) (int, error) { return le.Skip32(bs) }

// SkipInteger16 advances over a two-byte integer.
func SkipInteger16(bs []byte) (int, error) { return le.Skip16(bs) }

// SkipInteger8 advances over a one-byte integer.
func SkipInteger8(bs []byte) (int, error) { return le.Skip8(bs) }

// wordSize is the encoded width of an int, and of a uint.
//
// The width is a constant that is resolved at compile time. It used to be a
// package variable chosen in init, which forced every Int and Uint call through
// an indirect call the compiler could not inline or bounds-check-eliminate.
const wordSize = binco.IntWordBytes

// putWord writes an int in the wire word width.
func putWord(v int, bs []byte) int { return le.PutWord(bs, v) }

// readWord reads an int of the wire word width and reports ErrOverflow when the
// host int is narrower than the value, as it is on 32-bit builds.
func readWord(bs []byte) (int, int, error) { return le.GetWord(bs) }

// skipWord advances over an int of the wire word width.
func skipWord(bs []byte) (int, error) { return le.SkipWord(bs) }

// putWordUnsigned writes a uint in the wire word width.
func putWordUnsigned(v uint, bs []byte) int { return le.PutWordUnsigned(bs, v) }

// readWordUnsigned reads a uint of the wire word width and reports ErrOverflow
// when the host uint is narrower than the value.
func readWordUnsigned(bs []byte) (uint, int, error) { return le.GetWordUnsigned(bs) }

// skipWordUnsigned advances over a uint of the wire word width.
func skipWordUnsigned(bs []byte) (int, error) { return le.SkipWordUnsigned(bs) }
