// Package binco is a binary serialization toolkit for schema-driven formats.
//
// A codec turns a value into bytes and back. Serializers write into a buffer the
// caller owns, so a steady-state program can reuse one buffer and allocate
// nothing while encoding. The same format is available over byte streams in the
// subpackage wirefmt, is generated for user types by subpackage codegen, and supports
// shared pointers, type identifiers and validation through the ptrmap, typeid
// and composite packages.
//
// This package holds the pieces every codec shares: the Serializer and
// Marshaller interfaces, the Encode and Decode helpers, the error values, the
// type identifier, the validator interface, the pointer maps and the numeric
// constraints.
package binco
