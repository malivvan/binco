package scalar_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
)

func TestTruncatedInputs(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"float64.Unmarshal", func() error {
			for i := 0; i < binco.Num64RawSize; i++ {
				if _, _, err := scalar.Float64.Unmarshal(make([]byte, i)); err != binco.ErrTooSmallByteSlice {
					return err
				}
			}
			return nil
		}},
		{"float32.Unmarshal", func() error {
			for i := 0; i < binco.Num32RawSize; i++ {
				if _, _, err := scalar.Float32.Unmarshal(make([]byte, i)); err != binco.ErrTooSmallByteSlice {
					return err
				}
			}
			return nil
		}},
		{"time.Unmarshal", func() error {
			for i := 0; i < binco.Num64RawSize; i++ {
				if _, _, err := scalar.TimeUnix.Unmarshal(make([]byte, i)); err != binco.ErrTooSmallByteSlice {
					return err
				}
			}
			return nil
		}},
		{"SkipInteger64", func() error {
			if _, err := scalar.SkipInteger64(make([]byte, 7)); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
		{"SkipInteger32", func() error {
			if _, err := scalar.SkipInteger32(make([]byte, 3)); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
		{"SkipInteger16", func() error {
			if _, err := scalar.SkipInteger16(make([]byte, 1)); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
		{"SkipInteger8", func() error {
			if _, err := scalar.SkipInteger8(nil); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
		{"float64.Skip", func() error {
			if _, err := scalar.Float64.Skip(make([]byte, 7)); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
		{"float32.Skip", func() error {
			if _, err := scalar.Float32.Skip(make([]byte, 3)); err != binco.ErrTooSmallByteSlice {
				return err
			}
			return nil
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != nil {
				t.Fatalf("%v: got %v, want ErrTooSmallByteSlice", tc.name, err)
			}
		})
	}
}
