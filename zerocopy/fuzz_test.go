//go:build fuzz

package zerocopy_test

import (
	"os"
	"runtime/debug"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/zerocopy"
)

func TestMain(m *testing.M) {
	// A decoded zerocopy string or slice aliases the input, so an unbounded read
	// would be an out-of-bounds view; keep a hard memory ceiling as well.
	debug.SetMemoryLimit(256 << 20)
	os.Exit(m.Run())
}

// FuzzString decodes arbitrary bytes as a zero-copy string. The value aliases bs,
// so the check compares it against the bytes it covers.
func FuzzString(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 'a'}, {3, 'a', 'b', 'c'}, {0x80}, {0x80, 0x01},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		v, n, err := zerocopy.String.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err != nil {
			return
		}
		if got := zerocopy.String.Size(v); got != n {
			t.Fatalf("Size = %d, Unmarshal consumed %d", got, n)
		}
		if skipN, skipErr := zerocopy.String.Skip(bs); skipErr != nil || skipN != n {
			t.Fatalf("Skip = (%d, %v), Unmarshal consumed %d", skipN, skipErr, n)
		}
	})
}

// FuzzByteSlice is FuzzString for the zero-copy byte slice codec.
func FuzzByteSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 0xff}, {3, 1, 2, 3}, {0x80}, {0x80, 0x01},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		v, n, err := zerocopy.ByteSlice.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err != nil {
			return
		}
		if got := zerocopy.ByteSlice.Size(v); got != n {
			t.Fatalf("Size = %d, Unmarshal consumed %d", got, n)
		}
	})
}

// FuzzArray decodes arbitrary bytes as an array and checks the length rule.
func FuzzArray(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {2, 2, 4}, {1, 2}, {2, 2}, {0xff},
	} {
		f.Add(seed)
	}
	serializer := zerocopy.NewArraySerializer[[2]int, int](scalar.Int)
	f.Fuzz(func(t *testing.T, bs []byte) {
		v, n, err := serializer.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err != nil {
			return
		}
		if got := serializer.Size(v); got != n {
			t.Fatalf("Size = %d, Unmarshal consumed %d", got, n)
		}
	})
}

// FuzzScalars decodes arbitrary bytes at an odd offset, which is where an
// unaligned unsafe load used to break.
func FuzzScalars(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7, 8},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		buf := make([]byte, len(bs)+1)
		copy(buf[1:], bs)
		off := buf[1:]
		if _, n, err := zerocopy.Int64.Unmarshal(off); err == nil && n != binco.Num64RawSize {
			t.Fatalf("Int64 consumed %d bytes", n)
		}
		if _, n, err := zerocopy.Float32.Unmarshal(off); err == nil && n != binco.Num32RawSize {
			t.Fatalf("Float32 consumed %d bytes", n)
		}
		if _, n, err := zerocopy.Uint16.Unmarshal(off); err == nil && n != binco.Num16RawSize {
			t.Fatalf("Uint16 consumed %d bytes", n)
		}
	})
}
