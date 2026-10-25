// Package intropts configures an interface serializer and the registration of its implementations.
package intropts

import (
	"reflect"
)

// Options lists the implementations of an interface.
type Options struct {
	Impls      []reflect.Type
	Marshaller bool
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithImpl registers one implementation type.
func WithImpl(impl reflect.Type) SetOption {
	return WithImpls(impl)
}

// WithMarshaller delegates marshalling to the value instead of switching over the implementations.
func WithMarshaller() SetOption {
	return func(o *Options) {
		o.Marshaller = true
	}
}

// WithImpls registers several implementation types.
func WithImpls(impls ...reflect.Type) SetOption {
	return func(o *Options) {
		o.Impls = append(o.Impls, impls...)
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
