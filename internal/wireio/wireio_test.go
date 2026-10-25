package wireio_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/malivvan/binco/internal/wireio"
)

// TestReadFixed checks a complete read and the truncation rule: a value that has
// started but not finished is io.ErrUnexpectedEOF.
func TestReadFixed(t *testing.T) {
	body, n, err := wireio.ReadFixed(bytes.NewReader([]byte("hello")), 5)
	if err != nil || n != 5 || string(body) != "hello" {
		t.Fatalf("ReadFixed = (%q, %d, %v)", body, n, err)
	}

	if body, n, err = wireio.ReadFixed(bytes.NewReader([]byte("hi")), 5); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated ReadFixed = (%q, %d, %v)", body, n, err)
	} else if n != 2 {
		t.Fatalf("truncated ReadFixed read %d bytes, want 2", n)
	}

	if body, n, err = wireio.ReadFixed(bytes.NewReader(nil), 0); err != nil || n != 0 || len(body) != 0 {
		t.Fatalf("ReadFixed(0) = (%v, %d, %v)", body, n, err)
	}
}

// TestDiscardFixed checks that skipping consumes the same bytes as reading.
func TestDiscardFixed(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))
	n, err := wireio.DiscardFixed(r, 5)
	if err != nil || n != 5 {
		t.Fatalf("DiscardFixed = (%d, %v)", n, err)
	}
	rest, _ := io.ReadAll(r)
	if string(rest) != " world" {
		t.Fatalf("remaining = %q", rest)
	}
	if n, err = wireio.DiscardFixed(bytes.NewReader([]byte("hi")), 5); !errors.Is(err, io.ErrUnexpectedEOF) || n != 2 {
		t.Fatalf("truncated DiscardFixed = (%d, %v)", n, err)
	}
	if n, err = wireio.DiscardFixed(bytes.NewReader(nil), 0); err != nil || n != 0 {
		t.Fatalf("DiscardFixed(0) = (%d, %v)", n, err)
	}
}

// stallReader returns (0, nil) forever: it never makes progress and never fails.
type stallReader struct{ reads int }

func (r *stallReader) Read([]byte) (int, error) {
	r.reads++
	return 0, nil
}

// negativeReader reports a negative count, which the io.Reader contract forbids.
type negativeReader struct{}

func (negativeReader) Read([]byte) (int, error) { return -1, nil }

// errBoom is the reader failure the tests propagate.
var errBoom = errors.New("boom")

// boomReader returns its data and then errBoom instead of io.EOF.
type boomReader struct{ data []byte }

func (r *boomReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, errBoom
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

// shortWriter accepts one byte per call, whatever it is given.
type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return 1, nil
}

// TestCopyN covers every outcome of the bounded copy.
func TestCopyN(t *testing.T) {
	t.Run("copies every byte", func(t *testing.T) {
		var dst bytes.Buffer
		n, err := wireio.CopyN(&dst, bytes.NewReader([]byte("hello")), 5)
		if err != nil || n != 5 || dst.String() != "hello" {
			t.Fatalf("CopyN = (%d, %v) %q", n, err, dst.String())
		}
	})
	t.Run("truncation is unexpected EOF", func(t *testing.T) {
		n, err := wireio.CopyN(io.Discard, bytes.NewReader([]byte("hi")), 5)
		if !errors.Is(err, io.ErrUnexpectedEOF) || n != 2 {
			t.Fatalf("CopyN = (%d, %v), want (2, ErrUnexpectedEOF)", n, err)
		}
	})
	t.Run("nothing read is a clean EOF", func(t *testing.T) {
		n, err := wireio.CopyN(io.Discard, bytes.NewReader(nil), 4)
		if !errors.Is(err, io.EOF) || n != 0 {
			t.Fatalf("CopyN = (%d, %v), want (0, io.EOF)", n, err)
		}
	})
	t.Run("stalled reader is ErrNoProgress", func(t *testing.T) {
		r := &stallReader{}
		n, err := wireio.CopyN(io.Discard, r, 4)
		if !errors.Is(err, io.ErrNoProgress) || n != 0 {
			t.Fatalf("CopyN = (%d, %v), want (0, io.ErrNoProgress)", n, err)
		}
		if r.reads == 0 || r.reads > 1000 {
			t.Fatalf("CopyN made %d reads, want a bounded positive count", r.reads)
		}
	})
	t.Run("negative read is rejected", func(t *testing.T) {
		n, err := wireio.CopyN(io.Discard, negativeReader{}, 4)
		if err == nil || n != 0 || !strings.Contains(err.Error(), "negative") {
			t.Fatalf("CopyN = (%d, %v), want a negative-count error", n, err)
		}
	})
	t.Run("short write is reported", func(t *testing.T) {
		n, err := wireio.CopyN(shortWriter{}, bytes.NewReader([]byte("hello")), 5)
		if !errors.Is(err, io.ErrShortWrite) || n != 1 {
			t.Fatalf("CopyN = (%d, %v), want (1, io.ErrShortWrite)", n, err)
		}
	})
	t.Run("reader errors are propagated", func(t *testing.T) {
		n, err := wireio.CopyN(io.Discard, &boomReader{data: []byte("he")}, 4)
		if !errors.Is(err, errBoom) || n != 2 {
			t.Fatalf("CopyN = (%d, %v), want (2, boom)", n, err)
		}
	})
}

