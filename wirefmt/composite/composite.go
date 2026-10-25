// Package composite writes bools, strings, byte slices, pointers, slices and maps
// to a wire.
//
// Lengths use varnum.PositiveInt unless another length codec is passed through
// options. The NewValid constructors validate data while it is decoded, which is
// how a reader protects itself from a length that does not match the payload.
package composite
