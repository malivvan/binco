// Package ptrmap encodes values whose pointers matter.
//
// A plain pointer codec in package composite writes every pointer by value:
// shared references are duplicated and cycles cannot be represented at all. The
// codecs here give every pointer an identifier, write the pointed-to data only
// the first time an identifier appears, and resolve the identifiers back to the
// same pointers while decoding. Pass the same maps to every codec of one value
// and wrap the outermost codec with Wrap, which clears the maps after each call.
package ptrmap
