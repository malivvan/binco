package stopts

import tpopts "github.com/malivvan/binco/codegen/config/type"

// UnderlyingTimeOptions holds the unit used by a time.Time wrapper.
type UnderlyingTimeOptions struct {
	TimeUnit tpopts.TimeUnit
}

// SetUnderlyingTimeOption changes one field of UnderlyingTimeOptions.
type SetUnderlyingTimeOption func(o *UnderlyingTimeOptions)

// WithUnderlyingTimeUnit chooses the unit of a time.Time wrapper.
func WithUnderlyingTimeUnit(timeUnit tpopts.TimeUnit) SetUnderlyingTimeOption {
	return func(o *UnderlyingTimeOptions) {
		o.TimeUnit = timeUnit
	}
}

// UnderlyingTimeApply runs opts in order, skipping nil entries.
func UnderlyingTimeApply(opts []SetUnderlyingTimeOption, o *UnderlyingTimeOptions) {
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
}
