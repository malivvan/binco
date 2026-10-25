// Package fldopts configures one struct field.
package fldopts

import tpopts "github.com/malivvan/binco/codegen/config/type"

// Options holds the codec options and the ignore flag of a field.
type Options struct {
	Type   tpopts.Options
	Ignore bool
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithIgnore keeps the field out of the payload.
func WithIgnore() SetOption {
	return func(o *Options) {
		o.Ignore = true
	}
}

// WithType applies type options to the field.
func WithType(opts ...tpopts.SetOption) SetOption {
	return func(o *Options) {
		tpopts.Apply(&o.Type, opts...)
	}
}

// Apply runs opts in order, skipping nil entries.
func Apply(o *Options, opts ...SetOption) {
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
}
