package zerocopy_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	sarropts "github.com/malivvan/binco/wirefmt/config/array"
	sstropts "github.com/malivvan/binco/wirefmt/config/string"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
	snative "github.com/malivvan/binco/wirefmt/zerocopy"
	"github.com/malivvan/binco/zerocopy"
)

func encodeStream[T any](t *testing.T, s wirefmt.Serializer[T], v T) []byte {
	t.Helper()
	var buf bytes.Buffer
	n, err := s.Marshal(v, wirefmt.NewWriter(&buf))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if n != buf.Len() || n != s.Size(v) {
		t.Fatalf("Marshal wrote %d bytes, buffer has %d, Size returned %d",
			n, buf.Len(), s.Size(v))
	}
	return buf.Bytes()
}

func TestParity(t *testing.T) {
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	tests := []struct {
		name   string
		stream []byte
		buffer []byte
	}{
		{"bool", encodeStream(t, snative.Bool, true), binco.Encode(zerocopy.Bool, true)},
		{"byte", encodeStream(t, snative.Byte, 255), binco.Encode(zerocopy.Byte, 255)},
		{"uint8", encodeStream(t, snative.Uint8, 200), binco.Encode(zerocopy.Uint8, 200)},
		{"uint16", encodeStream(t, snative.Uint16, math.MaxUint16), binco.Encode(zerocopy.Uint16, math.MaxUint16)},
		{"uint32", encodeStream(t, snative.Uint32, math.MaxUint32), binco.Encode(zerocopy.Uint32, math.MaxUint32)},
		{"uint64", encodeStream(t, snative.Uint64, math.MaxUint64), binco.Encode(zerocopy.Uint64, math.MaxUint64)},
		{"uint", encodeStream(t, snative.Uint, 7), binco.Encode(zerocopy.Uint, 7)},
		{"int8", encodeStream(t, snative.Int8, -128), binco.Encode(zerocopy.Int8, -128)},
		{"int16", encodeStream(t, snative.Int16, -1), binco.Encode(zerocopy.Int16, -1)},
		{"int32", encodeStream(t, snative.Int32, -1), binco.Encode(zerocopy.Int32, -1)},
		{"int64", encodeStream(t, snative.Int64, math.MinInt64), binco.Encode(zerocopy.Int64, math.MinInt64)},
		{"int", encodeStream(t, snative.Int, -7), binco.Encode(zerocopy.Int, -7)},
		{"float64", encodeStream(t, snative.Float64, 1.5), binco.Encode(zerocopy.Float64, 1.5)},
		{"float32", encodeStream(t, snative.Float32, 1.5), binco.Encode(zerocopy.Float32, 1.5)},
		{"string", encodeStream(t, snative.String, "hello"), binco.Encode(zerocopy.String, "hello")},
		{"time", encodeStream(t, snative.TimeUnixUTC, ts), binco.Encode(zerocopy.TimeUnixUTC, ts)},
		{"time milli", encodeStream(t, snative.TimeUnixMilliUTC, ts), binco.Encode(zerocopy.TimeUnixMilliUTC, ts)},
		{"time micro", encodeStream(t, snative.TimeUnixMicroUTC, ts), binco.Encode(zerocopy.TimeUnixMicroUTC, ts)},
		{"time nano", encodeStream(t, snative.TimeUnixNanoUTC, ts), binco.Encode(zerocopy.TimeUnixNanoUTC, ts)},
	}
	for _, tc := range tests {
		if !bytes.Equal(tc.stream, tc.buffer) {
			t.Fatalf("%v: stream %v != buffer %v", tc.name, tc.stream, tc.buffer)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	bs := encodeStream(t, snative.String, "héllo")
	got, n, err := snative.String.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil || n != len(bs) || got != "héllo" {
		t.Fatalf("string: (%q, %d, %v)", got, n, err)
	}
	if n, err = snative.String.Skip(wirefmt.NewReader(bytes.NewReader(bs))); err != nil || n != len(bs) {
		t.Fatalf("string skip: (%d, %v)", n, err)
	}

	u, _, err := snative.Uint64.Unmarshal(wirefmt.NewReader(
		bytes.NewReader(encodeStream(t, snative.Uint64, math.MaxUint64))))
	if err != nil || u != math.MaxUint64 {
		t.Fatalf("uint64: (%v, %v)", u, err)
	}
	if _, _, err = snative.Bool.Unmarshal(wirefmt.NewReader(
		bytes.NewReader([]byte{2}))); err != binco.ErrWrongFormat {
		t.Fatalf("bool: %v", err)
	}
	if _, _, err = snative.String.Unmarshal(wirefmt.NewReader(
		bytes.NewReader([]byte{5, 'a'}))); err == nil {
		t.Fatal("truncated string returned no error")
	}
}

func TestArray(t *testing.T) {
	serializer := snative.NewArraySerializer[[3]int, int](wvarnum.Int)
	v := [3]int{1, 2, 3}
	bs := encodeStream(t, serializer, v)
	if want := []byte{3, 2, 4, 6}; !bytes.Equal(bs, want) {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}
	got, n, err := serializer.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil || n != len(bs) || got != v {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}

	sliceSerializer := wcomposite.NewSliceSerializer(wvarnum.Int)
	short := encodeStream(t, sliceSerializer, []int{1, 2})
	if _, _, err = serializer.Unmarshal(wirefmt.NewReader(bytes.NewReader(short))); err != binco.ErrWrongLength {
		t.Fatalf("Unmarshal of a shorter array returned %v, want ErrWrongLength", err)
	}
	long := encodeStream(t, sliceSerializer, []int{1, 2, 3, 4})
	if _, _, err = serializer.Unmarshal(wirefmt.NewReader(bytes.NewReader(long))); err != binco.ErrWrongLength {
		t.Fatalf("Unmarshal of a longer array returned %v, want ErrWrongLength", err)
	}
}

func TestSameBytesAsRaw(t *testing.T) {
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	if a, b := encodeStream(t, snative.TimeUnixNano, ts), encodeStream(t, wscalar.TimeUnixNano, ts); !bytes.Equal(a, b) {
		t.Fatalf("time nano: %v != %v", a, b)
	}
	if a, b := binco.Encode(zerocopy.Float64, 2.5), binco.Encode(scalar.Float64, 2.5); !bytes.Equal(a, b) {
		t.Fatalf("float64: %v != %v", a, b)
	}
}

func checkScalarStream[T comparable](t *testing.T, name string,
	s wirefmt.Serializer[T], ref binco.Serializer[T], values []T,
) {
	t.Helper()
	for _, v := range values {
		bs := encodeStream(t, s, v)
		if want := binco.Encode(ref, v); !bytes.Equal(bs, want) {
			t.Fatalf("%v: stream %v != buffer %v", name, bs, want)
		}
		got, n, err := s.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("%v: Unmarshal(%v) = (%v, %d, %v)", name, bs, got, n, err)
		}
		if n, err = s.Skip(wirefmt.NewReader(bytes.NewReader(bs))); err != nil || n != len(bs) {
			t.Fatalf("%v: Skip = (%d, %v)", name, n, err)
		}
	}
}

func TestAllScalars(t *testing.T) {
	checkScalarStream(t, "byte", snative.Byte, zerocopy.Byte, []byte{0, 1, 255})
	checkScalarStream(t, "uint8", snative.Uint8, zerocopy.Uint8, []uint8{0, 1, 255})
	checkScalarStream(t, "uint16", snative.Uint16, zerocopy.Uint16,
		[]uint16{0, 1, math.MaxUint16})
	checkScalarStream(t, "uint32", snative.Uint32, zerocopy.Uint32,
		[]uint32{0, 1, math.MaxUint32})
	checkScalarStream(t, "uint64", snative.Uint64, zerocopy.Uint64,
		[]uint64{0, 1, math.MaxUint64})
	checkScalarStream(t, "uint", snative.Uint, zerocopy.Uint, []uint{0, 1, math.MaxUint})
	checkScalarStream(t, "int8", snative.Int8, zerocopy.Int8,
		[]int8{math.MinInt8, -1, 0, math.MaxInt8})
	checkScalarStream(t, "int16", snative.Int16, zerocopy.Int16,
		[]int16{math.MinInt16, -1, math.MaxInt16})
	checkScalarStream(t, "int32", snative.Int32, zerocopy.Int32,
		[]int32{math.MinInt32, -1, math.MaxInt32})
	checkScalarStream(t, "int64", snative.Int64, zerocopy.Int64,
		[]int64{math.MinInt64, -1, math.MaxInt64})
	checkScalarStream(t, "int", snative.Int, zerocopy.Int,
		[]int{math.MinInt, -1, 0, math.MaxInt})
	checkScalarStream(t, "float32", snative.Float32, zerocopy.Float32,
		[]float32{0, 1.5, -1.5, math.MaxFloat32})
	checkScalarStream(t, "float64", snative.Float64, zerocopy.Float64,
		[]float64{0, 1.5, -1.5, math.MaxFloat64, math.SmallestNonzeroFloat64})
	checkScalarStream(t, "bool", snative.Bool, zerocopy.Bool, []bool{true, false})
}

func TestStringSerializers(t *testing.T) {
	for _, v := range []string{"", "a", "hello world", "héllo"} {
		bs := encodeStream(t, snative.String, v)
		if want := binco.Encode(zerocopy.String, v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%q) = %v, want %v", v, bs, want)
		}
		got, n, err := snative.String.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%q, %d, %v)", bs, got, n, err)
		}
	}

	rawLen := snative.NewStringSerializer(sstropts.WithLenSerializer(wscalar.Int))
	bs := encodeStream(t, rawLen, "abc")
	if len(bs) != 11 || bs[0] != 3 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}

	wantErr := errors.New("too long")
	valid := snative.NewStringSerializer(
		sstropts.WithLenValidator(binco.ValidatorFn[int](func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	long := encodeStream(t, valid, "abc")
	if _, _, err := valid.Unmarshal(wirefmt.NewReader(bytes.NewReader(long))); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, _, err := valid.Unmarshal(wirefmt.NewReader(
		bytes.NewReader(encodeStream(t, valid, "ab")))); err != nil || got != "ab" {
		t.Fatalf("validated Unmarshal = (%q, %v)", got, err)
	}

	if _, _, err := snative.String.Unmarshal(wirefmt.NewReader(
		bytes.NewReader([]byte{5, 'a'}))); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}
	if _, err := snative.String.Skip(wirefmt.NewReader(
		bytes.NewReader([]byte{5, 'a'}))); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Skip of a truncated string returned %v", err)
	}
}

