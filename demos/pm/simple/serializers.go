package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
)

func MakeTwoPtrsSer() binco.Serializer[TwoPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newTwoPtrsSer(ptrMap, revPtrMap))
}

func MakeEmbeddedPtrsSer() binco.Serializer[EmbeddedPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newEmbeddedPtrsSer(ptrMap, revPtrMap))
}

func newTwoPtrsSer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) twoPtrsSer {
	return twoPtrsSer{strPtrSer: ptrmap.NewPtrSer(ptrMap, revPtrMap, composite.String)}
}

func newEmbeddedPtrsSer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) embeddedPtrsSer {
	return embeddedPtrsSer{

		twoSer:  newTwoPtrsSer(ptrMap, revPtrMap),
		ptr3Ser: ptrmap.NewPtrSer(ptrMap, revPtrMap, composite.String),
	}
}

type twoPtrsSer struct {
	strPtrSer binco.Serializer[*string]
}

func (s twoPtrsSer) Marshal(v TwoPtrs, bs []byte) (n int) {
	n = s.strPtrSer.Marshal(v.Ptr1, bs)
	return n + s.strPtrSer.Marshal(v.Ptr2, bs[n:])
}

func (s twoPtrsSer) Unmarshal(bs []byte) (v TwoPtrs, n int, err error) {
	v.Ptr1, n, err = s.strPtrSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr2, n1, err = s.strPtrSer.Unmarshal(bs[n:])
	n += n1
	return
}

func (s twoPtrsSer) Size(v TwoPtrs) (size int) {
	return s.strPtrSer.Size(v.Ptr1) + s.strPtrSer.Size(v.Ptr2)
}

func (s twoPtrsSer) Skip(bs []byte) (n int, err error) {
	n, err = s.strPtrSer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.strPtrSer.Skip(bs[n:])
	return n + n1, err
}

type embeddedPtrsSer struct {
	twoSer  twoPtrsSer
	ptr3Ser binco.Serializer[*string]
}

func (s embeddedPtrsSer) Marshal(v EmbeddedPtrs, bs []byte) (n int) {
	n = s.twoSer.Marshal(v.TwoPtrs, bs)
	return n + s.ptr3Ser.Marshal(v.Ptr3, bs[n:])
}

func (s embeddedPtrsSer) Unmarshal(bs []byte) (v EmbeddedPtrs, n int, err error) {
	v.TwoPtrs, n, err = s.twoSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr3, n1, err = s.ptr3Ser.Unmarshal(bs[n:])
	n += n1
	return
}

func (s embeddedPtrsSer) Size(v EmbeddedPtrs) (size int) {
	return s.twoSer.Size(v.TwoPtrs) + s.ptr3Ser.Size(v.Ptr3)
}

func (s embeddedPtrsSer) Skip(bs []byte) (n int, err error) {
	n, err = s.twoSer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.ptr3Ser.Skip(bs[n:])
	return n + n1, err
}
