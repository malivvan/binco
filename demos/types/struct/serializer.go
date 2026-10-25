// FooSer is a hand-written codec; the generated code in examples/gen has the same shape.
package main

import (
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

var (
	FooSer  = fooSer{}
	intsSer = composite.NewSliceSer(varnum.Int)
)

type fooSer struct{}

func (s fooSer) Marshal(v Foo, bs []byte) (n int) {

	n = composite.String.Marshal(v.Str, bs)

	return n + intsSer.Marshal(v.Sl, bs[n:])
}

func (s fooSer) Unmarshal(bs []byte) (v Foo, n int, err error) {

	v.Str, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}

	var n1 int
	v.Sl, n1, err = intsSer.Unmarshal(bs[n:])

	n += n1
	return
}

func (s fooSer) Size(v Foo) (size int) {
	return composite.String.Size(v.Str) + intsSer.Size(v.Sl)
}

func (s fooSer) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = intsSer.Skip(bs[n:])
	return n + n1, err
}

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSer.Size(f)
}
