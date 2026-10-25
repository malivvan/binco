// Package composite encodes the types that need more than a fixed-width number:
// bools, strings, byte slices, pointers, slices and maps.
//
// Lengths are written with varnum.PositiveInt unless a serializer passes another
// length codec through options. Containers can be validated while decoding: the
// NewValid constructors build serializers that run their validators as soon as
// the length, key or element they describe has been decoded.
package composite
