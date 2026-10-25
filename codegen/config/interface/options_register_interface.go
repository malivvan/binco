package intropts

import (
	"reflect"

	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

// RegisterOptions collects the implementations passed to RegisterInterface.
type RegisterOptions struct {
	StructImpls      []StructImpl
	DefinedTypeImpls []DefinedTypeImpl
	Marshaller       bool
}

// NewRegisterOptions returns empty RegisterOptions.
func NewRegisterOptions() RegisterOptions {
	return RegisterOptions{
		StructImpls:      []StructImpl{},
		DefinedTypeImpls: []DefinedTypeImpl{},
	}
}

// StructImpl is a struct implementation together with its options.
type StructImpl struct {
	Type reflect.Type
	Opts []stopts.SetOption
}

// DefinedTypeImpl is a defined-type implementation together with its options.
type DefinedTypeImpl struct {
	Type reflect.Type
	Opts []tpopts.SetOption
}

// SetRegisterOption changes one field of RegisterOptions.
type SetRegisterOption func(o *RegisterOptions)

// WithStructImpl registers a struct implementation.
func WithStructImpl(t reflect.Type, opts ...stopts.SetOption) SetRegisterOption {
	impl := StructImpl{Type: t, Opts: opts}
	return func(o *RegisterOptions) {
		o.StructImpls = append(o.StructImpls, impl)
	}
}

// WithDefinedTypeImpl registers a defined-type implementation.
func WithDefinedTypeImpl(t reflect.Type, opts ...tpopts.SetOption) SetRegisterOption {
	impl := DefinedTypeImpl{Type: t, Opts: opts}
	return func(o *RegisterOptions) {
		o.DefinedTypeImpls = append(o.DefinedTypeImpls, impl)
	}
}

// WithRegisterMarshaller delegates marshalling to the value instead of switching over the implementations.
func WithRegisterMarshaller() SetRegisterOption {
	return func(o *RegisterOptions) {
		o.Marshaller = true
	}
}

// ApplyRegister runs opts in order, skipping nil entries.
func ApplyRegister(o *RegisterOptions, opts ...SetRegisterOption) {
	for i := range opts {
		if opts[i] != nil {
			opts[i](o)
		}
	}
}
