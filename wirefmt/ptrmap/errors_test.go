package ptrmap_test

import (
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/ptrmap"
	"github.com/malivvan/binco/wirefmt/varnum"
)

func TestPointerCodecStreamErrors(t *testing.T) {
	newSer := func() wirefmt.Serializer[*int] {
		return ptrmap.NewPtrSer(binco.NewPtrMap(), binco.NewReversePtrMap(), varnum.Int)
	}

	tests := []struct {
		name string
		fn   func() error
		want error
	}{
		{"unmarshal empty", func() error {
			_, _, err := newSer().Unmarshal(reader())
			return err
		}, io.EOF},
		{"unmarshal unknown flag", func() error {
			_, _, err := newSer().Unmarshal(reader(0xFF))
			return err
		}, binco.ErrWrongFormat},
		{"unmarshal truncated id", func() error {
			_, _, err := newSer().Unmarshal(reader(byte(binco.PtrMapping)))
			return err
		}, io.EOF},
		{"unmarshal truncated data", func() error {
			_, _, err := newSer().Unmarshal(reader(byte(binco.PtrMapping), 0))
			return err
		}, io.EOF},
		{"skip empty", func() error {
			_, err := newSer().Skip(reader())
			return err
		}, io.EOF},
		{"skip unknown flag", func() error {
			_, err := newSer().Skip(reader(0xFF))
			return err
		}, binco.ErrWrongFormat},
		{"skip truncated id", func() error {
			_, err := newSer().Skip(reader(byte(binco.PtrMapping)))
			return err
		}, io.EOF},
		{"skip truncated data", func() error {
			_, err := newSer().Skip(reader(byte(binco.PtrMapping), 0))
			return err
		}, io.EOF},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPointerCodecWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	ser := ptrmap.NewPtrSer(binco.NewPtrMap(), binco.NewReversePtrMap(), varnum.Int)
	v := 5

	if _, err := ser.Marshal(nil, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal(nil) returned %v, want %v", err, want)
	}
	if _, err := ser.Marshal(&v, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal(&v) returned %v, want %v", err, want)
	}
}
