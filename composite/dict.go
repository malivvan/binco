package composite

import (
	"github.com/malivvan/binco"
	mapopts "github.com/malivvan/binco/config/map"
	"github.com/malivvan/binco/varnum"
)

// NewMapSerializer returns a map serializer that encodes keys with keySerializer and values with valueSerializer.
func NewMapSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], opts ...mapopts.SetOption[T, V],
) dictCodec[T, V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)

	return newDictCodec(keySerializer, valueSerializer, o)
}

// NewValidMapSerializer returns a map serializer that runs its length, key and value validators while decoding.
func NewValidMapSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], opts ...mapopts.SetOption[T, V],
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
		dictCodec: newDictCodec(keySerializer, valueSerializer, o),
		lenVl:     lenVl,
		keyVl:     keyVl,
		valueVl:   valueVl,
	}
}

func newDictCodec[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], o mapopts.Options[T, V],
) dictCodec[T, V] {
	var lenSerializer binco.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return dictCodec[T, V]{lenSerializer, keySerializer, valueSerializer}
}

type dictCodec[T comparable, V any] struct {
	lenSerializer   binco.Serializer[int]
	keySerializer   binco.Serializer[T]
	valueSerializer binco.Serializer[V]
}

func (s dictCodec[T, V]) Marshal(v map[T]V, bs []byte) (n int) {
	n = s.lenSerializer.Marshal(len(v), bs)
	for k, v := range v {
		n += s.keySerializer.Marshal(k, bs[n:])
		n += s.valueSerializer.Marshal(v, bs[n:])
	}
	return
}

func (s dictCodec[T, V]) Unmarshal(bs []byte) (v map[T]V, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int

	var k T

	var val V

	v = make(map[T]V, length)
	for range length {
		k, n1, err = s.keySerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		val, n1, err = s.valueSerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		v[k] = val
	}
	return
}

func (s dictCodec[T, V]) Size(v map[T]V) (size int) {
	size = s.lenSerializer.Size(len(v))
	for k, v := range v {
		size += s.keySerializer.Size(k)
		size += s.valueSerializer.Size(v)
	}
	return
}

func (s dictCodec[T, V]) Skip(bs []byte) (n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
	if err != nil {
		return
	}
	if length < 0 {
		err = binco.ErrNegativeLength
		return
	}
	var n1 int
	for range length {
		n1, err = s.keySerializer.Skip(bs[n:])
		n += n1
		if err != nil {
			return
		}
		n1, err = s.valueSerializer.Skip(bs[n:])
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

func (s checkedDict[T, V]) Unmarshal(bs []byte) (v map[T]V, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(bs)
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

	var k T

	var val V

	v = make(map[T]V, length)
	for range length {
		k, n1, err = s.keySerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		if s.keyVl != nil {
			if err = s.keyVl.Validate(k); err != nil {
				return
			}
		}
		val, n1, err = s.valueSerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return
		}
		if s.valueVl != nil {
			if err = s.valueVl.Validate(val); err != nil {
				return
			}
		}
		v[k] = val
	}
	return
}
