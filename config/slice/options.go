// Package slopts provides options for the slice serializer.
package slopts

import (
	"github.com/malivvan/binco"
)

// Options collects the codecs and validators a serializer accepts.
type Options[T any] struct {
	LenSer binco.Serializer[int]
	LenVl  binco.Validator[int]
	ElemVl binco.Validator[T]
}

// SetOption changes one field of Options.
type SetOption[T any] func(o *Options[T])

// WithLenSer replaces the codec used for the length.
func WithLenSer[T any](lenSer binco.Serializer[int]) SetOption[T] {
	return func(o *Options[T]) { o.LenSer = lenSer }
}

// WithLenValidator sets the validator called with the decoded length.
func WithLenValidator[T any](lenVl binco.Validator[int]) SetOption[T] {
	return func(o *Options[T]) { o.LenVl = lenVl }
}

// WithElemValidator sets the validator called with every decoded element.
func WithElemValidator[T any](elemVl binco.Validator[T]) SetOption[T] {
	return func(o *Options[T]) { o.ElemVl = elemVl }
}

// Apply runs opts in order, skipping nil entries.
func Apply[T any](opts []SetOption[T], o *Options[T]) {
	for i := range opts {
		if opts[i] != nil {
			opts[i](o)
		}
	}
}
