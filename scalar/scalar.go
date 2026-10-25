// Package scalar encodes numbers and timestamps in a fixed number of bytes.
//
// Every value occupies the width of its Go type, in little-endian byte order:
// eight bytes for 64-bit types, four for 32-bit, two for 16-bit and one for
// 8-bit types. Fixed-width encoding costs more space than varnum for small
// numbers but needs no bit manipulation, which makes it the faster choice when
// payload size is not the bottleneck. Time values are stored as Unix timestamps;
// the UTC codecs return the decoded value in UTC.
package scalar
