package zerocopy

import (
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

// String serializes a string. Decoding returns a string that shares memory with the input buffer.
var String = NewStringSerializer()

// NewStringSerializer returns a string serializer whose length codec can be replaced through opts.
//
// A WithLenValidator option is honoured. The decoded string aliases the input
// buffer: it is only valid while that buffer is, and the buffer must not be
// reused or modified. Use composite.String when the value outlives the buffer.
func NewStringSerializer(opts ...stropts.SetOption) binco.Serializer[string] {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	s := newTextView(o)
	if o.LenVl != nil {
		return checkedText{s, o.LenVl}
	}
	return s
}

func newTextView(o stropts.Options) textView {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return textView{lenSerializer}
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
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return "", n, err
	}
	if len(body) == 0 {
		return "", end, nil
	}
	return unsafe_mod.String(&body[0], len(body)), end, nil
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
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return "", n, err
	}
	if s.lenVl != nil {
		if err := s.lenVl.Validate(length); err != nil {
			return "", n, err
		}
	}
	if len(body) == 0 {
		return "", end, nil
	}
	return unsafe_mod.String(&body[0], len(body)), end, nil
}

// fixedBody returns the length bytes that follow the decoded length and the
// offset of the first byte after them. binco.CheckLength keeps the arithmetic
// overflow-safe, so a hostile length cannot wrap into a slice panic.
func fixedBody(bs []byte, n, length int) (body []byte, end int, err error) {
	if err = binco.CheckLength(bs, n, length); err != nil {
		return nil, n, err
	}
	return bs[n : n+length], n + length, nil
}

// ByteSlice serializes a byte slice. Decoding returns a slice that shares memory with the input buffer.
var ByteSlice = NewByteSliceSerializer()

// NewByteSliceSerializer returns a byte slice serializer whose length codec can
// be replaced. A WithLenValidator option is honoured.
//
// The decoded slice aliases the input buffer; do not modify it and do not reuse
// the buffer while the slice is alive.
func NewByteSliceSerializer(opts ...bslopts.SetOption) binco.Serializer[[]byte] {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	s := newBlobView(o)
	if o.LenVl != nil {
		return checkedBlob{s, o.LenVl}
	}
	return s
}

func newBlobView(o bslopts.Options) blobView {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return blobView{lenSerializer}
}

type blobView struct {
	lenSerializer binco.Serializer[int]
}

func (s blobView) Marshal(v []byte, bs []byte) (n int) {
	return composite.MarshalByteSlice(v, s.lenSerializer, bs)
}

func (s blobView) Unmarshal(bs []byte) (v []byte, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return nil, n, err
	}
	if len(body) == 0 {
		return []byte{}, end, nil
	}
	return body, end, nil
}

func (s blobView) Size(v []byte) (size int) {
	return composite.SizeByteSlice(v, s.lenSerializer)
}

func (s blobView) Skip(bs []byte) (n int, err error) {
	return composite.SkipByteSlice(s.lenSerializer, bs)
}

type checkedBlob struct {
	blobView
	lenVl binco.Validator[int]
}

func (s checkedBlob) Unmarshal(bs []byte) (v []byte, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return nil, n, err
	}
	if s.lenVl != nil {
		if err := s.lenVl.Validate(length); err != nil {
			return nil, n, err
		}
	}
	if len(body) == 0 {
		return []byte{}, end, nil
	}
	return body, end, nil
}