type negativeLenSerializer struct{}

func (negativeLenSerializer) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, nil }
func (negativeLenSerializer) Unmarshal(r wirefmt.Reader) (int, int, error)       { return -1, 1, nil }
func (negativeLenSerializer) Size(v int) (size int)                              { return 1 }
func (negativeLenSerializer) Skip(r wirefmt.Reader) (n int, err error)           { return 1, nil }

func TestTimeAllUnits(t *testing.T) {
	sec := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	serializers := []struct {
		name   string
		stream wirefmt.Serializer[time.Time]
		buffer binco.Serializer[time.Time]
	}{
		{"unix", snative.TimeUnix, zerocopy.TimeUnix},
		{"unix milli", snative.TimeUnixMilli, zerocopy.TimeUnixMilli},
		{"unix micro", snative.TimeUnixMicro, zerocopy.TimeUnixMicro},
		{"unix nano", snative.TimeUnixNano, zerocopy.TimeUnixNano},
		{"unix utc", snative.TimeUnixUTC, zerocopy.TimeUnixUTC},
		{"unix milli utc", snative.TimeUnixMilliUTC, zerocopy.TimeUnixMilliUTC},
		{"unix micro utc", snative.TimeUnixMicroUTC, zerocopy.TimeUnixMicroUTC},
		{"unix nano utc", snative.TimeUnixNanoUTC, zerocopy.TimeUnixNanoUTC},
	}
	values := []time.Time{
		sec,
		sec.Add(456 * time.Millisecond),
		sec.Add(456789 * time.Microsecond),
		sec.Add(456789123 * time.Nanosecond),
	}
	for _, serializer := range serializers {
		for _, v := range values {
			bs := encodeStream(t, serializer.stream, v)
			if want := binco.Encode(serializer.buffer, v); !bytes.Equal(bs, want) {
				t.Fatalf("%v: stream %v != buffer %v", serializer.name, bs, want)
			}
			got, n, err := serializer.stream.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
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
			if n, err = serializer.stream.Skip(wirefmt.NewReader(bytes.NewReader(bs))); err != nil || n != len(bs) {
				t.Fatalf("%v: Skip = (%d, %v)", serializer.name, n, err)
			}
		}
	}

	v := time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("custom", 3*60*60))
	got, _, err := snative.TimeUnixUTC.Unmarshal(wirefmt.NewReader(
		bytes.NewReader(encodeStream(t, snative.TimeUnixUTC, v))))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Location() != time.UTC || !got.Equal(v) {
		t.Fatalf("TimeUnixUTC returned %v in %v", got, got.Location())
	}
}

