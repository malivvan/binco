package composite_test

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	mapopts "github.com/malivvan/binco/config/map"
	slopts "github.com/malivvan/binco/config/slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

// maxInt64Varint is a length prefix that decodes to MaxInt64.
var maxInt64Varint = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}

// TestHugeLengthDoesNotPanic covers S1: adding a decoded MaxInt64 length to the
// offset used to wrap and turn the bounds check into a slice panic.
func TestHugeLengthDoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"String", func() error {
			_, _, err := composite.String.Unmarshal(maxInt64Varint)
			return err
		}},
		{"ByteSlice", func() error {
			_, _, err := composite.ByteSlice.Unmarshal(maxInt64Varint)
			return err
		}},
		{"Slice", func() error {
			_, _, err := composite.NewSliceSerializer(varnum.Int).Unmarshal(maxInt64Varint)
			return err
		}},
		{"Map", func() error {
			_, _, err := composite.NewMapSerializer(varnum.Int, varnum.Int).Unmarshal(maxInt64Varint)
			return err
		}},
		{"StringSkip", func() error {
			_, err := composite.String.Skip(maxInt64Varint)
			return err
		}},
		{"ByteSliceSkip", func() error {
			_, err := composite.ByteSlice.Skip(maxInt64Varint)
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v", r)
				}
			}()
			// On a 32-bit host a MaxInt64 length does not even fit an int, so
			// the length codec reports ErrOverflow. Both are safe rejections;
			// what matters is that nothing panics and nothing is allocated.
			err := tc.fn()
			if !errors.Is(err, binco.ErrTooSmallByteSlice) && !errors.Is(err, binco.ErrOverflow) {
				t.Fatalf("got %v, want ErrTooSmallByteSlice or ErrOverflow", err)
			}
		})
	}
}

// TestHugeLengthAllocatesLittle covers S2: a length prefix must not be turned
// into an allocation before the elements are read.
func TestHugeLengthAllocatesLittle(t *testing.T) {
	// Five bytes that announce 1<<28 elements.
	payload := []byte{0x80, 0x80, 0x80, 0x80, 0x01}

	before := allocBytes()
	for range 3 {
		if _, _, err := composite.NewSliceSerializer(varnum.Int).Unmarshal(payload); err == nil {
			t.Fatal("Unmarshal of a truncated slice returned no error")
		}
		if _, _, err := composite.NewMapSerializer(varnum.Int, varnum.Int).Unmarshal(payload); err == nil {
			t.Fatal("Unmarshal of a truncated map returned no error")
		}
		if _, _, err := composite.ByteSlice.Unmarshal([]byte{0x80, 0x80, 0x80, 0x80, 0x01}); err == nil {
			t.Fatal("Unmarshal of a truncated blob returned no error")
		}
	}
	if grew := allocBytes() - before; grew > 8<<20 {
		t.Fatalf("decoding three truncated values allocated %d bytes", grew)
	}
}

// allocBytes reports how many bytes the process has allocated so far.
func allocBytes() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.TotalAlloc
}

// TestValidatorsAreHonoured covers S9: the plain constructors used to discard
// the validators that the option type accepted.
func TestValidatorsAreHonoured(t *testing.T) {
	tooLong := binco.ValidatorFn[int](func(n int) error {
		if n > 2 {
			return errors.New("too long")
		}
		return nil
	})
	long := binco.Encode(composite.String, "abcd")

	if _, _, err := composite.NewStringSerializer(stropts.WithLenValidator(tooLong)).
		Unmarshal(long); err == nil {
		t.Fatal("NewStringSerializer ignored its length validator")
	}
	if _, _, err := composite.NewByteSliceSerializer(
		bslopts.WithLenValidator(tooLong)).Unmarshal([]byte{4, 1, 2, 3, 4}); err == nil {
		t.Fatal("NewByteSliceSerializer ignored its length validator")
	}
	if _, _, err := composite.NewSliceSerializer(varnum.Int,
		slopts.WithLenValidator[int](tooLong)).Unmarshal([]byte{4, 1, 2, 3, 4}); err == nil {
		t.Fatal("NewSliceSerializer ignored its length validator")
	}
	if _, _, err := composite.NewMapSerializer(varnum.Int, varnum.Int,
		mapopts.WithLenValidator[int, int](tooLong)).Unmarshal([]byte{4, 1, 2, 3, 4}); err == nil {
		t.Fatal("NewMapSerializer ignored its length validator")
	}
}

// TestZeroSizeElementsAreBounded covers S10: elements that consume no bytes must
// not make a huge declared length spin.
func TestZeroSizeElementsAreBounded(t *testing.T) {
	huge := binco.MaxZeroSizeElements + 1
	payload := binco.Encode(varnum.PositiveInt, huge)

	start := time.Now()
	if _, _, err := composite.NewSliceSerializer(zeroSerializer{}).Unmarshal(payload); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Unmarshal returned %v, want ErrTooLargeLength", err)
	}
	if _, err := composite.NewSliceSerializer(zeroSerializer{}).Skip(payload); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Skip returned %v, want ErrTooLargeLength", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("decoding a zero-size container took %v", elapsed)
	}
}

// zeroSerializer encodes a value that occupies no bytes at all.
type zeroSerializer struct{}

func (zeroSerializer) Marshal(v struct{}, bs []byte) (n int)      { return 0 }
func (zeroSerializer) Unmarshal(bs []byte) (struct{}, int, error) { return struct{}{}, 0, nil }
func (zeroSerializer) Size(v struct{}) (size int)                 { return 0 }
func (zeroSerializer) Skip(bs []byte) (n int, err error)          { return 0, nil }
