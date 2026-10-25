package composite_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// failingCodec reports one byte per value and always fails to write. The sorted
// map collects entries before writing them, so it must not lose that error.
type failingCodec struct{ err error }

func (c failingCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, c.err }
func (failingCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	return 0, 0, nil
}
func (failingCodec) Size(v int) (size int) { return 1 }
func (failingCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return 0, nil
}

// TestSortedMapPropagatesElementError covers the error the sorting buffer used
// to swallow: a failing key or value codec must fail the whole Marshal.
func TestSortedMapPropagatesElementError(t *testing.T) {
	want := errors.New("encode failed")
	sorted := composite.NewSortedMapSerializer(varnum.Int, failingCodec{err: want})

	var buf bytes.Buffer
	if _, err := sorted.Marshal(map[int]int{1: 2}, wirefmt.NewWriter(&buf)); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
}

// shortCodec reports two bytes and writes one; the mismatch must be reported.
type shortCodec struct{}

func (shortCodec) Marshal(v int, w wirefmt.Writer) (n int, err error) {
	if err := w.WriteByte(byte(v)); err != nil {
		return 0, err
	}
	return 1, nil
}
func (shortCodec) Unmarshal(r wirefmt.Reader) (v int, n int, err error) {
	b, err := r.ReadByte()
	return int(b), 1, err
}
func (shortCodec) Size(v int) (size int)                    { return 2 }
func (shortCodec) Skip(r wirefmt.Reader) (n int, err error) { return 2, nil }

func TestSortedMapRejectsInconsistentSize(t *testing.T) {
	sorted := composite.NewSortedMapSerializer(varnum.Int, shortCodec{})

	var buf bytes.Buffer
	_, err := sorted.Marshal(map[int]int{1: 7}, wirefmt.NewWriter(&buf))
	if !errors.Is(err, binco.ErrWrongSize) {
		t.Fatalf("Marshal returned %v, want ErrWrongSize", err)
	}
}