// TestReadFull covers the single-value read used by the byte adapters.
func TestReadFull(t *testing.T) {
	if n, err := wireio.ReadFull(bytes.NewReader(nil), nil); err != nil || n != 0 {
		t.Fatalf("ReadFull(nil) = (%d, %v)", n, err)
	}

	buf := make([]byte, 4)
	if n, err := wireio.ReadFull(bytes.NewReader([]byte{1, 2, 3, 4}), buf); err != nil || n != 4 {
		t.Fatalf("ReadFull = (%d, %v)", n, err)
	}
	if n, err := wireio.ReadFull(bytes.NewReader([]byte{1}), buf); !errors.Is(err, io.ErrUnexpectedEOF) || n != 1 {
		t.Fatalf("partial ReadFull = (%d, %v)", n, err)
	}
	if n, err := wireio.ReadFull(bytes.NewReader(nil), buf); !errors.Is(err, io.EOF) || n != 0 {
		t.Fatalf("empty ReadFull = (%d, %v)", n, err)
	}
	if _, err := wireio.ReadFull(&stallReader{}, buf); !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("stalled ReadFull = %v, want ErrNoProgress", err)
	}
}

// TestReadFixedStallAndEnd covers the two failures the codec-facing wrapper maps.
func TestReadFixedStallAndEnd(t *testing.T) {
	if _, n, err := wireio.ReadFixed(&stallReader{}, 3); !errors.Is(err, io.ErrNoProgress) || n != 0 {
		t.Fatalf("stalled ReadFixed = (%d, %v)", n, err)
	}
	if _, _, err := wireio.ReadFixed(bytes.NewReader(nil), 3); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("empty ReadFixed = %v, want ErrUnexpectedEOF", err)
	}
}

// TestDiscardFixedStallAndEnd covers the same two failures for skipping.
func TestDiscardFixedStallAndEnd(t *testing.T) {
	if n, err := wireio.DiscardFixed(&stallReader{}, 3); !errors.Is(err, io.ErrNoProgress) || n != 0 {
		t.Fatalf("stalled DiscardFixed = (%d, %v)", n, err)
	}
	if _, err := wireio.DiscardFixed(bytes.NewReader(nil), 3); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("empty DiscardFixed = %v, want ErrUnexpectedEOF", err)
	}
}

// failingWriter fails every write.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errBoom }

// eofReader returns all of its data together with io.EOF in one call, which a
// file or a buffer at the end of its content may do.
type eofReader struct{ data []byte }

func (r *eofReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, io.EOF
}

// TestCopyNWriterAndEOFEdgeCases covers the two remaining outcomes: a write that
// fails, and a read that delivers the last bytes together with io.EOF.
func TestCopyNWriterAndEOFEdgeCases(t *testing.T) {
	n, err := wireio.CopyN(failingWriter{}, bytes.NewReader([]byte("hello")), 5)
	if !errors.Is(err, errBoom) || n != 0 {
		t.Fatalf("CopyN with a failing writer = (%d, %v), want (0, boom)", n, err)
	}

	var dst bytes.Buffer
	n, err = wireio.CopyN(&dst, &eofReader{data: []byte("hello")}, 5)
	if err != nil || n != 5 || dst.String() != "hello" {
		t.Fatalf("CopyN with data and EOF = (%d, %v) %q", n, err, dst.String())
	}
}
