package varnum_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
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
			t.Fatalf("%v: Unmarshal read %d bytes, want %d for %v", name, n2, size, v)
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("%v: round trip returned %v, want %v", name, got, v)
		}
		n3, err := s.Skip(bs)
		if err != nil {
			t.Fatalf("%v: Skip(%v): %v", name, bs, err)
		}
		if n3 != size {
			t.Fatalf("%v: Skip skipped %d bytes, want %d", name, n3, size)
		}

		if _, _, err = s.Unmarshal(append(bs, 0xFF)); err != nil {
			t.Fatalf("%v: Unmarshal with trailing bytes: %v", name, err)
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

func TestUnsigned(t *testing.T) {
	checkRoundTrip(t, "uint8", varnum.Uint8, []uint8{0, 1, 127, 128, 200, 255})
	checkRoundTrip(t, "uint16", varnum.Uint16, []uint16{0, 1, 127, 128, 300, 16383, 16384, math.MaxUint16})
	checkRoundTrip(t, "uint32", varnum.Uint32,
		[]uint32{0, 1, 127, 128, 1 << 28, 1<<28 - 1, math.MaxUint32})
	checkRoundTrip(t, "uint64", varnum.Uint64,
		[]uint64{0, 1, 127, 128, 1 << 56, 1<<56 - 1, math.MaxUint64})
	checkRoundTrip(t, "uint", varnum.Uint, []uint{0, 1, 1 << 28, math.MaxUint})
	checkRoundTrip(t, "byte", varnum.Byte, []byte{0, 1, 127, 128, 255})
}

func TestSigned(t *testing.T) {
	checkRoundTrip(t, "int8", varnum.Int8,
		[]int8{math.MinInt8, -64, -1, 0, 1, 63, math.MaxInt8})
	checkRoundTrip(t, "int16", varnum.Int16,
		[]int16{math.MinInt16, -1, 0, 1, math.MaxInt16})
	checkRoundTrip(t, "int32", varnum.Int32,
		[]int32{math.MinInt32, -1, 0, 1, math.MaxInt32})
	checkRoundTrip(t, "int64", varnum.Int64,
		[]int64{math.MinInt64, -1, 0, 1, math.MaxInt64})
	checkRoundTrip(t, "int", varnum.Int,
		[]int{math.MinInt, -1, 0, 1, math.MaxInt})
}

func TestPositiveSigned(t *testing.T) {
	checkRoundTrip(t, "positive int8", varnum.PositiveInt8, []int8{0, 1, 127})
	checkRoundTrip(t, "positive int16", varnum.PositiveInt16, []int16{0, 1, 300, math.MaxInt16})
	checkRoundTrip(t, "positive int32", varnum.PositiveInt32, []int32{0, 1, 300, math.MaxInt32})
	checkRoundTrip(t, "positive int64", varnum.PositiveInt64, []int64{0, 1, 300, math.MaxInt64})
	checkRoundTrip(t, "positive int", varnum.PositiveInt, []int{0, 1, 300, math.MaxInt})
}

func TestFloat(t *testing.T) {
	floats := []float64{0, 1.5, -1.5, 1e-9, math.MaxFloat64, math.SmallestNonzeroFloat64,
		math.Inf(1), math.Inf(-1)}
	checkRoundTrip(t, "float64", varnum.Float64, floats)
	checkRoundTrip(t, "float32", varnum.Float32,
		[]float32{0, 1.5, -1.5, 1e-9, math.MaxFloat32, math.SmallestNonzeroFloat32,
			float32(math.Inf(1)), float32(math.Inf(-1))})
}

func TestEncodings(t *testing.T) {
	checkEncoding(t, "uint64 0", varnum.Uint64, 0, []byte{0})
	checkEncoding(t, "uint64 127", varnum.Uint64, 127, []byte{127})
	checkEncoding(t, "uint64 128", varnum.Uint64, 128, []byte{128, 1})
	checkEncoding(t, "uint64 300", varnum.Uint64, 300, []byte{172, 2})
	checkEncoding(t, "uint8 255", varnum.Uint8, 255, []byte{255, 1})
	checkEncoding(t, "uint16 16384", varnum.Uint16, 16384, []byte{128, 128, 1})
	checkEncoding(t, "int64 0", varnum.Int64, 0, []byte{0})
	checkEncoding(t, "int64 -1", varnum.Int64, -1, []byte{1})
	checkEncoding(t, "int64 1", varnum.Int64, 1, []byte{2})
	checkEncoding(t, "int64 -2", varnum.Int64, -2, []byte{3})
	checkEncoding(t, "int64 63", varnum.Int64, 63, []byte{126})
	checkEncoding(t, "int64 -64", varnum.Int64, -64, []byte{127})
	checkEncoding(t, "int64 64", varnum.Int64, 64, []byte{128, 1})
	checkEncoding(t, "int64 300", varnum.Int64, 300, []byte{216, 4})
	checkEncoding(t, "positive int 300", varnum.PositiveInt, 300, []byte{172, 2})
}

func TestUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		fn   func() error
	}{
		{"uint8 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint8.Unmarshal([]byte{0x80, 0x02})
			return err
		}},
		{"uint16 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint16.Unmarshal([]byte{0xFF, 0xFF, 0x04})
			return err
		}},
		{"uint32 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint32.Unmarshal([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x10})
			return err
		}},
		{"uint64 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint64.Unmarshal([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x02})
			return err
		}},
		{"uint64 too many bytes", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint64.Unmarshal([]byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01})
			return err
		}},
		{"empty", binco.ErrTooSmallByteSlice, func() error {
			_, _, err := varnum.Uint64.Unmarshal(nil)
			return err
		}},
		{"truncated", binco.ErrTooSmallByteSlice, func() error {
			_, _, err := varnum.Uint64.Unmarshal([]byte{0x80})
			return err
		}},
		{"skip uint8 overflow", binco.ErrOverflow, func() error {
			_, err := varnum.Uint8.Skip([]byte{0x80, 0x02})
			return err
		}},
		{"skip empty", binco.ErrTooSmallByteSlice, func() error {
			_, err := varnum.Uint64.Skip(nil)
			return err
		}},
		{"skip truncated", binco.ErrTooSmallByteSlice, func() error {
			_, err := varnum.Uint64.Skip([]byte{0x80})
			return err
		}},
		{"int8 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Int8.Unmarshal([]byte{0x80, 0x02})
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != tc.err {
				t.Fatalf("got %v, want %v", err, tc.err)
			}
		})
	}
}

func TestMarshalPanicsOnSmallBuffer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Marshal did not panic on a too small buffer")
		}
	}()
	varnum.Uint64.Marshal(math.MaxUint64, make([]byte, 1))
}

func TestZigZag(t *testing.T) {
	for _, v := range []int64{0, 1, -1, 2, -2, math.MaxInt64, math.MinInt64} {
		if got := int64(varnum.DecodeZigZag(uint64(varnum.EncodeZigZag(v)))); got != v {
			t.Fatalf("ZigZag round trip of %d returned %d", v, got)
		}
	}
}
