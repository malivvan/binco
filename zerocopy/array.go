package zerocopy

import (
	"fmt"
	"reflect"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	arropts "github.com/malivvan/binco/config/array"
	"github.com/malivvan/binco/varnum"
)

// NewArraySerializer returns a serializer for the array type T, whose elements are encoded with the element serializer.
//
// T must be an array whose element type is exactly V; the constructor panics
// with binco.ErrWrongArrayType otherwise. Without that check an accidental
// `NewArraySerializer[[4]int32, int64]` would read and write out of bounds
// through unsafe.
//
// A WithElemValidator option is honoured; a WithLenSerializer option replaces
// the length codec.
func NewArraySerializer[T, V any](elemSerializer binco.Serializer[V],
	opts ...arropts.SetOption[V],
) binco.Serializer[T] {
	o := arropts.Options[V]{}
	arropts.Apply(opts, &o)

	return newArrayView[T, V](elemSerializer, o)
}

func newArrayView[T, V any](elemSerializer binco.Serializer[V],
	o arropts.Options[V],
) arrayView[T, V] {
	length := validateArrayType[T, V]()
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
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

// validateArrayType reports the length of T after checking that T really is an
// array of V. Misuse is a programming error, so it panics at construction time
// rather than corrupting memory later.
func validateArrayType[T, V any]() (length int) {
	at := reflect.TypeFor[T]()
	vt := reflect.TypeFor[V]()
	if at.Kind() != reflect.Array {
		panic(fmt.Errorf("%w: %v is not an array", binco.ErrWrongArrayType, at))
	}
	if at.Elem() != vt {
		panic(fmt.Errorf("%w: %v does not hold %v", binco.ErrWrongArrayType, at, vt))
	}
	return at.Len()
}

type arrayView[T, V any] struct {
	length         int
	elemSerializer binco.Serializer[V]
	lenSerializer  binco.Serializer[int]
	lenVl          binco.Validator[int]
	elemVl         binco.Validator[V]
}

func (s arrayView[T, V]) Marshal(v T, bs []byte) (n int) {
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	return composite.MarshalSlice(sl, s.elemSerializer, s.lenSerializer, bs)
}

// Unmarshal checks the decoded length before it decodes or allocates anything,
// and decodes straight into the array memory instead of into a temporary slice.
func (s arrayView[T, V]) Unmarshal(bs []byte) (v T, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
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
	// The length validator installed by newArrayView has already checked the
	// length against the array type, so only the empty array needs a shortcut.
	if s.length == 0 {
		return v, n, nil
	}
	sl := unsafe_mod.Slice((*V)(unsafe_mod.Pointer(&v)), s.length)
	var e V

	var n1 int

	for i := range s.length {
		e, n1, err = s.elemSerializer.Unmarshal(bs[n:])
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

func (s arrayView[T, V]) Skip(bs []byte) (n int, err error) {
	return composite.SkipSlice(bs, s.elemSerializer, s.lenSerializer)
}

func lengthLimit(length int) binco.ValidatorFn[int] {
	return func(t int) (err error) {
		if t != length {
			err = binco.ErrWrongLength
		}
		return
	}
}
