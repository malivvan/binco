package ptrmap

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Wrap returns a serializer that clears the pointer maps around every operation, so identifiers start at zero for each value.
func Wrap[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	serializer wirefmt.Serializer[T],
) scope[T] {
	return scope[T]{ptrMap, revPtrMap, serializer}
}

type scope[T any] struct {
	ptrMap     *binco.PtrMap
	revPtrMap  *binco.ReversePtrMap
	serializer wirefmt.Serializer[T]
}

func (p scope[T]) Marshal(v T, w wirefmt.Writer) (n int, err error) {
	defer func() {
		*p.ptrMap = *binco.NewPtrMap()
	}()
	return p.serializer.Marshal(v, w)
}

func (p scope[T]) Unmarshal(r wirefmt.Reader) (t T, n int, err error) {
	defer func() {
		*p.revPtrMap = *binco.NewReversePtrMap()
	}()
	return p.serializer.Unmarshal(r)
}

func (p scope[T]) Size(v T) int {
	defer func() {
		*p.ptrMap = *binco.NewPtrMap()
	}()
	return p.serializer.Size(v)
}

func (p scope[T]) Skip(r wirefmt.Reader) (n int, err error) {
	defer func() {
		*p.revPtrMap = *binco.NewReversePtrMap()
	}()
	return p.serializer.Skip(r)
}
