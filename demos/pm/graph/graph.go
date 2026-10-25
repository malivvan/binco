package main

import (
	"fmt"
	"slices"
	"strings"
)

func NewGraph[T comparable, V any]() Graph[T, V] {
	return Graph[T, V]{Vertices: make(map[T]*Vertex[T, V])}
}

type Graph[T comparable, V any] struct {
	Vertices map[T]*Vertex[T, V]
}

func (g Graph[T, V]) AddVertex(id T, val V) {
	g.Vertices[id] = &Vertex[T, V]{Val: val}
}

func (g Graph[T, V]) AddEdge(from, to T, weight int) error {
	fromVertex, ok := g.Vertices[from]
	if !ok {
		return fmt.Errorf("vertex %v not found", from)
	}
	toVertex, ok := g.Vertices[to]
	if !ok {
		return fmt.Errorf("vertex %v not found", to)
	}
	if fromVertex.Edges == nil {
		fromVertex.Edges = make(map[T]*Edge[T, V])
	}
	fromVertex.Edges[to] = &Edge[T, V]{Vertex: toVertex, Weight: weight}
	return nil
}

func (g Graph[T, V]) Render() string {
	var b strings.Builder
	for _, id := range sortedKeys(g.Vertices) {
		vertex := g.Vertices[id]
		fmt.Fprintf(&b, "  vertex %v: %v", id, vertex.Val)
		if len(vertex.Edges) > 0 {
			b.WriteString(" → ")
			edges := make([]string, 0, len(vertex.Edges))
			for _, to := range sortedKeys(vertex.Edges) {
				edges = append(edges, fmt.Sprintf("%v (weight %d)", to, vertex.Edges[to].Weight))
			}
			b.WriteString(strings.Join(edges, ", "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func sortedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b K) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})
	return keys
}
