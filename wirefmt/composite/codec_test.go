package composite_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	bord "github.com/malivvan/binco/composite"
	bvarint "github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
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

func reader(bs ...byte) wirefmt.Reader {
	return wirefmt.NewReader(bytes.NewReader(bs))
}

type failingLenSer struct{}

func (failingLenSer) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, nil }
func (failingLenSer) Unmarshal(r wirefmt.Reader) (int, int, error)       { return -1, 1, nil }
func (failingLenSer) Size(v int) (size int)                              { return 1 }
func (failingLenSer) Skip(r wirefmt.Reader) (n int, err error)           { return 1, nil }

type failingElemSer struct{ err error }

func (s failingElemSer) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, nil }
func (s failingElemSer) Unmarshal(r wirefmt.Reader) (int, int, error)       { return 0, 0, s.err }
func (s failingElemSer) Size(v int) (size int)                              { return 1 }
func (s failingElemSer) Skip(r wirefmt.Reader) (n int, err error)           { return 0, s.err }

func TestBool(t *testing.T) {
	for _, v := range []bool{true, false} {
		bs := encode(t, composite.Bool, v)
		want := byte(0)
		if v {
			want = 1
		}
		if len(bs) != 1 || bs[0] != want {
			t.Fatalf("Marshal(%v) = %v", v, bs)
		}

		got, n, err := composite.Bool.Unmarshal(reader(bs...))
		if err != nil || n != 1 || got != v {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}

		r := bytes.NewReader(append(bs, 0xAA))
		if n, err = composite.Bool.Skip(wirefmt.NewReader(r)); err != nil || n != 1 || r.Len() != 1 {
			t.Fatalf("Skip = (%d, %v), remaining %d", n, err, r.Len())
		}
	}

	if _, _, err := composite.Bool.Unmarshal(reader(2)); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal(2) returned %v, want ErrWrongFormat", err)
	}
	if _, err := composite.SkipBool(reader(2)); err != binco.ErrWrongFormat {
		t.Fatalf("SkipBool(2) returned %v, want ErrWrongFormat", err)
	}
	if _, _, err := composite.Bool.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream returned %v", err)
	}
	if _, err := composite.Bool.Skip(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream returned %v", err)
	}
	if got := composite.SizeBool(true); got != 1 {
		t.Fatalf("SizeBool = %d", got)
	}
}

