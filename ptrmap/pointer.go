package ptrmap

import (
	"reflect"
	"unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// NewPtrSerializer returns a pointer serializer that writes the pointed-to data once and refers to it by identifier afterwards.
//
// Every identifier remembers the type it was decoded as, and a reference is
// only resolved when that type matches the field that reads it. A payload can
// therefore not make a *string out of an *int: the mismatch is reported as
// binco.ErrWrongType instead of being cast with unsafe.Pointer.
func NewPtrSerializer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	baseSerializer binco.Serializer[T],
) ptrCodec[T] {
	return ptrCodec[T]{ptrMap, revPtrMap, baseSerializer, reflect.TypeFor[T]()}
}

type ptrCodec[T any] struct {
	ptrMap         *binco.PtrMap
	revPtrMap      *binco.ReversePtrMap
	baseSerializer binco.Serializer[T]
	typ            reflect.Type
}

func (s ptrCodec[T]) Marshal(v *T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, 1)
	if v == nil {
		bs[0] = byte(binco.PtrNil)
		return 1
	}
	bs[0] = byte(binco.PtrMapping)
	n = 1
	id, newOne := idFor(unsafe.Pointer(v), s.typ, s.ptrMap)
	n += varnum.PositiveInt.Marshal(id, bs[n:])
	if newOne {
		n += s.baseSerializer.Marshal(*v, bs[n:])
	}
	return
}

func (s ptrCodec[T]) Unmarshal(bs []byte) (v *T, n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTruncated
		return
	}
	if err = s.revPtrMap.Enter(); err != nil {
		return nil, 0, err
	}
	defer s.revPtrMap.Leave()
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
		ptr, typ, pst := s.revPtrMap.Get(id)
		if !pst || ptr == nil {
			var n1 int
			v, n1, err = resolveData(id, s.typ, s.baseSerializer, s.revPtrMap, bs[n:])
			n += n1
			return
		}
		if typ != s.typ {
			return nil, n, binco.ErrWrongType
		}
		v = (*T)(ptr)
		return
	default:
		err = binco.ErrWrongFormat
	}
	return
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	size = 1
	if v != nil {
		id, newOne := idFor(unsafe.Pointer(v), s.typ, s.ptrMap)
		size += varnum.PositiveInt.Size(id)
		if newOne {
			return size + s.baseSerializer.Size(*v)
		}
	}
	return
}

func (s ptrCodec[T]) Skip(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTruncated
		return
	}
	if err = s.revPtrMap.Enter(); err != nil {
		return 0, err
	}
	defer s.revPtrMap.Leave()
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
		_, _, pst := s.revPtrMap.Get(id)
		if !pst {
			s.revPtrMap.Put(id, nil, s.typ)
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

func resolveData[T any](id int, typ reflect.Type, serializer binco.Serializer[T],
	revPtrMap *binco.ReversePtrMap,
	bs []byte,
) (v *T, n int, err error) {
	var k T

	var n1 int

	// The identifier is registered before the value is decoded so that a
	// reference inside the value resolves to the value being built. That is what
	// restores cycles; the depth counter bounds the recursion.
	revPtrMap.Put(id, unsafe.Pointer(&k), typ)
	k, n1, err = serializer.Unmarshal(bs)
	n += n1
	if err != nil {
		return nil, n, err
	}
	v = &k
	return
}

func idFor(ptr unsafe.Pointer, typ reflect.Type, ptrMap *binco.PtrMap) (id int, newOne bool) {
	id, pst := ptrMap.Get(ptr, typ)
	if !pst {
		id = ptrMap.Put(ptr, typ)
		newOne = true
	}
	return
}
