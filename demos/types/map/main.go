// Command map encodes a map and decodes it back. Entry order is Go's map order, so only decoding is deterministic.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

func main() {
	var (
		value = map[int]string{1: "one", 2: "two", 3: "three"}

		ser = composite.NewMapSer(varnum.Int, composite.String)
	)

	bs := make([]byte, ser.Size(value))
	n := ser.Marshal(value, bs)
	fmt.Printf("value:    %v\n", sorted(value))
	fmt.Printf("size:     %d bytes\n", len(bs))
	fmt.Printf("encoded:  %v (wrote %d)\n", bs, n)

	back, n, err := ser.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded:  %v (read %d)\n", sorted(back), n)

	n, err = ser.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped:  %d bytes\n", n)

	empty, err := binco.Decode(ser, binco.Encode(ser, map[int]string(nil)))
	if err != nil {
		panic(err)
	}
	fmt.Printf("empty:    len=%d, nil=%v\n", len(empty), empty == nil)
}

func sorted(m map[int]string) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d:%q", k, m[k]))
	}
	return "{" + strings.Join(out, ", ") + "}"
}
