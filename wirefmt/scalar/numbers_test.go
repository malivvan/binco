package scalar_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	braw "github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/scalar"
)

func encode[T any](t *testing.T, s wirefmt.Serializer[T], v T) []byte {
	t.Helper()
	var buf bytes.Buffer
	n, err := s.Marshal(v, wirefmt.NewWriter(&buf))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if n != buf.Len() {
		t.Fatalf("Marshal reported %d bytes, wrote %d", n, buf.Len())
	}
	if n != s.Size(v) {
		t.Fatalf("Marshal wrote %d bytes, Size returned %d", n, s.Size(v))
	}
	return buf.Bytes()
}

func roundTrip[T any](t *testing.T, name string, s wirefmt.Serializer[T], values []T) {
	t.Helper()
	for _, v := range values {
		bs := encode(t, s, v)

		got, n, err := s.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
		if err != nil {
			t.Fatalf("%v: Unmarshal(%v): %v", name, bs, err)
		}
		if n != len(bs) {
			t.Fatalf("%v: Unmarshal read %d bytes, want %d", name, n, len(bs))
		}
		if !reflect.DeepEqual(got, v) {
			t.Fatalf("%v: round trip returned %v, want %v", name, got, v)
		}

		r := bytes.NewReader(append(append([]byte{}, bs...), 0xAA))
		if n, err = s.Skip(wirefmt.NewReader(r)); err != nil || n != len(bs) {
			t.Fatalf("%v: Skip returned (%d, %v), want (%d, nil)", name, n, err, len(bs))
		}
		if r.Len() != 1 {
			t.Fatalf("%v: Skip left %d bytes, want 1", name, r.Len())
		}
	}
}

func TestNumbers(t *testing.T) {
	roundTrip(t, "byte", scalar.Byte, []byte{0, 1, 127, 255})
	roundTrip(t, "uint8", scalar.Uint8, []uint8{0, 1, 255})
	roundTrip(t, "uint16", scalar.Uint16, []uint16{0, 1, math.MaxUint16})
	roundTrip(t, "uint32", scalar.Uint32, []uint32{0, 1, math.MaxUint32})
	roundTrip(t, "uint64", scalar.Uint64, []uint64{0, 1, math.MaxUint64})
	roundTrip(t, "uint", scalar.Uint, []uint{0, 1, math.MaxUint})
	roundTrip(t, "int8", scalar.Int8, []int8{math.MinInt8, -1, 0, math.MaxInt8})
	roundTrip(t, "int16", scalar.Int16, []int16{math.MinInt16, -1, 0, math.MaxInt16})
	roundTrip(t, "int32", scalar.Int32, []int32{math.MinInt32, -1, 0, math.MaxInt32})
	roundTrip(t, "int64", scalar.Int64, []int64{math.MinInt64, -1, 0, math.MaxInt64})
	roundTrip(t, "int", scalar.Int, []int{math.MinInt, -1, 0, math.MaxInt})
}

func TestFloats(t *testing.T) {
	roundTrip(t, "float64", scalar.Float64, []float64{0, 1.5, -1.5, math.MaxFloat64,
		math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1)})
	roundTrip(t, "float32", scalar.Float32, []float32{0, 1.5, -1.5, math.MaxFloat32,
		math.SmallestNonzeroFloat32, float32(math.Inf(1))})
}

func TestLayout(t *testing.T) {
	tests := []struct {
		name   string
		stream []byte
		buffer []byte
		want   []byte
	}{
		{"uint64 40", encode(t, scalar.Uint64, 40), binco.Encode(braw.Uint64, 40),
			[]byte{40, 0, 0, 0, 0, 0, 0, 0}},
		{"uint64 max", encode(t, scalar.Uint64, math.MaxUint64),
			binco.Encode(braw.Uint64, math.MaxUint64),
			[]byte{255, 255, 255, 255, 255, 255, 255, 255}},
		{"uint32 1", encode(t, scalar.Uint32, 1), binco.Encode(braw.Uint32, 1),
			[]byte{1, 0, 0, 0}},
		{"uint16 0x1234", encode(t, scalar.Uint16, 0x1234), binco.Encode(braw.Uint16, 0x1234),
			[]byte{0x34, 0x12}},
		{"int64 -1", encode(t, scalar.Int64, -1), binco.Encode(braw.Int64, -1),
			[]byte{255, 255, 255, 255, 255, 255, 255, 255}},
		{"float64 1.0", encode(t, scalar.Float64, 1.0), binco.Encode(braw.Float64, 1.0),
			[]byte{0, 0, 0, 0, 0, 0, 0xF0, 0x3F}},
		{"float32 1.5", encode(t, scalar.Float32, 1.5), binco.Encode(braw.Float32, 1.5),
			[]byte{0, 0, 0xC0, 0x3F}},
	}
	for _, tc := range tests {
		if !bytes.Equal(tc.stream, tc.buffer) {
			t.Fatalf("%v: stream %v != buffer %v", tc.name, tc.stream, tc.buffer)
		}
		if !bytes.Equal(tc.stream, tc.want) {
			t.Fatalf("%v: got %v, want %v", tc.name, tc.stream, tc.want)
		}
	}
}

