package composite

import (
	"fmt"
	"sort"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// NewMapSerializer returns a map serializer that encodes keys with keySerializer and values with valueSerializer.
//
// Length, key and value validators passed through opts are honoured.
func NewMapSerializer[T comparable, V any](keySerializer wirefmt.Serializer[T],
	valueSerializer wirefmt.Serializer[V], opts ...mapopts.SetOption[T, V],
) wirefmt.Serializer[map[T]V] {
	o := mapopts.Options[T, V]{}
	mapopts.Apply(opts, &o)
	return newMapSerializer(newDictCodec(keySerializer, valueSerializer, o), o)
}

// newMapSerializer wraps the plain codec when any validator is set.
func newMapSerializer[T comparable, V any](dict dictCodec[T, V],
	o mapopts.Options[T, V],
) wirefmt.Serializer[map[T]V] {
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
// different payload for the same value on every run. This serializer sorts the
// entries by their encoded key bytes before writing them; the wire format is
// otherwise identical and decodes with any map codec.
func NewSortedMapSerializer[T comparable, V any](keySerializer wirefmt.Serializer[T],
	valueSerializer wirefmt.Serializer[V], opts ...mapopts.SetOption[T, V],
) wirefmt.Serializer[map[T]V] {
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
	serializer wirefmt.Serializer[map[T]V]
	key        wirefmt.Serializer[T]
	value      wirefmt.Serializer[V]
	lenSer     wirefmt.Serializer[int]
}

type dictEntry struct {
	key   []byte
	value []byte
}

// encodeBuffer marshals one map entry into a buffer of exactly the size its
// serializer reported.
//
// An element serializer that fails or writes a different number of bytes than it
// reported would shift every following entry, so the error is returned to the
// caller (or, for the size mismatch, reported as one) rather than written out as
// a corrupt payload.
func encodeBuffer(size int, marshal func(*wirefmt.SliceWriter) (int, error)) ([]byte, error) {
	w := wirefmt.NewSliceWriter(make([]byte, size))
	n, err := marshal(w)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("%w: serializer wrote %d of %d bytes", binco.ErrWrongSize, n, size)
	}
	return w.Buf[:n], nil
}

func (s sortedDictCodec[T, V]) Marshal(v map[T]V, w wirefmt.Writer) (n int, err error) {
	entries := make([]dictEntry, 0, len(v))
	for k, val := range v {
		kb, encodeErr := encodeBuffer(s.key.Size(k), func(buf *wirefmt.SliceWriter) (int, error) {
			return s.key.Marshal(k, buf)
		})
		if encodeErr != nil {
			return 0, encodeErr
		}
		vb, encodeErr := encodeBuffer(s.value.Size(val), func(buf *wirefmt.SliceWriter) (int, error) {
			return s.value.Marshal(val, buf)
		})
		if encodeErr != nil {
			return 0, encodeErr
		}
		entries = append(entries, dictEntry{key: kb, value: vb})
	}
	sort.Slice(entries, func(i, j int) bool {
		return string(entries[i].key) < string(entries[j].key)
	})
	n, err = s.lenSer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	for _, e := range entries {
		n1, err = w.Write(e.key)
		n += n1
		if err != nil {
			return
		}
		n1, err = w.Write(e.value)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

func (s sortedDictCodec[T, V]) Unmarshal(r wirefmt.Reader) (map[T]V, int, error) {
	return s.serializer.Unmarshal(r)
}

func (s sortedDictCodec[T, V]) Size(v map[T]V) (size int) {
	return s.serializer.Size(v)
}

func (s sortedDictCodec[T, V]) Skip(r wirefmt.Reader) (int, error) {
	return s.serializer.Skip(r)
}

func newDictCodec[T comparable, V any](keySerializer wirefmt.Serializer[T],
	valueSerializer wirefmt.Serializer[V], o mapopts.Options[T, V],
) dictCodec[T, V] {
	var lenSerializer wirefmt.Serializer[int] = varnum.PositiveInt
	if o.LenSerializer != nil {
		lenSerializer = o.LenSerializer
	}
	return dictCodec[T, V]{lenSerializer, keySerializer, valueSerializer}
}

type dictCodec[T comparable, V any] struct {
	lenSerializer wirefmt.Serializer[int]
	keySerializer wirefmt.Serializer[T]
	valSerializer wirefmt.Serializer[V]
}

func (s dictCodec[T, V]) Marshal(v map[T]V, w wirefmt.Writer) (n int, err error) {
	n, err = s.lenSerializer.Marshal(len(v), w)
	if err != nil {
		return
	}
	var n1 int
	for k, v := range v {
		n1, err = s.keySerializer.Marshal(k, w)
		n += n1
		if err != nil {
			return
		}
		n1, err = s.valSerializer.Marshal(v, w)
		n += n1
		if err != nil {
			return
		}
	}
	return
}

func (s dictCodec[T, V]) Unmarshal(r wirefmt.Reader) (v map[T]V, n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return nil, n, binco.ErrNegativeLength
	}
	return s.unmarshalPairs(r, n, length, nil, nil, nil)
}

// unmarshalPairs decodes length key/value pairs. The validators may be nil.
//
// A pair that consumes no input stops the loop: further pairs cannot make
// progress, so the declared length would only spin.
func (s dictCodec[T, V]) unmarshalPairs(r wirefmt.Reader, n, length int,
	lenVl binco.Validator[int], keyVl binco.Validator[T], valueVl binco.Validator[V],
) (v map[T]V, end int, err error) {
	if lenVl != nil {
		if err = lenVl.Validate(length); err != nil {
			return nil, n, err
		}
	}
	v = make(map[T]V, hintLength(length))
	var n1 int

	var k T

	var p V

	for range length {
		k, n1, err = s.keySerializer.Unmarshal(r)
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
		p, n1, err = s.valSerializer.Unmarshal(r)
		n += n1
		if err != nil {
			return nil, n, err
		}
		if valueVl != nil {
			if err = valueVl.Validate(p); err != nil {
				return nil, n, err
			}
		}
		v[k] = p
		if keyLen == 0 && n1 == 0 {
			return v, n, nil
		}
	}
	return v, n, nil
}

func (s dictCodec[T, V]) Size(v map[T]V) (size int) {
	size = s.lenSerializer.Size(len(v))
	for k, v := range v {
		size += s.keySerializer.Size(k)
		size += s.valSerializer.Size(v)
	}
	return
}

func (s dictCodec[T, V]) Skip(r wirefmt.Reader) (n int, err error) {
	length, n, err := s.lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return n, binco.ErrNegativeLength
	}
	var n1 int
	for range length {
		n1, err = s.keySerializer.Skip(r)
		n += n1
		if err != nil {
			return
		}
		keyLen := n1
		n1, err = s.valSerializer.Skip(r)
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

func (s checkedDict[T, V]) Unmarshal(r wirefmt.Reader) (v map[T]V, n int,
	err error,
) {
	length, n, err := s.lenSerializer.Unmarshal(r)
	if err != nil {
		return
	}
	if length < 0 {
		return nil, n, binco.ErrNegativeLength
	}
	return s.unmarshalPairs(r, n, length, s.lenVl, s.keyVl, s.valueVl)
}
