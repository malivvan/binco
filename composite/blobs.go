package composite

import (
	"github.com/malivvan/binco"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
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
	var lenSer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return textCodec{lenSer}
}

type textCodec struct {
	lenSer binco.Serializer[int]
}

func (s textCodec) Marshal(v string, bs []byte) (n int) {
	return MarshalString(v, s.lenSer, bs)
}

func (s textCodec) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	l := n + length
	if len(bs) < l {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if length == 0 {
		return
	}
	return string(bs[n:l]), l, nil
}

func (s textCodec) Size(v string) (size int) {
	return SizeString(v, s.lenSer)
}

func (s textCodec) Skip(bs []byte) (n int, err error) {
	return SkipString(s.lenSer, bs)
}

type checkedText struct {
	textCodec
	lenVl binco.Validator[int]
}

func (s checkedText) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	l := n + length
	if len(bs) < l {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return
		}
	}
	if length == 0 {
		return
	}
	return string(bs[n:l]), l, nil
}

// MarshalString writes v, prefixed by its length, into bs.
func MarshalString(v string, lenSer binco.Serializer[int], bs []byte) (n int) {
	length := len(v)
	n = lenSer.Marshal(length, bs)
	if len(bs) < n+length {
		panic(binco.ErrTooSmallByteSlice)
	}
	return n + copy(bs[n:], v)
}

// SizeString reports the encoded size of v.
func SizeString(v string, lenSer binco.Serializer[int]) (size int) {
	length := len(v)
	return lenSer.Size(length) + length
}

// SkipString advances over an encoded string.
func SkipString(lenSer binco.Serializer[int], bs []byte) (n int, err error) {
	length, n, err := lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	if len(bs[n:]) < length {
		err = binco.ErrTooSmallByteSlice
		return
	}
	return n + length, nil
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
	var lenSer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return blobCodec{lenSer}
}

type blobCodec struct {
	lenSer binco.Serializer[int]
}

func (s blobCodec) Marshal(v []byte, bs []byte) (n int) {
	return MarshalByteSlice(v, s.lenSer, bs)
}

func (s blobCodec) Unmarshal(bs []byte) (v []byte, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	if len(bs) < n+length {
		err = binco.ErrTooSmallByteSlice
		return
	}
	v = make([]byte, length)
	n += copy(v, bs[n:])
	return
}

func (s blobCodec) Size(v []byte) (size int) {
	return SizeByteSlice(v, s.lenSer)
}

func (s blobCodec) Skip(bs []byte) (n int, err error) {
	return SkipByteSlice(s.lenSer, bs)
}

type checkedBlob struct {
	blobCodec
	lenVl binco.Validator[int]
}

func (s checkedBlob) Unmarshal(bs []byte) (v []byte, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	if len(bs) < n+length {
		err = binco.ErrTooSmallByteSlice
		return
	}
	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return
		}
	}
	v = make([]byte, length)
	if length == 0 {
		return
	}
	n += copy(v, bs[n:])
	return
}

// MarshalByteSlice writes v, prefixed by its length, into bs.
func MarshalByteSlice(v []byte, lenSer binco.Serializer[int], bs []byte) (n int) {
	length := len(v)
	n = lenSer.Marshal(length, bs)
	if len(bs) < n+length {
		panic(binco.ErrTooSmallByteSlice)
	}
	return n + copy(bs[n:], v)
}

// SizeByteSlice reports the encoded size of v.
func SizeByteSlice(v []byte, lenSer binco.Serializer[int]) (size int) {
	length := len(v)
	return lenSer.Size(length) + length
}

// SkipByteSlice advances over an encoded byte slice.
func SkipByteSlice(lenSer binco.Serializer[int], bs []byte) (n int, err error) {
	length, n, err := lenSer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	l := n + length
	if len(bs) < l {
		err = binco.ErrTooSmallByteSlice
		return
	}
	n = l
	return
}
