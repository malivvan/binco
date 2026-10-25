package lew_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/le"
	"github.com/malivvan/binco/internal/lew"
)

// TestWriteRead checks every width: the stream and the buffer implementations
// must produce the same bytes, and each value must read back.
func TestWriteRead(t *testing.T) {
	var buf bytes.Buffer
	if n, err := lew.Write8(&buf, uint8(0xff)); err != nil || n != 1 {
		t.Fatalf("Write8 = (%d, %v)", n, err)
	}
	if n, err := lew.Write16(&buf, uint16(0x0102)); err != nil || n != 2 {
		t.Fatalf("Write16 = (%d, %v)", n, err)
	}
	if n, err := lew.Write32(&buf, uint32(0x01020304)); err != nil || n != 4 {
		t.Fatalf("Write32 = (%d, %v)", n, err)
	}
	if n, err := lew.Write64(&buf, int64(-1234567890)); err != nil || n != 8 {
		t.Fatalf("Write64 = (%d, %v)", n, err)
	}

	want := make([]byte, 15)
	le.Put8(want, uint8(0xff))
	le.Put16(want[1:], uint16(0x0102))
	le.Put32(want[3:], uint32(0x01020304))
	le.Put64(want[7:], int64(-1234567890))
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("stream wrote %v, buffer wrote %v", buf.Bytes(), want)
	}

	r := bytes.NewReader(buf.Bytes())
	if got, n, err := lew.Read8[uint8](r); err != nil || n != 1 || got != 0xff {
		t.Fatalf("Read8 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := lew.Read16[uint16](r); err != nil || n != 2 || got != 0x0102 {
		t.Fatalf("Read16 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := lew.Read32[uint32](r); err != nil || n != 4 || got != 0x01020304 {
		t.Fatalf("Read32 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := lew.Read64[int64](r); err != nil || n != 8 || got != -1234567890 {
		t.Fatalf("Read64 = (%d, %d, %v)", got, n, err)
	}
}

// TestWord checks the host-width helpers on a stream.
func TestWord(t *testing.T) {
	var buf bytes.Buffer
	if n, err := lew.WriteWord(&buf, -300); err != nil || n != 8 {
		t.Fatalf("WriteWord = (%d, %v)", n, err)
	}
	word := append([]byte(nil), buf.Bytes()...)
	if got, n, err := lew.ReadWord(bytes.NewReader(word)); err != nil || n != 8 || got != -300 {
		t.Fatalf("ReadWord = (%d, %d, %v)", got, n, err)
	}
	if n, err := lew.SkipWord(bytes.NewReader(word)); err != nil || n != 8 {
		t.Fatalf("SkipWord = (%d, %v)", n, err)
	}

	buf.Reset()
	if n, err := lew.WriteWordUnsigned(&buf, 300); err != nil || n != 8 {
		t.Fatalf("WriteWordUnsigned = (%d, %v)", n, err)
	}
	uword := append([]byte(nil), buf.Bytes()...)
	if got, n, err := lew.ReadWordUnsigned(bytes.NewReader(uword)); err != nil || n != 8 || got != 300 {
		t.Fatalf("ReadWordUnsigned = (%d, %d, %v)", got, n, err)
	}
	if n, err := lew.SkipWordUnsigned(bytes.NewReader(uword)); err != nil || n != 8 {
		t.Fatalf("SkipWordUnsigned = (%d, %v)", n, err)
	}

	// The wire value 2^63 does not fit a 32-bit host; on this host the same
	// conversion is covered through le.WordInt, which ReadWord delegates to.
	big := make([]byte, 8)
	le.Put64(big, uint64(1)<<63)
	if strconv.IntSize == 32 {
		if _, _, err := lew.ReadWord(bytes.NewReader(big)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("ReadWord of 2^63 = %v, want ErrOverflow", err)
		}
		if _, err := lew.SkipWord(bytes.NewReader(big)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("SkipWord of 2^63 = %v, want ErrOverflow", err)
		}
		maxUint := make([]byte, 8)
		le.Put64(maxUint, uint64(math.MaxUint64))
		if _, _, err := lew.ReadWordUnsigned(bytes.NewReader(maxUint)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("ReadWordUnsigned of 2^64-1 = %v, want ErrOverflow", err)
		}
		if _, err := lew.SkipWordUnsigned(bytes.NewReader(maxUint)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("SkipWordUnsigned of 2^64-1 = %v, want ErrOverflow", err)
		}
	}
}

// TestShortInput checks the EOF rule for every reader and skipper: a value that
// started and was cut short is io.ErrUnexpectedEOF, while reading nothing at all
// is a clean io.EOF.
func TestShortInput(t *testing.T) {
	readers := []struct {
		name  string
		width int
		call  func(r io.Reader) (int, error)
	}{
		{"Read8", 1, func(r io.Reader) (int, error) { _, n, err := lew.Read8[uint8](r); return n, err }},
		{"Read16", 2, func(r io.Reader) (int, error) { _, n, err := lew.Read16[uint16](r); return n, err }},
		{"Read32", 4, func(r io.Reader) (int, error) { _, n, err := lew.Read32[uint32](r); return n, err }},
		{"Read64", 8, func(r io.Reader) (int, error) { _, n, err := lew.Read64[uint64](r); return n, err }},
		{"ReadWord", 8, func(r io.Reader) (int, error) { _, n, err := lew.ReadWord(r); return n, err }},
		{"ReadWordUnsigned", 8, func(r io.Reader) (int, error) { _, n, err := lew.ReadWordUnsigned(r); return n, err }},
	}
	for _, tc := range readers {
		t.Run(tc.name, func(t *testing.T) {
			if tc.width == 1 {
				// A one-byte value that is absent has read nothing at all.
				if _, err := tc.call(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
					t.Fatalf("%s of an empty stream = %v, want io.EOF", tc.name, err)
				}
				return
			}
			if _, err := tc.call(bytes.NewReader(make([]byte, tc.width-1))); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("%s of a truncated value = %v, want io.ErrUnexpectedEOF", tc.name, err)
			}
		})
	}

	skippers := []struct {
		name  string
		width int
		call  func(r io.Reader) (int, error)
	}{
		{"Skip8", 1, func(r io.Reader) (int, error) { return lew.Skip8(r) }},
		{"Skip16", 2, func(r io.Reader) (int, error) { return lew.Skip16(r) }},
		{"Skip32", 4, func(r io.Reader) (int, error) { return lew.Skip32(r) }},
		{"Skip64", 8, func(r io.Reader) (int, error) { return lew.Skip64(r) }},
		{"SkipWord", 8, func(r io.Reader) (int, error) { return lew.SkipWord(r) }},
		{"SkipWordUnsigned", 8, func(r io.Reader) (int, error) { return lew.SkipWordUnsigned(r) }},
	}
	for _, tc := range skippers {
		t.Run(tc.name, func(t *testing.T) {
			// Nothing read at all is a clean end of stream.
			if _, err := tc.call(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
				t.Fatalf("%s of an empty stream = %v, want io.EOF", tc.name, err)
			}
			// A partially present value is unexpected EOF. A one-byte value
			// has no partial read, so there is nothing to distinguish there.
			if tc.width == 1 {
				return
			}
			if _, err := tc.call(bytes.NewReader(make([]byte, tc.width-1))); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("%s of a truncated value = %v, want io.ErrUnexpectedEOF", tc.name, err)
			}
		})
	}
}

