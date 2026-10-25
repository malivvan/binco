package ptrmap

import (
	"unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewPtrSerializer returns a pointer serializer that writes the pointed-to data once and refers to it by identifier afterwards.
func NewPtrSerializer[T any](ptrMap *binco.PtrMap, revPtrMap *binco.ReversePtrMap,
	baseSerializer wirefmt.Serializer[T],
) ptrCodec[T] {
	return ptrCodec[T]{ptrMap, revPtrMap, baseSerializer}
}

type ptrCodec[T any] struct {
	ptrMap         *binco.PtrMap
	revPtrMap      *binco.ReversePtrMap
	baseSerializer wirefmt.Serializer[T]
}

func (s ptrCodec[T]) Marshal(v *T, w wirefmt.Writer) (n int, err error) {
	if v == nil {
		err = w.WriteByte(byte(binco.PtrNil))
		if err != nil {
			return
		}
		n++
		return
	}
	err = w.WriteByte(byte(binco.PtrMapping))
	if err != nil {
		return
	}
	n++
	id, newOne := idFor(unsafe.Pointer(v), s.ptrMap)
	var n1 int
	n1, err = varnum.PositiveInt.Marshal(id, w)
	n += n1
	if err != nil {
		return
	}
	if newOne {
		n1, err = s.baseSerializer.Marshal(*v, w)
		n += n1
	}
	return
}

func (s ptrCodec[T]) Unmarshal(r wirefmt.Reader) (v *T, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	switch b {
	case byte(binco.PtrNil):
		n = 1
		return
	case byte(binco.PtrMapping):
		var n1 int

		var id int

		n = 1
		id, n1, err = varnum.PositiveInt.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		ptr, _ := s.revPtrMap.Get(id)
		if ptr == nil {
			v, n1, err = s.unmarshalData(id, r)
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
		} else {
			return
		}
	}
	return
}

func (s ptrCodec[T]) Skip(r wirefmt.Reader) (n int,
	err error,
) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	switch b {
	case byte(binco.PtrNil):
		n = 1
		return
	case byte(binco.PtrMapping):
		n = 1
		var id int

		var n1 int

		id, n1, err = varnum.PositiveInt.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		_, pst := s.revPtrMap.Get(id)
		if !pst {
			s.revPtrMap.Put(id, nil)
			n1, err = s.baseSerializer.Skip(r)
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

func (s ptrCodec[T]) unmarshalData(id int, r wirefmt.Reader) (v *T, n int, err error) {
	var k T

	var n1 int

	s.revPtrMap.Put(id, unsafe.Pointer(&k))
	k, n1, err = s.baseSerializer.Unmarshal(r)
	n += n1
	if err != nil {
		return
	}
	v = &k
	return
}

func idFor(ptr unsafe.Pointer, mp *binco.PtrMap) (id int, newOne bool) {
	id, pst := mp.Get(ptr)
	if !pst {
		id = mp.Put(ptr)
		newOne = true
	}
	return
}
