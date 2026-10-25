package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
)

func MakeTwoPtrsSerializer() binco.Serializer[TwoPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newTwoPtrsSerializer(ptrMap, revPtrMap))
}

func MakeEmbeddedPtrsSerializer() binco.Serializer[EmbeddedPtrs] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newEmbeddedPtrsSerializer(ptrMap, revPtrMap))
}

func newTwoPtrsSerializer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) twoPtrsSerializer {
	return twoPtrsSerializer{strPtrSerializer: ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String)}
}

func newEmbeddedPtrsSerializer(ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap) embeddedPtrsSerializer {
	return embeddedPtrsSerializer{

		twoSerializer:  newTwoPtrsSerializer(ptrMap, revPtrMap),
		ptr3Serializer: ptrmap.NewPtrSerializer(ptrMap, revPtrMap, composite.String),
	}
}

type twoPtrsSerializer struct {
	strPtrSerializer binco.Serializer[*string]
}

func (s twoPtrsSerializer) Marshal(v TwoPtrs, bs []byte) (n int) {
	n = s.strPtrSerializer.Marshal(v.Ptr1, bs)
	return n + s.strPtrSerializer.Marshal(v.Ptr2, bs[n:])
}

func (s twoPtrsSerializer) Unmarshal(bs []byte) (v TwoPtrs, n int, err error) {
	v.Ptr1, n, err = s.strPtrSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr2, n1, err = s.strPtrSerializer.Unmarshal(bs[n:])
	n += n1
	return
}

func (s twoPtrsSerializer) Size(v TwoPtrs) (size int) {
	return s.strPtrSerializer.Size(v.Ptr1) + s.strPtrSerializer.Size(v.Ptr2)
}

func (s twoPtrsSerializer) Skip(bs []byte) (n int, err error) {
	n, err = s.strPtrSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.strPtrSerializer.Skip(bs[n:])
	return n + n1, err
}

type embeddedPtrsSerializer struct {
	twoSerializer  twoPtrsSerializer
	ptr3Serializer binco.Serializer[*string]
}

func (s embeddedPtrsSerializer) Marshal(v EmbeddedPtrs, bs []byte) (n int) {
	n = s.twoSerializer.Marshal(v.TwoPtrs, bs)
	return n + s.ptr3Serializer.Marshal(v.Ptr3, bs[n:])
}

func (s embeddedPtrsSerializer) Unmarshal(bs []byte) (v EmbeddedPtrs, n int, err error) {
	v.TwoPtrs, n, err = s.twoSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Ptr3, n1, err = s.ptr3Serializer.Unmarshal(bs[n:])
	n += n1
	return
}

func (s embeddedPtrsSerializer) Size(v EmbeddedPtrs) (size int) {
	return s.twoSerializer.Size(v.TwoPtrs) + s.ptr3Serializer.Size(v.Ptr3)
}

func (s embeddedPtrsSerializer) Skip(bs []byte) (n int, err error) {
	n, err = s.twoSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.ptr3Serializer.Skip(bs[n:])
	return n + n1, err
}
