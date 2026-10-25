package binco

import (
	"errors"
	"fmt"
)

// ErrorPrefix starts the text of every error raised by this module.
const ErrorPrefix = "binco: "

// ErrOverflow reports a value that does not fit the width of its type.
var ErrOverflow = errors.New(ErrorPrefix + "overflow")

// ErrNegativeLength reports a length that decoded as a negative number.
var ErrNegativeLength = errors.New(ErrorPrefix + "negative length")

// ErrWrongFormat reports a flag byte or shape the format does not define.
var ErrWrongFormat = errors.New(ErrorPrefix + "wrong format")

// ErrUnsupportedIntSize reports a host whose int is neither 32 nor 64 bits wide.
var ErrUnsupportedIntSize = errors.New(ErrorPrefix + "unsupported IntSize")

// ErrTooLargeLength reports a length above the maximum the type allows.
var ErrTooLargeLength = errors.New(ErrorPrefix + "too large length")

// ErrTooSmallByteSlice reports input that ends inside a value or an output buffer that is too short.
var ErrTooSmallByteSlice = errors.New(ErrorPrefix + "too small byte slice")

// ErrWrongLength reports a length that does not match the type, such as a short array.
var ErrWrongLength = errors.New(ErrorPrefix + "wrong length")

// ErrUnsupportedArch reports a big-endian host, which package zerocpy cannot serve.
var ErrUnsupportedArch = errors.New(ErrorPrefix + "unsupported architecture, " +
	"little-endian byte order is required")

// UnexpectedDTMError reports a type identifier that the reader does not know.
type UnexpectedDTMError struct {
	dtm DTM
}

// NewUnexpectedDTMError builds an UnexpectedDTMError for dtm.
func NewUnexpectedDTMError(dtm DTM) UnexpectedDTMError {
	return UnexpectedDTMError{dtm: dtm}
}

// DTM returns the identifier that was read.
func (e UnexpectedDTMError) DTM() DTM {
	return e.dtm
}

// Error implements error.
func (e UnexpectedDTMError) Error() string {
	return fmt.Sprintf(ErrorPrefix+"unexpected DTM %d", e.dtm)
}

// WrongDTMError reports a payload written as a different type than the one expected.
type WrongDTMError struct {
	expected DTM
	actual   DTM
}

// NewWrongDTMError builds a WrongDTMError.
func NewWrongDTMError(expected, actual DTM) WrongDTMError {
	return WrongDTMError{expected: expected, actual: actual}
}

// Expected returns the identifier the reader was built for.
func (e WrongDTMError) Expected() DTM {
	return e.expected
}

// Actual returns the identifier found in the payload.
func (e WrongDTMError) Actual() DTM {
	return e.actual
}

// Error implements error.
func (e WrongDTMError) Error() string {
	return fmt.Sprintf(ErrorPrefix+"wrong DTM, expected %d, got %d", e.expected, e.actual)
}
