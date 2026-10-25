package ptrmap

import "github.com/malivvan/binco"

// Wrap returns a serializer that clears the pointer maps around every operation, so identifiers start at zero for each value.
//
// The maps are cleared in place instead of being replaced by freshly allocated
// ones, and both are reset: a decoder that only cleared the reverse map would
// leave the encoder's identifiers from the previous value behind.
func Wrap[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	serializer binco.Serializer[T],
) scope[T] {
	return scope[T]{ptrMap, revPtrMap, serializer}
}

type scope[T any] struct {
	ptrMap     *binco.PtrMap
	revPtrMap  *binco.ReversePtrMap
	serializer binco.Serializer[T]
}

func (p scope[T]) Marshal(v T, bs []byte) (n int) {
	defer p.ptrMap.Clear()
	return p.serializer.Marshal(v, bs)
}

func (p scope[T]) Unmarshal(bs []byte) (t T, n int, err error) {
	defer p.revPtrMap.Clear()
	return p.serializer.Unmarshal(bs)
}

func (p scope[T]) Size(v T) int {
	defer p.ptrMap.Clear()
	return p.serializer.Size(v)
}

func (p scope[T]) Skip(bs []byte) (n int, err error) {
	defer p.revPtrMap.Clear()
	return p.serializer.Skip(bs)
}
