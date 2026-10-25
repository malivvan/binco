package scalar

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/lew"
	"github.com/malivvan/binco/wirefmt"
)

// The fixed-width codecs delegate the byte-level work to internal/lew, which
// package wirefmt/zerocpy shares. Each value is assembled in a local buffer and
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

// readLE64 reads a 64-bit value from little-endian bytes.
func readLE64[T binco.Integer64](r wirefmt.Reader) (T, int, error) {
	return lew.Read64[T](r)
}

// readLE32 reads a 32-bit value from little-endian bytes.
func readLE32[T binco.Integer32](r wirefmt.Reader) (T, int, error) {
	return lew.Read32[T](r)
}

// readLE16 reads a 16-bit value from little-endian bytes.
func readLE16[T binco.Integer16](r wirefmt.Reader) (T, int, error) {
	return lew.Read16[T](r)
}

// readLE8 reads an 8-bit value.
func readLE8[T binco.Integer8](r wirefmt.Reader) (T, int, error) {
	return lew.Read8[T](r)
}

// skipLE64 advances over eight bytes.
func skipLE64(r wirefmt.Reader) (int, error) { return lew.Skip64(r) }

// skipLE32 advances over four bytes.
func skipLE32(r wirefmt.Reader) (int, error) { return lew.Skip32(r) }

// skipLE16 advances over two bytes.
func skipLE16(r wirefmt.Reader) (int, error) { return lew.Skip16(r) }

// skipLE8 advances over one byte.
func skipLE8(r wirefmt.Reader) (int, error) { return lew.Skip8(r) }

// widthOf64 is the encoded width of a 64-bit value.
func widthOf64[T binco.Num64](t T) int {
	return binco.Num64RawSize
}

// widthOf32 is the encoded width of a 32-bit value.
func widthOf32[T binco.Num32](t T) int {
	return binco.Num32RawSize
}

// writeWord writes an int in the wire word width.
func writeWord(v int, w wirefmt.Writer) (int, error) { return lew.WriteWord(w, v) }

// readWord reads an int of the wire word width, reporting ErrOverflow when the
// host int is narrower than the value.
func readWord(r wirefmt.Reader) (int, int, error) { return lew.ReadWord(r) }

// wordSize reports the encoded width of an int.
func wordSize(v int) int { return binco.IntWordBytes }

// skipWord advances over an int of the wire word width.
func skipWord(r wirefmt.Reader) (int, error) { return lew.SkipWord(r) }

// writeWordUnsigned writes a uint in the wire word width.
func writeWordUnsigned(v uint, w wirefmt.Writer) (int, error) {
	return lew.WriteWordUnsigned(w, v)
}

// readWordUnsigned reads a uint of the wire word width, reporting ErrOverflow
// when the host uint is narrower than the value.
func readWordUnsigned(r wirefmt.Reader) (uint, int, error) {
	return lew.ReadWordUnsigned(r)
}

// wordUnsignedSize reports the encoded width of a uint.
func wordUnsignedSize(v uint) int { return binco.IntWordBytes }

// skipWordUnsigned advances over a uint of the wire word width.
func skipWordUnsigned(r wirefmt.Reader) (int, error) { return lew.SkipWordUnsigned(r) }
