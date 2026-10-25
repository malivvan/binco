// Package stropts provides options for the streaming string serializer.
package stropts

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
)

// Options for the string serializer.
// Options collects the codecs and validators a serializer accepts.
type Options struct {
	LenSerializer wirefmt.Serializer[int]
	LenVl         binco.Validator[int]
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithLenSerializer replaces the codec used for the length.
func WithLenSerializer(lenSerializer wirefmt.Serializer[int]) SetOption {
	return func(o *Options) { o.LenSerializer = lenSerializer }
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
