// Package composite encodes the types that need more than a fixed-width number:
// bools, strings, byte slices, pointers, slices and maps.
//
// Lengths are written with varnum.PositiveInt unless a serializer passes another
// length codec through options. Containers can be validated while decoding: a
// validator passed through options is run as soon as the length, key or element
// it describes has been decoded.
//
// Decoding is bounded by the input: a container grows as its elements are read
// instead of being allocated from the declared length, and a length whose
// elements consume no bytes at all is capped by binco.MaxZeroSizeElements.
// Slices of fixed-width scalars take a bulk copy path when the element codec
// reports binco.FixedWidth.
package composite
