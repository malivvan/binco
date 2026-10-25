package composite

import (
	"io"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// String serializes a string as a length-prefixed byte sequence.
var String = NewStringSer()

// NewStringSer returns a string serializer whose length codec can be replaced through opts.
func NewStringSer(opts ...stropts.SetOption) textCodec {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	return newTextCodec(o)
}

// NewValidStringSer returns a string serializer that runs its length validator while decoding.
func NewValidStringSer(opts ...stropts.SetOption) checkedText {
	o := stropts.Options{}
	stropts.Apply(opts, &o)
	return checkedText{newTextCodec(o), o.LenVl}
}

func newTextCodec(o stropts.Options) textCodec {
	var lenSer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return textCodec{lenSer}
}

type textCodec struct {
	lenSer wirefmt.Serializer[int]
}

func (s textCodec) Marshal(v string, w wirefmt.Writer) (n int, err error) {
	return MarshalString(v, s.lenSer, w)
}

func (s textCodec) Unmarshal(r wirefmt.Reader) (v string, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var c []byte

	var n1 int

	if length == 0 {
		return
	}
	c = make([]byte, length)
	n1, err = io.ReadFull(r, c)
	n += n1
	if err != nil {
		return
	}
	v = string(c)
	return
}

func (s textCodec) Size(v string) (n int) {
	return SizeString(v, s.lenSer)
}

func (s textCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return SkipString(r, s.lenSer)
}

type checkedText struct {
	textCodec
	lenVl binco.Validator[int]
}

func (s checkedText) Unmarshal(r wirefmt.Reader) (v string, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var c []byte

	var n1 int

	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return
		}
	}
	if length == 0 {
		return
	}
	c = make([]byte, length)
	n1, err = io.ReadFull(r, c)
	n += n1
	if err != nil {
		return
	}
	v = string(c)
	return
}

// MarshalString writes v, prefixed by its length, to w.
func MarshalString(v string, lenSer wirefmt.Serializer[int], w wirefmt.Writer) (n int, err error) {
	n, err = lenSer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = w.WriteString(v)
	n += n1
	return
}

// SizeString reports the encoded size of v.
func SizeString(v string, lenSer wirefmt.Serializer[int]) (n int) {
	length := len(v)
	return lenSer.Size(length) + length
}

// SkipString advances over an encoded string.
func SkipString(r wirefmt.Reader, lenSer wirefmt.Serializer[int]) (n int, err error) {
	length, n, err := lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int64
	n1, err = io.CopyN(io.Discard, r, int64(length))
	n += int(n1)
	return
}

// ByteSlice serializes a byte slice as a length-prefixed byte sequence.
var ByteSlice = NewByteSliceSer()

// NewByteSliceSer returns a byte slice serializer whose length codec can be replaced.
func NewByteSliceSer(opts ...bslopts.SetOption) blobCodec {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	return newBlobCodec(o)
}

// NewValidByteSliceSer returns a byte slice serializer that runs its length validator while decoding.
func NewValidByteSliceSer(opts ...bslopts.SetOption) checkedBlob {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	var lenVl binco.Validator[int]
	if o.LenVl != nil {
		lenVl = o.LenVl
	}
	return checkedBlob{newBlobCodec(o), lenVl}
}

func newBlobCodec(o bslopts.Options) blobCodec {
	var lenSer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return blobCodec{lenSer}
}

type blobCodec struct {
	lenSer wirefmt.Serializer[int]
}

func (s blobCodec) Marshal(v []byte, w wirefmt.Writer) (
	n int, err error,
) {
	n, err = s.lenSer.Marshal(len(v), w)
	if err != nil {
		return
	}
	n1, err := w.Write(v)
	n += n1
	return
}

func (s blobCodec) Unmarshal(r wirefmt.Reader) (v []byte, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	v = make([]byte, length)
	n1, err = io.ReadFull(r, v)
	n += n1
	return
}

func (s blobCodec) Size(v []byte) (size int) {
	length := len(v)
	return s.lenSer.Size(length) + length
}

func (s blobCodec) Skip(r wirefmt.Reader) (n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int64
	n1, err = io.CopyN(io.Discard, r, int64(length))
	n += int(n1)
	return
}

type checkedBlob struct {
	blobCodec
	lenVl binco.Validator[int]
}

func (s checkedBlob) Unmarshal(r wirefmt.Reader) (v []byte, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return
		}
	}
	var n1 int
	v = make([]byte, length)
	n1, err = io.ReadFull(r, v)
	n += n1
	return
}
