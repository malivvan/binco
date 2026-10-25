package zerocpy

// This package used to refuse to run on a big-endian host, because it converted
// scalars by reinterpreting the value's bytes in place. The conversions now go
// through encoding/binary with an explicit little-endian order, so the host byte
// order no longer matters.
//
// Strings and byte slices are still allocated while decoding: a stream cannot
// expose bytes it has not read yet.
