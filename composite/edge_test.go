package composite_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bsopts "github.com/malivvan/binco/config/byte_slice"
	mapopts "github.com/malivvan/binco/config/map"
	slopts "github.com/malivvan/binco/config/slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

var errTooLong = errors.New("too long")

func TestStringEdgeCases(t *testing.T) {
	// A buffer that cannot hold the value is a caller bug, so it panics.
	checkPanics(t, func() {
		composite.MarshalString("hello", varnum.PositiveInt, make([]byte, 2))
	})

	// Skip must not read past the payload.
	if _, err := composite.SkipString(varnum.PositiveInt, []byte{5, 'a'}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("SkipString returned %v, want ErrTooSmallByteSlice", err)
	}

	// The validating codec reports a rejected length and a short payload.
	valid := composite.NewValidStringSer(stropts.WithLenValidator(
		binco.ValidatorFn[int](func(int) error { return errTooLong })))
	if _, _, err := valid.Unmarshal([]byte{3, 'a', 'b', 'c'}); !errors.Is(err, errTooLong) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, errTooLong)
	}

	short := composite.NewValidStringSer(stropts.WithLenValidator(
		binco.ValidatorFn[int](func(int) error { return nil })))
	if _, _, err := short.Unmarshal([]byte{3, 'a'}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("validated Unmarshal of a short payload returned %v", err)
	}
}

func TestByteSliceEdgeCases(t *testing.T) {
	if _, err := composite.SkipByteSlice(varnum.PositiveInt, []byte{3, 1}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("SkipByteSlice returned %v, want ErrTooSmallByteSlice", err)
	}

	valid := composite.NewValidByteSliceSer(bsopts.WithLenValidator(
		binco.ValidatorFn[int](func(int) error { return errTooLong })))
	if _, _, err := valid.Unmarshal([]byte{3, 1, 2, 3}); !errors.Is(err, errTooLong) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, errTooLong)
	}

	short := composite.NewValidByteSliceSer(bsopts.WithLenValidator(
		binco.ValidatorFn[int](func(int) error { return nil })))
	if _, _, err := short.Unmarshal([]byte{3, 1}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("validated Unmarshal of a short payload returned %v", err)
	}
}

func TestPtrUnmarshalTruncated(t *testing.T) {
	ser := composite.NewPtrSer(varnum.Int)
	if _, _, err := ser.Unmarshal([]byte{byte(binco.PtrNotNil)}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated pointer returned %v", err)
	}
}

func TestValidContainersRejectBadData(t *testing.T) {
	reject := binco.ValidatorFn[int](func(int) error { return errTooLong })

	validSlice := composite.NewValidSliceSer(varnum.Int,
		slopts.WithElemValidator[int](reject))
	if _, _, err := validSlice.Unmarshal([]byte{2, 2, 4}); !errors.Is(err, errTooLong) {
		t.Fatalf("element validator returned %v, want %v", err, errTooLong)
	}

	validMap := composite.NewValidMapSer(varnum.Int, varnum.Int,
		mapopts.WithKeyValidator[int, int](reject))
	if _, _, err := validMap.Unmarshal([]byte{1, 2, 4}); !errors.Is(err, errTooLong) {
		t.Fatalf("key validator returned %v, want %v", err, errTooLong)
	}
}
