package composite

import (
	"sort"

	"github.com/malivvan/binco"
	mapopts "github.com/malivvan/binco/config/map"
	"github.com/malivvan/binco/varnum"
)

// NewMapSerializer returns a map serializer that encodes keys with keySerializer and values with valueSerializer.
//
// Length, key and value validators passed through opts are honoured; there is
// no separate constructor to remember.
func NewMapSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], opts ...mapopts.SetOption[T, V],
) binco.Serializer[map[T]V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)

	s := newDictCodec(keySerializer, valueSerializer, o)
	if o.LenVl != nil || o.KeyVl != nil || o.ValueVl != nil {
		return checkedDict[T, V]{
			dictCodec: s,
			lenVl:     o.LenVl,
			keyVl:     o.KeyVl,
			valueVl:   o.ValueVl,
		}
	}
	return s
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

// NewSortedMapSerializer is NewMapSerializer with a deterministic byte order.
//
// A Go map iterates in a random order, so the default codec produces a
// different payload for the same value on every run. That is fine for
// transport, but it defeats golden tests, content hashes and caches. This
// serializer sorts the entries by their encoded key bytes before writing them;
// the wire format is otherwise identical and decodes with any map codec.
func NewSortedMapSerializer[T comparable, V any](keySerializer binco.Serializer[T],
	valueSerializer binco.Serializer[V], opts ...mapopts.SetOption[T, V],
) binco.Serializer[map[T]V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)

	inner := NewMapSerializer(keySerializer, valueSerializer, opts...)
	return sortedDictCodec[T, V]{
		serializer: inner,
		key:        keySerializer,
		value:      valueSerializer,
		lenSer:     newDictCodec(keySerializer, valueSerializer, o).lenSerializer,
	}
}

type sortedDictCodec[T comparable, V any] struct {
	serializer binco.Serializer[map[T]V]
	key        binco.Serializer[T]
	value      binco.Serializer[V]
	lenSer     binco.Serializer[int]
}

type dictEntry struct {
	key   []byte
	value []byte
}

func (s sortedDictCodec[T, V]) Marshal(v map[T]V, bs []byte) (n int) {
	entries := make([]dictEntry, 0, len(v))
	for k, val := range v {
		kb := make([]byte, s.key.Size(k))
		s.key.Marshal(k, kb)
		vb := make([]byte, s.value.Size(val))
		s.value.Marshal(val, vb)
		entries = append(entries, dictEntry{key: kb, value: vb})
	}
	sort.Slice(entries, func(i, j int) bool {
		return string(entries[i].key) < string(entries[j].key)
	})
	n = s.lenSer.Marshal(len(v), bs)
	for _, e := range entries {
		binco.CheckWrite(bs, n, 0)
		n += copy(bs[n:], e.key)
		binco.CheckWrite(bs, n, 0)
		n += copy(bs[n:], e.value)
	}
	return
}

func (s sortedDictCodec[T, V]) Unmarshal(bs []byte) (map[T]V, int, error) {
	return s.serializer.Unmarshal(bs)
}

func (s sortedDictCodec[T, V]) Size(v map[T]V) (size int) {
	return s.serializer.Size(v)
}

func (s sortedDictCodec[T, V]) Skip(bs []byte) (int, error) {
	return s.serializer.Skip(bs)
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
		binco.CheckWrite(bs, n, 0)
		n += s.keySerializer.Marshal(k, bs[n:])
		binco.CheckWrite(bs, n, 0)
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
	v, n, err = s.unmarshalPairs(bs, n, length, nil, nil, nil)
	return
}

// unmarshalPairs decodes length key/value pairs. The validators may be nil.
//
// The map is sized from the bytes that are actually available, and a pair that
// consumes no input stops the loop: further pairs would repeat the same entry.
func (s dictCodec[T, V]) unmarshalPairs(bs []byte, n, length int,
	lenVl binco.Validator[int], keyVl binco.Validator[T], valueVl binco.Validator[V],
) (v map[T]V, end int, err error) {
	if lenVl != nil {
		if err = lenVl.Validate(length); err != nil {
			return nil, n, err
		}
	}
	// Each pair needs at least two bytes; the hint only avoids rehashing, the
	// actual size follows the decoded entries.
	hint := binco.AllocHint(length, binco.Remaining(bs, n)/2)
	v = make(map[T]V, hint)
	var n1 int

	var k T

	var val V

	for range length {
		k, n1, err = s.keySerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return nil, n, err
		}
		if keyVl != nil {
			if err = keyVl.Validate(k); err != nil {
				return nil, n, err
			}
		}
		keyLen := n1
		val, n1, err = s.valueSerializer.Unmarshal(bs[n:])
		n += n1
		if err != nil {
			return nil, n, err
		}
		if valueVl != nil {
			if err = valueVl.Validate(val); err != nil {
				return nil, n, err
			}
		}
		v[k] = val
		if keyLen == 0 && n1 == 0 {
			// Neither the key nor the value consumed input. Decoding further
			// pairs cannot make progress, so stop instead of spinning for as
			// long as the declared length says.
			return v, n, nil
		}
	}
	return v, n, nil
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
		keyLen := n1
		n1, err = s.valueSerializer.Skip(bs[n:])
		n += n1
		if err != nil {
			return
		}
		if keyLen == 0 && n1 == 0 {
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
	return s.unmarshalPairs(bs, n, length, s.lenVl, s.keyVl, s.valueVl)
}