func TestValidArray(t *testing.T) {
	wantErr := errors.New("elem")
	serializer := snative.NewArraySerializer[[2]int, int](wvarnum.Int,
		sarropts.WithElemValidator[int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := serializer.Unmarshal(wirefmt.NewReader(bytes.NewReader(
		encodeStream(t, serializer, [2]int{-1, 1})))); !errors.Is(err, wantErr) {
		t.Fatalf("element validator returned %v, want %v", err, wantErr)
	}
	if got, _, err := serializer.Unmarshal(wirefmt.NewReader(bytes.NewReader(
		encodeStream(t, serializer, [2]int{1, 2})))); err != nil || got != [2]int{1, 2} {
		t.Fatalf("Unmarshal = (%v, %v)", got, err)
	}

	rawLen := snative.NewArraySerializer[[2]int, int](wvarnum.Int,
		sarropts.WithLenSerializer[int](wscalar.Int))
	bs := encodeStream(t, rawLen, [2]int{1, 2})
	if len(bs) != 10 || bs[0] != 2 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
}

func TestScalarErrors(t *testing.T) {
	empty := func() wirefmt.Reader { return wirefmt.NewReader(bytes.NewReader(nil)) }
	bytesReader := func(bs ...byte) wirefmt.Reader {
		return wirefmt.NewReader(bytes.NewReader(bs))
	}
	tests := []struct {
		name string
		err  error
		fn   func() error
	}{
		{"uint64 empty", io.EOF, func() error {
			_, _, err := snative.Uint64.Unmarshal(empty())
			return err
		}},
		{"uint64 truncated", io.ErrUnexpectedEOF, func() error {
			_, _, err := snative.Uint64.Unmarshal(bytesReader(1, 2, 3))
			return err
		}},
		{"uint32 empty", io.EOF, func() error {
			_, _, err := snative.Uint32.Unmarshal(empty())
			return err
		}},
		{"uint16 truncated", io.ErrUnexpectedEOF, func() error {
			_, _, err := snative.Uint16.Unmarshal(bytesReader(1))
			return err
		}},
		{"uint8 empty", io.EOF, func() error {
			_, _, err := snative.Uint8.Unmarshal(empty())
			return err
		}},
		{"int64 empty", io.EOF, func() error {
			_, _, err := snative.Int64.Unmarshal(empty())
			return err
		}},
		{"int32 truncated", io.ErrUnexpectedEOF, func() error {
			_, _, err := snative.Int32.Unmarshal(bytesReader(1, 2))
			return err
		}},
		{"int16 empty", io.EOF, func() error {
			_, _, err := snative.Int16.Unmarshal(empty())
			return err
		}},
		{"byte empty", io.EOF, func() error {
			_, _, err := snative.Byte.Unmarshal(empty())
			return err
		}},
		{"float64 truncated", io.ErrUnexpectedEOF, func() error {
			_, _, err := snative.Float64.Unmarshal(bytesReader(1, 2))
			return err
		}},
		{"float32 empty", io.EOF, func() error {
			_, _, err := snative.Float32.Unmarshal(empty())
			return err
		}},
		{"time empty", io.EOF, func() error {
			_, _, err := snative.TimeUnix.Unmarshal(empty())
			return err
		}},
		{"bool empty", io.EOF, func() error {
			_, _, err := snative.Bool.Unmarshal(empty())
			return err
		}},
		{"skip uint64", io.EOF, func() error {
			_, err := snative.Uint64.Skip(empty())
			return err
		}},
		{"skip int32", io.ErrUnexpectedEOF, func() error {
			_, err := snative.Int32.Skip(bytesReader(1))
			return err
		}},
		{"skip time", io.EOF, func() error {
			_, err := snative.TimeUnixNano.Skip(empty())
			return err
		}},
		{"skip bool", io.EOF, func() error {
			_, err := snative.Bool.Skip(empty())
			return err
		}},
		{"negative length", binco.ErrNegativeLength, func() error {
			_, _, err := snative.NewStringSerializer(
				sstropts.WithLenSerializer(negativeLenSerializer{})).Unmarshal(bytesReader(0))
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); !errors.Is(err, tc.err) {
				t.Fatalf("got %v, want %v", err, tc.err)
			}
		})
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(p []byte) (int, error)       { return 0, w.err }
func (w failingWriter) WriteByte(c byte) error            { return w.err }
func (w failingWriter) WriteString(s string) (int, error) { return 0, w.err }

func TestWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	w := failingWriter{err: want}
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	for _, tc := range []struct {
		name string
		fn   func(w wirefmt.Writer) (int, error)
	}{
		{"bool", func(w wirefmt.Writer) (int, error) { return snative.Bool.Marshal(true, w) }},
		{"byte", func(w wirefmt.Writer) (int, error) { return snative.Byte.Marshal(1, w) }},
		{"uint64", func(w wirefmt.Writer) (int, error) { return snative.Uint64.Marshal(1, w) }},
		{"int32", func(w wirefmt.Writer) (int, error) { return snative.Int32.Marshal(1, w) }},
		{"float64", func(w wirefmt.Writer) (int, error) { return snative.Float64.Marshal(1.5, w) }},
		{"string", func(w wirefmt.Writer) (int, error) { return snative.String.Marshal("abc", w) }},
		{"time", func(w wirefmt.Writer) (int, error) { return snative.TimeUnixNano.Marshal(ts, w) }},
	} {
		if _, err := tc.fn(w); !errors.Is(err, want) {
			t.Fatalf("%v: got %v, want %v", tc.name, err, want)
		}
	}
}

func TestValidStringErrors(t *testing.T) {
	valid := snative.NewStringSerializer()
	reader := func(bs ...byte) wirefmt.Reader {
		return wirefmt.NewReader(bytes.NewReader(bs))
	}

	got, n, err := valid.Unmarshal(reader(0))
	if err != nil || n != 1 || got != "" {
		t.Fatalf("Unmarshal of an empty string = (%q, %d, %v)", got, n, err)
	}

	if _, _, err = valid.Unmarshal(reader(5, 'a')); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}

	if _, _, err = snative.NewStringSerializer(
		sstropts.WithLenSerializer(negativeLenSerializer{})).Unmarshal(reader(0)); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

func TestTimeErrors(t *testing.T) {
	empty := func() wirefmt.Reader { return wirefmt.NewReader(bytes.NewReader(nil)) }
	for _, tc := range []struct {
		name string
		fn   func() error
	}{
		{"milli", func() error {
			_, _, err := snative.TimeUnixMilli.Unmarshal(empty())
			return err
		}},
		{"micro", func() error {
			_, _, err := snative.TimeUnixMicro.Unmarshal(empty())
			return err
		}},
		{"nano", func() error {
			_, _, err := snative.TimeUnixNano.Unmarshal(empty())
			return err
		}},
		{"milli utc", func() error {
			_, _, err := snative.TimeUnixMilliUTC.Unmarshal(empty())
			return err
		}},
	} {
		if err := tc.fn(); !errors.Is(err, io.EOF) {
			t.Fatalf("%v: got %v, want io.EOF", tc.name, err)
		}
	}
}
