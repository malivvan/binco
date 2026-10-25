package zerocpy_test

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	arropts "github.com/malivvan/binco/config/array"
	bsopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/zerocpy"
)

var testTime = time.Date(2024, 5, 6, 7, 8, 9, 123456789, time.UTC)

func TestParityWithRaw(t *testing.T) {
	check := func(name string, nat []byte, ref []byte) {
		t.Helper()
		if !reflect.DeepEqual(nat, ref) {
			t.Fatalf("%v: native %v != reference %v", name, nat, ref)
		}
	}

	check("bool", binco.Encode(zerocpy.Bool, true), binco.Encode(composite.Bool, true))
	check("byte", binco.Encode(zerocpy.Byte, 255), binco.Encode(scalar.Byte, 255))
	check("uint8", binco.Encode(zerocpy.Uint8, 200), binco.Encode(scalar.Uint8, 200))
	check("uint16", binco.Encode(zerocpy.Uint16, math.MaxUint16), binco.Encode(scalar.Uint16, math.MaxUint16))
	check("uint32", binco.Encode(zerocpy.Uint32, math.MaxUint32), binco.Encode(scalar.Uint32, math.MaxUint32))
	check("uint64", binco.Encode(zerocpy.Uint64, math.MaxUint64), binco.Encode(scalar.Uint64, math.MaxUint64))
	check("uint", binco.Encode(zerocpy.Uint, 7), binco.Encode(scalar.Uint, 7))
	check("int8", binco.Encode(zerocpy.Int8, -128), binco.Encode(scalar.Int8, -128))
	check("int16", binco.Encode(zerocpy.Int16, -1), binco.Encode(scalar.Int16, -1))
	check("int32", binco.Encode(zerocpy.Int32, -1), binco.Encode(scalar.Int32, -1))
	check("int64", binco.Encode(zerocpy.Int64, math.MinInt64), binco.Encode(scalar.Int64, math.MinInt64))
	check("int", binco.Encode(zerocpy.Int, -7), binco.Encode(scalar.Int, -7))
	check("float64", binco.Encode(zerocpy.Float64, 1.5), binco.Encode(scalar.Float64, 1.5))
	check("float32", binco.Encode(zerocpy.Float32, 1.5), binco.Encode(scalar.Float32, 1.5))
	check("string", binco.Encode(zerocpy.String, "hello"), binco.Encode(composite.String, "hello"))
	check("byte slice", binco.Encode(zerocpy.ByteSlice, []byte{1, 2}), binco.Encode(composite.ByteSlice, []byte{1, 2}))

	for _, tc := range []struct {
		name string
		nat  []byte
		ref  []byte
	}{
		{"time", binco.Encode(zerocpy.TimeUnixUTC, testTime), binco.Encode(scalar.TimeUnixUTC, testTime)},
		{"time milli", binco.Encode(zerocpy.TimeUnixMilliUTC, testTime), binco.Encode(scalar.TimeUnixMilliUTC, testTime)},
		{"time micro", binco.Encode(zerocpy.TimeUnixMicroUTC, testTime), binco.Encode(scalar.TimeUnixMicroUTC, testTime)},
		{"time nano", binco.Encode(zerocpy.TimeUnixNanoUTC, testTime), binco.Encode(scalar.TimeUnixNanoUTC, testTime)},
	} {
		check(tc.name, tc.nat, tc.ref)
	}
}

func TestRoundTrip(t *testing.T) {
	if got, err := binco.Decode(zerocpy.Uint64, binco.Encode(zerocpy.Uint64, math.MaxUint64)); err != nil || got != math.MaxUint64 {
		t.Fatalf("uint64: (%v, %v)", got, err)
	}
	if got, err := binco.Decode(zerocpy.Int64, binco.Encode(zerocpy.Int64, math.MinInt64)); err != nil || got != math.MinInt64 {
		t.Fatalf("int64: (%v, %v)", got, err)
	}
	if got, err := binco.Decode(zerocpy.Float64, binco.Encode(zerocpy.Float64, -2.5)); err != nil || got != -2.5 {
		t.Fatalf("float64: (%v, %v)", got, err)
	}
	if got, err := binco.Decode(zerocpy.Bool, binco.Encode(zerocpy.Bool, false)); err != nil || got != false {
		t.Fatalf("bool: (%v, %v)", got, err)
	}
	if got, err := binco.Decode(zerocpy.String, binco.Encode(zerocpy.String, "héllo")); err != nil || got != "héllo" {
		t.Fatalf("string: (%v, %v)", got, err)
	}
}

