// Package wireio reads length-prefixed payloads from a stream.
//
// Both streaming text packages need the same two operations, and both need the
// same three rules: the buffer grows with the bytes that actually arrive, so a
// declared length that the stream never delivers costs nothing, a truncation is
// io.ErrUnexpectedEOF rather than a clean io.EOF, and a reader that stops
// producing data fails with io.ErrNoProgress instead of spinning forever.
package wireio

import (
	"bytes"
	"errors"
	"io"
)

// maxEmptyReads is the number of consecutive (0, nil) reads a stream may make
// before it is considered stalled. It is the bound io.ReadAll uses.
const maxEmptyReads = 100

// errNegativeRead reports a reader that returned a negative count.
var errNegativeRead = errors.New("wireio: reader returned a negative count")

// CopyN copies exactly n bytes from src to dst.
//
// written is the number of bytes copied, which is smaller than n when the stream
// ends early (io.ErrUnexpectedEOF) or stalls (io.ErrNoProgress).
func CopyN(dst io.Writer, src io.Reader, n int64) (written int64, err error) {
	var buf [4096]byte
	empty := 0
	for written < n {
		if empty >= maxEmptyReads {
			return written, io.ErrNoProgress
		}
		want := int64(len(buf))
		if remaining := n - written; remaining < want {
			want = remaining
		}
		m, rerr := src.Read(buf[:want])
		if m < 0 {
			return written, errNegativeRead
		}
		if m > 0 {
			empty = 0
			w, werr := dst.Write(buf[:m])
			written += int64(w)
			if werr != nil {
				return written, werr
			}
			if w != m {
				return written, io.ErrShortWrite
			}
		} else if rerr == nil {
			empty++
		}
		if rerr != nil {
			if rerr == io.EOF {
				// The same distinction as io.ReadFull: nothing read is a clean
				// end of stream, a partial value is a truncation. Callers that
				// have already consumed part of the value (the blob decoders,
				// which read a length first) map io.EOF themselves.
				if written > 0 && written < n {
					return written, io.ErrUnexpectedEOF
				}
				if written >= n {
					return written, nil
				}
				return written, io.EOF
			}
			return written, rerr
		}
	}
	return written, nil
}

// ReadFull reads exactly len(buf) bytes, with the same stall guard as CopyN.
func ReadFull(r io.Reader, buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return 0, nil
	}
	var n1 int64
	n1, err = CopyN(&sliceWriter{buf: buf}, r, int64(len(buf)))
	return int(n1), err
}

// sliceWriter writes into a fixed buffer without allocating.
type sliceWriter struct {
	buf []byte
	n   int
}

func (w *sliceWriter) Write(p []byte) (int, error) {
	n := copy(w.buf[w.n:], p)
	w.n += n
	return n, nil
}

// ReadFixed reads exactly length bytes. It returns the bytes plus the number
// read, which is smaller than length on a truncated stream.
func ReadFixed(r io.Reader, length int) (body []byte, n int, err error) {
	if length == 0 {
		return []byte{}, 0, nil
	}
	var buf bytes.Buffer
	n1, err := CopyN(&buf, r, int64(length))
	if err != nil {
		if err == io.EOF {
			// The length was already read, so the stream ends inside the value.
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
	n1, err := CopyN(io.Discard, r, int64(length))
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return int(n1), err
}
