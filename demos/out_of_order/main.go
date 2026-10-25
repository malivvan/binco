// Command out_of_order computes field offsets with Skip and then decodes the fields in reverse order.
package main

import (
	"fmt"

	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

var (
	idSerializer    = varnum.Int
	nameSerializer  = composite.String
	scoreSerializer = scalar.Int32
)

func main() {
	const (
		id    = 42
		name  = "out of order"
		score = int32(1234)
	)
	bs := encode(id, name, score)
	fmt.Printf("payload: %v (%d bytes)\n\n", bs, len(bs))

	idEnd, err := idSerializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	nameEnd, err := nameSerializer.Skip(bs[idEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("offsets: ID ends at %d, Name ends at %d\n\n", idEnd, idEnd+nameEnd)

	gotScore, _, err := scoreSerializer.Unmarshal(bs[idEnd+nameEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("score: %d\n", gotScore)

	gotName, _, err := nameSerializer.Unmarshal(bs[idEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("name:  %q\n", gotName)

	gotID, _, err := idSerializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("id:    %d\n", gotID)

	if _, err = idSerializer.Skip(bs); err != nil {
		panic(err)
	}
	if _, err = nameSerializer.Skip(bs[idEnd:]); err != nil {
		panic(err)
	}
	onlyScore, _, err := scoreSerializer.Unmarshal(bs[idEnd+nameEnd:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nonly the score: %d\n", onlyScore)

	if _, err = nameSerializer.Skip(bs[idEnd : idEnd+5]); err != nil {
		fmt.Printf("truncated payload: %v\n", err)
	}
}

func encode(id int, name string, score int32) []byte {
	bs := make([]byte, idSerializer.Size(id)+nameSerializer.Size(name)+scoreSerializer.Size(score))
	n := idSerializer.Marshal(id, bs)
	n += nameSerializer.Marshal(name, bs[n:])
	scoreSerializer.Marshal(score, bs[n:])
	return bs
}
