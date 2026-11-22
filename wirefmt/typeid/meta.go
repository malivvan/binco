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
	// The identifier is read through Unmarshal so that Skip applies the same
	// range checks: a value Unmarshal rejects is not silently skipped.
	_, n, err = s.Unmarshal(r)
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

func (optionalMetaCodec) Marshal(dtm binco.DTM, w wirefmt.Writer) (n int, err error) {
	if dtm < 0 {
		return 0, binco.ErrOverflow
	}
	return varnum.PositiveInt64.Marshal(int64(dtm), w)
}

func (optionalMetaCodec) Unmarshal(r wirefmt.Reader) (dtm binco.DTM, n int, err error) {
	num, n, err := varnum.PositiveInt64.Unmarshal(r)
	if err != nil {
		return
	}
	if int64(binco.DTM(num)) != num {
		return 0, n, binco.ErrOverflow
	}
	return binco.DTM(num), n, nil
}

func (optionalMetaCodec) Size(dtm binco.DTM) (size int) {
	if dtm < 0 {
		panic(binco.ErrOverflow)
	}
	return varnum.PositiveInt64.Size(int64(dtm))
}

func (optionalMetaCodec) Skip(r wirefmt.Reader) (n int, err error) {
	_, n, err = (optionalMetaCodec{}).Unmarshal(r)
	return
}
