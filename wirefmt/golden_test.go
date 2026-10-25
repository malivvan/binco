package wirefmt_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

// The buffer packages and wirefmt implement the same format twice (D3). These
// vectors run one set of values through both sides in both directions: a
// buffer payload must decode with the streaming codec, a stream payload must
// decode with the buffer codec, and the bytes must be identical. Any drift
// between the implementations turns into a failure here instead of a corrupted
// stream in production.
func TestGoldenVectorsCrossTheImplementations(t *testing.T) {
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)

	tests := []struct {
		name string
		// buffer encodes v; stream encodes the same v.
		buffer func() []byte
		stream func() []byte
		// decodeBuffer decodes with the streaming codec from buffer bytes.
		decodeStream func(bs []byte) error
		// decodeBuffer decodes with the buffer codec from stream bytes.
		decodeBuffer func(bs []byte) error
	}{
		{
			name:         "int",
			buffer:       func() []byte { return binco.Encode(varnum.Int, 300) },
			stream:       func() []byte { return streamBytes(t, wvarnum.Int, 300) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wvarnum.Int, bs, 300) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, varnum.Int, bs, 300) },
		},
		{
			name:         "uint64",
			buffer:       func() []byte { return binco.Encode(varnum.Uint64, 1<<40) },
			stream:       func() []byte { return streamBytes(t, wvarnum.Uint64, uint64(1<<40)) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wvarnum.Uint64, bs, uint64(1<<40)) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, varnum.Uint64, bs, uint64(1<<40)) },
		},
		{
			name:         "scalar int32",
			buffer:       func() []byte { return binco.Encode(scalar.Int32, -12345) },
			stream:       func() []byte { return streamBytes(t, wscalar.Int32, int32(-12345)) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wscalar.Int32, bs, int32(-12345)) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, scalar.Int32, bs, int32(-12345)) },
		},
		{
			name:         "float64",
			buffer:       func() []byte { return binco.Encode(scalar.Float64, 1.5) },
			stream:       func() []byte { return streamBytes(t, wscalar.Float64, 1.5) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wscalar.Float64, bs, 1.5) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, scalar.Float64, bs, 1.5) },
		},
		{
			name:         "time",
			buffer:       func() []byte { return binco.Encode(scalar.TimeUnixUTC, ts) },
			stream:       func() []byte { return streamBytes(t, wscalar.TimeUnixUTC, ts) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wscalar.TimeUnixUTC, bs, ts) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, scalar.TimeUnixUTC, bs, ts) },
		},
		{
			name:         "bool",
			buffer:       func() []byte { return binco.Encode(composite.Bool, true) },
			stream:       func() []byte { return streamBytes(t, wcomposite.Bool, true) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wcomposite.Bool, bs, true) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, composite.Bool, bs, true) },
		},
		{
			name:         "string",
			buffer:       func() []byte { return binco.Encode(composite.String, "hello") },
			stream:       func() []byte { return streamBytes(t, wcomposite.String, "hello") },
			decodeStream: func(bs []byte) error { return decodeStream(t, wcomposite.String, bs, "hello") },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, composite.String, bs, "hello") },
		},
		{
			name:         "byte slice",
			buffer:       func() []byte { return binco.Encode(composite.ByteSlice, []byte{1, 2, 3}) },
			stream:       func() []byte { return streamBytes(t, wcomposite.ByteSlice, []byte{1, 2, 3}) },
			decodeStream: func(bs []byte) error { return decodeStream(t, wcomposite.ByteSlice, bs, []byte{1, 2, 3}) },
			decodeBuffer: func(bs []byte) error { return decodeBuffer(t, composite.ByteSlice, bs, []byte{1, 2, 3}) },
		},
		{
			name: "slice",
			buffer: func() []byte {
				return binco.Encode(composite.NewSliceSerializer(varnum.Int), []int{1, -1, 300})
			},
			stream: func() []byte {
				return streamBytes(t, wcomposite.NewSliceSerializer(wvarnum.Int), []int{1, -1, 300})
			},
			decodeStream: func(bs []byte) error {
				return decodeStream(t, wcomposite.NewSliceSerializer(wvarnum.Int), bs, []int{1, -1, 300})
			},
			decodeBuffer: func(bs []byte) error {
				return decodeBuffer(t, composite.NewSliceSerializer(varnum.Int), bs, []int{1, -1, 300})
			},
		},
		{
			name: "sorted map",
			buffer: func() []byte {
				return binco.Encode(composite.NewSortedMapSerializer(varnum.Int, varnum.Int),
					map[int]int{2: 20, 1: 10, 3: 30})
			},
			stream: func() []byte {
				return streamBytes(t, wcomposite.NewSortedMapSerializer(wvarnum.Int, wvarnum.Int),
					map[int]int{2: 20, 1: 10, 3: 30})
			},
			decodeStream: func(bs []byte) error {
				return decodeStream(t, wcomposite.NewSortedMapSerializer(wvarnum.Int, wvarnum.Int),
					bs, map[int]int{2: 20, 1: 10, 3: 30})
			},
			decodeBuffer: func(bs []byte) error {
				return decodeBuffer(t, composite.NewSortedMapSerializer(varnum.Int, varnum.Int),
					bs, map[int]int{2: 20, 1: 10, 3: 30})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bufferBS := tc.buffer()
			streamBS := tc.stream()
			if !bytes.Equal(bufferBS, streamBS) {
				t.Fatalf("buffer bytes %v, stream bytes %v", bufferBS, streamBS)
			}
			if err := tc.decodeStream(bufferBS); err != nil {
				t.Fatalf("decoding buffer bytes with the stream codec: %v", err)
			}
			if err := tc.decodeBuffer(streamBS); err != nil {
				t.Fatalf("decoding stream bytes with the buffer codec: %v", err)
			}
		})
	}
}

func streamBytes[T any](t *testing.T, s wirefmt.Serializer[T], v T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := wirefmt.Encode(s, &buf, v); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return buf.Bytes()
}

func decodeStream[T any](t *testing.T, s wirefmt.Serializer[T], bs []byte, want T) error {
	t.Helper()
	got, n, err := s.Unmarshal(wirefmt.NewReader(bytes.NewReader(bs)))
	if err != nil {
		return err
	}
	if n != len(bs) {
		t.Fatalf("stream Unmarshal read %d of %d bytes", n, len(bs))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stream Unmarshal = %v, want %v", got, want)
	}
	return nil
}

func decodeBuffer[T any](t *testing.T, s binco.Serializer[T], bs []byte, want T) error {
	t.Helper()
	got, n, err := s.Unmarshal(bs)
	if err != nil {
		return err
	}
	if n != len(bs) {
		t.Fatalf("buffer Unmarshal read %d of %d bytes", n, len(bs))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buffer Unmarshal = %v, want %v", got, want)
	}
	return nil
}
