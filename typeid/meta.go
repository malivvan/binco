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
	return varnum.PositiveInt64.Skip(bs)
}
