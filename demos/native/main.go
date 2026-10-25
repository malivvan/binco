// Command native shows the zero-copy codecs: decoded strings and slices point into the input buffer, so they must be copied when they outlive it.
package main

import (
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/zerocpy"
)

func main() {

	value := "hello"
	fmt.Printf("zerocpy.String: %v\n", binco.Encode(zerocpy.String, value))
	fmt.Printf("composite.String:    %v\n\n", binco.Encode(composite.String, value))

	messages := []string{"first", "second"}
	buf := make([]byte, 64)

	var (
		aliased []string
		copied  []string
	)
	for _, msg := range messages {

		n := copy(buf, binco.Encode(zerocpy.String, msg))

		str, _, err := zerocpy.String.Unmarshal(buf[:n])
		if err != nil {
			panic(err)
		}

		fmt.Printf("processing %q\n", str)

		aliased = append(aliased, str)

		copied = append(copied, strings.Clone(str))
	}
	fmt.Printf("\naliased strings after the loop: %q\n", aliased)
	fmt.Printf("copied strings after the loop:  %q\n", copied)

	payload := binco.Encode(zerocpy.ByteSlice, []byte{1, 2, 3})
	data, _, err := zerocpy.ByteSlice.Unmarshal(payload)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\ndecoded bytes: %v\n", data)
	payload[1] = 9
	fmt.Printf("after modifying the payload: %v (the decoded slice changed)\n", data)

	safe, _, err := composite.ByteSlice.Unmarshal(payload)
	if err != nil {
		panic(err)
	}
	payload[1] = 7
	fmt.Printf("composite.ByteSlice copied %v and is unaffected by the payload change to %v\n",
		safe, payload)

	parts := make([]string, 0, 4)
	for _, v := range []uint64{0, 40, 1 << 40, ^uint64(0)} {
		bs := binco.Encode(zerocpy.Uint64, v)
		got, _, err := zerocpy.Uint64.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		parts = append(parts, fmt.Sprintf("%d (%d bytes)", got, len(bs)))
	}
	fmt.Printf("zerocpy.Uint64 round trips: %s\n", strings.Join(parts, ", "))

	arr := [3]int32{7, 8, 9}
	arrSerializer := zerocpy.NewArraySerializer[[3]int32, int32](scalar.Int32)
	arrBS := binco.Encode(arrSerializer, arr)
	gotArr, _, err := arrSerializer.Unmarshal(arrBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("native array: %v → %v (%d bytes)\n", arr, gotArr, len(arrBS))
}
