package zerocpy_test

import (
	"testing"

	"github.com/malivvan/binco"
	bsopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/zerocpy"
)

func TestEmptyText(t *testing.T) {
	// A zero length decodes to an empty value without touching the buffer.
	if v, n, err := zerocpy.String.Unmarshal([]byte{0}); err != nil || v != "" || n != 1 {
		t.Fatalf("String.Unmarshal = (%q, %d, %v)", v, n, err)
	}
	if v, n, err := zerocpy.ByteSlice.Unmarshal([]byte{0}); err != nil || len(v) != 0 || n != 1 {
		t.Fatalf("ByteSlice.Unmarshal = (%v, %d, %v)", v, n, err)
	}

	// The validating codecs take the same path after the validator ran.
	noop := binco.ValidatorFn[int](func(int) error { return nil })
	validText := zerocpy.NewValidStringSer(stropts.WithLenValidator(noop))
	if v, n, err := validText.Unmarshal([]byte{0}); err != nil || v != "" || n != 1 {
		t.Fatalf("validated String.Unmarshal = (%q, %d, %v)", v, n, err)
	}
	validBlob := zerocpy.NewValidByteSliceSer(bsopts.WithLenValidator(noop))
	if v, n, err := validBlob.Unmarshal([]byte{0}); err != nil || len(v) != 0 || n != 1 {
		t.Fatalf("validated ByteSlice.Unmarshal = (%v, %d, %v)", v, n, err)
	}
}

func TestTruncatedText(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"string length", func() error { _, _, err := zerocpy.String.Unmarshal(nil); return err }},
		{"string data", func() error { _, _, err := zerocpy.String.Unmarshal([]byte{3, 'a'}); return err }},
		{"byte slice data", func() error { _, _, err := zerocpy.ByteSlice.Unmarshal([]byte{2, 1}); return err }},
		{"validated string data", func() error {
			noop := binco.ValidatorFn[int](func(int) error { return nil })
			_, _, err := zerocpy.NewValidStringSer(stropts.WithLenValidator(noop)).
				Unmarshal([]byte{3, 'a'})
			return err
		}},
		{"validated byte slice data", func() error {
			noop := binco.ValidatorFn[int](func(int) error { return nil })
			_, _, err := zerocpy.NewValidByteSliceSer(bsopts.WithLenValidator(noop)).
				Unmarshal([]byte{2, 1})
			return err
		}},
		{"byte slice length", func() error {
			_, _, err := zerocpy.ByteSlice.Unmarshal(nil)
			return err
		}},
		{"validated string length", func() error {
			_, _, err := zerocpy.NewValidStringSer().Unmarshal(nil)
			return err
		}},
		{"validated byte slice length", func() error {
			_, _, err := zerocpy.NewValidByteSliceSer().Unmarshal(nil)
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != binco.ErrTooSmallByteSlice {
				t.Fatalf("got %v, want ErrTooSmallByteSlice", err)
			}
		})
	}
}
