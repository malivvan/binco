package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeLinkedListSerializer[T any](valSerializer binco.Serializer[T]) binco.Serializer[LinkedList[T]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newLinkedListSerializer(ptrMap, revPtrMap, valSerializer))
}

func newLinkedListSerializer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	valSerializer binco.Serializer[T],
) linkedListSerializer[T] {
	var (
		elem              elemSerializer[T]
		elemPtrSerializer binco.Serializer[*Elem[T]]
	)

	elem = elemSerializer[T]{valSerializer: valSerializer, elemPtrSerializer: &elemPtrSerializer}
	elemPtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, elem)

	return linkedListSerializer[T]{elemPtrSerializer: elemPtrSerializer}
}

type elemSerializer[T any] struct {
	valSerializer     binco.Serializer[T]
	elemPtrSerializer *binco.Serializer[*Elem[T]]
}

func (s elemSerializer[T]) Marshal(v Elem[T], bs []byte) (n int) {
	n = s.valSerializer.Marshal(v.Val, bs)
	n += (*s.elemPtrSerializer).Marshal(v.Prev, bs[n:])
	return n + (*s.elemPtrSerializer).Marshal(v.Next, bs[n:])
}

func (s elemSerializer[T]) Unmarshal(bs []byte) (v Elem[T], n int, err error) {
	v.Val, n, err = s.valSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Prev, n1, err = (*s.elemPtrSerializer).Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Next, n1, err = (*s.elemPtrSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s elemSerializer[T]) Size(v Elem[T]) (size int) {
	return s.valSerializer.Size(v.Val) +
		(*s.elemPtrSerializer).Size(v.Prev) +
		(*s.elemPtrSerializer).Size(v.Next)
}

func (s elemSerializer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.elemPtrSerializer).Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = (*s.elemPtrSerializer).Skip(bs[n:])
	return n + n1, err
}

type linkedListSerializer[T any] struct {
	elemPtrSerializer binco.Serializer[*Elem[T]]
}

func (s linkedListSerializer[T]) Marshal(v LinkedList[T], bs []byte) (n int) {
	n = s.elemPtrSerializer.Marshal(v.Head, bs)
	n += s.elemPtrSerializer.Marshal(v.Tail, bs[n:])
	return n + varnum.PositiveInt.Marshal(v.Len, bs[n:])
}

func (s linkedListSerializer[T]) Unmarshal(bs []byte) (v LinkedList[T], n int, err error) {
	v.Head, n, err = s.elemPtrSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Tail, n1, err = s.elemPtrSerializer.Unmarshal(bs[n:])
	n += n1
	if err != nil {
		return
	}
	v.Len, n1, err = varnum.PositiveInt.Unmarshal(bs[n:])
	n += n1
	return
}

func (s linkedListSerializer[T]) Size(v LinkedList[T]) (size int) {
	return s.elemPtrSerializer.Size(v.Head) +
		s.elemPtrSerializer.Size(v.Tail) +
		varnum.PositiveInt.Size(v.Len)
}

func (s linkedListSerializer[T]) Skip(bs []byte) (n int, err error) {
	n, err = s.elemPtrSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = s.elemPtrSerializer.Skip(bs[n:])
	n += n1
	if err != nil {
		return
	}
	n1, err = varnum.PositiveInt.Skip(bs[n:])
	return n + n1, err
}
