package main

import (
	"fmt"
	"reflect"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/typeid"
	"github.com/malivvan/binco/varnum"
)

const (
	CopyDTM binco.DTM = iota + 1
	InsertDTM
)

var (
	CopySerializer   = copySerializer{}
	InsertSerializer = insertSerializer{}

	CopyTypedSerializer   = typeid.NewSerializer[Copy](CopyDTM, CopySerializer)
	InsertTypedSerializer = typeid.NewSerializer[Insert](InsertDTM, InsertSerializer)

	InstructionSerializer = instructionSerializer{}
)

type copySerializer struct{}

func (copySerializer) Marshal(v Copy, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Start, bs)
	return n + varnum.Int.Marshal(v.End, bs[n:])
}

func (copySerializer) Unmarshal(bs []byte) (v Copy, n int, err error) {
	v.Start, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.End, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	return
}

func (copySerializer) Size(v Copy) (size int) {
	return varnum.Int.Size(v.Start) + varnum.Int.Size(v.End)
}

func (copySerializer) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = varnum.Int.Skip(bs[n:])
	return n + n1, err
}

type insertSerializer struct{}

func (insertSerializer) Marshal(v Insert, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (insertSerializer) Unmarshal(bs []byte) (v Insert, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (insertSerializer) Size(v Insert) (size int) {
	return composite.String.Size(v.Str)
}

func (insertSerializer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type instructionSerializer struct{}

func (instructionSerializer) Marshal(v Instruction, bs []byte) (n int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.MarshalTypedBinco(bs)
}

func (instructionSerializer) Size(v Instruction) (size int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.SizeTypedBinco()
}

func (instructionSerializer) Unmarshal(bs []byte) (v Instruction, n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		v, n1, err = CopyTypedSerializer.UnmarshalData(bs[n:])
	case InsertDTM:
		v, n1, err = InsertTypedSerializer.UnmarshalData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (instructionSerializer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		n1, err = CopyTypedSerializer.SkipData(bs[n:])
	case InsertDTM:
		n1, err = InsertTypedSerializer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}
