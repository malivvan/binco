// Package ptrmap writes pointer graphs to a wire without losing pointer identity.
//
// Every pointer gets an identifier; the pointed-to data is written the first time
// an identifier appears and later occurrences are just a reference. Shared
// references and cycles therefore survive the round trip. Share one pair of maps
// between all codecs of a value and wrap the outermost codec with Wrap.
package ptrmap
