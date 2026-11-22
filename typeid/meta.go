package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// DTMSerializer encodes and decodes type identifiers.
//
// Identifiers are positive: Marshal panics with binco.ErrOverflow for a zero or
// negative value, which would otherwise be written as a huge unsigned number
// and never match on decode.
var DTMSerializer = metaCodec{}

type metaCodec struct{}

func (s metaCodec) Marshal(dtm binco.DTM, bs []byte) (n int) {
	if dtm <= 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Marshal(int64(dtm), bs)
}

func (s metaCodec) Unmarshal(bs []byte) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt64.Unmarshal(bs)
	if err != nil {
		return
	}
	if num <= 0 || int64(binco.DTM(num)) != num {
		return 0, n, binco.ErrOverflow
	}
	dtm = binco.DTM(num)
	return
}

func (s metaCodec) Size(dtm binco.DTM) (size int) {
	if dtm <= 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Size(int64(dtm))
}

func (s metaCodec) Skip(bs []byte) (n int, err error) {
	// The identifier is read through Unmarshal so that Skip applies the same
	// range checks: a value Unmarshal rejects is not silently skipped.
	_, n, err = s.Unmarshal(bs)
	return
}

// OptionalDTMSerializer reads and writes the identifier an interface payload
// carries.
//
// It is DTMSerializer with one difference: identifier zero is reserved for a nil
// interface value, so it is written and decoded instead of being rejected. Every
// non-zero identifier behaves exactly like DTMSerializer and produces the same
// bytes.
var OptionalDTMSerializer = optionalMetaCodec{}

type optionalMetaCodec struct{}

func (s optionalMetaCodec) Marshal(dtm binco.DTM, bs []byte) (n int) {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Marshal(int64(dtm), bs)
}

func (s optionalMetaCodec) Unmarshal(bs []byte) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt64.Unmarshal(bs)
	if err != nil {
		return
	}
	if int64(binco.DTM(num)) != num {
		return 0, n, binco.ErrOverflow
	}
	return binco.DTM(num), n, nil
}

func (s optionalMetaCodec) Size(dtm binco.DTM) (size int) {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Size(int64(dtm))
}

// Skip validates exactly like Unmarshal, so the two agree on 32-bit hosts where
// an identifier above MaxInt does not fit binco.DTM.
func (s optionalMetaCodec) Skip(bs []byte) (n int, err error) {
	_, n, err = s.Unmarshal(bs)
	return
}
