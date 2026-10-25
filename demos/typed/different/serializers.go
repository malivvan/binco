package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	FooDTM binco.DTM = iota + 1
	BarDTM
)

var (
	FooSer = fooSer{}
	BarSer = barSer{}
)

var (
	FooTypedSer = typeid.NewSer[Foo](FooDTM, FooSer)
	BarTypedSer = typeid.NewSer[Bar](BarDTM, BarSer)
)

type fooSer struct{}

func (fooSer) Marshal(v Foo, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooSer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooSer) Size(v Foo) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooSer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type barSer struct{}

func (barSer) Marshal(v Bar, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (barSer) Unmarshal(bs []byte) (v Bar, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (barSer) Size(v Bar) (size int) {
	return composite.String.Size(v.Str)
}

func (barSer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}
