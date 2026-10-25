// Command marshal_func compares the explicit Size/Marshal pair with the binco.Marshal and binco.MarshalTyped shorthands.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {

	foo := Foo{Num: 10}
	bs := make([]byte, FooSer.Size(foo))
	n := FooSer.Marshal(foo, bs)
	fmt.Printf("explicit:      %v (wrote %d of %d bytes)\n", bs, n, len(bs))

	fmt.Printf("binco.Marshal:  %v\n", binco.Marshal(foo))

	bar := Bar{Str: "10"}
	fmt.Println("Foo and Bar marshal through the same call:")
	for _, m := range []binco.Marshaller{foo, bar} {
		fmt.Printf("  %-10T → %v\n", m, binco.Marshal(m))
	}

	fmt.Printf("\nbinco.Marshal:      %v\n", binco.Marshal(foo))
	fmt.Printf("binco.MarshalTyped: %v\n", binco.MarshalTyped(foo))

	back, err := binco.Decode(FooSer, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\ndecoded Foo: %+v\n", back)

	typedBS := binco.MarshalTyped(foo)
	back, err = binco.Decode(FooTypedSer, typedBS)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded Foo (typed, %d bytes): %+v\n", len(typedBS), back)
}
