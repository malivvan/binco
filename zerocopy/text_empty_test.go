package zerocopy_test

import (
	"testing"

	"github.com/malivvan/binco"
	bsopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/zerocopy"
)

func TestEmptyText(t *testing.T) {
	// A zero length decodes to an empty value without touching the buffer.
	if v, n, err := zerocopy.String.Unmarshal([]byte{0}); err != nil || v != "" || n != 1 {
		t.Fatalf("String.Unmarshal = (%q, %d, %v)", v, n, err)
	}
	if v, n, err := zerocopy.ByteSlice.Unmarshal([]byte{0}); err != nil || len(v) != 0 || n != 1 {
		t.Fatalf("ByteSlice.Unmarshal = (%v, %d, %v)", v, n, err)
	}

	// The validating codecs take the same path after the validator ran.
	noop := binco.ValidatorFn[int](func(int) error { return nil })
	validText := zerocopy.NewStringSerializer(stropts.WithLenValidator(noop))
	if v, n, err := validText.Unmarshal([]byte{0}); err != nil || v != "" || n != 1 {
		t.Fatalf("validated String.Unmarshal = (%q, %d, %v)", v, n, err)
	}
	validBlob := zerocopy.NewByteSliceSerializer(bsopts.WithLenValidator(noop))
	if v, n, err := validBlob.Unmarshal([]byte{0}); err != nil || len(v) != 0 || n != 1 {
		t.Fatalf("validated ByteSlice.Unmarshal = (%v, %d, %v)", v, n, err)
	}
}

func TestTruncatedText(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"string length", func() error { _, _, err := zerocopy.String.Unmarshal(nil); return err }},
		{"string data", func() error { _, _, err := zerocopy.String.Unmarshal([]byte{3, 'a'}); return err }},
		{"byte slice data", func() error { _, _, err := zerocopy.ByteSlice.Unmarshal([]byte{2, 1}); return err }},
		{"validated string data", func() error {
			noop := binco.ValidatorFn[int](func(int) error { return nil })
			_, _, err := zerocopy.NewStringSerializer(stropts.WithLenValidator(noop)).
				Unmarshal([]byte{3, 'a'})
			return err
		}},
		{"validated byte slice data", func() error {
			noop := binco.ValidatorFn[int](func(int) error { return nil })
			_, _, err := zerocopy.NewByteSliceSerializer(bsopts.WithLenValidator(noop)).
				Unmarshal([]byte{2, 1})
			return err
		}},
		{"byte slice length", func() error {
			_, _, err := zerocopy.ByteSlice.Unmarshal(nil)
			return err
		}},
		{"validated string length", func() error {
			_, _, err := zerocopy.NewStringSerializer().Unmarshal(nil)
			return err
		}},
		{"validated byte slice length", func() error {
			_, _, err := zerocopy.NewByteSliceSerializer().Unmarshal(nil)
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != binco.ErrTruncated {
				t.Fatalf("got %v, want ErrTruncated", err)
			}
		})
	}
}
