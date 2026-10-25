package zerocpy_test

import (
	"errors"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/zerocpy"
)

// TestArraySerializerRejectsMismatchedTypes covers S5: the constructor used to
// trust its two type parameters, so a mismatched pair read and wrote out of
// bounds through unsafe.
func TestArraySerializerRejectsMismatchedTypes(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"not an array", func() { zerocpy.NewArraySerializer[int32, int64](scalar.Int64) }},
		{"wrong element type", func() { zerocpy.NewArraySerializer[[4]int32, int64](scalar.Int64) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("NewArraySerializer accepted mismatched type parameters")
				}
				err, ok := r.(error)
				if !ok || !errors.Is(err, binco.ErrWrongArrayType) {
					t.Fatalf("panic value %v, want ErrWrongArrayType", r)
				}
			}()
			tc.fn()
		})
	}
}

// TestUnalignedAccess covers S6: a fixed-width value that starts at an odd
// offset must decode correctly on architectures that fault on unaligned access,
// so the codecs may not cast the slice to a wider pointer.
func TestUnalignedAccess(t *testing.T) {
	for offset := 1; offset <= 9; offset += 2 {
		buf := make([]byte, offset+8)
		scalar.Int64.Marshal(-123456789, buf[offset:])
		got, n, err := zerocpy.Int64.Unmarshal(buf[offset:])
		if err != nil || n != 8 || got != -123456789 {
			t.Fatalf("offset %d: Unmarshal = (%d, %d, %v)", offset, got, n, err)
		}
	}
}

// TestTimesDecodeToUTC covers S6: the non-UTC time codecs used to return
// time.Local, so the same payload compared differently across machines.
func TestTimesDecodeToUTC(t *testing.T) {
	v := time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("custom", 3*60*60))
	for name, s := range map[string]binco.Serializer[time.Time]{
		"TimeUnix":      zerocpy.TimeUnix,
		"TimeUnixMilli": zerocpy.TimeUnixMilli,
		"TimeUnixMicro": zerocpy.TimeUnixMicro,
	} {
		got, _, err := s.Unmarshal(binco.Encode(s, v))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got.Location() != time.UTC {
			t.Fatalf("%s returned a time in %v, want UTC", name, got.Location())
		}
		if !got.Equal(v) {
			t.Fatalf("%s returned %v, want %v", name, got, v)
		}
	}
}
