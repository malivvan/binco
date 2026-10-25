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
var String = NewStringSerializer()

// NewStringSerializer returns a string serializer whose length codec can be replaced through opts.
func NewStringSerializer(opts ...stropts.SetOption) textCodec {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	return newTextCodec(o)
}

// NewValidStringSerializer returns a string serializer that runs its length validator while decoding.
func NewValidStringSerializer(opts ...stropts.SetOption) checkedText {
	o := stropts.Options{}
	stropts.Apply(opts, &o)
	return checkedText{newTextCodec(o), o.LenVl}
}

func newTextCodec(o stropts.Options) textCodec {
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return textCodec{lenSerializer}
}

type textCodec struct {
	lenSerializer wirefmt.Serializer[int]
}

func (s textCodec) Marshal(v string, w wirefmt.Writer) (n int, err error) {
	return MarshalString(v, s.lenSerializer, w)
}

func (s textCodec) Unmarshal(r wirefmt.Reader) (v string, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
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
	return SizeString(v, s.lenSerializer)
}

func (s textCodec) Skip(r wirefmt.Reader) (n int, err error) {
	return SkipString(r, s.lenSerializer)
}

type checkedText struct {
	textCodec
	lenVl binco.Validator[int]
}

func (s checkedText) Unmarshal(r wirefmt.Reader) (v string, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
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
func MarshalString(v string, lenSerializer wirefmt.Serializer[int], w wirefmt.Writer) (n int, err error) {
	n, err = lenSerializer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	n1, err = w.WriteString(v)
	n += n1
	return
}

// SizeString reports the encoded size of v.
func SizeString(v string, lenSerializer wirefmt.Serializer[int]) (n int) {
	length := len(v)
	return lenSerializer.Size(length) + length
}

// SkipString advances over an encoded string.
func SkipString(r wirefmt.Reader, lenSerializer wirefmt.Serializer[int]) (n int, err error) {
	length, n, err := lenSerializer.Unmarshal(r)
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
var ByteSlice = NewByteSliceSerializer()

// NewByteSliceSerializer returns a byte slice serializer whose length codec can be replaced.
func NewByteSliceSerializer(opts ...bslopts.SetOption) blobCodec {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	return newBlobCodec(o)
}

// NewValidByteSliceSerializer returns a byte slice serializer that runs its length validator while decoding.
func NewValidByteSliceSerializer(opts ...bslopts.SetOption) checkedBlob {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	var lenVl binco.Validator[int]
	if o.LenVl != nil {
		lenVl = o.LenVl
	}
	return checkedBlob{newBlobCodec(o), lenVl}
}

func newBlobCodec(o bslopts.Options) blobCodec {
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return blobCodec{lenSerializer}
}

type blobCodec struct {
	lenSerializer wirefmt.Serializer[int]
}

func (s blobCodec) Marshal(v []byte, w wirefmt.Writer) (
	n int, err error,
) {
	n, err = s.lenSerializer.Marshal(len(v), w)
	if err != nil {
		return
	}
	n1, err := w.Write(v)
	n += n1
	return
}

func (s blobCodec) Unmarshal(r wirefmt.Reader) (v []byte, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
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
	return s.lenSerializer.Size(length) + length
}

func (s blobCodec) Skip(r wirefmt.Reader) (n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
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
	length, n, err := s.lenSerializer.Unmarshal(r)
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
