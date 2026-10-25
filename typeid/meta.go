package typeid

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// DTMSer encodes and decodes type identifiers.
var DTMSer = metaCodec{}

type metaCodec struct{}

func (s metaCodec) Marshal(dtm binco.DTM, bs []byte) (n int) {
	return varnum.PositiveInt.Marshal(int(dtm), bs)
}

func (s metaCodec) Unmarshal(bs []byte) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt.Unmarshal(bs)
	if err != nil {
		return
	}
	dtm = binco.DTM(num)
	return
}

func (s metaCodec) Size(dtm binco.DTM) (size int) {
	return varnum.PositiveInt.Size(int(dtm))
}

func (s metaCodec) Skip(bs []byte) (n int, err error) {
	return varnum.PositiveInt.Skip(bs)
}
