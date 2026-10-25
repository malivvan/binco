package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeGraphSer[T comparable, V any](keySer binco.Serializer[T],
	valSer binco.Serializer[V],
) binco.Serializer[Graph[T, V]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newGraphSer(ptrMap, revPtrMap, keySer, valSer))
}

func newGraphSer[T comparable, V any](ptrMap *binco.PtrMap,
	revPtrMap *binco.ReversePtrMap, keySer binco.Serializer[T], valSer binco.Serializer[V],
) graphSer[T, V] {
	var (
		edgesSer    binco.Serializer[map[T]*Edge[T, V]]
		verticesSer binco.Serializer[map[T]*Vertex[T, V]]

		edgePtrSer   binco.Serializer[*Edge[T, V]]
		vertexPtrSer binco.Serializer[*Vertex[T, V]]

		edge   = edgeSer[T, V]{vertexPtrSer: &vertexPtrSer}
		vertex = vertexSer[T, V]{valSer: valSer, edgesSer: &edgesSer}
	)
	edgePtrSer = ptrmap.NewPtrSer(ptrMap, revPtrMap, edge)
	vertexPtrSer = ptrmap.NewPtrSer(ptrMap, revPtrMap, vertex)

	edgesSer = composite.NewMapSer(keySer, edgePtrSer)
	verticesSer = composite.NewMapSer(keySer, vertexPtrSer)

	return graphSer[T, V]{verticesSer: verticesSer}
}

type edgeSer[T comparable, V any] struct {
	vertexPtrSer *binco.Serializer[*Vertex[T, V]]
}

func (s edgeSer[T, V]) Marshal(v Edge[T, V], bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Weight, bs)
	return n + (*s.vertexPtrSer).Marshal(v.Vertex, bs[n:])
}

func (s edgeSer[T, V]) Unmarshal(bs []byte) (v Edge[T, V], n int, err error) {
	v.Weight, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Vertex, n1, err = (*s.vertexPtrSer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s edgeSer[T, V]) Size(v Edge[T, V]) (size int) {
	return varnum.Int.Size(v.Weight) + (*s.vertexPtrSer).Size(v.Vertex)
}

func (s edgeSer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.vertexPtrSer).Skip(bs[n:])
	return n + n1, err
}

type vertexSer[T comparable, V any] struct {
	valSer   binco.Serializer[V]
	edgesSer *binco.Serializer[map[T]*Edge[T, V]]
}

func (s vertexSer[T, V]) Marshal(v Vertex[T, V], bs []byte) (n int) {
	n = s.valSer.Marshal(v.Val, bs)
	return n + (*s.edgesSer).Marshal(v.Edges, bs[n:])
}

func (s vertexSer[T, V]) Unmarshal(bs []byte) (v Vertex[T, V], n int, err error) {
	v.Val, n, err = s.valSer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Edges, n1, err = (*s.edgesSer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s vertexSer[T, V]) Size(v Vertex[T, V]) (size int) {
	return s.valSer.Size(v.Val) + (*s.edgesSer).Size(v.Edges)
}

func (s vertexSer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.edgesSer).Skip(bs[n:])
	return n + n1, err
}

type graphSer[T comparable, V any] struct {
	verticesSer binco.Serializer[map[T]*Vertex[T, V]]
}

func (s graphSer[T, V]) Marshal(v Graph[T, V], bs []byte) (n int) {
	return s.verticesSer.Marshal(v.Vertices, bs)
}

func (s graphSer[T, V]) Unmarshal(bs []byte) (v Graph[T, V], n int, err error) {
	v = Graph[T, V]{}
	v.Vertices, n, err = s.verticesSer.Unmarshal(bs)
	return
}

func (s graphSer[T, V]) Size(v Graph[T, V]) (size int) {
	return s.verticesSer.Size(v.Vertices)
}

func (s graphSer[T, V]) Skip(bs []byte) (n int, err error) {
	return s.verticesSer.Skip(bs)
}
