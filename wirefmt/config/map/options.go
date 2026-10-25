// Package mapopts provides options for the streaming map serializer.
package mapopts

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Options for the map serializer.
// Options collects the codecs and validators a serializer accepts.
type Options[T, V any] struct {
	LenSerializer wirefmt.Serializer[int]
	LenVl         binco.Validator[int]
	KeyVl         binco.Validator[T]
	ValueVl       binco.Validator[V]
}

// SetOption changes one field of Options.
type SetOption[T, V any] func(o *Options[T, V])

// WithLenSerializer replaces the codec used for the length.
func WithLenSerializer[T, V any](lenSerializer wirefmt.Serializer[int]) SetOption[T, V] {
	return func(o *Options[T, V]) { o.LenSerializer = lenSerializer }
}

// WithLenValidator sets the validator called with the decoded length.
func WithLenValidator[T, V any](lenVl binco.Validator[int]) SetOption[T, V] {
	return func(o *Options[T, V]) { o.LenVl = lenVl }
}

// WithKeyValidator sets the validator called with every decoded key.
func WithKeyValidator[T, V any](keyVl binco.Validator[T]) SetOption[T, V] {
	return func(o *Options[T, V]) { o.KeyVl = keyVl }
}

// WithValueValidator sets the validator called with every decoded value.
func WithValueValidator[T, V any](valVl binco.Validator[V]) SetOption[T, V] {
	return func(o *Options[T, V]) { o.ValueVl = valVl }
}

// Apply runs opts in order, skipping nil entries.
func Apply[T, V any](opts []SetOption[T, V], o *Options[T, V]) {
	for i := range opts {
		if opts[i] != nil {
			opts[i](o)
		}
	}
}
