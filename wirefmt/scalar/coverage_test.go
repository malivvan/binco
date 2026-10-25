package scalar_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/scalar"
)

// TestFixedWidth checks the width every streaming scalar and timestamp codec
// reports for the bulk slice path.
func TestFixedWidth(t *testing.T) {
	codecs := []struct {
		name string
		s    binco.FixedWidth
		want int
	}{
		{"Byte", scalar.Byte, 1},
		{"Uint8", scalar.Uint8, 1},
		{"Uint16", scalar.Uint16, 2},
		{"Uint32", scalar.Uint32, 4},
		{"Uint64", scalar.Uint64, 8},
		{"Uint", scalar.Uint, 8},
		{"Int8", scalar.Int8, 1},
		{"Int16", scalar.Int16, 2},
		{"Int32", scalar.Int32, 4},
		{"Int64", scalar.Int64, 8},
		{"Int", scalar.Int, 8},
		{"Float32", scalar.Float32, 4},
		{"Float64", scalar.Float64, 8},
		{"TimeUnix", scalar.TimeUnix, 8},
		{"TimeUnixMilli", scalar.TimeUnixMilli, 8},
		{"TimeUnixMicro", scalar.TimeUnixMicro, 8},
		{"TimeUnixNano", scalar.TimeUnixNano, 8},
		{"TimeUnixUTC", scalar.TimeUnixUTC, 8},
		{"TimeUnixMilliUTC", scalar.TimeUnixMilliUTC, 8},
		{"TimeUnixMicroUTC", scalar.TimeUnixMicroUTC, 8},
		{"TimeUnixNanoUTC", scalar.TimeUnixNanoUTC, 8},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.FixedWidth(); got != tc.want {
				t.Fatalf("FixedWidth() = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestTimeUnmarshalRejectsAnEmptyStream checks the error every timestamp codec
// returns before converting anything.
func TestTimeUnmarshalRejectsAnEmptyStream(t *testing.T) {
	codecs := []struct {
		name string
		s    wirefmt.Serializer[time.Time]
	}{
		{"TimeUnix", scalar.TimeUnix},
		{"TimeUnixMilli", scalar.TimeUnixMilli},
		{"TimeUnixMicro", scalar.TimeUnixMicro},
		{"TimeUnixNano", scalar.TimeUnixNano},
		{"TimeUnixUTC", scalar.TimeUnixUTC},
		{"TimeUnixMilliUTC", scalar.TimeUnixMilliUTC},
		{"TimeUnixMicroUTC", scalar.TimeUnixMicroUTC},
		{"TimeUnixNanoUTC", scalar.TimeUnixNanoUTC},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := tc.s.Unmarshal(bytes.NewReader(nil)); !errors.Is(err, io.EOF) {
				t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
			}
		})
	}
}
