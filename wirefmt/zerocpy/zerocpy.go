// Package zerocpy writes scalars to a wire with single machine-word writes.
//
// The bytes are identical to those of wirefmt/scalar, but scalars are converted with
// one memory access instead of shifting bytes. Strings are still allocated while
// decoding, because a wire cannot expose bytes it has not read yet. The package
// only runs on little-endian hosts and panics during initialization otherwise.
package zerocpy
