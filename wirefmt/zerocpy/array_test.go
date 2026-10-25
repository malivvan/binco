package zerocpy_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	sarropts "github.com/malivvan/binco/wirefmt/config/array"
	sstropts "github.com/malivvan/binco/wirefmt/config/string"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
	snative "github.com/malivvan/binco/wirefmt/zerocpy"
)

func TestArrayEdgeCases(t *testing.T) {
	// A custom length codec replaces the default positive varint.
	rawLen := snative.NewArraySerializer[[2]int, int](wvarnum.Int, sarropts.WithLenSerializer[int](wscalar.Int))
	bs := encodeStream(t, rawLen, [2]int{1, 2})
	if len(bs) != wscalar.Int.Size(0)+2 || bs[0] != 2 {
		t.Fatalf("Marshal with a length codec = %v", bs)
	}
	got, n, err := rawLen.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil || n != len(bs) || got != [2]int{1, 2} {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}

	// A zero-length array has no elements to decode.
	zero := snative.NewArraySerializer[[0]int, int](wvarnum.Int)
	zbs := encodeStream(t, zero, [0]int{})
	if len(zbs) != 1 || zbs[0] != 0 {
		t.Fatalf("Marshal of a zero-length array = %v", zbs)
	}
	if got, n, err := zero.Unmarshal(wirefmt.NewReader(bytes.NewReader(zbs))); err != nil ||
		n != 1 || got != [0]int{} {
		t.Fatalf("Unmarshal of a zero-length array = (%v, %d, %v)", got, n, err)
	}

	// Skip advances over the whole array.
	arr := snative.NewArraySerializer[[3]int, int](wvarnum.Int)
	abs := encodeStream(t, arr, [3]int{1, 2, 3})
	if n, err := arr.Skip(wirefmt.NewReader(bytes.NewReader(abs))); err != nil || n != len(abs) {
		t.Fatalf("Skip = (%d, %v), want %d bytes", n, err, len(abs))
	}
}

func TestEmptyWireText(t *testing.T) {
	if v, n, err := snative.String.Unmarshal(wirefmt.NewReader(bytes.NewReader([]byte{0}))); err != nil || v != "" || n != 1 {
		t.Fatalf("String.Unmarshal = (%q, %d, %v)", v, n, err)
	}

	noop := binco.ValidatorFn[int](func(int) error { return nil })
	valid := snative.NewValidStringSerializer(sstropts.WithLenValidator(noop))
	if v, n, err := valid.Unmarshal(wirefmt.NewReader(bytes.NewReader([]byte{0}))); err != nil || v != "" || n != 1 {
		t.Fatalf("validated String.Unmarshal = (%q, %d, %v)", v, n, err)
	}

	// A stream that ends before the length is complete.
	empty := func() wirefmt.Reader { return wirefmt.NewReader(bytes.NewReader(nil)) }
	if _, _, err := snative.String.Unmarshal(empty()); !errors.Is(err, io.EOF) {
		t.Fatalf("String.Unmarshal of an empty stream = %v, want io.EOF", err)
	}
	if _, _, err := valid.Unmarshal(empty()); !errors.Is(err, io.EOF) {
		t.Fatalf("validated String.Unmarshal of an empty stream = %v, want io.EOF", err)
	}
}
