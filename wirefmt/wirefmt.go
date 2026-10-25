// Package wirefmt provides the streaming form of every codec in this module.
//
// A streaming codec reads from an io.Reader and writes to an io.Writer instead of
// a byte slice, so a value never has to be held in memory as a whole. The format
// is identical to the buffer one: bytes written by a buffer codec are read by the
// matching streaming codec and the other way around. TestGoldenVectorsCrossThe
// Implementations pins that promise for every container.
//
// The streaming codecs are a second implementation of the same format, not a
// wrapper around the buffer ones: the buffer codecs need a length that is known
// in advance, while a stream must work with what has arrived. The price is that
// a change to the format has to be made twice. TestGoldenVectorsCrossThe
// Implementations is the guard against the two sides drifting apart: it encodes
// with one side and decodes with the other, in both directions, and compares the
// bytes.
//
// Two rules of the streaming API are worth remembering:
//
//   - Decode does not buffer a reader it cannot read single bytes from, because
//     it cannot know whether the caller needs the bytes it would read ahead. Use
//     NewDecoder when the codec owns the stream; it buffers once and reuses the
//     buffer, and does not swallow the next message.
//   - A value that is truncated in the middle returns io.ErrUnexpectedEOF, while
//     an empty stream returns io.EOF, so a clean end and a corrupt message are
//     distinguishable.
package wirefmt
