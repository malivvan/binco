// Package pkg declares the type generated with a custom serializer name and per-field options.
package pkg

import (
	"errors"
	"fmt"
)

const MaxStrLen = 100

var ErrTooLong = errors.New("string too long")

type Foo struct {
	Num int
	Str string
}

func ValidateLength(length int) error {
	if length > MaxStrLen {
		return fmt.Errorf("%w: %d > %d", ErrTooLong, length, MaxStrLen)
	}
	return nil
}