func TestZeroCopy(t *testing.T) {
	bs := binco.Encode(zerocpy.String, "abc")
	s, _, err := zerocpy.String.Unmarshal(bs)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got, want := s, "abc"; got != want {
		t.Fatalf("Unmarshal returned %q, want %q", got, want)
	}

	payload := binco.Encode(zerocpy.ByteSlice, []byte{1, 2, 3})
	b, _, err := zerocpy.ByteSlice.Unmarshal(payload)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	payload[1] = 9
	if b[0] != 9 {
		t.Fatal("decoded byte slice does not share memory with the payload")
	}
	if got, want := zerocpy.ByteSlice.Size([]byte{1, 2, 3}), 4; got != want {
		t.Fatalf("Size returned %d, want %d", got, want)
	}
}

func TestBoolErrors(t *testing.T) {
	if _, _, err := zerocpy.Bool.Unmarshal([]byte{2}); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal(2) returned %v", err)
	}
	if _, err := zerocpy.Bool.Skip(nil); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Skip(nil) returned %v", err)
	}
}

func TestArray(t *testing.T) {
	serializer := zerocpy.NewArraySerializer[[3]int, int](varnum.Int)
	v := [3]int{1, 2, 3}
	bs := binco.Encode(serializer, v)
	if want := []byte{3, 2, 4, 6}; !reflect.DeepEqual(bs, want) {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}
	if len(bs) != serializer.Size(v) {
		t.Fatalf("Size = %d, Marshal wrote %d", serializer.Size(v), len(bs))
	}
	got, n, err := serializer.Unmarshal(bs)
	if err != nil || n != len(bs) || got != v {
		t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
	}
	if n, err = serializer.Skip(bs); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	sliceSerializer := composite.NewSliceSerializer(varnum.Int)
	short := binco.Encode(sliceSerializer, []int{1, 2})
	if _, _, err = serializer.Unmarshal(short); err != binco.ErrWrongLength {
		t.Fatalf("Unmarshal of a shorter array returned %v, want ErrWrongLength", err)
	}

	long := binco.Encode(sliceSerializer, []int{1, 2, 3, 4})
	if _, _, err = serializer.Unmarshal(long); err != binco.ErrTooLargeLength {
		t.Fatalf("Unmarshal of a longer array returned %v, want ErrTooLargeLength", err)
	}

	zero := zerocpy.NewArraySerializer[[0]int, int](varnum.Int)
	zbs := binco.Encode(zero, [0]int{})
	if !reflect.DeepEqual(zbs, []byte{0}) {
		t.Fatalf("Marshal of a zero-length array = %v", zbs)
	}
	if _, _, err = zero.Unmarshal(zbs); err != nil {
		t.Fatalf("Unmarshal of a zero-length array: %v", err)
	}
}

func TestValidArray(t *testing.T) {
	wantErr := errors.New("elem")
	serializer := zerocpy.NewValidArraySerializer[[2]int, int](varnum.Int,
		arropts.WithElemValidator[int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := serializer.Unmarshal(binco.Encode(serializer, [2]int{-1, 1})); !errors.Is(err, wantErr) {
		t.Fatalf("element validator returned %v", err)
	}
	if got, err := binco.Decode(serializer, binco.Encode(serializer, [2]int{1, 2})); err != nil || got != [2]int{1, 2} {
		t.Fatalf("Unmarshal = (%v, %v)", got, err)
	}

	rawSerializer := zerocpy.NewValidArraySerializer[[2]int, int](varnum.Int,
		arropts.WithLenSerializer[int](scalar.Int))
	bs := binco.Encode(rawSerializer, [2]int{1, 2})
	if len(bs) != 10 || bs[0] != 2 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
}

type negativeLenSerializer struct{}

func (negativeLenSerializer) Marshal(v int, bs []byte) (n int)      { return 0 }
func (negativeLenSerializer) Unmarshal(bs []byte) (int, int, error) { return -1, 1, nil }
func (negativeLenSerializer) Size(v int) (size int)                 { return 1 }
func (negativeLenSerializer) Skip(bs []byte) (n int, err error)     { return 1, nil }

func checkScalar[T comparable](t *testing.T, name string, s, ref binco.Serializer[T], values []T) {
	t.Helper()
	for _, v := range values {
		bs := binco.Encode(s, v)
		if want := binco.Encode(ref, v); !bytes.Equal(bs, want) {
			t.Fatalf("%v: native %v != raw %v", name, bs, want)
		}
		if len(bs) != s.Size(v) {
			t.Fatalf("%v: Size = %d, Marshal wrote %d", name, s.Size(v), len(bs))
		}
		got, n, err := s.Unmarshal(bs)
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("%v: Unmarshal(%v) = (%v, %d, %v)", name, bs, got, n, err)
		}
		if n, err = s.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("%v: Skip(%v) = (%d, %v)", name, bs, n, err)
		}
	}
}

