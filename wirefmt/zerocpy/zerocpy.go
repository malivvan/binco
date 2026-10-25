// Package zerocpy provides the streaming counterparts of the zero-copy codecs.
//
// The bytes are identical to those of wirefmt/scalar, but fixed-width scalars
// are converted through a small local buffer and written with one call instead
// of one call per byte. Strings are allocated while decoding, because a wire
// cannot expose bytes it has not read yet, so this package offers no aliasing
// advantage over wirefmt/composite; it exists so that generated code can name
// the same package in both modes.
package zerocpy
