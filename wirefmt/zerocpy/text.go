package zerocpy

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/internal/wireio"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// String serializes a string; decoding allocates a fresh one, because a stream cannot hand out a view of unread bytes.
var String = NewStringSerializer()

// NewStringSerializer returns a string serializer whose length codec can be replaced through opts.
//
// A WithLenValidator option is honoured.
func NewStringSerializer(opts ...stropts.SetOption) wirefmt.Serializer[string] {
	o := stropts.Options{}
	stropts.Apply(opts, &o)

	s := newTextView(o)
	if o.LenVl != nil {
		return checkedText{s, o.LenVl}
	}
	return s
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

func (s textView) Marshal(v string, w wirefmt.Writer) (n int, err error) {
	return composite.MarshalString(v, s.lenSerializer, w)
}

func (s textView) Unmarshal(r wirefmt.Reader) (v string, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return "", n, binco.ErrNegativeLength
	}
	var n1 int
	var body []byte
	body, n1, err = wireio.ReadFixed(r, length)
	n += n1
	if err != nil {
		return
	}
	return string(body), n, nil
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
		return "", n, binco.ErrNegativeLength
	}
	if s.lenVl != nil {
		if err := s.lenVl.Validate(length); err != nil {
			return "", n, err
		}
	}
	var n1 int
	var body []byte
	body, n1, err = wireio.ReadFixed(r, length)
	n += n1
	if err != nil {
		return
	}
	return string(body), n, nil
}
