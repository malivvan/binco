# Pointer mapping (ptrmap)

A plain pointer serializer (`composite.NewPtrSer`) writes every pointer by value. That
cannot represent shared references (each copy is written again) and cannot
represent cycles at all. A `ptrmap` pointer serializer assigns an ID to every pointer
and writes the pointed-to data only the first time the ID occurs:

```
first occurrence:  [2] [ID] [data]
later occurrence:  [2] [ID]
nil pointer:       [1]
```

The receiver resolves the IDs back to the same pointers, so the pointer graph is
restored exactly.

- [simple](simple/): one pointer used by several fields.
- [linked_list](linked_list/): a doubly linked list.
- [graph](graph/): a cyclic graph.

## How to wire it up

```go
var (
	ptrMap    = binco.NewPtrMap()
	revPtrMap = binco.NewReversePtrMap()
	nodePtr   = ptrmap.NewPtrSer(ptrMap, revPtrMap, nodeSer) // *Node
)

var NodeSer = ptrmap.Wrap(ptrMap, revPtrMap, nodePtr) // clears the maps per call
```

Three rules:

1. **Share the maps.** Every `ptrmap` serializer that takes part in one value must
   use the same `PtrMap` and `ReversePtrMap`, or IDs will not match up.
2. **Wrap the top-level serializer.** `ptrmap.Wrap` clears the maps around every
   operation (`Marshal`, `Unmarshal`, `Size`, `Skip`), so IDs start at 0 again
   for the next value and `Size`/`Marshal` agree. Calling `Size` or `Marshal` on
   an inner serializer directly leaves IDs behind.
3. **Keep one wrapper per goroutine.** The maps are state, so a wrapped
   serializer must not be used concurrently for different values.

## simple

```
TwoPtrs: [2 0 30 116 104 101 ... 2 0]
  [2 0 ...] flag 2 (mapping), ID 0, then the string
  [2 0]     the same ID again, no string
  decoded: "the same pointer in two fields", "the same pointer in two fields"
  identity preserved: true
```

Two fields pointing at one string store it once; the second occurrence is two
bytes. The embedded example shows the same for three occurrences across a
composite struct, and a nil pointer is a single flag byte.

## linked_list

```
list:     [8, 9, 10, 11] (4 elements)
size:     23 bytes
decoded:  [8, 9, 10, 11] (read 23 bytes)
forward:  [8 9 10 11]
backward: [11 10 9 8]
prev/next links consistent: true
Head.Prev == nil: true, Tail.Next == nil: true
```

Only the head, the tail and the length are written as data; the elements are
reached through pointers and every element is written once. After decoding, the
list is a complete doubly linked list: walking forward and backward agrees, and
`next.Prev == current` everywhere. The example then removes an element to show
that the structure is fully functional, not just a copy of the values.

## graph

```
graph:
  vertex 1: one → 2 (weight 10)
  vertex 2: two → 3 (weight 20)
  vertex 3: three → 1 (weight 30)
size:    45 bytes

edge 1→2 → vertex 2: true
edge 2→3 → vertex 3: true
edge 3→1 → vertex 1: true
```

This is the hardest case: a vertex points at its edges and every edge points
back at a vertex, so the serializer graph itself is cyclic
(`vertexSer → edgeSer → vertexSer`). The serializers are wired through variables
that are assigned at the end of `newGraphSer`. Each vertex reached through an
edge is recognized as the vertex stored in `Vertices` - the identity checks after
decoding prove it.

## When not to use ptrmap

Pointer mapping costs a map lookup per pointer and adds an ID to every pointer on
the wire. If your data is a tree without shared references, the plain
`composite.NewPtrSer` is smaller and faster. Use `ptrmap` for linked lists, trees with
parent pointers, graphs, and any structure where two fields may alias.
