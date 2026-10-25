package composite

import (
	"sync/atomic"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Bool serializes a bool as a single byte.
var Bool = boolCodec{}

type boolCodec struct{}

func (boolCodec) Marshal(v bool, w wirefmt.Writer) (n int, err error) {
	if v {
		err = w.WriteByte(1)
	} else {
		err = w.WriteByte(0)
	}
	if err != nil {
		return
	}
	n = 1
	return
}

func (boolCodec) Unmarshal(r wirefmt.Reader) (v bool, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n = 1
	if b > 1 {
		err = binco.ErrWrongFormat
		return
	}
	v = b == 1
	return
}

func (boolCodec) Size(v bool) (size int) {
	return SizeBool(v)
}

func (boolCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return SkipBool(r)
}

// SizeBool reports the encoded size of a bool.
func SizeBool(v bool) (size int) {
	return 1
}

// SkipBool advances over an encoded bool.
func SkipBool(r wirefmt.Reader) (n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n = 1
	if b > 1 {
		err = binco.ErrWrongFormat
	}
	return
}

// NewPtrSerializer returns a pointer serializer that encodes nil as a flag byte.
//
// Unmarshalling a recursive type recurses once per non-nil pointer. The codec
// counts the active nesting and reports ErrTooDeep above
// binco.MaxPointerDepth, so hostile input cannot exhaust the goroutine stack.
func NewPtrSerializer[T any](baseSerializer wirefmt.Serializer[T]) ptrCodec[T] {
	return ptrCodec[T]{baseSerializer: baseSerializer, depth: new(atomic.Int64)}
}

type ptrCodec[T any] struct {
	baseSerializer wirefmt.Serializer[T]
	depth          *atomic.Int64
}

func (s ptrCodec[T]) enter() bool {
	// The failed increment is undone: without that, one ErrTooDeep would leave
	// the counter above the limit and every later call on this serializer would
	// fail too.
	if s.depth.Add(1) > binco.MaxPointerDepth {
		s.depth.Add(-1)
		return false
	}
	return true
}

func (s ptrCodec[T]) leave() {
	s.depth.Add(-1)
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
	if !s.enter() {
		return 0, binco.ErrTooDeep
	}
	defer s.leave()
	if err = w.WriteByte(byte(binco.PtrNotNil)); err != nil {
		return
	}
	n, err = s.baseSerializer.Marshal(*v, w)
	n += 1
	return
}

func (s ptrCodec[T]) Unmarshal(r wirefmt.Reader) (v *T, n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n++
	if b == byte(binco.PtrNil) {
		return
	}
	if b != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	if !s.enter() {
		return nil, n, binco.ErrTooDeep
	}
	defer s.leave()
	var n1 int
	k, n1, err := s.baseSerializer.Unmarshal(r)
	n += n1
	if err != nil {
		return
	}
	v = &k
	return
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	size = 1
	if v != nil {
		if !s.enter() {
			panic(binco.ErrTooDeep)
		}
		defer s.leave()
		size += s.baseSerializer.Size(*v)
	}
	return
}

func (s ptrCodec[T]) Skip(r wirefmt.Reader) (n int, err error) {
	b, err := r.ReadByte()
	if err != nil {
		return
	}
	n++
	if b == byte(binco.PtrNil) {
		return
	}
	if b != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	if !s.enter() {
		return n, binco.ErrTooDeep
	}
	defer s.leave()
	var n1 int
	n1, err = s.baseSerializer.Skip(r)
	n += n1
	return
}
