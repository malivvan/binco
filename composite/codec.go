package composite

import (
	"sync/atomic"

	"github.com/malivvan/binco"
)

// Bool serializes a bool as a single byte.
var Bool = boolCodec{}

type boolCodec struct{}

func (s boolCodec) Marshal(v bool, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, 1)
	if v {
		bs[0] = 1
	} else {
		bs[0] = 0
	}
	return 1
}

func (s boolCodec) Unmarshal(bs []byte) (v bool, n int, err error) {
	if len(bs) < 1 {
		return false, 0, binco.ErrTruncated
	}
	if bs[0] > 1 {
		return false, 0, binco.ErrWrongFormat
	}
	return bs[0] == 1, 1, nil
}

func (s boolCodec) Size(_ bool) (size int) {
	return 1
}

func (s boolCodec) Skip(bs []byte) (n int, err error) {
	return SkipBool(bs)
}

// SkipBool advances over an encoded bool.
func SkipBool(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		return 0, binco.ErrTruncated
	}
	if bs[0] == 0 || bs[0] == 1 {
		return 1, nil
	}
	return 0, binco.ErrWrongFormat
}

// NewPtrSerializer returns a pointer serializer that encodes nil as a flag byte.
//
// Unmarshalling a recursive type recurses once per non-nil pointer. The codec
// counts the active nesting and reports ErrTooDeep above
// binco.MaxPointerDepth, so hostile input cannot exhaust the goroutine stack.
// The counter is atomic, so the serializer stays safe for concurrent use.
func NewPtrSerializer[T any](baseSerializer binco.Serializer[T]) ptrCodec[T] {
	return ptrCodec[T]{baseSerializer: baseSerializer, depth: new(atomic.Int64)}
}

type ptrCodec[T any] struct {
	baseSerializer binco.Serializer[T]
	depth          *atomic.Int64
}

func (s ptrCodec[T]) enter() bool {
	return s.depth.Add(1) <= binco.MaxPointerDepth
}

func (s ptrCodec[T]) leave() {
	s.depth.Add(-1)
}

func (s ptrCodec[T]) Marshal(v *T, bs []byte) (n int) {
	binco.CheckWrite(bs, 0, 1)
	if v == nil {
		bs[0] = byte(binco.PtrNil)
		n = 1
		return
	}
	if !s.enter() {
		panic(binco.ErrTooDeep)
	}
	defer s.leave()
	bs[0] = byte(binco.PtrNotNil)
	return 1 + s.baseSerializer.Marshal(*v, bs[1:])
}

func (s ptrCodec[T]) Unmarshal(bs []byte) (v *T, n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTruncated
		return
	}
	if bs[0] == byte(binco.PtrNil) {
		n = 1
		return
	}
	if bs[0] != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	if !s.enter() {
		err = binco.ErrTooDeep
		return
	}
	defer s.leave()
	k, n, err := s.baseSerializer.Unmarshal(bs[1:])
	if err != nil {
		n = 1 + n
		return
	}
	return &k, 1 + n, err
}

func (s ptrCodec[T]) Size(v *T) (size int) {
	if v != nil {
		if !s.enter() {
			panic(binco.ErrTooDeep)
		}
		defer s.leave()
		return 1 + s.baseSerializer.Size(*v)
	}
	return 1
}

func (s ptrCodec[T]) Skip(bs []byte) (n int, err error) {
	if len(bs) < 1 {
		err = binco.ErrTruncated
		return
	}
	if bs[0] == byte(binco.PtrNil) {
		n = 1
		return
	}
	if bs[0] != byte(binco.PtrNotNil) {
		err = binco.ErrWrongFormat
		return
	}
	if !s.enter() {
		return 0, binco.ErrTooDeep
	}
	defer s.leave()
	n, err = s.baseSerializer.Skip(bs[1:])
	return 1 + n, err
}