func TestString(t *testing.T) {
	for _, v := range []string{"", "a", "hello world", "héllo"} {
		bs := encode(t, composite.String, v)

		if want := binco.Encode(bord.String, v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%q) = %v, want %v", v, bs, want)
		}

		got, n, err := composite.String.Unmarshal(reader(bs...))
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%q, %d, %v)", bs, got, n, err)
		}

		r := bytes.NewReader(append(append([]byte{}, bs...), 0xAA))
		if n, err = composite.String.Skip(wirefmt.NewReader(r)); err != nil || n != len(bs) || r.Len() != 1 {
			t.Fatalf("Skip(%v) = (%d, %v), remaining %d", bs, n, err, r.Len())
		}
	}

	if bs := encode(t, composite.String, "hello world"); bs[0] != 11 {
		t.Fatalf("length prefix = %d, want 11", bs[0])
	}

	if _, _, err := composite.String.Unmarshal(reader(5, 'a')); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}
	if _, err := composite.String.Skip(reader(5, 'a')); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of a truncated string returned %v", err)
	}

	ser := composite.NewStringSer(stropts.WithLenSer(wscalar.Int))
	bs := encode(t, ser, "abc")
	if len(bs) != 11 || bs[0] != 3 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
	if got, err := decode(t, ser, bs); err != nil || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %v)", got, err)
	}

	wantErr := errors.New("too long")
	valid := composite.NewValidStringSer(stropts.WithLenValidator(binco.ValidatorFn[int](
		func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(reader(encode(t, valid, "abc")...)); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, err := decode(t, valid, encode(t, valid, "ab")); err != nil || got != "ab" {
		t.Fatalf("validated Unmarshal = (%q, %v)", got, err)
	}
}

func TestByteSlice(t *testing.T) {
	for _, v := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		bs := encode(t, composite.ByteSlice, v)
		if want := binco.Encode(bord.ByteSlice, v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%v) = %v, want %v", v, bs, want)
		}

		got, n, err := composite.ByteSlice.Unmarshal(reader(bs...))
		if err != nil || n != len(bs) || !reflect.DeepEqual(got, append([]byte{}, v...)) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}

		if n, err = composite.ByteSlice.Skip(reader(bs...)); err != nil || n != len(bs) {
			t.Fatalf("Skip(%v) = (%d, %v)", bs, n, err)
		}
	}

	if _, _, err := composite.ByteSlice.Unmarshal(reader(3, 1)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unmarshal of a truncated slice returned %v", err)
	}
	if _, err := composite.ByteSlice.Skip(reader(3, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of a truncated slice returned %v", err)
	}

	wantErr := errors.New("too long")
	valid := composite.NewValidByteSliceSer(bslopts.WithLenValidator(binco.ValidatorFn[int](
		func(n int) error {
			if n > 2 {
				return wantErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(reader(encode(t, valid, []byte{1, 2, 3})...)); !errors.Is(err, wantErr) {
		t.Fatalf("validated Unmarshal returned %v, want %v", err, wantErr)
	}
	if got, err := decode(t, valid, encode(t, valid, []byte{1})); err != nil || !bytes.Equal(got, []byte{1}) {
		t.Fatalf("validated Unmarshal = (%v, %v)", got, err)
	}
}

func TestPtr(t *testing.T) {
	ser := composite.NewPtrSer(wvarnum.Int)

	var nilPtr *int
	if bs := encode(t, ser, nilPtr); !bytes.Equal(bs, []byte{byte(binco.PtrNil)}) {
		t.Fatalf("Marshal(nil) = %v", bs)
	}
	if got, err := decode(t, ser, []byte{byte(binco.PtrNil)}); err != nil || got != nil {
		t.Fatalf("Unmarshal(nil) = (%v, %v)", got, err)
	}

	v := 5
	bs := encode(t, ser, &v)
	if want := binco.Encode(bord.NewPtrSer(bvarint.Int), &v); !bytes.Equal(bs, want) {
		t.Fatalf("Marshal(&5) = %v, want %v", bs, want)
	}
	got, n, err := ser.Unmarshal(reader(bs...))
	if err != nil || n != len(bs) || got == nil || *got != v {
		t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
	}
	if n, err = ser.Skip(reader(bs...)); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	if _, _, err = ser.Unmarshal(reader(byte(binco.PtrMapping))); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal of a mapping flag returned %v", err)
	}
	if _, err = ser.Skip(reader(byte(binco.PtrMapping))); err != binco.ErrWrongFormat {
		t.Fatalf("Skip of a mapping flag returned %v", err)
	}
	if _, _, err = ser.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream returned %v", err)
	}
	if _, err = ser.Skip(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream returned %v", err)
	}

	if _, _, err = ser.Unmarshal(reader(byte(binco.PtrNotNil), 0x80)); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of a truncated value returned %v", err)
	}
}

func TestSlice(t *testing.T) {
	ser := composite.NewSliceSer(wvarnum.Int)
	for _, v := range [][]int{nil, {}, {0}, {1, -1, 300}} {
		bs := encode(t, ser, v)
		if want := binco.Encode(bord.NewSliceSer(bvarint.Int), v); !bytes.Equal(bs, want) {
			t.Fatalf("Marshal(%v) = %v, want %v", v, bs, want)
		}
		got, n, err := ser.Unmarshal(reader(bs...))
		if err != nil || n != len(bs) || !reflect.DeepEqual(got, append([]int{}, v...)) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = ser.Skip(reader(bs...)); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}

	wantErr := errors.New("elem")
	failing := composite.NewSliceSer(failingElemSer{err: wantErr})
	if _, _, err := failing.Unmarshal(reader(1, 0)); !errors.Is(err, wantErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, wantErr)
	}
	if _, err := failing.Skip(reader(1, 0)); !errors.Is(err, wantErr) {
		t.Fatalf("Skip returned %v, want %v", err, wantErr)
	}

	lenErr, elemErr := errors.New("len"), errors.New("elem")
	valid := composite.NewValidSliceSer(wvarnum.Int,
		slopts.WithLenValidator[int](binco.ValidatorFn[int](func(n int) error {
			if n > 1 {
				return lenErr
			}
			return nil
		})),
		slopts.WithElemValidator[int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return elemErr
			}
			return nil
		})))
	if _, _, err := valid.Unmarshal(reader(encode(t, ser, []int{1, 2})...)); !errors.Is(err, lenErr) {
		t.Fatalf("length validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(reader(encode(t, ser, []int{-1})...)); !errors.Is(err, elemErr) {
		t.Fatalf("element validator returned %v", err)
	}

	rawLen := composite.NewSliceSer(wvarnum.Int, slopts.WithLenSer[int](wscalar.Int))
	if bs := encode(t, rawLen, []int{7}); len(bs) != 9 || bs[0] != 1 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
}

