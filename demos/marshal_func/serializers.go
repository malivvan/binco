package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const FooDTM binco.DTM = 1

var (
	FooSerializer = fooSerializer{}
	BarSerializer = barSerializer{}

	FooTypedSerializer = typeid.NewSerializer[Foo](FooDTM, FooSerializer)
)

type fooSerializer struct{}

func (fooSerializer) Marshal(v Foo, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooSerializer) Size(v Foo) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooSerializer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type barSerializer struct{}

func (barSerializer) Marshal(v Bar, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (barSerializer) Unmarshal(bs []byte) (v Bar, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (barSerializer) Size(v Bar) (size int) {
	return composite.String.Size(v.Str)
}

func (barSerializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}
