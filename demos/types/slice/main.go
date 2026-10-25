// Command slice encodes a slice with the default length codec and with a fixed-width one, and shows how nil and empty slices behave.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

func main() {
	var (
		value = []int{1, 2, 3}

		ser = composite.NewSliceSer(varnum.Int)

		rawLenSer = composite.NewSliceSer(varnum.Int, slopts.WithLenSer[int](scalar.Int))
	)

	bs := binco.Encode(ser, value)
	fmt.Printf("value:     %v\n", value)
	fmt.Printf("encoded:   %v (%d bytes)\n", bs, len(bs))

	back, n, err := ser.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:   %v (%d bytes)\n", back, n)

	bs = binco.Encode(rawLenSer, value)
	fmt.Printf("raw len:   %v (%d bytes)\n", bs, len(bs))

	for _, v := range [][]int{nil, {}} {
		back, err = binco.Decode(ser, binco.Encode(ser, v))
		if err != nil {
			panic(err)
		}
		fmt.Printf("nil=%v → len=%d, nil=%v\n", v == nil, len(back), back == nil)
	}

	n, err = ser.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:   %d bytes\n", n)
}
