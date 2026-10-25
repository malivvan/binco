package composite

import (
	"github.com/malivvan/binco"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

// String serializes a string as a length-prefixed byte sequence.
var String = NewStringSerializer()

// NewStringSerializer returns a string serializer whose length codec can be
// replaced through opts.
//
// A WithLenValidator option is honoured, so a length can be bounded before it
// is used.
func NewStringSerializer(opts ...stropts.SetOption) binco.Serializer[string] {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	s := newTextCodec(o)
	if o.LenVl != nil {
		return checkedText{s, o.LenVl}
	}
	return s
}

func newTextCodec(o stropts.Options) textCodec {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return textCodec{lenSerializer}
}

type textCodec struct {
	lenSerializer binco.Serializer[int]
}

func (s textCodec) Marshal(v string, bs []byte) (n int) {
	return MarshalString(v, s.lenSerializer, bs)
}

func (s textCodec) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return "", n, err
	}
	return string(body), end, nil
}

func (s textCodec) Size(v string) (size int) {
	return SizeString(v, s.lenSerializer)
}

func (s textCodec) Skip(bs []byte) (n int, err error) {
	return SkipString(s.lenSerializer, bs)
}

type checkedText struct {
	textCodec
	lenVl binco.Validator[int]
}

func (s checkedText) Unmarshal(bs []byte) (v string, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
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
	return string(body), end, nil
}

// fixedBody returns the length bytes that follow the decoded length and the
// offset of the first byte after them.
//
// The bounds check is centralised in binco.CheckLength, which never adds n and
// length before comparing, so a length of MaxInt64 is rejected instead of
// wrapping into a slice panic.
func fixedBody(bs []byte, n, length int) (body []byte, end int, err error) {
	if err = binco.CheckLength(bs, n, length); err != nil {
		return nil, n, err
	}
	return bs[n : n+length], n + length, nil
}

// MarshalString writes v, prefixed by its length, into bs.
func MarshalString(v string, lenSerializer binco.Serializer[int], bs []byte) (n int) {
	length := len(v)
	n = lenSerializer.Marshal(length, bs)
	if len(bs)-n < length {
		panic(binco.ErrBufferTooSmall)
	}
	return n + copy(bs[n:], v)
}

// SizeString reports the encoded size of v.
func SizeString(v string, lenSerializer binco.Serializer[int]) (size int) {
	length := len(v)
	return lenSerializer.Size(length) + length
}

// SkipString advances over an encoded string.
func SkipString(lenSerializer binco.Serializer[int], bs []byte) (n int, err error) {
	length, n, err := lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if err = binco.CheckLength(bs, n, length); err != nil {
		return n, err
	}
	return n + length, nil
}

// ByteSlice serializes a byte slice as a length-prefixed byte sequence.
var ByteSlice = NewByteSliceSerializer()

// NewByteSliceSerializer returns a byte slice serializer whose length codec can
// be replaced. A WithLenValidator option is honoured.
func NewByteSliceSerializer(opts ...bslopts.SetOption) binco.Serializer[[]byte] {
	o := bslopts.Options{}
	bslopts.Apply(opts, &o)

	s := newBlobCodec(o)
	if o.LenVl != nil {
		return checkedBlob{s, o.LenVl}
	}
	return s
}

func newBlobCodec(o bslopts.Options) blobCodec {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return blobCodec{lenSerializer}
}

type blobCodec struct {
	lenSerializer binco.Serializer[int]
}

func (s blobCodec) Marshal(v []byte, bs []byte) (n int) {
	return MarshalByteSlice(v, s.lenSerializer, bs)
}

func (s blobCodec) Unmarshal(bs []byte) (v []byte, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	// CheckLength also caps what make below may allocate: a hostile length can
	// never exceed the bytes that were actually supplied.
	body, end, err := fixedBody(bs, n, length)
	if err != nil {
		return nil, n, err
	}
	if length == 0 {
		return []byte{}, end, nil
	}
	v = make([]byte, length)
	copy(v, body)
	return v, end, nil
}

func (s blobCodec) Size(v []byte) (size int) {
	return SizeByteSlice(v, s.lenSerializer)
}

func (s blobCodec) Skip(bs []byte) (n int, err error) {
	return SkipByteSlice(s.lenSerializer, bs)
}

type checkedBlob struct {
	blobCodec
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
	if length == 0 {
		return []byte{}, end, nil
	}
	v = make([]byte, length)
	copy(v, body)
	return v, end, nil
}

// MarshalByteSlice writes v, prefixed by its length, into bs.
func MarshalByteSlice(v []byte, lenSerializer binco.Serializer[int], bs []byte) (n int) {
	length := len(v)
	n = lenSerializer.Marshal(length, bs)
	if len(bs)-n < length {
		panic(binco.ErrBufferTooSmall)
	}
	return n + copy(bs[n:], v)
}

// SizeByteSlice reports the encoded size of v.
func SizeByteSlice(v []byte, lenSerializer binco.Serializer[int]) (size int) {
	length := len(v)
	return lenSerializer.Size(length) + length
}

// SkipByteSlice advances over an encoded byte slice.
func SkipByteSlice(lenSerializer binco.Serializer[int], bs []byte) (n int, err error) {
	length, n, err := lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if err = binco.CheckLength(bs, n, length); err != nil {
		return n, err
	}
	return n + length, nil
}
