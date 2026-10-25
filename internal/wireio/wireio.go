// Package wireio reads length-prefixed payloads from a stream.
//
// Both streaming text packages need the same two operations, and both need the
// same rule: the buffer grows with the bytes that actually arrive, so a
// declared length that the stream never delivers costs nothing, and a
// truncation is io.ErrUnexpectedEOF rather than a clean io.EOF.
package wireio

import (
	"bytes"
	"io"
)

// ReadFixed reads exactly length bytes. It returns the bytes plus the number
// read, which is smaller than length on a truncated stream.
func ReadFixed(r io.Reader, length int) (body []byte, n int, err error) {
	if length == 0 {
		return []byte{}, 0, nil
	}
	var buf bytes.Buffer
	var n1 int64
	n1, err = io.CopyN(&buf, r, int64(length))
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, int(n1), err
	}
	return buf.Bytes(), int(n1), nil
}

// DiscardFixed skips exactly length bytes.
func DiscardFixed(r io.Reader, length int) (n int, err error) {
	if length == 0 {
		return 0, nil
	}
	var n1 int64
	n1, err = io.CopyN(io.Discard, r, int64(length))
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return int(n1), err
}
