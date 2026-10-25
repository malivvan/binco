package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// DTMSer reads and writes type identifiers.
var DTMSer = metaCodec{}

type metaCodec struct{}

func (s metaCodec) Marshal(dtm binco.DTM, w wirefmt.Writer) (n int, err error) {
	return varnum.PositiveInt.Marshal(int(dtm), w)
}

func (s metaCodec) Unmarshal(r wirefmt.Reader) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt.Unmarshal(r)
	if err != nil {
		return
	}
	dtm = binco.DTM(num)
	return
}

func (s metaCodec) Size(dtm binco.DTM) (size int) {
	return varnum.PositiveInt.Size(int(dtm))
}

func (s metaCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return varnum.PositiveInt.Skip(r)
}
