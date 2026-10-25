package varnum_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	bvarint "github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/varnum"
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

func TestUnsigned(t *testing.T) {
	roundTrip(t, "uint8", varnum.Uint8, []uint8{0, 1, 127, 128, 200, 255})
	roundTrip(t, "uint16", varnum.Uint16,
		[]uint16{0, 1, 127, 128, 300, 16383, 16384, math.MaxUint16})
	roundTrip(t, "uint32", varnum.Uint32,
		[]uint32{0, 1, 128, 1 << 28, 1<<28 - 1, math.MaxUint32})
	roundTrip(t, "uint64", varnum.Uint64,
		[]uint64{0, 1, 128, 1 << 56, 1<<56 - 1, math.MaxUint64})
	roundTrip(t, "uint", varnum.Uint, []uint{0, 1, 1 << 28, math.MaxUint})
	roundTrip(t, "byte", varnum.Byte, []byte{0, 1, 127, 128, 255})
}

func TestSigned(t *testing.T) {
	roundTrip(t, "int8", varnum.Int8,
		[]int8{math.MinInt8, -64, -1, 0, 1, 63, math.MaxInt8})
	roundTrip(t, "int16", varnum.Int16,
		[]int16{math.MinInt16, -1, 0, 1, math.MaxInt16})
	roundTrip(t, "int32", varnum.Int32,
		[]int32{math.MinInt32, -1, 0, 1, math.MaxInt32})
	roundTrip(t, "int64", varnum.Int64,
		[]int64{math.MinInt64, -1, 0, 1, math.MaxInt64})
	roundTrip(t, "int", varnum.Int, []int{math.MinInt, -1, 0, 1, math.MaxInt})
}

func TestPositiveSigned(t *testing.T) {
	roundTrip(t, "positive int8", varnum.PositiveInt8, []int8{0, 1, 127})
	roundTrip(t, "positive int16", varnum.PositiveInt16,
		[]int16{0, 1, 300, math.MaxInt16})
	roundTrip(t, "positive int32", varnum.PositiveInt32,
		[]int32{0, 1, 300, math.MaxInt32})
	roundTrip(t, "positive int64", varnum.PositiveInt64,
		[]int64{0, 1, 300, math.MaxInt64})
	roundTrip(t, "positive int", varnum.PositiveInt, []int{0, 1, 300, math.MaxInt})
}

func TestFloat(t *testing.T) {
	roundTrip(t, "float64", varnum.Float64, []float64{0, 1.5, -1.5, 1e-9,
		math.MaxFloat64, math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1)})
	roundTrip(t, "float32", varnum.Float32, []float32{0, 1.5, -1.5, 1e-9,
		math.MaxFloat32, math.SmallestNonzeroFloat32,
		float32(math.Inf(1)), float32(math.Inf(-1))})
}

