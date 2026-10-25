package zerocopy_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	arropts "github.com/malivvan/binco/wirefmt/config/array"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
	"github.com/malivvan/binco/wirefmt/zerocopy"
)

// TestFixedWidth checks the bulk-path width every streaming scalar and timestamp
// codec reports.
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

// TestArraySerializerRejectsMismatchedTypes checks the construction-time type
// check: a non-array and an array of the wrong element type both panic.
func TestArraySerializerRejectsMismatchedTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"not an array", func() { zerocopy.NewArraySerializer[int, int](varnum.Int) }},
		{"wrong element", func() { zerocopy.NewArraySerializer[[2]int, string](zerocopy.String) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, binco.ErrWrongArrayType) {
					t.Fatalf("panic value %v, want ErrWrongArrayType", err)
				}
			}()
			tc.call()
		})
	}
}

// TestArrayUnmarshalErrors covers the array decoder's early returns.
func TestArrayUnmarshalErrors(t *testing.T) {
	arrays := zerocopy.NewArraySerializer[[2]int, int](varnum.Int)
	if _, _, err := arrays.Unmarshal(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
	}

	// A signed length codec can produce a negative length.
	signed := zerocopy.NewArraySerializer[[2]int, int](varnum.Int,
		arropts.WithLenSerializer[int](varnum.Int))
	if _, _, err := signed.Unmarshal(bytes.NewReader([]byte{1})); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}

	// The built-in validator rejects a length that does not match the type.
	if _, _, err := arrays.Unmarshal(bytes.NewReader([]byte{1})); !errors.Is(err, binco.ErrWrongLength) {
		t.Fatalf("Unmarshal of a shorter array = %v, want ErrWrongLength", err)
	}

	// A matching length with no element bytes reports the element error.
	if _, _, err := arrays.Unmarshal(bytes.NewReader([]byte{2})); err == nil {
		t.Fatal("Unmarshal of a length without elements returned nil")
	}
}

// TestCheckedTextReportsAnEmptyStream checks the validating string codec's error
// return for an empty input and for a negative length.
//
// The streaming tree has no zero-copy byte slice codec: a stream cannot hand out
// a view of bytes it has not read yet.
func TestCheckedTextReportsAnEmptyStream(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	text := zerocopy.NewStringSerializer(stropts.WithLenValidator(never))
	if _, _, err := text.Unmarshal(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
		t.Fatalf("checked string Unmarshal of an empty stream = %v, want io.EOF", err)
	}

	negativeText := zerocopy.NewStringSerializer(stropts.WithLenSerializer(varnum.Int),
		stropts.WithLenValidator(never))
	if _, _, err := negativeText.Unmarshal(bytes.NewReader([]byte{1})); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("checked string Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}

	// A body that never arrives is unexpected EOF.
	if _, _, err := text.Unmarshal(bytes.NewReader([]byte{5})); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("checked string Unmarshal of a truncated body = %v, want io.ErrUnexpectedEOF", err)
	}
}
