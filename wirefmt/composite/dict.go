package composite

import (
	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewMapSer returns a map serializer for keySer and valueSer.
func NewMapSer[T comparable, V any](keySer wirefmt.Serializer[T],
	valueSer wirefmt.Serializer[V], opts ...mapopts.SetOption[T, V],
) dictCodec[T, V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)

	return newDictCodec(keySer, valueSer, o)
}

// NewValidMapSer returns a map serializer that runs its length, key and value validators while decoding.
func NewValidMapSer[T comparable, V any](keySer wirefmt.Serializer[T],
	valueSer wirefmt.Serializer[V], opts ...mapopts.SetOption[T, V],
) checkedDict[T, V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)

	var lenVl binco.Validator[int]

	var keyVl binco.Validator[T]

	var valueVl binco.Validator[V]

	if o.LenVl != nil {
		lenVl = o.LenVl
	}
	if o.KeyVl != nil {
		keyVl = o.KeyVl
	}
	if o.ValueVl != nil {
		valueVl = o.ValueVl
	}

	return checkedDict[T, V]{
		dictCodec: newDictCodec(keySer, valueSer, o),
		lenVl:     lenVl,
		keyVl:     keyVl,
		valueVl:   valueVl,
	}
}

func newDictCodec[T comparable, V any](keySer wirefmt.Serializer[T],
	valueSer wirefmt.Serializer[V], o mapopts.Options[T, V],
) dictCodec[T, V] {
	var lenSer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSer != nil {
		lenSer = o.LenSer
	}
	return dictCodec[T, V]{lenSer, keySer, valueSer}
}

type dictCodec[T comparable, V any] struct {
	lenSer wirefmt.Serializer[int]
	keySer wirefmt.Serializer[T]
	valSer wirefmt.Serializer[V]
}

func (s dictCodec[T, V]) Marshal(v map[T]V, w wirefmt.Writer) (n int, err error) {
	n, err = s.lenSer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	for k, v := range v {
		n1, err = s.keySer.Marshal(k, w)
		n += n1
		if err != nil {
			return
		}
		n1, err = s.valSer.Marshal(v, w)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

func (s dictCodec[T, V]) Unmarshal(r wirefmt.Reader) (v map[T]V, n int, err error) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int

	var k T

	var p V

	v = make(map[T]V, length)
	for range length {
		k, n1, err = s.keySer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		p, n1, err = s.valSer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		v[k] = p
	}
	return
}

func (s dictCodec[T, V]) Size(v map[T]V) (size int) {
	size = s.lenSer.Size(len(v))
	for k, v := range v {
		size += s.keySer.Size(k)
		size += s.valSer.Size(v)
	}
	return
}

func (s dictCodec[T, V]) Skip(r wirefmt.Reader) (
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
	var n1 int
	for range length {
		n1, err = s.keySer.Skip(r)
		n += n1
		if err != nil {
			return
		}
		n1, err = s.valSer.Skip(r)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

type checkedDict[T comparable, V any] struct {
	dictCodec[T, V]
	lenVl   binco.Validator[int]
	keyVl   binco.Validator[T]
	valueVl binco.Validator[V]
}

func (s checkedDict[T, V]) Unmarshal(r wirefmt.Reader) (v map[T]V, n int,
	err error,
) {
	length, n, err := s.lenSer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int

	var k T

	var p V

	if s.lenVl != nil {
		if err = s.lenVl.Validate(length); err != nil {
			return
		}
	}
	v = make(map[T]V, length)
	for range length {
		k, n1, err = s.keySer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		if s.keyVl != nil {
			if err = s.keyVl.Validate(k); err != nil {
				return
			}
		}
		p, n1, err = s.valSer.Unmarshal(r)
		n += n1
		if err != nil {
			return
		}
		if s.valueVl != nil {
			if err = s.valueVl.Validate(p); err != nil {
				return
			}
		}
		v[k] = p
	}
	return
}
