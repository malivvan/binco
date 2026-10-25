//go:build fuzz

package varnum_test

import (
	"testing"

	"github.com/malivvan/binco/varnum"
)

// FuzzUnmarshal covers the buffer varint codecs: canonical form, overflow and
// the consumed count, for arbitrary bytes.
func FuzzUnmarshal(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {0x7f}, {0x80}, {0x80, 0x01}, {0xff, 0xff},
		{0x80, 0x80, 0x80, 0x00},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		check(t, "Uint64", varnum.Uint64, bs)
		check(t, "Int64", varnum.Int64, bs)
		check(t, "PositiveInt", varnum.PositiveInt, bs)
	})
}

func check[T any](t *testing.T, name string, ser interface {
	Unmarshal([]byte) (T, int, error)
	Skip([]byte) (int, error)
	Size(T) int
	Marshal(T, []byte) int
}, bs []byte) {
	t.Helper()
	v, n, err := ser.Unmarshal(bs)
	if n < 0 || n > len(bs) {
		t.Fatalf("%s: Unmarshal consumed %d of %d bytes", name, n, len(bs))
	}
	if err != nil {
		if skipN, skipErr := ser.Skip(bs); skipErr != err || skipN != n {
			t.Fatalf("%s: Skip = (%d, %v) after Unmarshal returned (%d, %v)",
				name, skipN, skipErr, n, err)
		}
		return
	}
	size := ser.Size(v)
	out := make([]byte, size)
	if written := ser.Marshal(v, out); written != size {
		t.Fatalf("%s: Marshal wrote %d, Size reported %d", name, written, size)
	}
	if n != size {
		t.Fatalf("%s: Unmarshal consumed %d, Size reported %d", name, n, size)
	}
}
