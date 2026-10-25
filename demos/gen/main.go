//go:generate go run ./gen

// Command gen uses the serializers generated into pkg.
package main

import (
	"fmt"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/demos/gen/pkg"
)

func main() {

	num := pkg.Int(300)
	fmt.Printf("Int:     %v (varint)\n", binco.Encode(pkg.IntBinco, num))

	foo := pkg.Foo[pkg.Int]{S: "hello world", T: 5}
	bs := make([]byte, pkg.FooBinco.Size(foo))
	n := pkg.FooBinco.Marshal(foo, bs)
	fmt.Printf("Foo:     %v (%d bytes)\n", bs, n)

	back, n, err := pkg.FooBinco.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v (read %d bytes)\n", back, n)

	n, err = pkg.FooBinco.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)

	var ser binco.Serializer[pkg.Int] = pkg.IntBinco
	fmt.Printf("through binco.Serializer[int]: size=%d\n", ser.Size(num))
}
