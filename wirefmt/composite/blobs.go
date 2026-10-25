package composite

import (
	"bytes"
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
//
// A WithLenValidator option is honoured here as well as in
// NewValidStringSerializer.
func NewStringSerializer(opts ...stropts.SetOption) wirefmt.Serializer[string] {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	s := newTextCodec(o)
	if o.LenVl != nil {
		return checkedText{s, o.LenVl}
	}
	return s
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
		return "", n, binco.ErrNegativeLength
	}
	var n1 int
	var body []byte
	body, n1, err = readFixed(r, length)
	n += n1
	if err != nil {
		return
	}
	return string(body), n, nil
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
		return "", n, binco.ErrNegativeLength
	}
	if s.lenVl != nil {
		if err := s.lenVl.Validate(length); err != nil {
			return "", n, err
		}
	}
	var n1 int
	var body []byte
	body, n1, err = readFixed(r, length)
	n += n1
	if err != nil {
		return
	}
	return string(body), n, nil
}

// readFixed reads exactly length bytes from r.
//
// The result grows as the bytes arrive instead of being allocated from the
// declared length, so a stream that announces one gigabyte and then ends after
// five bytes costs five bytes and not one gigabyte. A truncation is reported as
// io.ErrUnexpectedEOF: the value has already started, so the stream did not end
// cleanly between two values.
func readFixed(r wirefmt.Reader, length int) (body []byte, n int, err error) {
	if length == 0 {
		return []byte{}, 0, nil
	}
	var buf bytes.Buffer
	var n1 int64
	n1, err = io.CopyN(&buf, r, int64(length))
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, int(n1), err
	}
	return buf.Bytes(), int(n1), nil
}

// discardFixed skips exactly length bytes, reporting a truncation as
// io.ErrUnexpectedEOF.
func discardFixed(r wirefmt.Reader, length int) (n int, err error) {
	if length == 0 {
		return 0, nil
	}
	var n1 int64
	n1, err = io.CopyN(io.Discard, r, int64(length))
	if err != nil && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return int(n1), err
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
		return n, binco.ErrNegativeLength
	}
	var n1 int
	n1, err = discardFixed(r, length)
	n += n1
	return
}

// ByteSlice serializes a byte slice as a length-prefixed byte sequence.
var ByteSlice = NewByteSliceSerializer()

// NewByteSliceSerializer returns a byte slice serializer whose length codec can be replaced.
func NewByteSliceSerializer(opts ...bslopts.SetOption) wirefmt.Serializer[[]byte] {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	s := newBlobCodec(o)
	if o.LenVl != nil {
		return checkedBlob{s, o.LenVl}
	}
	return s
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

func (s blobCodec) Marshal(v []byte, w wirefmt.Writer) (n int, err error) {
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
		return nil, n, binco.ErrNegativeLength
	}
	var n1 int
	v, n1, err = readFixed(r, length)
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
		return n, binco.ErrNegativeLength
	}
	var n1 int
	n1, err = discardFixed(r, length)
	n += n1
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
		return nil, n, binco.ErrNegativeLength
	}
	if s.lenVl != nil {
		if err := s.lenVl.Validate(length); err != nil {
			return nil, n, err
		}
	}
	var n1 int
	v, n1, err = readFixed(r, length)
	n += n1
	return
}
