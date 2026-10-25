package wirefmt_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
)

// TestSliceWriter checks the slice-backed Writer: successful writes, the short
// buffer rule, and that Bytes reports exactly what was written.
func TestSliceWriter(t *testing.T) {
	w := wirefmt.NewSliceWriter(make([]byte, 4))
	if err := w.WriteByte('a'); err != nil {
		t.Fatalf("WriteByte: %v", err)
	}
	if n, err := w.Write([]byte("bc")); err != nil || n != 2 {
		t.Fatalf("Write = (%d, %v)", n, err)
	}
	if n, err := w.WriteString("d"); err != nil || n != 1 {
		t.Fatalf("WriteString = (%d, %v)", n, err)
	}
	if got := string(w.Bytes()); got != "abcd" {
		t.Fatalf("Bytes = %q, want %q", got, "abcd")
	}

	if err := w.WriteByte('e'); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("WriteByte past the end = %v, want io.ErrShortBuffer", err)
	}
	if _, err := w.Write([]byte("e")); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("Write past the end = %v, want io.ErrShortBuffer", err)
	}
	if _, err := w.WriteString("e"); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("WriteString past the end = %v, want io.ErrShortBuffer", err)
	}

	// A negative offset is a short buffer too.
	w.N = -1
	if err := w.WriteByte('x'); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("WriteByte at a negative offset = %v, want io.ErrShortBuffer", err)
	}
	if _, err := w.Write(nil); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("Write at a negative offset = %v, want io.ErrShortBuffer", err)
	}
	if _, err := w.WriteString(""); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("WriteString at a negative offset = %v, want io.ErrShortBuffer", err)
	}
}

// TestEncoderPaths covers every way an Encoder can be built, plus Flush and
// Reset on a buffered and an unbuffered encoder.
func TestEncoderPaths(t *testing.T) {
	payload := streamBytes(t, wcomposite.String, "hi")

	// A plain io.Writer is buffered.
	var plain bytes.Buffer
	e := wirefmt.NewEncoder(&plain)
	if _, err := wirefmt.EncodeTo(e, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo: %v", err)
	}
	if err := e.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !bytes.Equal(plain.Bytes(), payload) {
		t.Fatalf("buffered encoder wrote %v, want %v", plain.Bytes(), payload)
	}

	// A *bufio.Writer is used directly.
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	e2 := wirefmt.NewEncoder(bw)
	if e2.Writer() != bw {
		t.Fatal("NewEncoder did not reuse the *bufio.Writer")
	}
	if _, err := wirefmt.EncodeTo(e2, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo: %v", err)
	}
	if err := e2.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), payload) {
		t.Fatalf("bufio encoder wrote %v, want %v", buf.Bytes(), payload)
	}

	// A Writer (bytes.Buffer) is used directly, so Flush has nothing to do.
	var direct bytes.Buffer
	e3 := wirefmt.NewEncoder(&direct)
	if e3.Writer() != &direct {
		t.Fatal("NewEncoder did not reuse the Writer")
	}
	if _, err := wirefmt.EncodeTo(e3, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo: %v", err)
	}
	if err := e3.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !bytes.Equal(direct.Bytes(), payload) {
		t.Fatalf("direct encoder wrote %v, want %v", direct.Bytes(), payload)
	}

	// Reset retargets a buffered encoder.
	var other bytes.Buffer
	e.Reset(&other)
	if _, err := wirefmt.EncodeTo(e, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo after Reset: %v", err)
	}
	if err := e.Flush(); err != nil {
		t.Fatalf("Flush after Reset: %v", err)
	}
	if !bytes.Equal(other.Bytes(), payload) {
		t.Fatalf("encoder after Reset wrote %v, want %v", other.Bytes(), payload)
	}

	// Reset rebuilds an unbuffered encoder.
	direct.Reset()
	e3.Reset(&direct)
	if e3.Writer() != &direct {
		t.Fatal("Reset did not rebuild the encoder around the Writer")
	}
}

// TestDecoderPaths covers every way a Decoder can be built, plus Reset on a
// buffered and an unbuffered decoder.
func TestDecoderPaths(t *testing.T) {
	payload := streamBytes(t, wcomposite.String, "hi")
	decode := func(t *testing.T, d *wirefmt.Decoder) string {
		t.Helper()
		v, _, err := wirefmt.DecodeFrom(d, wcomposite.String)
		if err != nil || v != "hi" {
			t.Fatalf("DecodeFrom = (%q, %v)", v, err)
		}
		return v
	}

	// A plain io.Reader is buffered.
	d := wirefmt.NewDecoder(bytes.NewReader(payload))
	decode(t, d)
	d.Reset(bytes.NewReader(payload))
	decode(t, d)

	// A *bufio.Reader is used directly.
	br := bufio.NewReader(bytes.NewReader(payload))
	d2 := wirefmt.NewDecoder(br)
	if d2.Reader() != br {
		t.Fatal("NewDecoder did not reuse the *bufio.Reader")
	}
	decode(t, d2)
	d2.Reset(bufio.NewReader(bytes.NewReader(payload)))
	decode(t, d2)

	// A Reader (bytes.Reader) is used directly, and Reset rebuilds the decoder.
	sr := bytes.NewReader(payload)
	d3 := wirefmt.NewDecoder(sr)
	if d3.Reader() != sr {
		t.Fatal("NewDecoder did not reuse the Reader")
	}
	decode(t, d3)
	d3.Reset(bytes.NewReader(payload))
	decode(t, d3)
}

// TestEncoderWrapsAPlainWriter checks the buffering branch of NewEncoder and the
// buffered branch of Reset.
func TestEncoderWrapsAPlainWriter(t *testing.T) {
	want := streamBytes(t, wcomposite.String, "hi")

	var buf bytes.Buffer
	e := wirefmt.NewEncoder(onlyWriter{w: &buf})
	if _, err := wirefmt.EncodeTo(e, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo: %v", err)
	}
	if err := e.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("wrapped encoder wrote %v, want %v", buf.Bytes(), want)
	}

	var other bytes.Buffer
	e.Reset(onlyWriter{w: &other})
	if _, err := wirefmt.EncodeTo(e, wcomposite.String, "hi"); err != nil {
		t.Fatalf("EncodeTo after Reset: %v", err)
	}
	if err := e.Flush(); err != nil {
		t.Fatalf("Flush after Reset: %v", err)
	}
	if !bytes.Equal(other.Bytes(), want) {
		t.Fatalf("encoder after Reset wrote %v, want %v", other.Bytes(), want)
	}
}

// TestDecoderWrapsAPlainReader checks the buffering branch of NewDecoder.
func TestDecoderWrapsAPlainReader(t *testing.T) {
	payload := streamBytes(t, wcomposite.String, "hi")
	d := wirefmt.NewDecoder(onlyReader{r: bytes.NewReader(payload)})
	if v, _, err := wirefmt.DecodeFrom(d, wcomposite.String); err != nil || v != "hi" {
		t.Fatalf("DecodeFrom = (%q, %v)", v, err)
	}
	d.Reset(bytes.NewReader(payload))
	if v, _, err := wirefmt.DecodeFrom(d, wcomposite.String); err != nil || v != "hi" {
		t.Fatalf("DecodeFrom after Reset = (%q, %v)", v, err)
	}
}
