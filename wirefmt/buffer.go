package wirefmt

import "io"

// SliceWriter adapts a byte slice to Writer.
//
// It is what makes it possible to use a streaming codec where a buffer codec is
// expected: the codec writes into Buf, and N reports how much it wrote. A write
// past the end returns io.ErrShortBuffer instead of panicking.
type SliceWriter struct {
	Buf []byte
	N   int
}

// NewSliceWriter returns a SliceWriter over bs.
func NewSliceWriter(bs []byte) *SliceWriter {
	return &SliceWriter{Buf: bs}
}

// WriteByte appends one byte.
func (w *SliceWriter) WriteByte(c byte) error {
	if w.N < 0 || w.N >= len(w.Buf) {
		return io.ErrShortBuffer
	}
	w.Buf[w.N] = c
	w.N++
	return nil
}

// Write appends p.
func (w *SliceWriter) Write(p []byte) (n int, err error) {
	if w.N < 0 || len(p) > len(w.Buf)-w.N {
		return 0, io.ErrShortBuffer
	}
	copy(w.Buf[w.N:], p)
	w.N += len(p)
	return len(p), nil
}

// WriteString appends s.
func (w *SliceWriter) WriteString(s string) (n int, err error) {
	if w.N < 0 || len(s) > len(w.Buf)-w.N {
		return 0, io.ErrShortBuffer
	}
	copy(w.Buf[w.N:], s)
	w.N += len(s)
	return len(s), nil
}

// Bytes returns everything written so far.
func (w *SliceWriter) Bytes() []byte {
	return w.Buf[:w.N]
}
