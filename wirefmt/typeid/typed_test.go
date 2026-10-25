package typeid_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/typeid"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

const dtm binco.DTM = 7

func reader(bs ...byte) wirefmt.Reader {
	return wirefmt.NewReader(bytes.NewReader(bs))
}

func encode[T any](t *testing.T, s wirefmt.Serializer[T], v T) []byte {
	t.Helper()
	var buf bytes.Buffer
	n, err := s.Marshal(v, wirefmt.NewWriter(&buf))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if n != buf.Len() || n != s.Size(v) {
		t.Fatalf("Marshal reported %d bytes, wrote %d, Size returned %d",
			n, buf.Len(), s.Size(v))
	}
	return buf.Bytes()
}

func TestSerializer(t *testing.T) {
	serializer := typeid.NewSerializer(dtm, wvarnum.Int)
	if serializer.DTM() != dtm {
		t.Fatalf("DTM returned %d", serializer.DTM())
	}

	bs := encode(t, serializer, 300)
	if len(bs) != 3 || bs[0] != byte(dtm) {
		t.Fatalf("Marshal = %v", bs)
	}

	got, n, err := serializer.Unmarshal(reader(bs...))
	if err != nil || n != len(bs) || got != 300 {
		t.Fatalf("Unmarshal = (%d, %d, %v)", got, n, err)
	}

	if n, err = serializer.Skip(reader(bs...)); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	data, n, err := serializer.UnmarshalData(reader(bs[1:]...))
	if err != nil || n != 2 || data != 300 {
		t.Fatalf("UnmarshalData = (%d, %d, %v)", data, n, err)
	}
	if n, err = serializer.SkipData(reader(bs[1:]...)); err != nil || n != 2 {
		t.Fatalf("SkipData = (%d, %v)", n, err)
	}
}

func TestWrongDTM(t *testing.T) {
	serializer := typeid.NewSerializer(dtm, wvarnum.Int)
	bs := encode(t, typeid.NewSerializer(dtm+1, wvarnum.Int), 1)

	_, n, err := serializer.Unmarshal(reader(bs...))
	if err == nil {
		t.Fatal("Unmarshal of a foreign DTM returned no error")
	}
	wrong, ok := errors.AsType[binco.WrongDTMError](err)
	if !ok {
		t.Fatalf("Unmarshal returned %v, want WrongDTMError", err)
	}
	if wrong.Expected() != dtm || wrong.Actual() != dtm+1 {
		t.Fatalf("WrongDTMError carries (%d, %d)", wrong.Expected(), wrong.Actual())
	}
	if n != 1 {
		t.Fatalf("Unmarshal read %d bytes, want 1", n)
	}
	if _, err = serializer.Skip(reader(bs...)); err == nil {
		t.Fatal("Skip of a foreign DTM returned no error")
	}

	if _, _, err = serializer.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream returned %v", err)
	}
	if _, _, err = serializer.Unmarshal(reader(0x80)); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of a truncated DTM returned %v", err)
	}
}

func TestDTMSerializer(t *testing.T) {
	for _, v := range []binco.DTM{1, 300, binco.DTM(1 << 20)} {
		bs := encode(t, typeid.DTMSerializer, v)

		got, n, err := typeid.DTMSerializer.Unmarshal(reader(bs...))
		if err != nil || n != len(bs) || got != v {
			t.Fatalf("Unmarshal(%v) = (%d, %d, %v)", bs, got, n, err)
		}

		if n, err = typeid.DTMSerializer.Skip(reader(bs...)); err != nil || n != len(bs) {
			t.Fatalf("Skip = (%d, %v)", n, err)
		}
	}
	if _, _, err := typeid.DTMSerializer.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream returned %v", err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(p []byte) (int, error)       { return 0, w.err }
func (w failingWriter) WriteByte(c byte) error            { return w.err }
func (w failingWriter) WriteString(s string) (int, error) { return 0, w.err }

func TestWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	w := failingWriter{err: want}

	if _, err := typeid.DTMSerializer.Marshal(dtm, w); !errors.Is(err, want) {
		t.Fatalf("DTMSerializer.Marshal returned %v, want %v", err, want)
	}
	if _, err := typeid.NewSerializer(dtm, wvarnum.Int).Marshal(1, w); !errors.Is(err, want) {
		t.Fatalf("Serializer.Marshal returned %v, want %v", err, want)
	}
}
