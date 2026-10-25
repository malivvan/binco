// Package tpopts configures how a type is generated.
//
// Every option is ignored by the types it does not apply to, so a value can be
// passed to any Add* method.
package tpopts

// Options holds the codec choices of one type.
type Options struct {
	TimeUnit TimeUnit

	NumEncoding   NumEncoding
	Validator     string
	LenEnc        NumEncoding
	LenValidator  string
	ElemValidator string
	KeyValidator  string
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithNumEncoding chooses the encoding of a number.
func WithNumEncoding(numEnc NumEncoding) SetOption {
	return func(o *Options) {
		o.NumEncoding = numEnc
	}
}

// WithTimeUnit chooses the unit a time.Time is stored in.
func WithTimeUnit(timeUnit TimeUnit) SetOption {
	return func(o *Options) {
		o.TimeUnit = timeUnit
	}
}

// WithValidator names a function called with the decoded value.
func WithValidator(vl string) SetOption {
	return func(o *Options) {
		o.Validator = vl
	}
}

// WithLenEncoding chooses the encoding of a length.
func WithLenEncoding(lenEnc NumEncoding) SetOption {
	return func(o *Options) {
		o.LenEnc = lenEnc
	}
}

// WithLenValidator names a function called with the decoded length.
func WithLenValidator(vl string) SetOption {
	return func(o *Options) {
		o.LenValidator = vl
	}
}

// WithElemValidator names a function called with every decoded element.
func WithElemValidator(vl string) SetOption {
	return func(o *Options) {
		o.ElemValidator = vl
	}
}

// WithKeyValidator names a function called with every decoded key.
func WithKeyValidator(v string) SetOption {
	return func(o *Options) {
		o.KeyValidator = v
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
