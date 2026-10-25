package zerocpy

// This package used to refuse to run on a big-endian host, because it read and
// wrote scalars by reinterpreting the bytes of the value in place. The codecs
// now convert through encoding/binary with an explicit little-endian order, so
// the byte order of the host no longer matters and no such check is needed.
//
// The zero-copy property is unaffected: strings and byte slices still alias the
// input buffer, and arrays are still decoded by reinterpreting their memory.
