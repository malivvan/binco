package binco

// Validator inspects a decoded value, or a decoded length, and can reject the payload.
type Validator[T any] interface {
	Validate(t T) error
}

// ValidatorFn adapts a function to the Validator interface.
type ValidatorFn[T any] func(t T) (err error)

// Validate calls the function.
func (fn ValidatorFn[T]) Validate(t T) (err error) {
	return fn(t)
}
