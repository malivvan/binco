package ptrmap

import "github.com/malivvan/binco"

// Wrap returns a serializer that clears the pointer maps around every operation, so identifiers start at zero for each value.
func Wrap[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	ser binco.Serializer[T],
) scope[T] {
	return scope[T]{ptrMap, revPtrMap, ser}
}

type scope[T any] struct {
	ptrMap    *binco.PtrMap
	revPtrMap *binco.ReversePtrMap
	ser       binco.Serializer[T]
}

func (p scope[T]) Marshal(v T, bs []byte) (n int) {
	defer func() {
		*p.ptrMap = *binco.NewPtrMap()
	}()
	return p.ser.Marshal(v, bs)
}

func (p scope[T]) Unmarshal(bs []byte) (t T, n int, err error) {
	defer func() {
		*p.revPtrMap = *binco.NewReversePtrMap()
	}()
	return p.ser.Unmarshal(bs)
}

func (p scope[T]) Size(v T) int {
	defer func() {
		*p.ptrMap = *binco.NewPtrMap()
	}()
	return p.ser.Size(v)
}

func (p scope[T]) Skip(bs []byte) (n int, err error) {
	defer func() {
		*p.revPtrMap = *binco.NewReversePtrMap()
	}()
	return p.ser.Skip(bs)
}
