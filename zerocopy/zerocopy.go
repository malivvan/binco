// Package zerocopy encodes and decodes through direct memory access.
//
// The codecs here produce the same bytes as scalar and composite, but they hand
// out strings and byte slices that point into the input buffer instead of
// copying it. That makes them the fastest option in this module, at one cost: a
// decoded string or slice is only valid while its buffer is, and must not be
// modified. Reusing the buffer (for example through a sync.Pool of read
// buffers) silently mutates values that were already decoded, and any small
// string keeps the whole buffer alive.
//
// Fixed-width scalars are converted with encoding/binary, so the package works
// on any byte order and on architectures that fault on unaligned access. Only
// the array codec reinterprets memory, and it validates at construction that the
// type parameters really describe an array of the element type.
package zerocopy
