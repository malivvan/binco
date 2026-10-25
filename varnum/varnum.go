// Package varnum encodes integers in a compact, variable-length form.
//
// Unsigned values are written seven bits per byte, least significant group
// first, with the high bit set on every byte but the last. Signed values are
// ZigZag-mapped first, so that -1, 1, -2 and 2 all fit in a single byte. The
// PositiveInt codecs skip that mapping and are meant for lengths, type
// identifiers and other values that cannot be negative. Floats are encoded as
// the unsigned varint of their IEEE 754 bit pattern.
package varnum
