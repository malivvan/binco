package varnum_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

func TestTruncatedEncodings(t *testing.T) {
	// A leading continuation byte with nothing after it: every codec runs out
	// of input before the encoding ends.
	truncated := []byte{0x80}

	tests := []struct {
		name string
		fn   func() error
	}{
		{"uint8", func() error { _, _, err := varnum.Uint8.Unmarshal(truncated); return err }},
		{"uint16", func() error { _, _, err := varnum.Uint16.Unmarshal(truncated); return err }},
		{"uint32", func() error { _, _, err := varnum.Uint32.Unmarshal(truncated); return err }},
		{"uint64", func() error { _, _, err := varnum.Uint64.Unmarshal(truncated); return err }},
		{"uint", func() error { _, _, err := varnum.Uint.Unmarshal(truncated); return err }},
		{"int8", func() error { _, _, err := varnum.Int8.Unmarshal(truncated); return err }},
		{"int16", func() error { _, _, err := varnum.Int16.Unmarshal(truncated); return err }},
		{"int32", func() error { _, _, err := varnum.Int32.Unmarshal(truncated); return err }},
		{"int64", func() error { _, _, err := varnum.Int64.Unmarshal(truncated); return err }},
		{"int", func() error { _, _, err := varnum.Int.Unmarshal(truncated); return err }},
		{"positive int8", func() error { _, _, err := varnum.PositiveInt8.Unmarshal(truncated); return err }},
		{"positive int64", func() error { _, _, err := varnum.PositiveInt64.Unmarshal(truncated); return err }},
		{"positive int", func() error { _, _, err := varnum.PositiveInt.Unmarshal(truncated); return err }},
		{"float64", func() error { _, _, err := varnum.Float64.Unmarshal(truncated); return err }},
		{"float32", func() error { _, _, err := varnum.Float32.Unmarshal(truncated); return err }},
		{"skip uint64", func() error { _, err := varnum.Uint64.Skip(truncated); return err }},
		{"skip negative", func() error { _, err := varnum.PositiveInt8.Skip([]byte{0xFF}); return err }},
		{"skip float64", func() error { _, err := varnum.Float64.Skip(truncated); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != binco.ErrTooSmallByteSlice {
				t.Fatalf("got %v, want ErrTooSmallByteSlice", err)
			}
		})
	}
}

func TestOverlongEncodings(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"uint8", func() error { _, _, err := varnum.Uint8.Unmarshal([]byte{0x80, 0x02}); return err }},
		{"uint16", func() error { _, _, err := varnum.Uint16.Unmarshal([]byte{0x80, 0x80, 0x04}); return err }},
		{"uint32", func() error {
			_, _, err := varnum.Uint32.Unmarshal([]byte{0x80, 0x80, 0x80, 0x80, 0x10})
			return err
		}},
		{"uint64", func() error {
			_, _, err := varnum.Uint64.Unmarshal([]byte{
				0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x02,
			})
			return err
		}},
		{"int8", func() error { _, _, err := varnum.Int8.Unmarshal([]byte{0x80, 0x02}); return err }},
		{"positive int8", func() error {
			_, _, err := varnum.PositiveInt8.Unmarshal([]byte{0x80, 0x02})
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != binco.ErrOverflow {
				t.Fatalf("got %v, want ErrOverflow", err)
			}
		})
	}
}
