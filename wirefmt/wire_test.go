package wirefmt_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	sstropts "github.com/malivvan/binco/wirefmt/config/string"
	wptrmap "github.com/malivvan/binco/wirefmt/ptrmap"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
	snative "github.com/malivvan/binco/wirefmt/zerocpy"
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

func roundTrip[T any](t *testing.T, s wirefmt.Serializer[T], v T) T {
	t.Helper()
	bs := encode(t, s, v)

	got, n, err := s.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(bs) {
		t.Fatalf("Unmarshal read %d bytes, want %d", n, len(bs))
	}

	r := bytes.NewReader(append(append([]byte{}, bs...), 0xAA))
	if n, err = s.Skip(wirefmt.NewReader(r)); err != nil || n != len(bs) {
		t.Fatalf("Skip returned (%d, %v), want (%d, nil)", n, err, len(bs))
	}
	if r.Len() != 1 {
		t.Fatalf("Skip left %d bytes in the stream, want 1", r.Len())
	}
	return got
}

func checkParity(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("%v: stream bytes %v differ from buffer bytes %v", name, got, want)
	}
}

func TestParityWithBuffer(t *testing.T) {
	checkParity(t, "bool", encode(t, wcomposite.Bool, true), binco.Encode(composite.Bool, true))
	checkParity(t, "string", encode(t, wcomposite.String, "hello"), binco.Encode(composite.String, "hello"))
	checkParity(t, "byte slice", encode(t, wcomposite.ByteSlice, []byte{1, 2}),
		binco.Encode(composite.ByteSlice, []byte{1, 2}))
	checkParity(t, "varint int64", encode(t, wvarnum.Int64, -300), binco.Encode(varnum.Int64, -300))
	checkParity(t, "varint uint64", encode(t, wvarnum.Uint64, 1<<40), binco.Encode(varnum.Uint64, uint64(1<<40)))
	checkParity(t, "raw int64", encode(t, wscalar.Int64, -300), binco.Encode(scalar.Int64, -300))
	checkParity(t, "raw float64", encode(t, wscalar.Float64, 1.5), binco.Encode(scalar.Float64, 1.5))

	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	checkParity(t, "time", encode(t, wscalar.TimeUnixUTC, ts), binco.Encode(scalar.TimeUnixUTC, ts))

	ptr := 7
	checkParity(t, "ptr", encode(t, wcomposite.NewPtrSer(wvarnum.Int), &ptr),
		binco.Encode(composite.NewPtrSer(varnum.Int), &ptr))

	checkParity(t, "slice", encode(t, wcomposite.NewSliceSer(wvarnum.Int), []int{1, -1, 300}),
		binco.Encode(composite.NewSliceSer(varnum.Int), []int{1, -1, 300}))

	checkParity(t, "map", encode(t, wcomposite.NewMapSer(wcomposite.String, wvarnum.Int), map[string]int{"a": 1}),
		binco.Encode(composite.NewMapSer(composite.String, varnum.Int), map[string]int{"a": 1}))

	checkParity(t, "typed", encode(t, wtypeid.NewSer(7, wvarnum.Int), 300),
		binco.Encode(typeid.NewSer(7, varnum.Int), 300))
}

func TestRoundTrip(t *testing.T) {
	if got := roundTrip(t, wcomposite.Bool, false); got != false {
		t.Fatalf("bool round trip returned %v", got)
	}
	if got := roundTrip(t, wcomposite.String, "héllo"); got != "héllo" {
		t.Fatalf("string round trip returned %q", got)
	}
	if got := roundTrip(t, wcomposite.ByteSlice, []byte{1, 2, 3}); !reflect.DeepEqual(got, []byte{1, 2, 3}) {
		t.Fatalf("byte slice round trip returned %v", got)
	}
	if got := roundTrip(t, wvarnum.Int64, -1); got != -1 {
		t.Fatalf("int64 round trip returned %v", got)
	}
	if got := roundTrip(t, wvarnum.Float32, 1.5); got != 1.5 {
		t.Fatalf("float32 round trip returned %v", got)
	}
	if got := roundTrip(t, wcomposite.NewSliceSer(wvarnum.Int), []int{1, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("slice round trip returned %v", got)
	}
	if got := roundTrip(t, wcomposite.NewMapSer(wcomposite.String, wvarnum.Int), map[string]int{"a": 1}); !reflect.DeepEqual(got, map[string]int{"a": 1}) {
		t.Fatalf("map round trip returned %v", got)
	}
}

type onlyWriter struct{ w io.Writer }

func (w onlyWriter) Write(p []byte) (int, error) { return w.w.Write(p) }

type onlyReader struct{ r io.Reader }

func (r onlyReader) Read(p []byte) (int, error) { return r.r.Read(p) }

func TestAdapters(t *testing.T) {
	var buf bytes.Buffer
	w := wirefmt.NewWriter(onlyWriter{w: &buf})
	s := wcomposite.NewStringSer()
	if _, err := s.Marshal("abc", w); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{3, 'a', 'b', 'c'}) {
		t.Fatalf("Marshal wrote %v", buf.Bytes())
	}

	r := onlyReader{r: bytes.NewReader([]byte{3, 'a', 'b', 'c', 0xAA})}
	got, n, err := s.Unmarshal(wirefmt.NewReader(r))
	if err != nil || n != 4 || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %d, %v)", got, n, err)
	}
}

