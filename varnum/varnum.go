// Package varnum encodes integers in a compact, variable-length form.
//
// Unsigned values are written seven bits per byte, least significant group
// first, with the high bit set on every byte but the last. Signed values are
// ZigZag-mapped first, so that -1, 1, -2 and 2 all fit in a single byte. The
// PositiveInt codecs skip that mapping and are meant for lengths, type
// identifiers and other values that cannot be negative; a decoded value above
// the type's maximum is reported as binco.ErrOverflow instead of wrapping, and
// marshalling a negative value panics with the same error. Floats are encoded as
// the unsigned varint of their IEEE 754 bit pattern.
//
// Encodings are canonical: a multi-byte value whose final byte is zero is
// rejected with binco.ErrNonCanonical, so a value has exactly one byte
// representation. Int and Uint use a 64-bit wire width on every architecture, so
// payloads are portable; a value the host cannot hold is ErrOverflow.
package varnum
