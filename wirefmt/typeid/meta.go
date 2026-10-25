package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// DTMSerializer reads and writes type identifiers.
//
// Identifiers are positive: Marshal returns binco.ErrOverflow for a zero or
// negative value, which would otherwise be written as a huge unsigned number.
var DTMSerializer = metaCodec{}

type metaCodec struct{}

func (s metaCodec) Marshal(dtm binco.DTM, w wirefmt.Writer) (n int, err error) {
	if dtm <= 0 {
		return 0, binco.ErrOverflow
	}
	return varnum.PositiveInt64.Marshal(int64(dtm), w)
}

func (s metaCodec) Unmarshal(r wirefmt.Reader) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt64.Unmarshal(r)
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

func (s metaCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return varnum.PositiveInt64.Skip(r)
}
