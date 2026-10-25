package zerocopy_test

import (
	"errors"
	"testing"
	"time"

	"github.com/malivvan/binco"
	arropts "github.com/malivvan/binco/config/array"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/zerocopy"
)

// TestFixedWidth checks the bulk-path width every scalar and timestamp codec
// reports. It must agree with Size, or a slice copy would use the wrong stride.
func TestFixedWidth(t *testing.T) {
	codecs := []struct {
		name string
		s    binco.FixedWidth
		want int
	}{
		{"Byte", zerocopy.Byte, 1},
		{"Uint8", zerocopy.Uint8, 1},
		{"Uint16", zerocopy.Uint16, 2},
		{"Uint32", zerocopy.Uint32, 4},
		{"Uint64", zerocopy.Uint64, 8},
		{"Uint", zerocopy.Uint, 8},
		{"Int8", zerocopy.Int8, 1},
		{"Int16", zerocopy.Int16, 2},
		{"Int32", zerocopy.Int32, 4},
		{"Int64", zerocopy.Int64, 8},
		{"Int", zerocopy.Int, 8},
		{"TimeUnix", zerocopy.TimeUnix, 8},
		{"TimeUnixMilli", zerocopy.TimeUnixMilli, 8},
		{"TimeUnixMicro", zerocopy.TimeUnixMicro, 8},
		{"TimeUnixNano", zerocopy.TimeUnixNano, 8},
		{"TimeUnixUTC", zerocopy.TimeUnixUTC, 8},
		{"TimeUnixMilliUTC", zerocopy.TimeUnixMilliUTC, 8},
		{"TimeUnixMicroUTC", zerocopy.TimeUnixMicroUTC, 8},
		{"TimeUnixNanoUTC", zerocopy.TimeUnixNanoUTC, 8},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.FixedWidth(); got != tc.want {
				t.Fatalf("FixedWidth() = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestTimeUnmarshalRejectsTruncatedInput checks the error return of the timestamp
// codecs before they convert anything.
func TestTimeUnmarshalRejectsTruncatedInput(t *testing.T) {
	codecs := []struct {
		name string
		s    binco.Serializer[time.Time]
	}{
		{"TimeUnix", zerocopy.TimeUnix},
		{"TimeUnixMilli", zerocopy.TimeUnixMilli},
		{"TimeUnixMicro", zerocopy.TimeUnixMicro},
		{"TimeUnixNano", zerocopy.TimeUnixNano},
		{"TimeUnixUTC", zerocopy.TimeUnixUTC},
		{"TimeUnixMilliUTC", zerocopy.TimeUnixMilliUTC},
		{"TimeUnixMicroUTC", zerocopy.TimeUnixMicroUTC},
		{"TimeUnixNanoUTC", zerocopy.TimeUnixNanoUTC},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := tc.s.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
				t.Fatalf("Unmarshal(nil) = %v, want ErrTruncated", err)
			}
		})
	}
}

// TestCheckedTextReportsATruncatedLength checks the validating string and blob
// codecs' error return for an empty input.
func TestCheckedTextReportsATruncatedLength(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	text := zerocopy.NewStringSerializer(stropts.WithLenValidator(never))
	blob := zerocopy.NewByteSliceSerializer(bslopts.WithLenValidator(never))

	if _, _, err := text.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("checked string Unmarshal(nil) = %v, want ErrTruncated", err)
	}
	if _, _, err := blob.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("checked blob Unmarshal(nil) = %v, want ErrTruncated", err)
	}
}

// TestArrayUnmarshalErrors covers the array decoder's early returns: a missing
// length, a negative length, a mismatched length and the empty array.
func TestArrayUnmarshalErrors(t *testing.T) {
	arrays := zerocopy.NewArraySerializer[[2]int, int](varnum.Int)
	if _, _, err := arrays.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Unmarshal(nil) = %v, want ErrTruncated", err)
	}
	if _, _, err := arrays.Unmarshal(binco.Encode(varnum.PositiveInt, 1)); !errors.Is(err, binco.ErrWrongLength) {
		t.Fatalf("Unmarshal of a shorter array = %v, want ErrWrongLength", err)
	}

	// A signed length codec can produce a negative length; the array codec must
	// reject it before reading anything.
	signed := zerocopy.NewArraySerializer[[2]int, int](varnum.Int,
		arropts.WithLenSerializer[int](varnum.Int))
	if _, _, err := signed.Unmarshal(binco.Encode(varnum.Int, -1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}

	// The empty array accepts exactly the zero length, and its built-in length
	// validator rejects anything else.
	empty := zerocopy.NewArraySerializer[[0]int, int](varnum.Int)
	if _, n, err := empty.Unmarshal(binco.Encode(varnum.PositiveInt, 0)); err != nil || n != 1 {
		t.Fatalf("Unmarshal of an empty array = (%d, %v)", n, err)
	}
	if _, _, err := empty.Unmarshal(binco.Encode(varnum.PositiveInt, 1)); !errors.Is(err, binco.ErrWrongLength) {
		t.Fatalf("Unmarshal of a non-empty payload into [0]int = %v, want ErrWrongLength", err)
	}
}
