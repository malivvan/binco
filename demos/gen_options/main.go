//go:generate go run ./gen

// Command gen_options uses a serializer generated with a custom name, a fixed-width number codec and a length validator.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/demos/gen_options/pkg"
)

func main() {

	foo := pkg.Foo{Num: 1, Str: "hello"}
	bs := binco.Encode(pkg.CustomFooBinco, foo)
	fmt.Printf("value:   %+v\n", foo)
	fmt.Printf("size:    %d bytes (Num alone takes 8: raw, not varint)\n", len(bs))
	fmt.Printf("encoded: %v\n", bs)

	back, err := binco.Decode(pkg.CustomFooBinco, bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %+v\n\n", back)

	long := pkg.Foo{Num: 1, Str: strings.Repeat("a", pkg.MaxStrLen+1)}
	longBS := binco.Encode(pkg.CustomFooBinco, long)
	_, err = binco.Decode(pkg.CustomFooBinco, longBS)
	fmt.Printf("string of %d characters: %v\n", len(long.Str), err)
	fmt.Printf("errors.Is(err, pkg.ErrTooLong): %v\n", errors.Is(err, pkg.ErrTooLong))

	ok := pkg.Foo{Num: 2, Str: strings.Repeat("a", pkg.MaxStrLen)}
	back, err = binco.Decode(pkg.CustomFooBinco, binco.Encode(pkg.CustomFooBinco, ok))
	if err != nil {
		panic(err)
	}
	fmt.Printf("string of %d characters: accepted, %q…\n", len(back.Str), back.Str[:8])
}