func TestEncodeDecodeHelpers(t *testing.T) {
	var buf bytes.Buffer
	n, err := wirefmt.Encode(wvarnum.Int64, &buf, 300)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if n != 2 || !bytes.Equal(buf.Bytes(), []byte{216, 4}) {
		t.Fatalf("Encode wrote %v", buf.Bytes())
	}

	if _, err = wirefmt.Encode(wcomposite.String, &buf, "hi"); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, n, err := wirefmt.Decode(wvarnum.Int64, &buf)
	if err != nil || got != 300 || n != 2 {
		t.Fatalf("Decode = (%d, %d, %v)", got, n, err)
	}
	s, n, err := wirefmt.Decode(wcomposite.String, &buf)
	if err != nil || s != "hi" || n != 3 {
		t.Fatalf("Decode = (%q, %d, %v)", s, n, err)
	}
}

func TestPointerMapping(t *testing.T) {
	var ptrMap = binco.NewPtrMap()

	var revPtrMap = binco.NewReversePtrMap()

	var ptrSer = wptrmap.NewPtrSer(ptrMap, revPtrMap, wvarnum.Int)

	var listSer = wptrmap.Wrap(ptrMap, revPtrMap, wcomposite.NewSliceSer(ptrSer))

	a, b := 10, 20
	bs := encode(t, listSer, []*int{&a, &a, &b})
	got, n, err := listSer.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil || n != len(bs) {
		t.Fatalf("Unmarshal = (_, %d, %v)", n, err)
	}
	if *got[0] != 10 || *got[2] != 20 || got[0] != got[1] || got[0] == got[2] {
		t.Fatalf("unexpected result %v", got)
	}
	var buf bytes.Buffer
	if _, err = listSer.Marshal([]*int{&a, &a, &b}, wirefmt.NewWriter(&buf)); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{3, byte(binco.PtrMapping), 0, 20,
		byte(binco.PtrMapping), 0, byte(binco.PtrMapping), 1, 40}) {
		t.Fatalf("Marshal = %v", buf.Bytes())
	}
}

