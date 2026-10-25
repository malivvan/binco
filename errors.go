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

// ErrTooDeep reports a value nested deeper than MaxDepth. It bounds the
// recursion of pointer and reference codecs, so that hostile input cannot
// exhaust the stack.
var ErrTooDeep = errors.New(ErrorPrefix + "too deep")

// ErrNonCanonical reports a varint that is not written in its shortest form.
// Overlong encodings are rejected so that a value has exactly one byte
// representation, which keeps hashes, signatures and byte-wise comparisons
// meaningful.
var ErrNonCanonical = errors.New(ErrorPrefix + "non-canonical varint")

// ErrWrongSize reports a codec that wrote a different number of bytes than its
// Size method promised. It is a bug in the codec, not in the data; codecs that
// cannot return an error panic with it.
var ErrWrongSize = errors.New(ErrorPrefix + "size mismatch")

// ErrTrailingBytes reports input that continues after a complete value.
var ErrTrailingBytes = errors.New(ErrorPrefix + "trailing bytes")

// ErrWrongType reports a value that was encoded under one type and is read
// back as another, such as a pointer identifier that referred to an int.
var ErrWrongType = errors.New(ErrorPrefix + "wrong type")

// ErrNilInterface reports a nil interface value where a concrete
// implementation was required.
var ErrNilInterface = errors.New(ErrorPrefix + "nil interface")

// ErrWrongArrayType reports that NewArraySerializer was built with type
// parameters that do not describe an array of the element type.
var ErrWrongArrayType = errors.New(ErrorPrefix + "wrong array type")

// ErrSchemaMismatch reports a payload whose schema fingerprint does not match
// the one the serializer was generated for.
var ErrSchemaMismatch = errors.New(ErrorPrefix + "schema mismatch")

// ErrDuplicateDTM reports two types that were assigned the same type
// identifier.
var ErrDuplicateDTM = errors.New(ErrorPrefix + "duplicate DTM")

// MaxDepth is the deepest nesting of pointers, references and arrays a codec
// accepts before returning ErrTooDeep.
const MaxDepth = 1000

// MaxPointerDepth bounds the recursion of the plain pointer codec.
//
// Unlike ptrmap, which counts depth per decoded value, the plain pointer codec
// has no per-value state to hang the counter on, so it shares an atomic
// counter per serializer instance. The limit is therefore higher than MaxDepth
// (process-wide concurrent nesting must stay below it) but far below the depth
// at which the goroutine stack would overflow.
const MaxPointerDepth = 10000

// ErrUnsupportedIntSize reports a host whose int is neither 32 nor 64 bits wide.
var ErrUnsupportedIntSize = errors.New(ErrorPrefix + "unsupported IntSize")

// ErrTooLargeLength reports a length above the maximum the type allows.
var ErrTooLargeLength = errors.New(ErrorPrefix + "too large length")

// ErrTruncated reports input that ends inside a value.
//
// It is the error a decoder returns when the bytes that arrived are not enough
// for the value the schema describes. Encoders report a short output buffer with
// ErrBufferTooSmall instead, so the two failures are distinguishable.
var ErrTruncated = errors.New(ErrorPrefix + "truncated input")

// ErrBufferTooSmall reports an output buffer that cannot hold the value.
//
// Marshal has no error return, so it panics with this value. Every codec uses it,
// which lets a caller recognise a short buffer with one recover.
var ErrBufferTooSmall = errors.New(ErrorPrefix + "buffer too small")

// ErrWrongLength reports a length that does not match the type, such as a short array.
var ErrWrongLength = errors.New(ErrorPrefix + "wrong length")

// ErrUnsupportedArch reports a big-endian host, which package zerocopy cannot serve.
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
