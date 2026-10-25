// The implementations carry their own marshalling, so the interface codec never names them.
package main

import "fmt"

type Instruction interface {
	Do()
}

type Copy struct {
	Start int
	End   int
}

func (c Copy) Do() {
	fmt.Printf("copy from %d to %d\n", c.Start, c.End)
}

func (c Copy) MarshalTypedBinco(bs []byte) (n int) {
	return CopyTypedSerializer.Marshal(c, bs)
}

func (c Copy) SizeTypedBinco() (size int) {
	return CopyTypedSerializer.Size(c)
}

type Insert struct {
	Str string
}

func (i Insert) Do() {
	fmt.Printf("insert %q\n", i.Str)
}

func (i Insert) MarshalTypedBinco(bs []byte) (n int) {
	return InsertTypedSerializer.Marshal(i, bs)
}

func (i Insert) SizeTypedBinco() (size int) {
	return InsertTypedSerializer.Size(i)
}
