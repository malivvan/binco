// Package zerocpy encodes and decodes through direct memory access.
//
// The codecs here produce the same bytes as fixed and composite, but they read
// and write scalars with a single machine word access, and they hand out strings
// and byte slices that point into the input buffer instead of copying it. That
// makes them the fastest option in this module, at two costs: a decoded string
// or slice is only valid while its buffer is, and must not be modified, and the
// package only works on little-endian hosts (it panics during initialization
// otherwise, so a non-conforming payload can never be produced silently).
//
// Arrays are serialized here as well, because decoding into an array means
// reinterpreting the decoded elements as the array type.
package zerocpy
