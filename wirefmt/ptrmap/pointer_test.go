package ptrmap_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	wptrmap "github.com/malivvan/binco/wirefmt/ptrmap"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

func encode[T any](t *testing.T, s wirefmt.Serializer[T], v T) []byte {
	t.Helper()
	var buf bytes.Buffer
	n, err := s.Marshal(v, wirefmt.NewWriter(&buf))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if n != buf.Len() || n != s.Size(v) {
		t.Fatalf("Marshal reported %d bytes, wrote %d, Size returned %d", n, buf.Len(), s.Size(v))
	}
	return buf.Bytes()
}

func reader(bs ...byte) wirefmt.Reader {
	return wirefmt.NewReader(bytes.NewReader(bs))
}

func TestPointerMapping(t *testing.T) {
	var ptrMap = binco.NewPtrMap()

	var revPtrMap = binco.NewReversePtrMap()

	var ptrCodec = wptrmap.NewPtrSer(ptrMap, revPtrMap, wvarnum.Int)

	var listSer = wptrmap.Wrap(ptrMap, revPtrMap, composite.NewSliceSer(ptrCodec))

	a, b := 10, 20
	v := []*int{&a, &a, &b}

	bs := encode(t, listSer, v)

	want := []byte{3, byte(binco.PtrMapping), 0, 20, byte(binco.PtrMapping), 0,
		byte(binco.PtrMapping), 1, 40}
	if !bytes.Equal(bs, want) {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}
	got, n, err := listSer.Unmarshal(reader(bs...))
	if err != nil || n != len(bs) {
		t.Fatalf("Unmarshal = (_, %d, %v)", n, err)
	}
	if len(got) != 3 || *got[0] != 10 || *got[2] != 20 {
		t.Fatalf("unexpected values %v", got)
	}
	if got[0] != got[1] {
		t.Fatal("pointer identity was not preserved")
	}
	if got[0] == got[2] {
		t.Fatal("distinct pointers were merged")
	}

	if ptrMap.Len() != 0 || revPtrMap.Len() != 0 {
		t.Fatalf("maps were not reset: %d, %d", ptrMap.Len(), revPtrMap.Len())
	}

	nilBS := encode(t, listSer, []*int{nil})
	if !bytes.Equal(nilBS, []byte{1, byte(binco.PtrNil)}) {
		t.Fatalf("Marshal of [nil] = %v", nilBS)
	}
	got, _, err = listSer.Unmarshal(reader(nilBS...))
	if err != nil || got[0] != nil {
		t.Fatalf("Unmarshal of [nil] = (%v, %v)", got, err)
	}
}

type node struct {
	Value int
	Next  *node
}

type nodeSer struct {
	next wirefmt.Serializer[*node]
}

func (s *nodeSer) Marshal(v node, w wirefmt.Writer) (n int, err error) {
	n, err = wvarnum.Int.Marshal(v.Value, w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.next.Marshal(v.Next, w)
	return n + n1, err
}

func (s *nodeSer) Unmarshal(r wirefmt.Reader) (v node, n int, err error) {
	v.Value, n, err = wvarnum.Int.Unmarshal(r)
	if err != nil {
		return
	}
	var n1 int
	v.Next, n1, err = s.next.Unmarshal(r)
	n += n1
	return
}

func (s *nodeSer) Size(v node) (size int) {
	return wvarnum.Int.Size(v.Value) + s.next.Size(v.Next)
}

func (s *nodeSer) Skip(r wirefmt.Reader) (n int, err error) {
	n, err = wvarnum.Int.Skip(r)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.next.Skip(r)
	return n + n1, err
}

func TestCyclicGraph(t *testing.T) {
	var ptrMap = binco.NewPtrMap()

	var revPtrMap = binco.NewReversePtrMap()

	var ns = &nodeSer{}

	ns.next = wptrmap.NewPtrSer(ptrMap, revPtrMap, ns)
	ser := wptrmap.Wrap(ptrMap, revPtrMap, ns.next)

	a := &node{Value: 1}
	b := &node{Value: 2, Next: a}
	a.Next = b

	bs := encode(t, ser, a)
	got, n, err := ser.Unmarshal(reader(bs...))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(bs) {
		t.Fatalf("Unmarshal read %d bytes, want %d", n, len(bs))
	}
	if got == nil || got.Value != 1 || got.Next == nil || got.Next.Value != 2 {
		t.Fatalf("unexpected graph %+v", got)
	}
	if got.Next.Next != got {
		t.Fatal("the cycle was not restored")
	}
	if _, err = ser.Skip(reader(bs...)); err != nil {
		t.Fatalf("Skip: %v", err)
	}
}

func TestPtrErrors(t *testing.T) {
	ser := wptrmap.NewPtrSer(binco.NewPtrMap(), binco.NewReversePtrMap(), wvarnum.Int)
	if _, _, err := ser.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream returned %v", err)
	}
	if _, err := ser.Skip(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream returned %v", err)
	}
	for _, bs := range [][]byte{{byte(binco.PtrNotNil), 0}, {3, 0}} {
		if _, _, err := ser.Unmarshal(reader(bs...)); err != binco.ErrWrongFormat {
			t.Fatalf("Unmarshal(%v) returned %v, want ErrWrongFormat", bs, err)
		}
		if _, err := ser.Skip(reader(bs...)); err != binco.ErrWrongFormat {
			t.Fatalf("Skip(%v) returned %v, want ErrWrongFormat", bs, err)
		}
	}

	if _, _, err := ser.Unmarshal(reader(byte(binco.PtrMapping), 0x80)); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of a truncated id returned %v", err)
	}
}

func TestWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	ser := wptrmap.Wrap(binco.NewPtrMap(), binco.NewReversePtrMap(),
		wptrmap.NewPtrSer(binco.NewPtrMap(), binco.NewReversePtrMap(), wvarnum.Int))
	v := 1
	if _, err := ser.Marshal(&v, failingWriter{err: want}); !errors.Is(err, want) {
		t.Fatalf("Marshal returned %v, want %v", err, want)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write(p []byte) (int, error)       { return 0, w.err }
func (w failingWriter) WriteByte(c byte) error            { return w.err }
func (w failingWriter) WriteString(s string) (int, error) { return 0, w.err }