func TestEncodings(t *testing.T) {
	tests := []struct {
		name   string
		stream []byte
		buffer []byte
	}{
		{"uint64 0", encode(t, varnum.Uint64, 0), binco.Encode(bvarint.Uint64, 0)},
		{"uint64 128", encode(t, varnum.Uint64, 128), binco.Encode(bvarint.Uint64, 128)},
		{"uint64 max", encode(t, varnum.Uint64, math.MaxUint64),
			binco.Encode(bvarint.Uint64, math.MaxUint64)},
		{"uint8 255", encode(t, varnum.Uint8, 255), binco.Encode(bvarint.Uint8, 255)},
		{"uint16 16384", encode(t, varnum.Uint16, 16384),
			binco.Encode(bvarint.Uint16, 16384)},
		{"int64 -1", encode(t, varnum.Int64, -1), binco.Encode(bvarint.Int64, -1)},
		{"int64 300", encode(t, varnum.Int64, 300), binco.Encode(bvarint.Int64, 300)},
		{"int64 -300", encode(t, varnum.Int64, -300), binco.Encode(bvarint.Int64, -300)},
		{"positive int 300", encode(t, varnum.PositiveInt, 300),
			binco.Encode(bvarint.PositiveInt, 300)},
		{"float64 1.5", encode(t, varnum.Float64, 1.5),
			binco.Encode(bvarint.Float64, 1.5)},
		{"byte 7", encode(t, varnum.Byte, 7), binco.Encode(bvarint.Byte, 7)},
	}
	for _, tc := range tests {
		if !bytes.Equal(tc.stream, tc.buffer) {
			t.Fatalf("%v: stream %v != buffer %v", tc.name, tc.stream, tc.buffer)
		}
	}

	for _, tc := range []struct {
		name string
		got  []byte
		want []byte
	}{
		{"uint64 127", encode(t, varnum.Uint64, 127), []byte{127}},
		{"uint64 128", encode(t, varnum.Uint64, 128), []byte{128, 1}},
		{"uint64 300", encode(t, varnum.Uint64, 300), []byte{172, 2}},
		{"int64 0", encode(t, varnum.Int64, 0), []byte{0}},
		{"int64 -1", encode(t, varnum.Int64, -1), []byte{1}},
		{"int64 1", encode(t, varnum.Int64, 1), []byte{2}},
		{"int64 63", encode(t, varnum.Int64, 63), []byte{126}},
		{"int64 -64", encode(t, varnum.Int64, -64), []byte{127}},
		{"int64 64", encode(t, varnum.Int64, 64), []byte{128, 1}},
		{"int64 300", encode(t, varnum.Int64, 300), []byte{216, 4}},
	} {
		if !bytes.Equal(tc.got, tc.want) {
			t.Fatalf("%v: got %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

func TestUnmarshalErrors(t *testing.T) {
	reader := func(bs ...byte) wirefmt.Reader {
		return wirefmt.NewReader(bytes.NewReader(bs))
	}
	tests := []struct {
		name string
		err  error
		fn   func() error
	}{
		{"uint8 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint8.Unmarshal(reader(0x80, 0x02))
			return err
		}},
		{"uint16 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint16.Unmarshal(reader(0xFF, 0xFF, 0x04))
			return err
		}},
		{"uint32 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint32.Unmarshal(reader(0xFF, 0xFF, 0xFF, 0xFF, 0x10))
			return err
		}},
		{"uint64 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Uint64.Unmarshal(reader(
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x02))
			return err
		}},
		{"empty", io.EOF, func() error {
			_, _, err := varnum.Uint64.Unmarshal(reader())
			return err
		}},
		{"truncated", io.EOF, func() error {
			_, _, err := varnum.Uint64.Unmarshal(reader(0x80))
			return err
		}},
		{"skip uint8 overflow", binco.ErrOverflow, func() error {
			_, err := varnum.Uint8.Skip(reader(0x80, 0x02))
			return err
		}},
		{"skip empty", io.EOF, func() error {
			_, err := varnum.Uint64.Skip(reader())
			return err
		}},
		{"skip truncated", io.EOF, func() error {
			_, err := varnum.Uint64.Skip(reader(0x80))
			return err
		}},
		{"int8 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Int8.Unmarshal(reader(0x80, 0x02))
			return err
		}},
		{"float64 empty", io.EOF, func() error {
			_, _, err := varnum.Float64.Unmarshal(reader())
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

func TestWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	for _, tc := range []struct {
		name string
		fn   func(w wirefmt.Writer) (int, error)
	}{
		{"uint64", func(w wirefmt.Writer) (int, error) { return varnum.Uint64.Marshal(300, w) }},
		{"uint8", func(w wirefmt.Writer) (int, error) { return varnum.Uint8.Marshal(200, w) }},
		{"int64", func(w wirefmt.Writer) (int, error) { return varnum.Int64.Marshal(-300, w) }},
		{"float32", func(w wirefmt.Writer) (int, error) { return varnum.Float32.Marshal(1.5, w) }},
	} {
		if _, err := tc.fn(failingWriter{err: want}); !errors.Is(err, want) {
			t.Fatalf("%v: got %v, want %v", tc.name, err, want)
		}
	}
}

func TestWriterErrorInTheMiddle(t *testing.T) {
	want := errors.New("write failed")

	if _, err := varnum.Uint64.Marshal(300, &flakyWriter{remaining: 1, err: want}); !errors.Is(err, want) {
		t.Fatalf("Uint64: got %v, want %v", err, want)
	}

	if _, err := varnum.Int64.Marshal(-300, &flakyWriter{remaining: 1, err: want}); !errors.Is(err, want) {
		t.Fatalf("Int64: got %v, want %v", err, want)
	}

	if _, err := varnum.Uint8.Marshal(7, &flakyWriter{remaining: 0, err: want}); !errors.Is(err, want) {
		t.Fatalf("Uint8: got %v, want %v", err, want)
	}
}

func TestUnmarshalErrorPropagation(t *testing.T) {
	reader := func(bs ...byte) wirefmt.Reader {
		return wirefmt.NewReader(bytes.NewReader(bs))
	}
	tests := []struct {
		name string
		err  error
		fn   func() error
	}{
		{"int16 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Int16.Unmarshal(reader(0xFF, 0xFF, 0x04))
			return err
		}},
		{"int32 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Int32.Unmarshal(reader(0xFF, 0xFF, 0xFF, 0xFF, 0x10))
			return err
		}},
		{"int64 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.Int64.Unmarshal(reader(
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x02))
			return err
		}},
		{"int empty", io.EOF, func() error {
			_, _, err := varnum.Int.Unmarshal(reader())
			return err
		}},
		{"float32 empty", io.EOF, func() error {
			_, _, err := varnum.Float32.Unmarshal(reader())
			return err
		}},
		{"positive int8 overflow", binco.ErrOverflow, func() error {
			_, _, err := varnum.PositiveInt8.Unmarshal(reader(0x80, 0x02))
			return err
		}},
		{"skip positive int8 overflow", binco.ErrOverflow, func() error {
			_, err := varnum.PositiveInt8.Skip(reader(0x80, 0x02))
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

func TestZigZag(t *testing.T) {
	for _, v := range []int64{0, 1, -1, 2, -2, math.MaxInt64, math.MinInt64} {
		if got := int64(varnum.DecodeZigZag(uint64(varnum.EncodeZigZag(v)))); got != v {
			t.Fatalf("ZigZag round trip of %d returned %d", v, got)
		}
	}
}
