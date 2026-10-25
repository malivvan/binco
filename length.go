package binco

// CheckLength reports whether a value of length bytes that starts at offset n
// fits entirely in bs.
//
// The comparison is written so that it cannot overflow: length is never added
// to n before the bounds check. Callers use it before slicing bs[n:n+length],
// which is why a hostile length such as MaxInt64 turns into
// ErrTooSmallByteSlice instead of a slice bounds panic.
func CheckLength(bs []byte, n, length int) error {
	if length < 0 {
		return ErrNegativeLength
	}
	if n < 0 || n > len(bs) {
		return ErrTooSmallByteSlice
	}
	if length > len(bs)-n {
		return ErrTooSmallByteSlice
	}
	return nil
}

// Remaining reports how many bytes follow offset n, or zero when n is past the
// end of bs.
func Remaining(bs []byte, n int) int {
	if n < 0 || n > len(bs) {
		return 0
	}
	return len(bs) - n
}

// AllocHint caps a length that came from the input by the number of bytes that
// are actually left to decode.
//
// A codec that reads a length before it reads the elements would otherwise let
// five bytes of input ask for a multi-gigabyte allocation. Element codecs
// consume at least one byte each, so a length above Remaining(bs, n) cannot be
// honoured; using the smaller value makes a hostile length fail with
// ErrTooSmallByteSlice after a bounded allocation instead of an out-of-memory
// crash. The returned value is never negative.
func AllocHint(length, remaining int) int {
	if length <= 0 {
		return 0
	}
	if length > remaining {
		return remaining
	}
	return length
}

// AllocHintElem is AllocHint for elements of a known minimum width. A length of
// n elements needs at least n*elemSize bytes; the multiplication is checked for
// overflow, and a length that cannot fit is capped by remaining/elemSize.
func AllocHintElem(length, remaining, elemSize int) int {
	if length <= 0 {
		return 0
	}
	if elemSize <= 1 {
		return AllocHint(length, remaining)
	}
	if remaining < 0 {
		return 0
	}
	fit := remaining / elemSize
	if length > fit {
		return fit
	}
	return length
}

// CheckLengthElem reports whether length elements of at least elemSize bytes
// can be present in bs after offset n. It is the validating counterpart of
// AllocHintElem and is used where an element cannot consume zero bytes.
func CheckLengthElem(bs []byte, n, length, elemSize int) error {
	if length < 0 {
		return ErrNegativeLength
	}
	if n < 0 || n > len(bs) {
		return ErrTooSmallByteSlice
	}
	if elemSize <= 1 {
		if length > len(bs)-n {
			return ErrTooSmallByteSlice
		}
		return nil
	}
	// length <= (len(bs)-n)/elemSize without multiplying length by elemSize.
	if length > (len(bs)-n)/elemSize {
		return ErrTooSmallByteSlice
	}
	return nil
}

// CheckWrite panics with ErrTooSmallByteSlice when a marshaller is about to
// write need bytes at offset n of a buffer that cannot hold them.
//
// Marshal has no error return, so a short output buffer is reported by panic.
// Every codec in this module panics with this sentinel rather than with a raw
// runtime index error, so a caller can recover and tell the two cases apart.
func CheckWrite(bs []byte, n, need int) {
	if n < 0 || need < 0 || n > len(bs) || need > len(bs)-n {
		panic(ErrTooSmallByteSlice)
	}
}

// Depth is a recursion counter shared by the codecs of one serialization. It
// is not safe for concurrent use; neither are pointer maps.
type Depth struct {
	n int
}

// MaxZeroSizeElements is the largest container that may be decoded when its
// elements consume no input bytes at all.
//
// A length is normally bounded by the input: every element takes at least one
// byte, so a length above the remaining bytes fails. Elements that consume
// nothing (empty structs, [0]T, codecs that always return n == 0) break that
// argument, and a length of 2^60 would then spin for hours. Such containers are
// capped here and rejected with ErrTooLargeLength above the cap.
const MaxZeroSizeElements = 1 << 20

// Enter records one level of nesting and reports ErrTooDeep once more than
// MaxDepth levels are active.
func (d *Depth) Enter() error {
	if d.n >= MaxDepth {
		return ErrTooDeep
	}
	d.n++
	return nil
}

// Leave ends one level of nesting.
func (d *Depth) Leave() {
	if d.n > 0 {
		d.n--
	}
}

// Reset returns the counter to zero.
func (d *Depth) Reset() {
	d.n = 0
}