func TestTime(t *testing.T) {
	sec := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	values := []time.Time{
		sec,
		sec.Add(456 * time.Millisecond),
		sec.Add(456789 * time.Microsecond),
		sec.Add(456789123 * time.Nanosecond),
	}
	serializers := []struct {
		name   string
		stream wirefmt.Serializer[time.Time]
		buffer binco.Serializer[time.Time]
	}{
		{"unix", scalar.TimeUnix, braw.TimeUnix},
		{"unix milli", scalar.TimeUnixMilli, braw.TimeUnixMilli},
		{"unix micro", scalar.TimeUnixMicro, braw.TimeUnixMicro},
		{"unix nano", scalar.TimeUnixNano, braw.TimeUnixNano},
		{"unix utc", scalar.TimeUnixUTC, braw.TimeUnixUTC},
		{"unix milli utc", scalar.TimeUnixMilliUTC, braw.TimeUnixMilliUTC},
		{"unix micro utc", scalar.TimeUnixMicroUTC, braw.TimeUnixMicroUTC},
		{"unix nano utc", scalar.TimeUnixNanoUTC, braw.TimeUnixNanoUTC},
	}
	for _, ser := range serializers {
		for _, v := range values {
			streamBS := encode(t, ser.stream, v)
			if bufferBS := binco.Encode(ser.buffer, v); !bytes.Equal(streamBS, bufferBS) {
				t.Fatalf("%v: stream %v != buffer %v", ser.name, streamBS, bufferBS)
			}

			got, n, err := ser.stream.Unmarshal(wirefmt.NewReader(bytes.NewReader(streamBS)))
			if err != nil {
				t.Fatalf("%v: Unmarshal: %v", ser.name, err)
			}
			if n != len(streamBS) {
				t.Fatalf("%v: Unmarshal read %d bytes, want %d", ser.name, n, len(streamBS))
			}
			if n, err = ser.stream.Skip(wirefmt.NewReader(bytes.NewReader(streamBS))); err != nil || n != len(streamBS) {
				t.Fatalf("%v: Skip returned (%d, %v), want (%d, nil)", ser.name, n, err, len(streamBS))
			}

			if ser.name == "unix nano" || ser.name == "unix nano utc" {
				if !got.Equal(v) {
					t.Fatalf("%v: round trip returned %v, want %v", ser.name, got, v)
				}
			} else if got.Unix() != v.Unix() {
				t.Fatalf("%v: round trip returned %v, want %v", ser.name, got, v)
			}
		}
	}

	v := time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("custom", 3*60*60))
	got, _, err := scalar.TimeUnixUTC.Unmarshal(wirefmt.NewReader(
		bytes.NewReader(encode(t, scalar.TimeUnixUTC, v))))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Location() != time.UTC || !got.Equal(v) {
		t.Fatalf("TimeUnixUTC returned %v in %v", got, got.Location())
	}
}

