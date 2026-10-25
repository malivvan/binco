// Package bslopts provides options for the byte slice serializer.
package bslopts

import (
	"github.com/malivvan/binco"
)

// Options collects the codecs and validators a serializer accepts.
type Options struct {
	LenSer binco.Serializer[int]
	LenVl  binco.Validator[int]
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithLenSer replaces the codec used for the length.
func WithLenSer(lenSer binco.Serializer[int]) SetOption {
	return func(o *Options) { o.LenSer = lenSer }
}

// WithLenValidator sets the validator called with the decoded length.
func WithLenValidator(lenVl binco.Validator[int]) SetOption {
	return func(o *Options) { o.LenVl = lenVl }
}

// Apply runs opts in order, skipping nil entries.
func Apply(opts []SetOption, o *Options) {
	for i := range opts {
		if opts[i] != nil {
			opts[i](o)
		}
	}
}
