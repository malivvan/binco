package binco_test

import (
	"errors"
	"math"
	"testing"

	"github.com/malivvan/binco"
)

// TestCheckLength pins the overflow-safe length check every decoder relies on.
func TestCheckLength(t *testing.T) {
	tests := []struct {
		name   string
		bs     []byte
		n      int
		length int
		want   error
	}{
		{"exact", []byte("abc"), 0, 3, nil},
		{"offset", []byte("abc"), 1, 2, nil},
		{"empty", nil, 0, 0, nil},
		{"zero length", []byte("abc"), 3, 0, nil},
		{"negative length", []byte("abc"), 0, -1, binco.ErrNegativeLength},
		{"negative offset", []byte("abc"), -1, 0, binco.ErrTruncated},
		{"offset past the end", []byte("abc"), 4, 0, binco.ErrTruncated},
		{"longer than the input", []byte("abc"), 0, 4, binco.ErrTruncated},
		{"huge length", []byte("abc"), 0, math.MaxInt, binco.ErrTruncated},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := binco.CheckLength(tc.bs, tc.n, tc.length); !errors.Is(err, tc.want) {
				t.Fatalf("CheckLength(%q, %d, %d) = %v, want %v", tc.bs, tc.n, tc.length, err, tc.want)
			}
		})
	}
}

// TestRemaining checks the byte count after an offset, including offsets that
// are not inside the buffer.
func TestRemaining(t *testing.T) {
	bs := []byte("abc")
	for _, tc := range []struct {
		n    int
		want int
	}{
		{0, 3},
		{1, 2},
		{3, 0},
		{-1, 0},
		{4, 0},
	} {
		if got := binco.Remaining(bs, tc.n); got != tc.want {
			t.Fatalf("Remaining(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
	if got := binco.Remaining(nil, 0); got != 0 {
		t.Fatalf("Remaining of nil = %d, want 0", got)
	}
}

// TestAllocHint checks that a decoded length can only shrink the allocation,
// never grow it beyond the bytes that are actually present.
func TestAllocHint(t *testing.T) {
	for _, tc := range []struct {
		length, remaining, want int
	}{
		{0, 10, 0},
		{-1, 10, 0},
		{5, 10, 5},
		{10, 10, 10},
		{15, 10, 10},
		{math.MaxInt, 3, 3},
	} {
		if got := binco.AllocHint(tc.length, tc.remaining); got != tc.want {
			t.Fatalf("AllocHint(%d, %d) = %d, want %d", tc.length, tc.remaining, got, tc.want)
		}
	}
}

// TestCheckLengthElem checks the element-aware variant used by the fixed-width
// slice fast path, including the overflow-free multiplication.
func TestCheckLengthElem(t *testing.T) {
	tests := []struct {
		name            string
		bs              []byte
		n, length, size int
		want            error
	}{
		{"negative length", []byte("abcd"), 0, -1, 2, binco.ErrNegativeLength},
		{"negative offset", []byte("abcd"), -1, 1, 2, binco.ErrTruncated},
		{"offset past the end", []byte("abcd"), 5, 1, 2, binco.ErrTruncated},
		{"one-byte elements fit", []byte("abcd"), 1, 3, 1, nil},
		{"one-byte elements too many", []byte("abcd"), 0, 5, 1, binco.ErrTruncated},
		{"two-byte elements fit", []byte("abcd"), 0, 2, 2, nil},
		{"two-byte elements too many", []byte("abcd"), 0, 3, 2, binco.ErrTruncated},
		{"zero-size elements", []byte("abcd"), 0, 5, 0, binco.ErrTruncated},
		{"huge length", []byte("abcd"), 0, math.MaxInt, 8, binco.ErrTruncated},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := binco.CheckLengthElem(tc.bs, tc.n, tc.length, tc.size); !errors.Is(err, tc.want) {
				t.Fatalf("CheckLengthElem(%d, %d, %d) = %v, want %v", tc.n, tc.length, tc.size, err, tc.want)
			}
		})
	}
}

// TestCheckWrite checks that every way of describing a short buffer panics with
// the documented sentinel, and that a valid write does not panic.
func TestCheckWrite(t *testing.T) {
	binco.CheckWrite(make([]byte, 3), 1, 2)
	binco.CheckWrite(nil, 0, 0)

	for _, tc := range []struct {
		name string
		n    int
		need int
		bs   []byte
	}{
		{"negative offset", -1, 0, make([]byte, 3)},
		{"negative need", 0, -1, make([]byte, 3)},
		{"offset past the end", 4, 0, make([]byte, 3)},
		{"need past the end", 2, 2, make([]byte, 3)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
					t.Fatalf("panic value %v, want ErrBufferTooSmall", err)
				}
			}()
			binco.CheckWrite(tc.bs, tc.n, tc.need)
		})
	}
}

// TestDepth checks the recursion counter used by the pointer codecs.
func TestDepth(t *testing.T) {
	var d binco.Depth
	for range binco.MaxDepth {
		if err := d.Enter(); err != nil {
			t.Fatalf("Enter at level %d: %v", binco.MaxDepth, err)
		}
	}
	if err := d.Enter(); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Enter past MaxDepth = %v, want ErrTooDeep", err)
	}
	d.Leave()
	if err := d.Enter(); err != nil {
		t.Fatalf("Enter after Leave = %v", err)
	}
	d.Reset()
	if err := d.Enter(); err != nil {
		t.Fatalf("Enter after Reset = %v", err)
	}

	// Leave on a fresh counter is a no-op rather than an underflow.
	var empty binco.Depth
	empty.Leave()
	if err := empty.Enter(); err != nil {
		t.Fatalf("Enter after Leave at zero = %v", err)
	}
}
