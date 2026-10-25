package ptrmap_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func TestPointerCodecErrors(t *testing.T) {
	newSerializer := func() binco.Serializer[*int] {
		return ptrmap.NewPtrSerializer(binco.NewPtrMap(), binco.NewReversePtrMap(), varnum.Int)
	}

	tests := []struct {
		name string
		fn   func() error
		want error
	}{
		{"unmarshal empty", func() error {
			_, _, err := newSerializer().Unmarshal(nil)
			return err
		}, binco.ErrTooSmallByteSlice},
		{"unmarshal unknown flag", func() error {
			_, _, err := newSerializer().Unmarshal([]byte{0xFF})
			return err
		}, binco.ErrWrongFormat},
		{"unmarshal truncated id", func() error {
			_, _, err := newSerializer().Unmarshal([]byte{byte(binco.PtrMapping)})
			return err
		}, binco.ErrTooSmallByteSlice},
		{"unmarshal truncated data", func() error {
			_, _, err := newSerializer().Unmarshal([]byte{byte(binco.PtrMapping), 0})
			return err
		}, binco.ErrTooSmallByteSlice},
		{"skip empty", func() error {
			_, err := newSerializer().Skip(nil)
			return err
		}, binco.ErrTooSmallByteSlice},
		{"skip unknown flag", func() error {
			_, err := newSerializer().Skip([]byte{0xFF})
			return err
		}, binco.ErrWrongFormat},
		{"skip truncated id", func() error {
			_, err := newSerializer().Skip([]byte{byte(binco.PtrMapping)})
			return err
		}, binco.ErrTooSmallByteSlice},
		{"skip truncated data", func() error {
			_, err := newSerializer().Skip([]byte{byte(binco.PtrMapping), 0})
			return err
		}, binco.ErrTooSmallByteSlice},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != tc.want {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPointerCodecNil(t *testing.T) {
	serializer := ptrmap.NewPtrSerializer(binco.NewPtrMap(), binco.NewReversePtrMap(), varnum.Int)

	bs := binco.Encode(serializer, (*int)(nil))
	if len(bs) != 1 || bs[0] != byte(binco.PtrNil) {
		t.Fatalf("Marshal(nil) = %v", bs)
	}
	got, n, err := serializer.Unmarshal(bs)
	if err != nil || got != nil || n != 1 {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}
	if size := serializer.Size(nil); size != 1 {
		t.Fatalf("Size(nil) = %d, want 1", size)
	}
	if n, err := serializer.Skip(bs); err != nil || n != 1 {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}
}
