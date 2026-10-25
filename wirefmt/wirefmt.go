// Package wirefmt provides the streaming form of every codec in this module.
//
// A streaming codec reads from an io.Reader and writes to an io.Writer instead of
// a byte slice, so a value never has to be held in memory as a whole. The format
// is identical to the buffer one: bytes written by a buffer codec are read by the
// matching streaming codec and the other way around.
package wirefmt
