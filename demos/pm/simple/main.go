// Command simple shows how pointer mapping keeps one pointer shared by several fields.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	two := NewTwoPtrs("the same pointer in two fields")
	bs := binco.Encode(MakeTwoPtrsSerializer(), two)
	fmt.Printf("TwoPtrs: %v\n", bs)
	fmt.Printf("  [2 0 ...] flag 2 (mapping), ID 0, then the string\n")
	fmt.Printf("  [2 0]     the same ID again, no string\n")

	back := decodeTwoPtrs(bs)
	fmt.Printf("  decoded: %q, %q\n", *back.Ptr1, *back.Ptr2)
	fmt.Printf("  identity preserved: %v\n\n", back.Ptr1 == back.Ptr2)

	embedded := NewEmbeddedPtrs("the same pointer in a composite struct")
	bs = binco.Encode(MakeEmbeddedPtrsSerializer(), embedded)
	fmt.Printf("EmbeddedPtrs: %v (%d bytes)\n", bs, len(bs))

	backE := decodeEmbeddedPtrs(bs)
	fmt.Printf("  decoded: %q, %q, %q\n", *backE.Ptr1, *backE.Ptr2, *backE.Ptr3)
	fmt.Printf("  identity preserved: %v (all three are the same pointer)\n",
		backE.Ptr1 == backE.Ptr2 && backE.Ptr1 == backE.Ptr3)

	var nilPtrs TwoPtrs
	bs = binco.Encode(MakeTwoPtrsSerializer(), nilPtrs)
	fmt.Printf("\nnil pointers: %v\n", bs)
	back = decodeTwoPtrs(bs)
	fmt.Printf("  decoded nil: %v, %v\n", back.Ptr1 == nil, back.Ptr2 == nil)
}

func NewTwoPtrs(s string) TwoPtrs {
	ptr := &s
	return TwoPtrs{Ptr1: ptr, Ptr2: ptr}
}

func NewEmbeddedPtrs(s string) EmbeddedPtrs {
	ptr := &s
	return EmbeddedPtrs{
		Ptr1: ptr,
		Ptr2: ptr,
		Ptr3: ptr,
	}
}

func decodeTwoPtrs(bs []byte) TwoPtrs {
	v, err := binco.Decode(MakeTwoPtrsSerializer(), bs)
	if err != nil {
		panic(err)
	}
	return v
}

func decodeEmbeddedPtrs(bs []byte) EmbeddedPtrs {
	v, err := binco.Decode(MakeEmbeddedPtrsSerializer(), bs)
	if err != nil {
		panic(err)
	}
	return v
}
