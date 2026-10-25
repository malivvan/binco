package zerocopy

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/le"
)

// The fixed-width codecs delegate the byte-level work to internal/le, which
// package scalar shares. Converting through encoding/binary with an explicit
// little-endian order keeps the codecs correct on any host byte order and at an
// unaligned offset.

// wordSize is the encoded width of an int, and of a uint.
const wordSize = binco.IntWordBytes

// writeLE64 stores a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, bs []byte) int { return le.Put64(bs, t) }

// writeLE32 stores a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, bs []byte) int { return le.Put32(bs, t) }

// writeLE16 stores a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, bs []byte) int { return le.Put16(bs, t) }

// writeLE8 stores an 8-bit value.
func writeLE8[T binco.Integer8](t T, bs []byte) int { return le.Put8(bs, t) }

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](bs []byte) (T, int, error) { return le.Get64[T](bs) }

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](bs []byte) (T, int, error) { return le.Get32[T](bs) }

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](bs []byte) (T, int, error) { return le.Get16[T](bs) }

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](bs []byte) (T, int, error) { return le.Get8[T](bs) }

// writeWord stores an int of the wire word size.
func writeWord(v int, bs []byte) int { return le.PutWord(bs, v) }

// loadWord reads an int of the wire word size, reporting ErrOverflow when the
// host int is narrower than the 64-bit wire value.
func loadWord(bs []byte) (int, int, error) { return le.GetWord(bs) }

// skipWord advances over an int of the wire word size.
//
// It reads the value, so a host that cannot hold it reports the same
// ErrOverflow that Unmarshal does.
func skipWord(bs []byte) (int, error) { return le.SkipWord(bs) }

// writeWordUnsigned stores a uint of the wire word size.
func writeWordUnsigned(v uint, bs []byte) int { return le.PutWordUnsigned(bs, v) }

// loadWordUnsigned reads a uint of the wire word size, reporting ErrOverflow
// when the host uint is narrower than the 64-bit wire value.
func loadWordUnsigned(bs []byte) (uint, int, error) { return le.GetWordUnsigned(bs) }

// skipWordUnsigned advances over a uint of the wire word size.
func skipWordUnsigned(bs []byte) (int, error) { return le.SkipWordUnsigned(bs) }
