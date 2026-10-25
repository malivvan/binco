// Package scalar writes fixed-width numbers and timestamps to a wire.
//
// Values are written in little-endian byte order, using the width of their Go
// type. The encoding matches the buffer codecs of package scalar, so the two can
// be mixed freely. Time values are Unix timestamps; the UTC codecs return the
// decoded value in UTC.
package scalar
