package binco

import "unsafe"

// FixedWidth is implemented by serializers whose encoded form of one value is
// exactly its little-endian memory representation.
//
// The slice codecs use it to replace one interface call per element with a
// single copy when the element type has that width. Implement it only for a
// codec that really writes the raw value: a codec that zig-zags, prefixes or
// copies must not claim it.
type FixedWidth interface {
	// FixedWidth reports the encoded size of one value in bytes.
	FixedWidth() int
}

// HostLittleEndian reports whether the host stores multi-byte values least
// significant byte first.
//
// The bulk slice path copies memory directly, which is only the wire format on a
// little-endian host; elsewhere the codecs fall back to the element serializer.
var HostLittleEndian = func() bool {
	var probe uint16 = 1
	return *(*byte)(unsafe.Pointer(&probe)) == 1
}()
