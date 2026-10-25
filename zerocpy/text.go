package zerocpy

import (
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

// String serializes a string. Decoding returns a string that shares memory with the input buffer.
var String = NewStringSer()

// NewStringSer returns a string serializer whose length codec can be replaced through opts.
func NewStringSer(opts ...stropts.SetOption) textView {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	return newTextView(o)
}

// NewValidStringSer returns a string serializer that runs its length validator while decoding.
func NewValidStringSer(opts ...stropts.SetOption) checkedText {
	o := stropts.Options{}
	stropts.Apply(opts, &o)
	return checkedText{newTextView(o), o.LenVl}
}

func newTextView(o stropts.Options) textView {
	var lenSer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return textView{lenSer}
}

type textView struct {
	len binco.Serializer[int]
}

func (s textView) Marshal(v string, bs []byte) (n int) {
	return composite.MarshalString(v, s.len, bs)
}

func (s textView) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.len.Unmarshal(bs)
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
	return unsafe_mod.String(&bs[n], length), l, nil
}

func (s textView) Size(v string) (size int) {
	return composite.SizeString(v, s.len)
}

func (s textView) Skip(bs []byte) (n int, err error) {
	return composite.SkipString(s.len, bs)
}

type checkedText struct {
	textView
	lenVl binco.Validator[int]
}

func (s checkedText) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.len.Unmarshal(bs)
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
	return unsafe_mod.String(&bs[n], length), l, nil
}

// ByteSlice serializes a byte slice. Decoding returns a slice that shares memory with the input buffer.
var ByteSlice = NewByteSliceSer()

// NewByteSliceSer returns a byte slice serializer whose length codec can be replaced.
func NewByteSliceSer(opts ...bslopts.SetOption) blobView {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	return newBlobView(o)
}

// NewValidByteSliceSer returns a byte slice serializer that runs its length validator while decoding.
func NewValidByteSliceSer(opts ...bslopts.SetOption) checkedBlob {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	var lenVl binco.Validator[int]
	if o.LenVl != nil {
		lenVl = o.LenVl
	}
	return checkedBlob{newBlobView(o), lenVl}
}

func newBlobView(o bslopts.Options) blobView {
	var lenSer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return blobView{lenSer}
}

type blobView struct {
	lenSer binco.Serializer[int]
}

func (s blobView) Marshal(v []byte, bs []byte) (n int) {
	return composite.MarshalByteSlice(v, s.lenSer, bs)
}

func (s blobView) Unmarshal(bs []byte) (v []byte, n int, err error) {
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
		return []byte{}, l, nil
	}
	return unsafe_mod.Slice(&bs[n], length), l, nil
}

func (s blobView) Size(v []byte) (size int) {
	return composite.SizeByteSlice(v, s.lenSer)
}

func (s blobView) Skip(bs []byte) (n int, err error) {
	return composite.SkipByteSlice(s.lenSer, bs)
}

type checkedBlob struct {
	blobView
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
		return []byte{}, l, nil
	}
	return unsafe_mod.Slice(&bs[n], length), l, nil
}
