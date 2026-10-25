// FooSerializer is a hand-written codec; the generated code in examples/gen has the same shape.
package main

import (
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

var (
	FooSerializer  = fooSerializer{}
	intsSerializer = composite.NewSliceSerializer(varnum.Int)
)

type fooSerializer struct{}

func (s fooSerializer) Marshal(v Foo, bs []byte) (n int) {

	n = composite.String.Marshal(v.Str, bs)

	return n + intsSerializer.Marshal(v.Sl, bs[n:])
}

func (s fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {

	v.Str, n, err = composite.String.Unmarshal(bs)
	if err != nil {
		return
	}

	var n1 int
	v.Sl, n1, err = intsSerializer.Unmarshal(bs[n:])

	n += n1
	return
}

func (s fooSerializer) Size(v Foo) (size int) {
	return composite.String.Size(v.Str) + intsSerializer.Size(v.Sl)
}

func (s fooSerializer) Skip(bs []byte) (n int, err error) {
	n, err = composite.String.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = intsSerializer.Skip(bs[n:])
	return n + n1, err
}

func (f Foo) MarshalBinco(bs []byte) (n int) {
	return FooSerializer.Marshal(f, bs)
}

func (f Foo) SizeBinco() (size int) {
	return FooSerializer.Size(f)
}
