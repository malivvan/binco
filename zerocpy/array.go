package zerocpy

import (
	"reflect"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	arropts "github.com/malivvan/binco/config/array"
	"github.com/malivvan/binco/varnum"
)

// NewArraySer returns a serializer for the array type T, whose elements are encoded with elemSer.
func NewArraySer[T, V any](elemSer binco.Serializer[V],
	opts ...arropts.SetOption[V],
) (s arrayView[T, V]) {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSer binco.Serializer[int] = varnum.PositiveInt

	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return arrayView[T, V]{
		length:  length,
		elemSer: elemSer,
		lenSer:  lenSer,
		lenVl:   lengthLimit(length),
	}
}

// NewValidArraySer is NewArraySer with element validation.
func NewValidArraySer[T, V any](elemSer binco.Serializer[V],
	opts ...arropts.SetOption[V],
) arrayView[T, V] {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSer binco.Serializer[int] = varnum.PositiveInt

	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return arrayView[T, V]{
		length:  length,
		elemSer: elemSer,
		lenSer:  lenSer,
		lenVl:   lengthLimit(length),
		elemVl:  o.ElemVl,
	}
}

type arrayView[T, V any] struct {
	length  int
	elemSer binco.Serializer[V]
	lenSer  binco.Serializer[int]
	lenVl   binco.Validator[int]
	elemVl  binco.Validator[V]
}

func (s arrayView[T, V]) Marshal(v T, bs []byte) (n int) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.MarshalSlice(sl, s.elemSer, s.lenSer, bs)
}

func (s arrayView[T, V]) Unmarshal(bs []byte) (v T, n int, err error) {
	sl, n, err := composite.UnmarshalValidSlice(bs, s.elemSer, s.lenSer, s.lenVl, s.elemVl)
	if err != nil {
		return
	}

	if len(sl) != s.length {
		err = binco.ErrWrongLength
		return
	}
	if s.length == 0 {
		return
	}
	v = *(*T)(unsafe_mod.Pointer(unsafe_mod.SliceData(sl)))
	return
}

func (s arrayView[T, V]) Size(v T) (size int) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.SizeSlice(sl, s.elemSer, s.lenSer)
}

func (s arrayView[T, V]) Skip(bs []byte) (n int, err error) {
	return composite.SkipSlice(bs, s.elemSer, s.lenSer)
}

func lengthLimit(length int) binco.ValidatorFn[int] {
	return func(t int) (err error) {
		if t > length {
			err = binco.ErrTooLargeLength
		}
		return
	}
}
