package main

import (
	"fmt"
	"strconv"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	FooV1DTM binco.DTM = iota + 1
	FooV2DTM
)

var (
	FooV1Ser = fooV1Ser{}
	FooV2Ser = fooV2Ser{}

	FooV1TypedSer = typeid.NewSer[FooV1](FooV1DTM, FooV1Ser)
	FooV2TypedSer = typeid.NewSer[FooV2](FooV2DTM, FooV2Ser)

	FooSer = fooSer{}
)

type fooV1Ser struct{}

func (fooV1Ser) Marshal(v FooV1, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooV1Ser) Unmarshal(bs []byte) (v FooV1, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooV1Ser) Size(v FooV1) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooV1Ser) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type fooV2Ser struct{}

func (fooV2Ser) Marshal(v FooV2, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (fooV2Ser) Unmarshal(bs []byte) (v FooV2, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (fooV2Ser) Size(v FooV2) (size int) {
	return composite.String.Size(v.Str)
}

func (fooV2Ser) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type fooSer struct{}

func (fooSer) Marshal(v Foo, bs []byte) (n int) {
	return FooV2TypedSer.Marshal(FooV2(v), bs)
}

func (fooSer) Size(v Foo) (size int) {
	return FooV2TypedSer.Size(FooV2(v))
}

func (fooSer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	dtm, n, err := typeid.DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		var old FooV1
		old, n1, err = FooV1TypedSer.UnmarshalData(bs[n:])
		if err != nil {
			return
		}
		v = migrateFooV1(old)
	case FooV2DTM:
		var cur FooV2
		cur, n1, err = FooV2TypedSer.UnmarshalData(bs[n:])
		v = Foo(cur)
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (fooSer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		n1, err = FooV1TypedSer.SkipData(bs[n:])
	case FooV2DTM:
		n1, err = FooV2TypedSer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	return n + n1, err
}

func migrateFooV1(v FooV1) Foo {
	return Foo(FooV2{Str: strconv.Itoa(v.Num)})
}
