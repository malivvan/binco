package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeLinkedListSer[T any](valSer binco.Serializer[T]) binco.Serializer[LinkedList[T]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newLinkedListSer(ptrMap, revPtrMap, valSer))
}

func newLinkedListSer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	valSer binco.Serializer[T],
) linkedListSer[T] {
	var (
		elem       elemSer[T]
		elemPtrSer binco.Serializer[*Elem[T]]
	)

	elem = elemSer[T]{valSer: valSer, elemPtrSer: &elemPtrSer}
	elemPtrSer = ptrmap.NewPtrSer(ptrMap, revPtrMap, elem)

	return linkedListSer[T]{elemPtrSer: elemPtrSer}
}

type elemSer[T any] struct {
	valSer     binco.Serializer[T]
	elemPtrSer *binco.Serializer[*Elem[T]]
}

func (s elemSer[T]) Marshal(v Elem[T], bs []byte) (n int) {
	n = s.valSer.Marshal(v.Val, bs)
	n += (*s.elemPtrSer).Marshal(v.Prev, bs[n:])
	return n + (*s.elemPtrSer).Marshal(v.Next, bs[n:])
}

func (s elemSer[T]) Unmarshal(bs []byte) (v Elem[T], n int, err error) {
	v.Val, n, err = s.valSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Prev, n1, err = (*s.elemPtrSer).Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Next, n1, err = (*s.elemPtrSer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s elemSer[T]) Size(v Elem[T]) (size int) {
	return s.valSer.Size(v.Val) +
		(*s.elemPtrSer).Size(v.Prev) +
		(*s.elemPtrSer).Size(v.Next)
}

func (s elemSer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.elemPtrSer).Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = (*s.elemPtrSer).Skip(bs[n:])
	return n + n1, err
}

type linkedListSer[T any] struct {
	elemPtrSer binco.Serializer[*Elem[T]]
}

func (s linkedListSer[T]) Marshal(v LinkedList[T], bs []byte) (n int) {
	n = s.elemPtrSer.Marshal(v.Head, bs)
	n += s.elemPtrSer.Marshal(v.Tail, bs[n:])
	return n + varnum.PositiveInt.Marshal(v.Len, bs[n:])
}

func (s linkedListSer[T]) Unmarshal(bs []byte) (v LinkedList[T], n int, err error) {
	v.Head, n, err = s.elemPtrSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Tail, n1, err = s.elemPtrSer.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Len, n1, err = varnum.PositiveInt.Unmarshal(bs[n:])
	n += n1
	return
}

func (s linkedListSer[T]) Size(v LinkedList[T]) (size int) {
	return s.elemPtrSer.Size(v.Head) +
		s.elemPtrSer.Size(v.Tail) +
		varnum.PositiveInt.Size(v.Len)
}

func (s linkedListSer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.elemPtrSer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.elemPtrSer.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = varnum.PositiveInt.Skip(bs[n:])
	return n + n1, err
}
