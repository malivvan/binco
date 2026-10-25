package scalar_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
)

func checkRoundTrip[T any](t *testing.T, name string, s binco.Serializer[T], values []T) {
	t.Helper()
	for _, v := range values {
		size := s.Size(v)
		bs := make([]byte, size)
		n := s.Marshal(v, bs)
		if n != size {
			t.Fatalf("%v: Marshal wrote %d bytes, Size returned %d for %v", name, n, size, v)
		}
		got, n2, err := s.Unmarshal(bs)
		if err != nil {
			t.Fatalf("%v: Unmarshal(%v): %v", name, bs, err)
		}
		if n2 != size {
			t.Fatalf("%v: Unmarshal read %d bytes, want %d", name, n2, size)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("%v: round trip returned %v, want %v", name, got, v)
		}
		n3, err := s.Skip(bs)
		if err != nil || n3 != size {
			t.Fatalf("%v: Skip returned (%d, %v), want (%d, nil)", name, n3, err, size)
		}
	}
}

func checkEncoding[T any](t *testing.T, name string, s binco.Serializer[T], v T, want []byte) {
	t.Helper()
	got := make([]byte, s.Size(v))
	s.Marshal(v, got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v: Marshal(%v) = %v, want %v", name, v, got, want)
	}
}

func TestNumbers(t *testing.T) {
	checkRoundTrip(t, "byte", scalar.Byte, []byte{0, 1, 127, 255})
	checkRoundTrip(t, "uint8", scalar.Uint8, []uint8{0, 1, 255})
	checkRoundTrip(t, "uint16", scalar.Uint16, []uint16{0, 1, math.MaxUint16})
	checkRoundTrip(t, "uint32", scalar.Uint32, []uint32{0, 1, math.MaxUint32})
	checkRoundTrip(t, "uint64", scalar.Uint64, []uint64{0, 1, math.MaxUint64})
	checkRoundTrip(t, "uint", scalar.Uint, []uint{0, 1, math.MaxUint})
	checkRoundTrip(t, "int8", scalar.Int8, []int8{math.MinInt8, -1, 0, math.MaxInt8})
	checkRoundTrip(t, "int16", scalar.Int16, []int16{math.MinInt16, -1, 0, math.MaxInt16})
	checkRoundTrip(t, "int32", scalar.Int32, []int32{math.MinInt32, -1, 0, math.MaxInt32})
	checkRoundTrip(t, "int64", scalar.Int64, []int64{math.MinInt64, -1, 0, math.MaxInt64})
	checkRoundTrip(t, "int", scalar.Int, []int{math.MinInt, -1, 0, math.MaxInt})
}

func TestFloats(t *testing.T) {
	checkRoundTrip(t, "float64", scalar.Float64,
		[]float64{0, 1.5, -1.5, math.MaxFloat64, math.SmallestNonzeroFloat64, math.Inf(1)})
	checkRoundTrip(t, "float32", scalar.Float32,
		[]float32{0, 1.5, -1.5, math.MaxFloat32, math.SmallestNonzeroFloat32, float32(math.Inf(-1))})
}

func TestLittleEndianLayout(t *testing.T) {
	checkEncoding(t, "uint64 40", scalar.Uint64, 40,
		[]byte{40, 0, 0, 0, 0, 0, 0, 0})
	checkEncoding(t, "uint32 1", scalar.Uint32, 1, []byte{1, 0, 0, 0})
	checkEncoding(t, "uint16 0x1234", scalar.Uint16, 0x1234, []byte{0x34, 0x12})
	checkEncoding(t, "byte 7", scalar.Byte, 7, []byte{7})
	checkEncoding(t, "int64 -1", scalar.Int64, -1,
		[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	checkEncoding(t, "float64 1.0", scalar.Float64, 1.0,
		[]byte{0, 0, 0, 0, 0, 0, 0xF0, 0x3F})
}

func checkTimeRoundTrip(t *testing.T, name string, s binco.Serializer[time.Time],
	values []time.Time,
) {
	t.Helper()
	for _, v := range values {
		size := s.Size(v)
		bs := make([]byte, size)
		if n := s.Marshal(v, bs); n != size {
			t.Fatalf("%v: Marshal wrote %d bytes, Size returned %d", name, n, size)
		}
		got, n, err := s.Unmarshal(bs)
		if err != nil {
			t.Fatalf("%v: Unmarshal: %v", name, err)
		}
		if n != size {
			t.Fatalf("%v: Unmarshal read %d bytes, want %d", name, n, size)
		}
		if !got.Equal(v) {
			t.Fatalf("%v: round trip returned %v, want %v", name, got, v)
		}
		if n, err = s.Skip(bs); err != nil || n != size {
			t.Fatalf("%v: Skip returned (%d, %v), want (%d, nil)", name, n, err, size)
		}
	}
}

func TestTime(t *testing.T) {
	sec := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	milli := sec.Add(456 * time.Millisecond)
	micro := sec.Add(456789 * time.Microsecond)
	nano := sec.Add(456789123 * time.Nanosecond)

	checkTimeRoundTrip(t, "time unix", scalar.TimeUnix, []time.Time{sec})
	checkTimeRoundTrip(t, "time unix milli", scalar.TimeUnixMilli, []time.Time{milli})
	checkTimeRoundTrip(t, "time unix micro", scalar.TimeUnixMicro, []time.Time{micro})
	checkTimeRoundTrip(t, "time unix nano", scalar.TimeUnixNano, []time.Time{nano})
	checkTimeRoundTrip(t, "time unix utc", scalar.TimeUnixUTC, []time.Time{sec})
	checkTimeRoundTrip(t, "time unix milli utc", scalar.TimeUnixMilliUTC, []time.Time{milli})
	checkTimeRoundTrip(t, "time unix micro utc", scalar.TimeUnixMicroUTC, []time.Time{micro})
	checkTimeRoundTrip(t, "time unix nano utc", scalar.TimeUnixNanoUTC, []time.Time{nano})

	bs := binco.Encode(scalar.TimeUnix, sec.Add(500*time.Millisecond))
	got, err := binco.Decode(scalar.TimeUnix, bs)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !got.Equal(sec) {
		t.Fatalf("TimeUnix returned %v, want %v", got, sec)
	}

	loc := time.FixedZone("custom", 3*60*60)
	v := time.Date(2024, 5, 6, 7, 8, 9, 0, loc)
	got, err = binco.Decode(scalar.TimeUnixUTC, binco.Encode(scalar.TimeUnixUTC, v))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("TimeUnixUTC returned a value in %v, want UTC", got.Location())
	}
	if !got.Equal(v) {
		t.Fatalf("TimeUnixUTC returned %v, want %v", got, v)
	}
}

func TestErrors(t *testing.T) {
	if _, _, err := scalar.Uint64.Unmarshal(make([]byte, 7)); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a short buffer returned %v", err)
	}
	if _, err := scalar.Uint64.Skip(make([]byte, 7)); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Skip of a short buffer returned %v", err)
	}
	if _, _, err := scalar.Int32.Unmarshal(make([]byte, 3)); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a short buffer returned %v", err)
	}
	if _, _, err := scalar.Int16.Unmarshal(make([]byte, 1)); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a short buffer returned %v", err)
	}
	if _, _, err := scalar.Int8.Unmarshal(nil); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of an empty buffer returned %v", err)
	}
}

func TestMarshalPanicsOnSmallBuffer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Marshal did not panic on a too small buffer")
		}
	}()
	scalar.Uint64.Marshal(1, make([]byte, 7))
}
