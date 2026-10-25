// Package arropts provides options for the streaming array serializer.
package arropts

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Options for the array serializer.
// Options collects the codecs and validators a serializer accepts.
type Options[T any] struct {
	LenSerializer wirefmt.Serializer[int]
	ElemVl        binco.Validator[T]
}

// SetOption changes one field of Options.
type SetOption[T any] func(o *Options[T])

// WithLenSerializer replaces the codec used for the length.
func WithLenSerializer[T any](lenSerializer wirefmt.Serializer[int]) SetOption[T] {
	return func(o *Options[T]) { o.LenSerializer = lenSerializer }
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
