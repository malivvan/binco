package zerocpy

import (
	"reflect"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	arropts "github.com/malivvan/binco/wirefmt/config/array"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewArraySerializer returns a serializer for the array type T, whose elements are encoded with the element serializer.
func NewArraySerializer[T, V any](elemSerializer wirefmt.Serializer[V],
	opts ...arropts.SetOption[V],
) (s arrayView[T, V]) {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return arrayView[T, V]{
		length:         length,
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
		lenVl:          lengthLimit(length),
	}
}

// NewValidArraySerializer is NewArraySerializer with element validation.
func NewValidArraySerializer[T, V any](elemSerializer wirefmt.Serializer[V],
	opts ...arropts.SetOption[V],
) arrayView[T, V] {
	var o = arropts.Options[V]{}

	var t = reflect.TypeFor[T]()

	var length = t.Len()

	arropts.Apply(opts, &o)
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return arrayView[T, V]{
		length:         length,
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
		lenVl:          lengthLimit(length),
		elemVl:         o.ElemVl,
	}
}

type arrayView[T, V any] struct {
	length         int
	elemSerializer wirefmt.Serializer[V]
	lenSerializer  wirefmt.Serializer[int]
	lenVl          binco.Validator[int]
	elemVl         binco.Validator[V]
}

func (s arrayView[T, V]) Marshal(v T, w wirefmt.Writer) (n int, err error) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.MarshalSlice(sl, s.elemSerializer, s.lenSerializer, w)
}

func (s arrayView[T, V]) Unmarshal(r wirefmt.Reader) (v T, n int, err error) {
	sl, n, err := composite.UnmarshalValidSlice(s.elemSerializer, s.lenSerializer, s.lenVl, s.elemVl,
		r)
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
	return composite.SizeSlice(sl, s.elemSerializer, s.lenSerializer)
}

func (s arrayView[T, V]) Skip(r wirefmt.Reader) (n int, err error) {
	return composite.SkipSlice(s.elemSerializer, s.lenSerializer, r)
}

func lengthLimit(length int) binco.ValidatorFn[int] {
	return func(t int) (err error) {
		if t > length {
			err = binco.ErrTooLargeLength
		}
		return
	}
}
