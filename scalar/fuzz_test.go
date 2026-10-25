//go:build fuzz

package scalar_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
)

// FuzzFixedWidth decodes arbitrary bytes with the fixed-width codecs. None of
// them may panic, read past the input or report a consumed count outside it.
func FuzzFixedWidth(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7, 8},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		check(t, "Uint64", scalar.Uint64, bs, binco.Num64RawSize)
		check(t, "Int32", scalar.Int32, bs, binco.Num32RawSize)
		check(t, "Float64", scalar.Float64, bs, binco.Num64RawSize)
	})
}

func check[T any](t *testing.T, name string, ser binco.Serializer[T], bs []byte, width int) {
	t.Helper()
	_, n, err := ser.Unmarshal(bs)
	if n < 0 || n > len(bs) {
		t.Fatalf("%s: Unmarshal consumed %d of %d bytes", name, n, len(bs))
	}
	if err == nil && n != width {
		t.Fatalf("%s: Unmarshal consumed %d, want %d", name, n, width)
	}
	if err != nil && len(bs) >= width {
		t.Fatalf("%s: Unmarshal of %d bytes returned %v", name, len(bs), err)
	}
	if skipN, skipErr := ser.Skip(bs); skipN < 0 || skipN > len(bs) || (err == nil) != (skipErr == nil) {
		t.Fatalf("%s: Skip = (%d, %v) after Unmarshal returned %v", name, skipN, skipErr, err)
	}
}

// FuzzTime decodes arbitrary bytes as a time.Time.
func FuzzTime(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 2, 3, 4, 5, 6, 7, 8}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		check(t, "TimeUnix", scalar.TimeUnix, bs, binco.Num64RawSize)
		check(t, "TimeUnixNano", scalar.TimeUnixNano, bs, binco.Num64RawSize)
	})
}
