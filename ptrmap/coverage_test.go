package ptrmap_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

// chain is a recursive type: its codec delegates the next pointer to the same
// pointer codec, so skipping one value recurses once per node.
type chain struct {
	next *chain
}

type chainCodec struct {
	ptr binco.Serializer[*chain]
}

func (c *chainCodec) Marshal(v chain, bs []byte) (n int) { return c.ptr.Marshal(v.next, bs) }
func (c *chainCodec) Unmarshal(bs []byte) (v chain, n int, err error) {
	v.next, n, err = c.ptr.Unmarshal(bs)
	return
}
func (c *chainCodec) Size(v chain) (size int)           { return c.ptr.Size(v.next) }
func (c *chainCodec) Skip(bs []byte) (n int, err error) { return c.ptr.Skip(bs) }

// TestSkipBoundsTheReferenceDepth checks the nesting limit on the skip path:
// Skip enters the shared depth counter just like Unmarshal, so hostile input
// cannot recurse without bound.
func TestSkipBoundsTheReferenceDepth(t *testing.T) {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	c := &chainCodec{}
	c.ptr = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, c)

	// Each level is a mapping reference with a fresh identifier, which makes the
	// base serializer decode (and therefore recurse) at every level.
	var payload []byte
	for id := range binco.MaxDepth + 2 {
		payload = append(payload, byte(binco.PtrMapping))
		payload = append(payload, binco.Encode(varnum.PositiveInt, id)...)
	}
	if _, err := c.Skip(payload); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Skip returned %v, want ErrTooDeep", err)
	}

	// The counter is not left behind: a shallow value still skips.
	revPtrMap.Clear()
	if _, err := c.Skip([]byte{byte(binco.PtrNil)}); err != nil {
		t.Fatalf("Skip after ErrTooDeep = %v", err)
	}
}
