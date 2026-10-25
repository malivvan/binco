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
	FooV1Serializer = fooV1Serializer{}
	FooV2Serializer = fooV2Serializer{}

	FooV1TypedSerializer = typeid.NewSerializer[FooV1](FooV1DTM, FooV1Serializer)
	FooV2TypedSerializer = typeid.NewSerializer[FooV2](FooV2DTM, FooV2Serializer)

	FooSerializer = fooSerializer{}
)

type fooV1Serializer struct{}

func (fooV1Serializer) Marshal(v FooV1, bs []byte) (n int) {
	return varnum.Int.Marshal(v.Num, bs)
}

func (fooV1Serializer) Unmarshal(bs []byte) (v FooV1, n int, err error) {
	v.Num, n, err = varnum.Int.Unmarshal(bs)
	return
}

func (fooV1Serializer) Size(v FooV1) (size int) {
	return varnum.Int.Size(v.Num)
}

func (fooV1Serializer) Skip(bs []byte) (n int, err error) {
	return varnum.Int.Skip(bs)
}

type fooV2Serializer struct{}

func (fooV2Serializer) Marshal(v FooV2, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (fooV2Serializer) Unmarshal(bs []byte) (v FooV2, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (fooV2Serializer) Size(v FooV2) (size int) {
	return composite.String.Size(v.Str)
}

func (fooV2Serializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type fooSerializer struct{}

func (fooSerializer) Marshal(v Foo, bs []byte) (n int) {
	return FooV2TypedSerializer.Marshal(FooV2(v), bs)
}

func (fooSerializer) Size(v Foo) (size int) {
	return FooV2TypedSerializer.Size(FooV2(v))
}

func (fooSerializer) Unmarshal(bs []byte) (v Foo, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		var old FooV1
		old, n1, err = FooV1TypedSerializer.UnmarshalData(bs[n:])
		if err != nil {
			return
		}
		v = migrateFooV1(old)
	case FooV2DTM:
		var cur FooV2
		cur, n1, err = FooV2TypedSerializer.UnmarshalData(bs[n:])
		v = Foo(cur)
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (fooSerializer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case FooV1DTM:
		n1, err = FooV1TypedSerializer.SkipData(bs[n:])
	case FooV2DTM:
		n1, err = FooV2TypedSerializer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected version: %w", binco.NewUnexpectedDTMError(dtm))
	}
	return n + n1, err
}

func migrateFooV1(v FooV1) Foo {
	return Foo(FooV2{Str: strconv.Itoa(v.Num)})
}
