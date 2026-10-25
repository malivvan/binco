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
	var lenSer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return textView{lenSer}
}

type textView struct {
	lenSer wirefmt.Serializer[int]
}

func (s textView) Marshal(v string, w wirefmt.Writer) (n int,
	err error,
) {
	return composite.MarshalString(v, s.lenSer, w)
}

func (s textView) Unmarshal(r wirefmt.Reader) (v string,
	n int, err error,
) {
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
	v = unsafe_mod.String(&c[0], len(c))
	return
}

func (s textView) Size(v string) (n int) {
	return composite.SizeString(v, s.lenSer)
}

func (s textView) Skip(r wirefmt.Reader) (n int, err error) {
	return composite.SkipString(r, s.lenSer)
}

type checkedText struct {
	textView
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
	v = unsafe_mod.String(&c[0], len(c))
	return
}
