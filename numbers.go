package binco

import "strconv"

// Num64RawSize is the width of a 64-bit value in bytes.
const Num64RawSize = 8

// Num32RawSize is the width of a 32-bit value in bytes.
const Num32RawSize = 4

// Num16RawSize is the width of a 16-bit value in bytes.
const Num16RawSize = 2

// Num8RawSize is the width of an 8-bit value in bytes.
const Num8RawSize = 1

// Signed constrains a type parameter to the signed integer types.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned constrains a type parameter to the unsigned integer types.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer64 permits the integer types that are 64 bits wide.
type Integer64 interface {
	~uint | ~uint64 | ~int | ~int64
}

// Integer32 permits the integer types that are 32 bits wide.
type Integer32 interface {
	~uint | ~uint32 | ~int | ~int32
}

// Integer16 permits the integer types that are 16 bits wide.
type Integer16 interface {
	~uint16 | ~int16
}

// Integer8 permits the integer types that are 8 bits wide.
type Integer8 interface {
	~uint8 | ~int8
}

// Num64 permits 64-bit integers and float64.
type Num64 interface {
	~float64 | Integer64
}

// Num32 permits 32-bit integers and float32.
type Num32 interface {
	~float32 | Integer32
}

// Uint64MaxVarintLen is the longest varint encoding of a uint64.
const Uint64MaxVarintLen = 10

// Uint32MaxVarintLen is the longest varint encoding of a uint32.
const Uint32MaxVarintLen = 5

// Uint16MaxVarintLen is the longest varint encoding of a uint16.
const Uint16MaxVarintLen = 3

// Uint8MaxVarintLen is the longest varint encoding of a uint8.
const Uint8MaxVarintLen = 2

// Uint64MaxLastByte is the largest final byte of a uint64 varnum.
const Uint64MaxLastByte byte = 1

// Uint32MaxLastByte is the largest final byte of a uint32 varnum.
const Uint32MaxLastByte byte = 15

// Uint16MaxLastByte is the largest final byte of a uint16 varnum.
const Uint16MaxLastByte byte = 3

// Uint8MaxLastByte is the largest final byte of a uint8 varnum.
const Uint8MaxLastByte byte = 1

var varintWordLen int

var varintWordLast byte

func init() {
	maxLen, maxLast, ok := wordLimits(strconv.IntSize)
	if !ok {
		panic("unsupported IntSize")
	}
	varintWordLen, varintWordLast = maxLen, maxLast
}

// wordLimits returns the varint limits of an int of the given size in bits.
func wordLimits(intSize int) (maxLen int, maxLast byte, ok bool) {
	switch intSize {
	case 64:
		return Uint64MaxVarintLen, Uint64MaxLastByte, true
	case 32:
		return Uint32MaxVarintLen, Uint32MaxLastByte, true
	case 16:
		return Uint16MaxVarintLen, Uint16MaxLastByte, true
	case 8:
		return Uint8MaxVarintLen, Uint8MaxLastByte, true
	}
	return 0, 0, false
}

// UintMaxVarintLen is the longest varint encoding on this host.
func UintMaxVarintLen() int {
	return varintWordLen
}

// UintMaxLastByte is the largest final varint byte on this host.
func UintMaxLastByte() byte {
	return varintWordLast
}
