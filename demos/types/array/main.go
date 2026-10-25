// Command array encodes an array, decodes it, skips it and shows that a payload whose length differs from the array type is rejected.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/zerocopy"
)

func main() {
	var (
		value      = [3]int{1, 2, 3}
		serializer = zerocopy.NewArraySerializer[[3]int, int](varnum.Int)
	)

	size := serializer.Size(value)
	bs := make([]byte, size)
	n := serializer.Marshal(value, bs)

	fmt.Printf("value:   %v\n", value)
	fmt.Printf("size:    %d bytes\n", size)
	fmt.Printf("encoded: %v (wrote %d)\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %v (read %d)\n", back, n)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	shorter := binco.Encode(composite.NewSliceSerializer(varnum.Int), []int{1, 2})
	if _, _, err = serializer.Unmarshal(shorter); err != nil {
		fmt.Printf("shorter payload: %v\n", err)
	}
}
