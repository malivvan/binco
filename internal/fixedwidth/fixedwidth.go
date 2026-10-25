// Package fixedwidth decides whether a codec can be applied in bulk.
//
// A slice of a fixed-width scalar can be encoded and decoded with a single copy
// instead of one serializer call per element. That is only valid when the codec
// writes the raw little-endian value (binco.FixedWidth), the host is
// little-endian, and the Go type has exactly the encoded size.
package fixedwidth

import (
	"unsafe"

	"github.com/malivvan/binco"
)

// Of reports the encoded width of one value when serializer writes raw
// little-endian values of exactly the size of T.
func Of[T any](serializer any) (width int, ok bool) {
	fixed, isFixed := serializer.(binco.FixedWidth)
	if !isFixed || !binco.HostLittleEndian {
		return 0, false
	}
	width = fixed.FixedWidth()
	if width <= 0 || int(unsafe.Sizeof(*new(T))) != width {
		return 0, false
	}
	return width, true
}
