package zerocpy

import (
	"io"
	unsafe_mod "unsafe"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// String serializes a string; decoding allocates a fresh one, because a stream cannot hand out a view of unread bytes.
var String = NewStringSerializer()

// NewStringSerializer returns a string serializer whose length codec can be replaced through opts.
func NewStringSerializer(opts ...stropts.SetOption) textView {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	return newTextView(o)
}

// NewValidStringSerializer returns a string serializer that runs its length validator while decoding.
func NewValidStringSerializer(opts ...stropts.SetOption) checkedText {
	o := stropts.Options{}
	stropts.Apply(opts, &o)
	return checkedText{newTextView(o), o.LenVl}
}

func newTextView(o stropts.Options) textView {
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return textView{lenSerializer}
}

type textView struct {
	lenSerializer wirefmt.Serializer[int]
}

func (s textView) Marshal(v string, w wirefmt.Writer) (n int,
	err error,
) {
	return composite.MarshalString(v, s.lenSerializer, w)
}

func (s textView) Unmarshal(r wirefmt.Reader) (v string,
	n int, err error,
) {
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
	v = unsafe_mod.String(&c[0], len(c))
	return
}

func (s textView) Size(v string) (n int) {
	return composite.SizeString(v, s.lenSerializer)
}

func (s textView) Skip(r wirefmt.Reader) (n int, err error) {
	return composite.SkipString(r, s.lenSerializer)
}

type checkedText struct {
	textView
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
	v = unsafe_mod.String(&c[0], len(c))
	return
}