func TestErrors(t *testing.T) {
	reader := func(bs ...byte) wirefmt.Reader {
		return wirefmt.NewReader(bytes.NewReader(bs))
	}
	tests := []struct {
		name string
		err  error
		fn   func() error
	}{
		{"uint64 empty", io.EOF, func() error {
			_, _, err := scalar.Uint64.Unmarshal(reader())
			return err
		}},

		{"uint16 truncated", io.EOF, func() error {
			_, _, err := scalar.Uint16.Unmarshal(reader(1))
			return err
		}},
		{"int8 empty", io.EOF, func() error {
			_, _, err := scalar.Int8.Unmarshal(reader())
			return err
		}},
		{"float64 truncated", io.EOF, func() error {
			_, _, err := scalar.Float64.Unmarshal(reader(1, 2, 3))
			return err
		}},
		{"time empty", io.EOF, func() error {
			_, _, err := scalar.TimeUnix.Unmarshal(reader())
			return err
		}},
		{"skip empty", io.EOF, func() error {
			_, err := scalar.Uint64.Skip(reader())
			return err
		}},
		{"skip truncated", io.EOF, func() error {
			_, err := scalar.Int32.Skip(reader(1, 2))
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

type flakyWriter struct {
	remaining int
	err       error
}

func (w *flakyWriter) WriteByte(c byte) error {
	if w.remaining <= 0 {
		return w.err
	}
	w.remaining--
	return nil
}

func (w *flakyWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, w.err
	}
	w.remaining--
	return len(p), nil
}

func (w *flakyWriter) WriteString(s string) (int, error) {
	if w.remaining <= 0 {
		return 0, w.err
	}
	w.remaining--
	return len(s), nil
}

type failingWriter struct{ err error }

func (w failingWriter) Write(p []byte) (int, error)       { return 0, w.err }
func (w failingWriter) WriteByte(c byte) error            { return w.err }
func (w failingWriter) WriteString(s string) (int, error) { return 0, w.err }

func TestPartialReads(t *testing.T) {
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	tests := []struct {
		name  string
		fn    func(r wirefmt.Reader) (int, error)
		bytes []byte
	}{
		{"uint64", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Uint64.Unmarshal(r)
			return n, err
		}, encode(t, scalar.Uint64, math.MaxUint64)},
		{"int32", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Int32.Unmarshal(r)
			return n, err
		}, encode(t, scalar.Int32, math.MinInt32)},
		{"float64", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Float64.Unmarshal(r)
			return n, err
		}, encode(t, scalar.Float64, 1.5)},
		{"time", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.TimeUnix.Unmarshal(r)
			return n, err
		}, encode(t, scalar.TimeUnix, ts)},
	}
	for _, tc := range tests {
		for length := 0; length < len(tc.bytes); length++ {
			n, err := tc.fn(wirefmt.NewReader(bytes.NewReader(tc.bytes[:length])))
			if !errors.Is(err, io.EOF) {
				t.Fatalf("%v truncated to %d bytes: got (%d, %v), want io.EOF", tc.name, length, n, err)
			}
			if n != length {
				t.Fatalf("%v truncated to %d bytes: read %d bytes", tc.name, length, n)
			}
		}

		if _, err := tc.fn(wirefmt.NewReader(bytes.NewReader(tc.bytes))); err != nil {
			t.Fatalf("%v: %v", tc.name, err)
		}
	}
}

func TestWriterErrorInTheMiddle(t *testing.T) {
	want := errors.New("write failed")
	for _, tc := range []struct {
		name      string
		remaining int
		fn        func(w wirefmt.Writer) (int, error)
	}{
		{"uint64", 3, func(w wirefmt.Writer) (int, error) { return scalar.Uint64.Marshal(1, w) }},
		{"int32", 3, func(w wirefmt.Writer) (int, error) { return scalar.Int32.Marshal(1, w) }},
		{"uint16", 1, func(w wirefmt.Writer) (int, error) { return scalar.Uint16.Marshal(1, w) }},
		{"float64", 3, func(w wirefmt.Writer) (int, error) { return scalar.Float64.Marshal(1.5, w) }},
	} {

		w := &flakyWriter{remaining: tc.remaining, err: want}
		if _, err := tc.fn(w); !errors.Is(err, want) {
			t.Fatalf("%v: got %v, want %v", tc.name, err, want)
		}
	}
}

func TestWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	for _, tc := range []struct {
		name string
		fn   func(w wirefmt.Writer) (int, error)
	}{
		{"byte", func(w wirefmt.Writer) (int, error) { return scalar.Byte.Marshal(1, w) }},
		{"uint16", func(w wirefmt.Writer) (int, error) { return scalar.Uint16.Marshal(1, w) }},
		{"uint64", func(w wirefmt.Writer) (int, error) { return scalar.Uint64.Marshal(1, w) }},
		{"int", func(w wirefmt.Writer) (int, error) { return scalar.Int.Marshal(1, w) }},
		{"float64", func(w wirefmt.Writer) (int, error) { return scalar.Float64.Marshal(1.5, w) }},
		{"time", func(w wirefmt.Writer) (int, error) { return scalar.TimeUnix.Marshal(ts, w) }},
	} {
		if _, err := tc.fn(failingWriter{err: want}); !errors.Is(err, want) {
			t.Fatalf("%v: got %v, want %v", tc.name, err, want)
		}
	}
}
