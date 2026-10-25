package tpopts

import "fmt"

// The numeric encodings a type can use.
const (
	NumEncodingUndefined NumEncoding = iota

	NumEncodingVarint

	NumEncodingVarintPositive

	NumEncodingRaw
)

// NumEncoding selects how a number is written.
type NumEncoding int

// Package names the codec package a number encoding lives in.
func (e NumEncoding) Package() (pkg string) {
	switch e {
	case NumEncodingUndefined, NumEncodingVarint, NumEncodingVarintPositive:
		return "varnum"
	case NumEncodingRaw:
		return "scalar"
	default:
		panic(fmt.Errorf("undefined %d NumEncoding", e))
	}
}

// LenSerializer names the codec a length of this encoding uses.
func (e NumEncoding) LenSerializer() string {
	switch e {
	case NumEncodingUndefined, NumEncodingVarintPositive:
		return "varnum.PositiveInt"
	case NumEncodingVarint:
		return "varnum.Int"
	case NumEncodingRaw:
		return "scalar.Int"
	default:
		panic(fmt.Errorf("undefined %d NumEncoding", e))
	}
}
