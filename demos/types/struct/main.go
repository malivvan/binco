// Command struct uses the hand-written FooSerializer serializer to show the shape of an encoded struct.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
)

func main() {
	v := Foo{
		Str: "hello, binco",
		Sl:  []int{1, 2, 3},
	}

	bs := make([]byte, FooSerializer.Size(v))
	n := FooSerializer.Marshal(v, bs)
	fmt.Printf("value:   %+v\n", v)
	fmt.Printf("size:    %d bytes\n", len(bs))
	fmt.Printf("encoded: %v (wrote %d)\n", bs, n)

	back, n, err := FooSerializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v (read %d)\n", back, n)

	n, err = FooSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	bs = binco.Marshal(v)
	fmt.Printf("binco.Marshal: %v\n", bs)

	if _, _, err = FooSerializer.Unmarshal(bs[:len(bs)-1]); err != nil {
		fmt.Printf("truncated payload: %v\n", err)
	}
}
