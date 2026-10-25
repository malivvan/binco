// Package varnum writes variable-length integers to a wire.
//
// The encoding matches the buffer codecs of package varnum: unsigned LEB128,
// with ZigZag in front for signed values. The PositiveInt codecs skip the ZigZag
// step and are meant for lengths, type identifiers and other values that cannot
// be negative.
package varnum