func TestAllScalars(t *testing.T) {
	checkScalar(t, "byte", zerocpy.Byte, scalar.Byte, []byte{0, 1, 127, 255})
	checkScalar(t, "uint8", zerocpy.Uint8, scalar.Uint8, []uint8{0, 1, 200, 255})
	checkScalar(t, "uint16", zerocpy.Uint16, scalar.Uint16,
		[]uint16{0, 1, 300, math.MaxUint16})
	checkScalar(t, "uint32", zerocpy.Uint32, scalar.Uint32,
		[]uint32{0, 1, 1 << 20, math.MaxUint32})
	checkScalar(t, "uint64", zerocpy.Uint64, scalar.Uint64,
		[]uint64{0, 1, 1 << 40, math.MaxUint64})
	checkScalar(t, "uint", zerocpy.Uint, scalar.Uint, []uint{0, 1, 1 << 20, math.MaxUint})
	checkScalar(t, "int8", zerocpy.Int8, scalar.Int8,
		[]int8{math.MinInt8, -1, 0, 1, math.MaxInt8})
	checkScalar(t, "int16", zerocpy.Int16, scalar.Int16,
		[]int16{math.MinInt16, -1, 0, math.MaxInt16})
	checkScalar(t, "int32", zerocpy.Int32, scalar.Int32,
		[]int32{math.MinInt32, -1, 0, math.MaxInt32})
	checkScalar(t, "int64", zerocpy.Int64, scalar.Int64,
		[]int64{math.MinInt64, -1, 0, math.MaxInt64})
	checkScalar(t, "int", zerocpy.Int, scalar.Int, []int{math.MinInt, -1, 0, math.MaxInt})
	checkScalar(t, "float32", zerocpy.Float32, scalar.Float32,
		[]float32{0, 1.5, -1.5, math.MaxFloat32, float32(math.Inf(-1))})
	checkScalar(t, "float64", zerocpy.Float64, scalar.Float64,
		[]float64{0, 1.5, -1.5, math.MaxFloat64, math.SmallestNonzeroFloat64,
			math.Inf(1)})
}

