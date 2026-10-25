// Command versioning reads payloads written by an older version and converts them to the current type.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	v1 := FooV1{Num: 10}
	v1BS := binco.Encode(FooV1TypedSer, v1)
	fmt.Printf("V1 payload: %v (%d bytes)\n", v1BS, len(v1BS))

	v2 := FooV2{Str: "current"}
	v2BS := binco.Encode(FooV2TypedSer, v2)
	fmt.Printf("V2 payload: %v (%d bytes)\n\n", v2BS, len(v2BS))

	for _, bs := range [][]byte{v1BS, v2BS} {
		foo, n, err := FooSer.Unmarshal(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("decoded %d bytes: %q\n", n, foo)
	}

	foo := Foo(FooV2{Str: "rewritten"})
	bs := binco.Encode(FooSer, foo)
	fmt.Printf("\nmarshaled %q → %v\n", foo, bs)
	back, err := binco.Decode(FooSer, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded back: %q\n", back)

	fmt.Printf("binco.Marshal:      %v\n", binco.Marshal(foo))
	fmt.Printf("binco.MarshalTyped: %v\n", binco.MarshalTyped(v1))

	for _, bs := range [][]byte{v1BS, v2BS} {
		n, err := FooSer.Skip(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("skipped %d of %d bytes\n", n, len(bs))
	}

	unknown := append([]byte{99}, v2BS[1:]...)
	if _, _, err = FooSer.Unmarshal(unknown); err != nil {
		fmt.Printf("\nunknown version: %v\n", err)
	}
}
