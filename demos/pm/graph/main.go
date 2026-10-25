// Command graph encodes a cyclic graph. Edges point back at vertices, and the pointer identities survive the round trip.
package main

import (
	"fmt"

	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

func main() {
	graph := CyclicGraph()
	serializer := MakeGraphSerializer[int, string](varnum.Int, composite.String)

	bs := make([]byte, serializer.Size(graph))
	n := serializer.Marshal(graph, bs)
	fmt.Printf("graph:\n%s", graph.Render())
	fmt.Printf("size:    %d bytes\n", len(bs))
	fmt.Printf("encoded: %v (wrote %d)\n\n", bs, n)

	back, n, err := serializer.Unmarshal(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded %d bytes:\n%s\n", n, back.Render())

	v1, v2, v3 := back.Vertices[1], back.Vertices[2], back.Vertices[3]
	fmt.Printf("edge 1→2 → vertex 2: %v\n", v1.Edges[2].Vertex == v2)
	fmt.Printf("edge 2→3 → vertex 3: %v\n", v2.Edges[3].Vertex == v3)
	fmt.Printf("edge 3→1 → vertex 1: %v\n", v3.Edges[1].Vertex == v1)

	n, err = serializer.Skip(bs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("skipped: %d bytes\n", n)
}

func CyclicGraph() Graph[int, string] {
	g := NewGraph[int, string]()
	g.AddVertex(1, "one")
	g.AddVertex(2, "two")
	g.AddVertex(3, "three")

	edges := [][3]int{{1, 2, 10}, {2, 3, 20}, {3, 1, 30}}
	for _, e := range edges {
		if err := g.AddEdge(e[0], e[1], e[2]); err != nil {
			panic(err)
		}
	}
	return g
}