// TestSuccessfulSkips checks that every skipper consumes its width.
func TestSuccessfulSkips(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(r io.Reader) (int, error)
		want int
	}{
		{"Skip8", func(r io.Reader) (int, error) { return lew.Skip8(r) }, 1},
		{"Skip16", func(r io.Reader) (int, error) { return lew.Skip16(r) }, 2},
		{"Skip32", func(r io.Reader) (int, error) { return lew.Skip32(r) }, 4},
		{"Skip64", func(r io.Reader) (int, error) { return lew.Skip64(r) }, 8},
	} {
		if n, err := tc.call(bytes.NewReader(make([]byte, tc.want))); err != nil || n != tc.want {
			t.Fatalf("%s = (%d, %v), want (%d, nil)", tc.name, n, err, tc.want)
		}
	}
}

// errorWriter accepts nothing and returns a sentinel error.
type errorWriter struct{ err error }

func (w errorWriter) Write(p []byte) (int, error) { return 0, w.err }

// TestWriterErrors checks that a failing writer surfaces its own error.
func TestWriterErrors(t *testing.T) {
	want := errors.New("disk full")
	w := errorWriter{err: want}
	writers := []struct {
		name string
		call func() (int, error)
	}{
		{"Write8", func() (int, error) { return lew.Write8(w, uint8(1)) }},
		{"Write16", func() (int, error) { return lew.Write16(w, uint16(1)) }},
		{"Write32", func() (int, error) { return lew.Write32(w, uint32(1)) }},
		{"Write64", func() (int, error) { return lew.Write64(w, int64(1)) }},
		{"WriteWord", func() (int, error) { return lew.WriteWord(w, 1) }},
		{"WriteWordUnsigned", func() (int, error) { return lew.WriteWordUnsigned(w, 1) }},
	}
	for _, tc := range writers {
		if _, err := tc.call(); !errors.Is(err, want) {
			t.Fatalf("%s = %v, want %v", tc.name, err, want)
		}
	}
}
