package zerocpy

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/lew"
	"github.com/malivvan/binco/wirefmt"
)

// The fixed-width codecs delegate the byte-level work to internal/lew, which
// package wirefmt/scalar shares. Each value is assembled in a local buffer and
// written with one call, or read with one io.ReadFull.

// writeLE64 writes a 64-bit value in little-endian order.
func writeLE64[T binco.Integer64](t T, w wirefmt.Writer) (int, error) {
	return lew.Write64(w, t)
}

// writeLE32 writes a 32-bit value in little-endian order.
func writeLE32[T binco.Integer32](t T, w wirefmt.Writer) (int, error) {
	return lew.Write32(w, t)
}

// writeLE16 writes a 16-bit value in little-endian order.
func writeLE16[T binco.Integer16](t T, w wirefmt.Writer) (int, error) {
	return lew.Write16(w, t)
}

// writeLE8 writes an 8-bit value.
func writeLE8[T binco.Integer8](t T, w wirefmt.Writer) (int, error) {
	return lew.Write8(w, t)
}

// loadLE64 reads a 64-bit value from little-endian bytes.
func loadLE64[T binco.Integer64](r wirefmt.Reader) (T, int, error) {
	return lew.Read64[T](r)
}

// loadLE32 reads a 32-bit value from little-endian bytes.
func loadLE32[T binco.Integer32](r wirefmt.Reader) (T, int, error) {
	return lew.Read32[T](r)
}

// loadLE16 reads a 16-bit value from little-endian bytes.
func loadLE16[T binco.Integer16](r wirefmt.Reader) (T, int, error) {
	return lew.Read16[T](r)
}

// loadLE8 reads an 8-bit value.
func loadLE8[T binco.Integer8](r wirefmt.Reader) (T, int, error) {
	return lew.Read8[T](r)
}

// writeWord writes an int in the wire word width.
func writeWord(v int, w wirefmt.Writer) (int, error) { return lew.WriteWord(w, v) }

// loadWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func loadWord(r wirefmt.Reader) (int, int, error) { return lew.ReadWord(r) }

// wordSize reports the encoded width of an int.
func wordSize(v int) int { return binco.IntWordBytes }

// skipWord advances over an int of the wire word width.
func skipWord(r wirefmt.Reader) (int, error) { return lew.SkipWord(r) }

// writeWordUnsigned writes a uint in the wire word width.
func writeWordUnsigned(v uint, w wirefmt.Writer) (int, error) {
	return lew.WriteWordUnsigned(w, v)
}

// loadWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func loadWordUnsigned(r wirefmt.Reader) (uint, int, error) {
	return lew.ReadWordUnsigned(r)
}

// wordUnsignedSize reports the encoded width of a uint.
func wordUnsignedSize(v uint) int { return binco.IntWordBytes }

// skipWordUnsigned advances over a uint of the wire word width.
func skipWordUnsigned(r wirefmt.Reader) (int, error) { return lew.SkipWordUnsigned(r) }
