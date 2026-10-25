package ptrmap

import (
	"unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// NewPtrSerializer returns a pointer serializer that writes the pointed-to data once and refers to it by identifier afterwards.
func NewPtrSerializer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	baseSerializer binco.Serializer[T],
) ptrCodec[T] {
	return ptrCodec[T]{ptrMap, revPtrMap, baseSerializer}
}

type ptrCodec[T any] struct {
	ptrMap         *binco.PtrMap
	revPtrMap      *binco.ReversePtrMap
	baseSerializer binco.Serializer[T]
}

func (s ptrCodec[T]) Marshal(v *T, bs []byte) (n int) {
	if v == nil {
		bs[0] = byte(binco.PtrNil)
		return 1
	}
	bs[0] = byte(binco.PtrMapping)
	n = 1
	id, newOne := idFor(unsafe.Pointer(v), s.ptrMap)
	n += varnum.PositiveInt.Marshal(id, bs[n:])
	if newOne {
		n += s.baseSerializer.Marshal(*v, bs[n:])
	}
	return
}

func (s ptrCodec[T]) Unmarshal(bs []byte) (v *T, n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	switch bs[0] {
	case byte(binco.PtrNil):
		n = 1
		return
	case byte(binco.PtrMapping):
		var n1 int

		var id int

		n = 1
		id, n1, err = varnum.PositiveInt.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		ptr, _ := s.revPtrMap.Get(id)
		if ptr == nil {
			v, n1, err = resolveData(id, s.baseSerializer, s.revPtrMap, bs[n:])
			n += n1
		} else {
			v = (*T)(ptr)
		}
	default:
		err = binco.ErrWrongFormat
	}
	return
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	size = 1
	if v != nil {
		id, newOne := idFor(unsafe.Pointer(v), s.ptrMap)
		size += varnum.PositiveInt.Size(id)
		if newOne {
			return size + s.baseSerializer.Size(*v)
		}
	}
	return
}

func (s ptrCodec[T]) Skip(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTooSmallByteSlice
		return
	}
	switch bs[0] {
	case byte(binco.PtrNil):
		n = 1
		return
	case byte(binco.PtrMapping):
		n = 1
		var id int

		var n1 int

		id, n1, err = varnum.PositiveInt.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		_, pst := s.revPtrMap.Get(id)
		if !pst {
			s.revPtrMap.Put(id, nil)
			n1, err = s.baseSerializer.Skip(bs[n:])
			n += n1
			if err != nil {
				return
			}
		}
	default:
		err = binco.ErrWrongFormat
	}
	return
}

func resolveData[T any](id int, serializer binco.Serializer[T],
	revPtrMap *binco.ReversePtrMap,
	bs []byte,
) (v *T, n int, err error) {
	var k T

	var n1 int

	revPtrMap.Put(id, unsafe.Pointer(&k))
	k, n1, err = serializer.Unmarshal(bs)
	n += n1
	if err != nil {
		return
	}
	v = &k
	return
}

func idFor(ptr unsafe.Pointer, ptrMap *binco.PtrMap) (id int, newOne bool) {
	id, pst := ptrMap.Get(ptr)
	if !pst {
		id = ptrMap.Put(ptr)
		newOne = true
	}
	return
}
