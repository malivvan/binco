package wireio_test

import (
	"bytes"
	"errors"
	"io"
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
