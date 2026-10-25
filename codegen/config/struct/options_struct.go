// Package stopts configures a struct type.
package stopts

import (
	fldopts "github.com/malivvan/binco/codegen/config/field"
)

// Options holds the validator, the per-field options and the underlying time options of a struct.
type Options struct {
	Validator      string
	Fields         []fldopts.Options
	UnderlyingTime []SetUnderlyingTimeOption
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithValidator names a function called with the decoded struct.
func WithValidator(validator string) SetOption {
	return func(o *Options) {
		o.Validator = validator
	}
}

// WithField appends the options of the next field; pass one call per field.
func WithField(opts ...fldopts.SetOption) SetOption {
	return func(o *Options) {
		fopts := fldopts.Options{}
		fldopts.Apply(&fopts, opts...)
		o.Fields = append(o.Fields, fopts)
	}
}

// WithUnderlyingTime treats the struct as a wrapper around time.Time.
func WithUnderlyingTime(opts ...SetUnderlyingTimeOption) SetOption {
	return func(o *Options) {
		if o.UnderlyingTime == nil {
			o.UnderlyingTime = make([]SetUnderlyingTimeOption, 0)
		}
		o.UnderlyingTime = append(o.UnderlyingTime, opts...)
	}
}

// Apply runs opts in order, skipping nil entries.
func Apply(opts []SetOption, o *Options) {
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
}
