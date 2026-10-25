package le_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/le"
)

// TestPutAndGet checks that every width writes little-endian bytes, reports its
// size and reads back, including at an unaligned offset.
func TestPutAndGet(t *testing.T) {
	bs := make([]byte, 16)
	if n := le.Put8(bs, uint8(0xff)); n != 1 || bs[0] != 0xff {
		t.Fatalf("Put8 = (%d, %v)", n, bs[:1])
	}
	if n := le.Put16(bs[2:], uint16(0x0102)); n != 2 {
		t.Fatalf("Put16 size = %d", n)
	}
	if n := le.Put32(bs[4:], uint32(0x01020304)); n != 4 {
		t.Fatalf("Put32 size = %d", n)
	}
	if n := le.Put64(bs[8:], int64(-1234567890)); n != 8 {
		t.Fatalf("Put64 size = %d", n)
	}

	if got, n, err := le.Get8[uint8](bs); err != nil || n != 1 || got != 0xff {
		t.Fatalf("Get8 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := le.Get16[uint16](bs[2:]); err != nil || n != 2 || got != 0x0102 {
		t.Fatalf("Get16 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := le.Get32[uint32](bs[4:]); err != nil || n != 4 || got != 0x01020304 {
		t.Fatalf("Get32 = (%#x, %d, %v)", got, n, err)
	}
	if got, n, err := le.Get64[int64](bs[8:]); err != nil || n != 8 || got != -1234567890 {
		t.Fatalf("Get64 = (%d, %d, %v)", got, n, err)
	}

	// Little-endian, whatever the host order is.
	if !bytesEqual(bs[2:4], []byte{2, 1}) || !bytesEqual(bs[4:8], []byte{4, 3, 2, 1}) {
		t.Fatalf("not little-endian: %v", bs)
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestShortInput checks that every reader and skipper reports the missing bytes.
func TestShortInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func() (int, error)
	}{
		{"Get8", func() (int, error) { _, n, err := le.Get8[uint8](nil); return n, err }},
		{"Get16", func() (int, error) { _, n, err := le.Get16[uint16](make([]byte, 1)); return n, err }},
		{"Get32", func() (int, error) { _, n, err := le.Get32[uint32](make([]byte, 3)); return n, err }},
		{"Get64", func() (int, error) { _, n, err := le.Get64[uint64](make([]byte, 7)); return n, err }},
		{"Skip8", func() (int, error) { return le.Skip8(nil) }},
		{"Skip16", func() (int, error) { return le.Skip16(make([]byte, 1)) }},
		{"Skip32", func() (int, error) { return le.Skip32(make([]byte, 3)) }},
		{"Skip64", func() (int, error) { return le.Skip64(make([]byte, 7)) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.call(); !errors.Is(err, binco.ErrTruncated) {
				t.Fatalf("%s = %v, want ErrTruncated", tc.name, err)
			}
		})
	}

	// The successful skips report the width.
	for _, tc := range []struct {
		name string
		call func() (int, error)
		want int
	}{
		{"Skip8", func() (int, error) { return le.Skip8(make([]byte, 1)) }, 1},
		{"Skip16", func() (int, error) { return le.Skip16(make([]byte, 2)) }, 2},
		{"Skip32", func() (int, error) { return le.Skip32(make([]byte, 4)) }, 4},
		{"Skip64", func() (int, error) { return le.Skip64(make([]byte, 8)) }, 8},
	} {
		if n, err := tc.call(); err != nil || n != tc.want {
			t.Fatalf("%s = (%d, %v), want (%d, nil)", tc.name, n, err, tc.want)
		}
	}
}

// TestWord checks the host-width helpers: the wire width is always 64 bits, and
// Skip applies the same overflow rule as the reader.
func TestWord(t *testing.T) {
	bs := make([]byte, 8)
	if n := le.PutWord(bs, -300); n != 8 {
		t.Fatalf("PutWord size = %d", n)
	}
	if got, n, err := le.GetWord(bs); err != nil || n != 8 || got != -300 {
		t.Fatalf("GetWord = (%d, %d, %v)", got, n, err)
	}
	if n, err := le.SkipWord(bs); err != nil || n != 8 {
		t.Fatalf("SkipWord = (%d, %v)", n, err)
	}

	if n := le.PutWordUnsigned(bs, 300); n != 8 {
		t.Fatalf("PutWordUnsigned size = %d", n)
	}
	if got, n, err := le.GetWordUnsigned(bs); err != nil || n != 8 || got != 300 {
		t.Fatalf("GetWordUnsigned = (%d, %d, %v)", got, n, err)
	}
	if n, err := le.SkipWordUnsigned(bs); err != nil || n != 8 {
		t.Fatalf("SkipWordUnsigned = (%d, %v)", n, err)
	}

	// Truncated input reaches the error return of every word helper.
	if _, _, err := le.GetWord(make([]byte, 7)); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("GetWord short = %v", err)
	}
	if _, err := le.SkipWord(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("SkipWord short = %v", err)
	}
	if _, _, err := le.GetWordUnsigned(make([]byte, 7)); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("GetWordUnsigned short = %v", err)
	}
	if _, err := le.SkipWordUnsigned(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("SkipWordUnsigned short = %v", err)
	}
}

// TestPutShortBuffer checks that a writer reports a short buffer with the usual
// sentinel instead of a runtime panic.
func TestPutShortBuffer(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"Put8", func() { le.Put8(nil, uint8(1)) }},
		{"Put16", func() { le.Put16(make([]byte, 1), uint16(1)) }},
		{"Put32", func() { le.Put32(make([]byte, 3), uint32(1)) }},
		{"Put64", func() { le.Put64(make([]byte, 7), int64(1)) }},
		{"PutWord", func() { le.PutWord(nil, 1) }},
		{"PutWordUnsigned", func() { le.PutWordUnsigned(nil, 1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if err, ok := recover().(error); !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
					t.Fatalf("panic value %v, want ErrBufferTooSmall", recover())
				}
			}()
			tc.call()
		})
	}
}