func TestMap(t *testing.T) {
	ser := composite.NewMapSer(composite.String, wvarnum.Int)
	for _, v := range []map[string]int{{}, {"a": 1}, {"a": 1, "b": -2}} {
		bs := encode(t, ser, v)
		got, n, err := ser.Unmarshal(reader(bs...))
		if err != nil || n != len(bs) || !reflect.DeepEqual(got, v) {
			t.Fatalf("Unmarshal(%v) = (%v, %d, %v)", bs, got, n, err)
		}
		if n, err = ser.Skip(reader(bs...)); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}

	single := map[string]int{"a": 1}
	if got, want := encode(t, ser, single),
		binco.Encode(bord.NewMapSer(bord.String, bvarint.Int), single); !bytes.Equal(got, want) {
		t.Fatalf("Marshal = %v, want %v", got, want)
	}

	keyErr, valErr := errors.New("key"), errors.New("value")
	failingKey := composite.NewMapSer(failingElemSer{err: keyErr}, wvarnum.Int)
	if _, _, err := failingKey.Unmarshal(reader(1, 0)); !errors.Is(err, keyErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, keyErr)
	}
	if _, err := failingKey.Skip(reader(1, 0)); !errors.Is(err, keyErr) {
		t.Fatalf("Skip returned %v, want %v", err, keyErr)
	}
	failingValue := composite.NewMapSer(wvarnum.Int, failingElemSer{err: valErr})
	if _, _, err := failingValue.Unmarshal(reader(1, 0)); !errors.Is(err, valErr) {
		t.Fatalf("Unmarshal returned %v, want %v", err, valErr)
	}

	lenErr := errors.New("len")
	valid := composite.NewValidMapSer(wvarnum.Int, wvarnum.Int,
		mapopts.WithLenValidator[int, int](binco.ValidatorFn[int](func(n int) error {
			if n > 1 {
				return lenErr
			}
			return nil
		})),
		mapopts.WithKeyValidator[int, int](binco.ValidatorFn[int](func(k int) error {
			if k < 0 {
				return keyErr
			}
			return nil
		})),
		mapopts.WithValueValidator[int, int](binco.ValidatorFn[int](func(v int) error {
			if v < 0 {
				return valErr
			}
			return nil
		})))

	intMap := composite.NewMapSer(wvarnum.Int, wvarnum.Int)
	if _, _, err := valid.Unmarshal(reader(encode(t, intMap, map[int]int{1: 1, 2: 2})...)); !errors.Is(err, lenErr) {
		t.Fatalf("length validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(reader(encode(t, intMap, map[int]int{-1: 1})...)); !errors.Is(err, keyErr) {
		t.Fatalf("key validator returned %v", err)
	}
	if _, _, err := valid.Unmarshal(reader(encode(t, intMap, map[int]int{1: -1})...)); !errors.Is(err, valErr) {
		t.Fatalf("value validator returned %v", err)
	}

	rawLen := composite.NewMapSer(wvarnum.Int, wvarnum.Int, mapopts.WithLenSer[int, int](wscalar.Int))
	if bs := encode(t, rawLen, map[int]int{1: 2}); len(bs) < 9 || bs[0] != 1 {
		t.Fatalf("Marshal with a scalar.Int length = %v", bs)
	}
}

func TestNegativeLength(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"string", func() error {
			_, _, err := composite.NewStringSer(stropts.WithLenSer(failingLenSer{})).Unmarshal(reader(0))
			return err
		}},
		{"byte slice", func() error {
			_, _, err := composite.NewByteSliceSer(bslopts.WithLenSer(failingLenSer{})).Unmarshal(reader(0))
			return err
		}},
		{"slice", func() error {
			_, _, err := composite.NewSliceSer(wvarnum.Int, slopts.WithLenSer[int](failingLenSer{})).Unmarshal(reader(0))
			return err
		}},
		{"map", func() error {
			_, _, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int,
				mapopts.WithLenSer[int, int](failingLenSer{})).Unmarshal(reader(0))
			return err
		}},
		{"skip string", func() error {
			_, err := composite.NewStringSer(stropts.WithLenSer(failingLenSer{})).Skip(reader(0))
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != binco.ErrNegativeLength {
				t.Fatalf("got %v, want ErrNegativeLength", err)
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
	tests := []struct {
		name string
		fn   func(w wirefmt.Writer) error
	}{
		{"bool", func(w wirefmt.Writer) error { _, err := composite.Bool.Marshal(true, w); return err }},
		{"string", func(w wirefmt.Writer) error { _, err := composite.String.Marshal("abc", w); return err }},
		{"byte slice", func(w wirefmt.Writer) error {
			_, err := composite.ByteSlice.Marshal([]byte{1, 2}, w)
			return err
		}},
		{"ptr", func(w wirefmt.Writer) error {
			v := 5
			_, err := composite.NewPtrSer(wvarnum.Int).Marshal(&v, w)
			return err
		}},
		{"slice", func(w wirefmt.Writer) error {
			_, err := composite.NewSliceSer(wvarnum.Int).Marshal([]int{1}, w)
			return err
		}},
		{"map", func(w wirefmt.Writer) error {
			_, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Marshal(map[int]int{1: 2}, w)
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(w); !errors.Is(err, want) {
				t.Fatalf("got %v, want %v", err, want)
			}
		})
	}
}

func decode[T any](t *testing.T, s wirefmt.Serializer[T], bs []byte) (T, error) {
	t.Helper()
	got, n, err := s.Unmarshal(reader(bs...))
	if err == nil && n != len(bs) {
		t.Fatalf("Unmarshal read %d bytes, want %d", n, len(bs))
	}
	return got, err
}
