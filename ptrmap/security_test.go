package ptrmap_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

// TestTypeConfusionIsRejected covers S3: an identifier defined by one type used
// to be handed to a codec of another type and cast with unsafe.Pointer.
func TestTypeConfusionIsRejected(t *testing.T) {
	ptrMap := binco.NewPtrMap()
	revPtrMap := binco.NewReversePtrMap()

	intSerializer := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, varnum.Int)
	stringSerializer := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String)

	// Define identifier 0 as an *int.
	v, _, err := intSerializer.Unmarshal([]byte{byte(binco.PtrMapping), 0, 10})
	if err != nil {
		t.Fatalf("Unmarshal of the definition returned %v", err)
	}
	if *v != 5 {
		t.Fatalf("Unmarshal returned %d, want 5", *v)
	}

	// Read the same identifier back as a *string.
	if _, _, err = stringSerializer.Unmarshal([]byte{byte(binco.PtrMapping), 0}); !errors.Is(err, binco.ErrWrongType) {
		t.Fatalf("Unmarshal of a cross-typed reference returned %v, want ErrWrongType", err)
	}
}

// inner is a struct whose first field shares the struct's address.
type inner struct{ Value int64 }

type innerValueSer struct{}

func (innerValueSer) Marshal(v inner, bs []byte) (n int) {
	return varnum.Int64.Marshal(v.Value, bs)
}

func (innerValueSer) Unmarshal(bs []byte) (v inner, n int, err error) {
	v.Value, n, err = varnum.Int64.Unmarshal(bs)
	return
}

func (innerValueSer) Size(v inner) (size int) {
	return varnum.Int64.Size(v.Value)
}

func (innerValueSer) Skip(bs []byte) (n int, err error) {
	return varnum.Int64.Skip(bs)
}

// TestSameAddressDifferentTypes covers the encode side of S3: a struct and its
// first field share an address and used to receive the same identifier, so the
// second value was written as a reference to the wrong type.
func TestSameAddressDifferentTypes(t *testing.T) {
	ptrMap := binco.NewPtrMap()
	revPtrMap := binco.NewReversePtrMap()

	innerSerializer := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, innerValueSer{})
	intSerializer := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, varnum.Int64)

	o := &inner{Value: 7}
	binco.Encode(innerSerializer, o)
	binco.Encode(intSerializer, &o.Value)

	if ptrMap.Len() != 2 {
		t.Fatalf("the pointer map holds %d entries, want 2 (address and type are both part of the key)",
			ptrMap.Len())
	}
}

// TestPlainPointerDepthIsBounded covers S4 for the plain pointer codec: a
// payload of nested PtrNotNil values used to recurse until the goroutine stack
// died. The node type and its serializer come from pointer_test.go.
func TestPlainPointerDepthIsBounded(t *testing.T) {
	ns := &nodeSerializer{}
	ns.next = composite.NewPtrSerializer[node](ns)

	// Each node costs one PtrNotNil byte and one value byte.
	deep := make([]byte, 2*(binco.MaxPointerDepth+2))
	if _, _, err := ns.next.Unmarshal(deep); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Unmarshal of a deeply nested value returned %v, want ErrTooDeep", err)
	}
	if _, err := ns.next.Skip(deep); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Skip of a deeply nested value returned %v, want ErrTooDeep", err)
	}
}

// TestReferenceDepthIsBounded covers S4 for the shared-pointer codec, whose
// depth is counted per decoded value on the reverse map.
func TestReferenceDepthIsBounded(t *testing.T) {
	ptrMap := binco.NewPtrMap()
	revPtrMap := binco.NewReversePtrMap()

	ns := &nodeSerializer{}
	ns.next = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, ns)

	// Each level is a PtrMapping flag, a fresh identifier and a value byte.
	depth := binco.MaxDepth + 2
	bs := make([]byte, 0, depth*4)
	for i := range depth {
		bs = append(bs, byte(binco.PtrMapping))
		bs = append(bs, binco.Encode(varnum.PositiveInt, i)...)
		bs = append(bs, 0)
	}
	if _, _, err := ns.next.Unmarshal(bs); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Unmarshal of %d nested references returned %v, want ErrTooDeep", depth, err)
	}
}
