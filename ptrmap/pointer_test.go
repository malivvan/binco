package ptrmap_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func TestPointerMapping(t *testing.T) {
	var ptrMap = binco.NewPtrMap()

	var revPtrMap = binco.NewReversePtrMap()

	var ptrCodec = ptrmap.NewPtrSer(ptrMap, revPtrMap, varnum.Int)

	var listSer = ptrmap.Wrap(ptrMap, revPtrMap, composite.NewSliceSer(ptrCodec))

	a, b := 10, 20
	v := []*int{&a, &a, &b}

	bs := binco.Encode(listSer, v)

	want := []byte{3, byte(binco.PtrMapping), 0, 20, byte(binco.PtrMapping), 0,
		byte(binco.PtrMapping), 1, 40}
	if !reflect.DeepEqual(bs, want) {
		t.Fatalf("Marshal = %v, want %v", bs, want)
	}

	got, n, err := listSer.Unmarshal(bs)
	if err != nil || n != len(bs) {
		t.Fatalf("Unmarshal = (_, %d, %v)", n, err)
	}
	if len(got) != 3 {
		t.Fatalf("Unmarshal returned %d elements", len(got))
	}
	if *got[0] != 10 || *got[2] != 20 {
		t.Fatalf("unexpected values %d, %d", *got[0], *got[2])
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

	nilBS := binco.Encode(listSer, []*int{nil})
	if !reflect.DeepEqual(nilBS, []byte{1, byte(binco.PtrNil)}) {
		t.Fatalf("Marshal of [nil] = %v", nilBS)
	}
	got, _, err = listSer.Unmarshal(nilBS)
	if err != nil || got[0] != nil {
		t.Fatalf("Unmarshal of [nil] = (%v, %v)", got, err)
	}
}

type node struct {
	Value int
	Next  *node
}

type nodeSer struct {
	next binco.Serializer[*node]
}

func (s *nodeSer) Marshal(v node, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Value, bs)
	return n + s.next.Marshal(v.Next, bs[n:])
}

func (s *nodeSer) Unmarshal(bs []byte) (v node, n int, err error) {
	v.Value, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Next, n1, err = s.next.Unmarshal(bs[n:])
	n += n1
	return
}

func (s *nodeSer) Size(v node) (size int) {
	return varnum.Int.Size(v.Value) + s.next.Size(v.Next)
}

func (s *nodeSer) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.next.Skip(bs[n:])
	return n + n1, err
}

func TestCyclicGraph(t *testing.T) {
	var ptrMap = binco.NewPtrMap()

	var revPtrMap = binco.NewReversePtrMap()

	var ns = &nodeSer{}

	ns.next = ptrmap.NewPtrSer(ptrMap, revPtrMap, ns)
	ser := ptrmap.Wrap(ptrMap, revPtrMap, ns.next)

	a := &node{Value: 1}
	b := &node{Value: 2, Next: a}
	a.Next = b

	bs := binco.Encode(ser, a)
	got, n, err := ser.Unmarshal(bs)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != ser.Size(a) {
		t.Fatalf("Unmarshal read %d bytes, Size returned %d", n, ser.Size(a))
	}
	if got == nil || got.Value != 1 || got.Next == nil || got.Next.Value != 2 {
		t.Fatalf("unexpected graph %+v", got)
	}
	if got.Next.Next != got {
		t.Fatal("the cycle was not restored: Next.Next is not the root node")
	}
	if _, err = ser.Skip(bs); err != nil {
		t.Fatalf("Skip: %v", err)
	}
}

func TestPtrErrors(t *testing.T) {
	ser := ptrmap.NewPtrSer(binco.NewPtrMap(), binco.NewReversePtrMap(), varnum.Int)
	if _, _, err := ser.Unmarshal(nil); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal(nil) returned %v", err)
	}
	if _, err := ser.Skip(nil); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Skip(nil) returned %v", err)
	}
	for _, bs := range [][]byte{{byte(binco.PtrNotNil), 0}, {3, 0}} {
		if _, _, err := ser.Unmarshal(bs); err != binco.ErrWrongFormat {
			t.Fatalf("Unmarshal(%v) returned %v, want ErrWrongFormat", bs, err)
		}
		if _, err := ser.Skip(bs); err != binco.ErrWrongFormat {
			t.Fatalf("Skip(%v) returned %v, want ErrWrongFormat", bs, err)
		}
	}

	if _, _, err := ser.Unmarshal([]byte{byte(binco.PtrMapping), 0x80}); err != binco.ErrTooSmallByteSlice {
		t.Fatalf("Unmarshal of a truncated id returned %v", err)
	}
}
