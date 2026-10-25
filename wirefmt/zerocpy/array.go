package zerocpy

import (
	"fmt"
	"reflect"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	arropts "github.com/malivvan/binco/wirefmt/config/array"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewArraySerializer returns a serializer for the array type T, whose elements are encoded with the element serializer.
//
// T must be an array whose element type is exactly V; the constructor panics
// with binco.ErrWrongArrayType otherwise, so mismatched type parameters cannot
// read or write out of bounds through unsafe. A WithElemValidator option is
// honoured.
func NewArraySerializer[T, V any](elemSerializer wirefmt.Serializer[V],
	opts ...arropts.SetOption[V],
) wirefmt.Serializer[T] {
	o := arropts.Options[V]{}
	arropts.Apply(opts, &o)
	return newArrayView[T, V](elemSerializer, o, o.ElemVl)
}

// NewValidArraySerializer is NewArraySerializer with element validation.
func NewValidArraySerializer[T, V any](elemSerializer wirefmt.Serializer[V],
	opts ...arropts.SetOption[V],
) arrayView[T, V] {
	o := arropts.Options[V]{}
	arropts.Apply(opts, &o)
	return newArrayView[T, V](elemSerializer, o, o.ElemVl)
}

func newArrayView[T, V any](elemSerializer wirefmt.Serializer[V],
	o arropts.Options[V], elemVl binco.Validator[V],
) arrayView[T, V] {
	length := validateArrayType[T, V]()
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return arrayView[T, V]{
		length:         length,
		elemSerializer: elemSerializer,
		lenSerializer:  lenSerializer,
		lenVl:          lengthLimit(length),
		elemVl:         elemVl,
	}
}

// validateArrayType reports the length of T after checking that T really is an
// array of V.
func validateArrayType[T, V any]() (length int) {
	at := reflect.TypeFor[T]()
	vt := reflect.TypeFor[V]()
	if at.Kind() != reflect.Array {
		panic(fmt.Errorf("%w: %v is not an array", binco.ErrWrongArrayType, at))
	}
	if at.Elem() != vt {
		panic(fmt.Errorf("%w: %v does not hold %v", binco.ErrWrongArrayType, at, vt))
	}
	if at.Len()*int(vt.Size()) != int(at.Size()) {
		panic(fmt.Errorf("%w: %v is not laid out as %d values of %v",
			binco.ErrWrongArrayType, at, at.Len(), vt))
	}
	return at.Len()
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

// Unmarshal checks the decoded length before it decodes anything, and fills the
// array memory directly instead of going through a temporary slice.
func (s arrayView[T, V]) Unmarshal(r wirefmt.Reader) (v T, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return v, n, binco.ErrNegativeLength
	}
	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return v, n, err
		}
	}
	if length != s.length {
		return v, n, binco.ErrWrongLength
	}
	if s.length == 0 {
		return v, n, nil
	}
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	var e V

	var n1 int

	for i := range s.length {
		e, n1, err = s.elemSerializer.Unmarshal(r)
		n += n1
		if err != nil {
			return v, n, err
		}
		if s.elemVl != nil {
			if err = s.elemVl.Validate(e); err != nil {
				return v, n, err
			}
		}
		sl[i] = e
	}
	return v, n, nil
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
		if t != length {
			err = binco.ErrWrongLength
		}
		return
	}
}
