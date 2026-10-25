package composite

import (
	"fmt"
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
	return newMapSerializer(newDictCodec(keySerializer, valueSerializer, o), o)
}

// newMapSerializer wraps the plain codec when any validator is set.
func newMapSerializer[T comparable, V any](dict dictCodec[T, V],
	o mapopts.Options[T, V],
) binco.Serializer[map[T]V] {
	if o.LenVl == nil && o.KeyVl == nil && o.ValueVl == nil {
		return dict
	}
	return checkedDict[T, V]{
		dictCodec: dict,
		lenVl:     o.LenVl,
		keyVl:     o.KeyVl,
		valueVl:   o.ValueVl,
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

	dict := newDictCodec(keySerializer, valueSerializer, o)
	return sortedDictCodec[T, V]{
		serializer: newMapSerializer(dict, o),
		key:        keySerializer,
		value:      valueSerializer,
		lenSer:     dict.lenSerializer,
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

// encodeEntry marshals one map entry into a buffer of exactly the size its
// serializer reported.
//
// Marshal has no error return, so a codec that writes a different number of
// bytes than it reported is reported with a panic: writing it out would shift
// every following entry and corrupt the payload.
func encodeEntry[T any](serializer binco.Serializer[T], v T) []byte {
	bs := make([]byte, serializer.Size(v))
	if n := serializer.Marshal(v, bs); n != len(bs) {
		panic(fmt.Errorf("%w: serializer wrote %d of %d bytes", binco.ErrWrongSize, n, len(bs)))
	}
	return bs
}

func (s sortedDictCodec[T, V]) Marshal(v map[T]V, bs []byte) (n int) {
	entries := make([]dictEntry, 0, len(v))
	for k, val := range v {
		entries = append(entries, dictEntry{
			key:   encodeEntry(s.key, k),
			value: encodeEntry(s.value, val),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return string(entries[i].key) < string(entries[j].key)
	})
	n = s.lenSer.Marshal(len(v), bs)
	for _, e := range entries {
		binco.CheckWrite(bs, n, len(e.key))
		n += copy(bs[n:], e.key)
		binco.CheckWrite(bs, n, len(e.value))
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