func TestTyped(t *testing.T) {
	ser := wtypeid.NewSer(7, wvarnum.Int)
	bs := encode(t, ser, 300)
	if len(bs) != 3 || bs[0] != 7 {
		t.Fatalf("Marshal = %v", bs)
	}
	got, _, err := ser.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil || got != 300 {
		t.Fatalf("Unmarshal = (%d, %v)", got, err)
	}

	foreign := encode(t, wtypeid.NewSer(8, wvarnum.Int), 1)
	_, _, err = ser.Unmarshal(wirefmt.NewReader(bytes.NewReader(foreign)))
	wrong, ok := errors.AsType[binco.WrongDTMError](err)
	if !ok {
		t.Fatalf("Unmarshal returned %v, want WrongDTMError", err)
	}
	if wrong.Expected() != 7 || wrong.Actual() != 8 {
		t.Fatalf("WrongDTMError carries (%d, %d)", wrong.Expected(), wrong.Actual())
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(p []byte) (int, error)       { return 0, w.err }
func (w failingWriter) WriteByte(c byte) error            { return w.err }
func (w failingWriter) WriteString(s string) (int, error) { return 0, w.err }

func TestWriterErrors(t *testing.T) {
	want := errors.New("boom")
	if _, err := wscalar.Int64.Marshal(1, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
	if _, err := wcomposite.Bool.Marshal(true, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
	if _, err := wcomposite.String.Marshal("abc", failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
	if _, err := wvarnum.Uint64.Marshal(300, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
}

func TestReaderErrors(t *testing.T) {
	empty := func() wirefmt.Reader { return wirefmt.NewReader(bytes.NewReader(nil)) }
	if _, _, err := wscalar.Int64.Unmarshal(empty()); err == nil {
		t.Fatal("Unmarshal of an empty stream returned no error")
	}
	if _, err := wcomposite.String.Skip(empty()); err == nil {
		t.Fatal("Skip of an empty stream returned no error")
	}
	if _, _, err := wcomposite.Bool.Unmarshal(empty()); err == nil {
		t.Fatal("Unmarshal of an empty stream returned no error")
	}

	if _, _, err := wcomposite.Bool.Unmarshal(wirefmt.NewReader(bytes.NewReader([]byte{2}))); err != binco.ErrWrongFormat {
		t.Fatalf("Unmarshal(2) returned %v", err)
	}

	if _, _, err := wcomposite.String.Unmarshal(wirefmt.NewReader(bytes.NewReader([]byte{5, 'a'}))); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Unmarshal of a truncated string returned %v", err)
	}

	if _, _, err := wcomposite.NewStringSer(sstropts.WithLenSer(streamStringLenSer{})).Unmarshal(
		wirefmt.NewReader(bytes.NewReader([]byte{0}))); err != binco.ErrNegativeLength {
		t.Fatalf("Unmarshal returned %v, want ErrNegativeLength", err)
	}
}

type streamStringLenSer struct{}

func (s streamStringLenSer) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, nil }
func (s streamStringLenSer) Unmarshal(r wirefmt.Reader) (int, int, error)       { return -1, 1, nil }
func (s streamStringLenSer) Size(v int) (size int)                              { return 1 }
func (s streamStringLenSer) Skip(r wirefmt.Reader) (n int, err error)           { return 1, nil }

type onlyFailingWriter struct{ err error }

func (w onlyFailingWriter) Write(p []byte) (int, error) { return 0, w.err }

func TestWriterAdapterWrite(t *testing.T) {

	var buf bytes.Buffer
	n, err := snative.Uint64.Marshal(1, wirefmt.NewWriter(onlyWriter{w: &buf}))
	if err != nil || n != 8 {
		t.Fatalf("Marshal = (%d, %v)", n, err)
	}
	if buf.Len() != 8 {
		t.Fatalf("wrote %d bytes, want 8", buf.Len())
	}
}

func TestReaderAdapter(t *testing.T) {

	r := wirefmt.NewReader(onlyReader{r: bytes.NewReader(nil)})
	if _, err := r.ReadByte(); !errors.Is(err, io.EOF) {
		t.Fatalf("ReadByte returned %v, want io.EOF", err)
	}
	if n, err := r.Read(make([]byte, 1)); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("Read returned (%d, %v), want (0, io.EOF)", n, err)
	}

	bs := encode(t, wcomposite.String, "abc")
	r = wirefmt.NewReader(onlyReader{r: bytes.NewReader(append(append([]byte{}, bs...), 0xAA))})
	got, n, err := wcomposite.String.Unmarshal(r)
	if err != nil || n != len(bs) || got != "abc" {
		t.Fatalf("Unmarshal = (%q, %d, %v)", got, n, err)
	}
}

func TestEncodePaths(t *testing.T) {

	var plain bytes.Buffer
	n, err := wirefmt.Encode(wcomposite.String, onlyWriter{w: &plain}, "hi")
	if err != nil || n != 3 || !bytes.Equal(plain.Bytes(), []byte{2, 'h', 'i'}) {
		t.Fatalf("Encode = (%d, %v), buffer %v", n, err, plain.Bytes())
	}

	buffered := bufio.NewWriter(&plain)
	if _, err = wirefmt.Encode(wcomposite.String, buffered, "ho"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if err = buffered.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	want := errors.New("boom")
	if _, err = wirefmt.Encode(wcomposite.String, failingWriter{err: want}, "hi"); !errors.Is(err, want) {
		t.Fatalf("Encode returned %v, want %v", err, want)
	}

	if _, err = wirefmt.Encode(wcomposite.String, onlyFailingWriter{err: want}, "hi"); !errors.Is(err, want) {
		t.Fatalf("Encode returned %v, want %v", err, want)
	}
}

func TestDecodePaths(t *testing.T) {
	bs := encode(t, wcomposite.String, "hi")

	got, n, err := wirefmt.Decode(wcomposite.String, onlyReader{r: bytes.NewReader(bs)})
	if err != nil || n != len(bs) || got != "hi" {
		t.Fatalf("Decode = (%q, %d, %v)", got, n, err)
	}

	got, n, err = wirefmt.Decode(wcomposite.String, bytes.NewReader(bs))
	if err != nil || n != len(bs) || got != "hi" {
		t.Fatalf("Decode = (%q, %d, %v)", got, n, err)
	}

	_, _, err = wirefmt.Decode(wcomposite.String, onlyReader{r: bytes.NewReader(nil)})
	if !errors.Is(err, io.EOF) {
		t.Fatalf("Decode returned %v, want io.EOF", err)
	}
}
