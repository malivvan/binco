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
	CopySer   = copySer{}
	InsertSer = insertSer{}

	CopyTypedSer   = typeid.NewSer[Copy](CopyDTM, CopySer)
	InsertTypedSer = typeid.NewSer[Insert](InsertDTM, InsertSer)

	InstructionSer = instructionSer{}
)

type copySer struct{}

func (copySer) Marshal(v Copy, bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Start, bs)
	return n + varnum.Int.Marshal(v.End, bs[n:])
}

func (copySer) Unmarshal(bs []byte) (v Copy, n int, err error) {
	v.Start, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.End, n1, err = varnum.Int.Unmarshal(bs[n:])
	n += n1
	return
}

func (copySer) Size(v Copy) (size int) {
	return varnum.Int.Size(v.Start) + varnum.Int.Size(v.End)
}

func (copySer) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = varnum.Int.Skip(bs[n:])
	return n + n1, err
}

type insertSer struct{}

func (insertSer) Marshal(v Insert, bs []byte) (n int) {
	return composite.String.Marshal(v.Str, bs)
}

func (insertSer) Unmarshal(bs []byte) (v Insert, n int, err error) {
	v.Str, n, err = composite.String.Unmarshal(bs)
	return
}

func (insertSer) Size(v Insert) (size int) {
	return composite.String.Size(v.Str)
}

func (insertSer) Skip(bs []byte) (n int, err error) {
	return composite.String.Skip(bs)
}

type instructionSer struct{}

func (instructionSer) Marshal(v Instruction, bs []byte) (n int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.MarshalTypedBinco(bs)
}

func (instructionSer) Size(v Instruction) (size int) {
	m, ok := v.(binco.MarshallerTyped)
	if !ok {
		panic(fmt.Sprintf("%v doesn't implement binco.MarshallerTyped", reflect.TypeOf(v)))
	}
	return m.SizeTypedBinco()
}

func (instructionSer) Unmarshal(bs []byte) (v Instruction, n int, err error) {
	dtm, n, err := typeid.DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		v, n1, err = CopyTypedSer.UnmarshalData(bs[n:])
	case InsertDTM:
		v, n1, err = InsertTypedSer.UnmarshalData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}

func (instructionSer) Skip(bs []byte) (n int, err error) {
	dtm, n, err := typeid.DTMSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	switch dtm {
	case CopyDTM:
		n1, err = CopyTypedSer.SkipData(bs[n:])
	case InsertDTM:
		n1, err = InsertTypedSer.SkipData(bs[n:])
	default:
		err = fmt.Errorf("unexpected instruction: %w", binco.NewUnexpectedDTMError(dtm))
	}
	n += n1
	return
}