func TestStringSerializers(t *testing.T) {

	for _, v := range []string{"", "a", "hello world", "héllo"} {
		bs := binco.Encode(zerocpy.String, v)
		if want := binco.Encode(composite.String, v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%q) = %v, want %v", v, bs, want)
		}
		if len(bs) != zerocpy.String.Size(v) {
			t.Fatalf("Size(%q) = %d, Marshal wrote %d", v, zerocpy.String.Size(v), len(bs))
		}
		got, n, err := zerocpy.String.Unmarshal(bs)
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%q, %d, %v)", bs, got, n, err)
		}
		if n, err = zerocpy.String.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}

	rawLen := zerocpy.NewStringSerializer(stropts.WithLenSerializer(scalar.Int))
	bs := binco.Encode(rawLen, "abc")
	if len(bs) != 11 || bs[0] != 3 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
	if got, err := binco.Decode(rawLen, bs); err != nil || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %v)", got, err)
	}

	wantErr := errors.New("too long")
	valid := zerocpy.NewValidStringSerializer(stropts.WithLenValidator(binco.ValidatorFn[int](
		func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(binco.Encode(valid, "abc")); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, err := binco.Decode(valid, binco.Encode(valid, "ab")); err != nil || got != "ab" {
		t.Fatalf("validated Unmarshal = (%q, %v)", got, err)
	}

	if _, _, err := zerocpy.String.Unmarshal([]byte{5, 'a'}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}
	if _, err := zerocpy.String.Skip([]byte{5, 'a'}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Skip of a truncated string returned %v", err)
	}
	neg := zerocpy.NewStringSerializer(stropts.WithLenSerializer(negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
	if _, err := neg.Skip([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Skip returned %v, want ErrNegativeLength", err)
	}
}

func TestByteSliceSerializers(t *testing.T) {
	for _, v := range [][]byte{nil, {}, {1, 2, 3}} {
		bs := binco.Encode(zerocpy.ByteSlice, v)
		if want := binco.Encode(composite.ByteSlice, v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%v) = %v, want %v", v, bs, want)
		}
		got, n, err := zerocpy.ByteSlice.Unmarshal(bs)
		if err != nil || n != len(bs) || !bytes.Equal(got, v) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = zerocpy.ByteSlice.Skip(bs); err != nil || n != len(bs) {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}

	rawLen := zerocpy.NewByteSliceSerializer(bsopts.WithLenSerializer(scalar.Int))
	bs := binco.Encode(rawLen, []byte{1, 2})
	if len(bs) != 10 || bs[0] != 2 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}

	wantErr := errors.New("too long")
	valid := zerocpy.NewValidByteSliceSerializer(bsopts.WithLenValidator(binco.ValidatorFn[int](
		func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(binco.Encode(valid, []byte{1, 2, 3})); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, err := binco.Decode(valid, binco.Encode(valid, []byte{1})); err != nil || !bytes.Equal(got, []byte{1}) {
		t.Fatalf("validated Unmarshal = (%v, %v)", got, err)
	}
	if got, err := binco.Decode(valid, binco.Encode(valid, nil)); err != nil || len(got) != 0 {
		t.Fatalf("validated Unmarshal of an empty slice = (%v, %v)", got, err)
	}

	if _, _, err := zerocpy.ByteSlice.Unmarshal([]byte{3, 1}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated slice returned %v", err)
	}
	if _, err := zerocpy.ByteSlice.Skip([]byte{3, 1}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Skip of a truncated slice returned %v", err)
	}
	neg := zerocpy.NewByteSliceSerializer(bsopts.WithLenSerializer(negativeLenSerializer{}))
	if _, _, err := neg.Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

func TestTimeAllUnits(t *testing.T) {
	sec := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	serializers := []struct {
		name   string
		native binco.Serializer[time.Time]
		ref    binco.Serializer[time.Time]
	}{
		{"unix", zerocpy.TimeUnix, scalar.TimeUnix},
		{"unix milli", zerocpy.TimeUnixMilli, scalar.TimeUnixMilli},
		{"unix micro", zerocpy.TimeUnixMicro, scalar.TimeUnixMicro},
		{"unix nano", zerocpy.TimeUnixNano, scalar.TimeUnixNano},
		{"unix utc", zerocpy.TimeUnixUTC, scalar.TimeUnixUTC},
		{"unix milli utc", zerocpy.TimeUnixMilliUTC, scalar.TimeUnixMilliUTC},
		{"unix micro utc", zerocpy.TimeUnixMicroUTC, scalar.TimeUnixMicroUTC},
		{"unix nano utc", zerocpy.TimeUnixNanoUTC, scalar.TimeUnixNanoUTC},
	}
	values := []time.Time{
		sec,
		sec.Add(456 * time.Millisecond),
		sec.Add(456789 * time.Microsecond),
		sec.Add(456789123 * time.Nanosecond),
	}
	for _, serializer := range serializers {
		for _, v := range values {
			bs := binco.Encode(serializer.native, v)
			if want := binco.Encode(serializer.ref, v); !bytes.Equal(bs, want) {
				t.Fatalf("%v: native %v != raw %v", serializer.name, bs, want)
			}
			if len(bs) != serializer.native.Size(v) {
				t.Fatalf("%v: Size = %d, Marshal wrote %d", serializer.name, serializer.native.Size(v), len(bs))
			}
			got, n, err := serializer.native.Unmarshal(bs)
			if err != nil || n != len(bs) {
				t.Fatalf("%v: Unmarshal = (%v, %d, %v)", serializer.name, got, n, err)
			}

			if serializer.name == "unix nano" || serializer.name == "unix nano utc" {
				if !got.Equal(v) {
					t.Fatalf("%v: round trip returned %v, want %v", serializer.name, got, v)
				}
			} else if got.Unix() != v.Unix() {
				t.Fatalf("%v: round trip returned %v, want %v", serializer.name, got, v)
			}
			if n, err = serializer.native.Skip(bs); err != nil || n != len(bs) {
				t.Fatalf("%v: Skip = (%d, %v)", serializer.name, n, err)
			}
		}
	}

	v := time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("custom", 3*60*60))
	got, err := binco.Decode(zerocpy.TimeUnixUTC, binco.Encode(zerocpy.TimeUnixUTC, v))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Location() != time.UTC || !got.Equal(v) {
		t.Fatalf("TimeUnixUTC returned %v in %v", got, got.Location())
	}
}

func TestScalarErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"uint64 unmarshal", func() error {
			_, _, err := zerocpy.Uint64.Unmarshal(make([]byte, 7))
			return err
		}},
		{"uint64 skip", func() error {
			_, err := zerocpy.Uint64.Skip(make([]byte, 7))
			return err
		}},
		{"uint32 unmarshal", func() error {
			_, _, err := zerocpy.Uint32.Unmarshal(make([]byte, 3))
			return err
		}},
		{"uint16 unmarshal", func() error {
			_, _, err := zerocpy.Uint16.Unmarshal(make([]byte, 1))
			return err
		}},
		{"int64 unmarshal", func() error {
			_, _, err := zerocpy.Int64.Unmarshal(nil)
			return err
		}},
		{"int32 skip", func() error {
			_, err := zerocpy.Int32.Skip(make([]byte, 3))
			return err
		}},
		{"int16 skip", func() error {
			_, err := zerocpy.Int16.Skip(nil)
			return err
		}},
		{"int8 unmarshal", func() error {
			_, _, err := zerocpy.Int8.Unmarshal(nil)
			return err
		}},
		{"byte unmarshal", func() error {
			_, _, err := zerocpy.Byte.Unmarshal(nil)
			return err
		}},
		{"byte skip", func() error {
			_, err := zerocpy.Byte.Skip(nil)
			return err
		}},
		{"float64 unmarshal", func() error {
			_, _, err := zerocpy.Float64.Unmarshal(make([]byte, 7))
			return err
		}},
		{"float32 skip", func() error {
			_, err := zerocpy.Float32.Skip(make([]byte, 3))
			return err
		}},
		{"time unmarshal", func() error {
			_, _, err := zerocpy.TimeUnix.Unmarshal(make([]byte, 7))
			return err
		}},
		{"array unmarshal", func() error {
			serializer := zerocpy.NewArraySerializer[[2]int, int](varnum.Int)
			_, _, err := serializer.Unmarshal([]byte{2, 2})
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); !errors.Is(err, binco.ErrTooSmallByteSlice) {
				t.Fatalf("got %v, want ErrTooSmallByteSlice", err)
			}
		})
	}
}

func TestValidStringErrors(t *testing.T) {
	valid := zerocpy.NewValidStringSerializer()

	got, n, err := valid.Unmarshal([]byte{0})
	if err != nil || n != 1 || got != "" {
		t.Fatalf("Unmarshal of an empty string = (%q, %d, %v)", got, n, err)
	}

	if _, _, err = valid.Unmarshal([]byte{5, 'a'}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}

	if _, _, err = zerocpy.NewValidStringSerializer(
		stropts.WithLenSerializer(negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

func TestValidByteSliceErrors(t *testing.T) {
	valid := zerocpy.NewValidByteSliceSerializer()

	got, n, err := valid.Unmarshal([]byte{0})
	if err != nil || n != 1 || len(got) != 0 {
		t.Fatalf("Unmarshal of an empty slice = (%v, %d, %v)", got, n, err)
	}
	if _, _, err = valid.Unmarshal([]byte{3, 1}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated slice returned %v", err)
	}
	if _, _, err = zerocpy.NewValidByteSliceSerializer(
		bsopts.WithLenSerializer(negativeLenSerializer{})).Unmarshal([]byte{0}); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

func TestMoreScalarErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"bool empty", func() error {
			_, _, err := zerocpy.Bool.Unmarshal(nil)
			return err
		}},
		{"float32 unmarshal", func() error {
			_, _, err := zerocpy.Float32.Unmarshal(make([]byte, 3))
			return err
		}},
		{"time milli", func() error {
			_, _, err := zerocpy.TimeUnixMilli.Unmarshal(make([]byte, 7))
			return err
		}},
		{"time micro", func() error {
			_, _, err := zerocpy.TimeUnixMicro.Unmarshal(make([]byte, 7))
			return err
		}},
		{"time nano", func() error {
			_, _, err := zerocpy.TimeUnixNano.Unmarshal(make([]byte, 7))
			return err
		}},
		{"time utc", func() error {
			_, _, err := zerocpy.TimeUnixUTC.Unmarshal(make([]byte, 7))
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); !errors.Is(err, binco.ErrTooSmallByteSlice) {
				t.Fatalf("got %v, want ErrTooSmallByteSlice", err)
			}
		})
	}
}

func TestArrayWithLengthSerializer(t *testing.T) {

	serializer := zerocpy.NewArraySerializer[[2]int, int](varnum.Int, arropts.WithLenSerializer[int](scalar.Int))
	bs := binco.Encode(serializer, [2]int{1, 2})
	if len(bs) != 10 || bs[0] != 2 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
	if got, err := binco.Decode(serializer, bs); err != nil || got != [2]int{1, 2} {
		t.Fatalf("Unmarshal = (%v, %v)", got, err)
	}
}
