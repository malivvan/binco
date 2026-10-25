// Package arropts provides options for the array serializer.
package arropts

import (
	"github.com/malivvan/binco"
)

// Options collects the codecs and validators a serializer accepts.
type Options[T any] struct {
	LenSer binco.Serializer[int]
	ElemVl binco.Validator[T]
}

// SetOption changes one field of Options.
type SetOption[T any] func(o *Options[T])

// WithLenSer replaces the codec used for the length.
func WithLenSer[T any](lenSer binco.Serializer[int]) SetOption[T] {
	return func(o *Options[T]) { o.LenSer = lenSer }
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
