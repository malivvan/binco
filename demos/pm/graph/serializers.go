package main

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

func MakeGraphSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valSerializer binco.Serializer[V],
) binco.Serializer[Graph[T, V]] {
	ptrMap, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	return ptrmap.Wrap(ptrMap, revPtrMap, newGraphSerializer(ptrMap, revPtrMap, keySerializer, valSerializer))
}

func newGraphSerializer[T comparable, V any](ptrMap *binco.PtrMap,
	revPtrMap *binco.ReversePtrMap, keySerializer binco.Serializer[T], valSerializer binco.Serializer[V],
) graphSerializer[T, V] {
	var (
		edgesSerializer    binco.Serializer[map[T]*Edge[T, V]]
		verticesSerializer binco.Serializer[map[T]*Vertex[T, V]]

		edgePtrSerializer   binco.Serializer[*Edge[T, V]]
		vertexPtrSerializer binco.Serializer[*Vertex[T, V]]

		edge   = edgeSerializer[T, V]{vertexPtrSerializer: &vertexPtrSerializer}
		vertex = vertexSerializer[T, V]{valSerializer: valSerializer, edgesSerializer: &edgesSerializer}
	)
	edgePtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, edge)
	vertexPtrSerializer = ptrmap.NewPtrSerializer(ptrMap, revPtrMap, vertex)

	edgesSerializer = composite.NewMapSerializer(keySerializer, edgePtrSerializer)
	verticesSerializer = composite.NewMapSerializer(keySerializer, vertexPtrSerializer)

	return graphSerializer[T, V]{verticesSerializer: verticesSerializer}
}

type edgeSerializer[T comparable, V any] struct {
	vertexPtrSerializer *binco.Serializer[*Vertex[T, V]]
}

func (s edgeSerializer[T, V]) Marshal(v Edge[T, V], bs []byte) (n int) {
	n = varnum.Int.Marshal(v.Weight, bs)
	return n + (*s.vertexPtrSerializer).Marshal(v.Vertex, bs[n:])
}

func (s edgeSerializer[T, V]) Unmarshal(bs []byte) (v Edge[T, V], n int, err error) {
	v.Weight, n, err = varnum.Int.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Vertex, n1, err = (*s.vertexPtrSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s edgeSerializer[T, V]) Size(v Edge[T, V]) (size int) {
	return varnum.Int.Size(v.Weight) + (*s.vertexPtrSerializer).Size(v.Vertex)
}

func (s edgeSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = varnum.Int.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.vertexPtrSerializer).Skip(bs[n:])
	return n + n1, err
}

type vertexSerializer[T comparable, V any] struct {
	valSerializer   binco.Serializer[V]
	edgesSerializer *binco.Serializer[map[T]*Edge[T, V]]
}

func (s vertexSerializer[T, V]) Marshal(v Vertex[T, V], bs []byte) (n int) {
	n = s.valSerializer.Marshal(v.Val, bs)
	return n + (*s.edgesSerializer).Marshal(v.Edges, bs[n:])
}

func (s vertexSerializer[T, V]) Unmarshal(bs []byte) (v Vertex[T, V], n int, err error) {
	v.Val, n, err = s.valSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	var n1 int
	v.Edges, n1, err = (*s.edgesSerializer).Unmarshal(bs[n:])
	n += n1
	return
}

func (s vertexSerializer[T, V]) Size(v Vertex[T, V]) (size int) {
	return s.valSerializer.Size(v.Val) + (*s.edgesSerializer).Size(v.Edges)
}

func (s vertexSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	n, err = s.valSerializer.Skip(bs)
	if err != nil {
		return
	}
	var n1 int
	n1, err = (*s.edgesSerializer).Skip(bs[n:])
	return n + n1, err
}

type graphSerializer[T comparable, V any] struct {
	verticesSerializer binco.Serializer[map[T]*Vertex[T, V]]
}

func (s graphSerializer[T, V]) Marshal(v Graph[T, V], bs []byte) (n int) {
	return s.verticesSerializer.Marshal(v.Vertices, bs)
}

func (s graphSerializer[T, V]) Unmarshal(bs []byte) (v Graph[T, V], n int, err error) {
	v = Graph[T, V]{}
	v.Vertices, n, err = s.verticesSerializer.Unmarshal(bs)
	return
}

func (s graphSerializer[T, V]) Size(v Graph[T, V]) (size int) {
	return s.verticesSerializer.Size(v.Vertices)
}

func (s graphSerializer[T, V]) Skip(bs []byte) (n int, err error) {
	return s.verticesSerializer.Skip(bs)
}
