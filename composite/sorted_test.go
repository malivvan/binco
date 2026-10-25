package composite_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/varnum"
)

// TestSortedMapIsDeterministic checks the reason the sorted codec exists: the
// same map produces the same bytes on every run, and the result still decodes
// with an ordinary map codec.
func TestSortedMapIsDeterministic(t *testing.T) {
	value := map[int]string{3: "c", 1: "a", 2: "b"}
	sorted := composite.NewSortedMapSerializer(varnum.Int, composite.String)

	want := binco.Encode(sorted, value)
	for range 20 {
		if got := binco.Encode(sorted, value); !bytes.Equal(got, want) {
			t.Fatalf("sorted map bytes changed: %v, want %v", got, want)
		}
	}
	if got := sorted.Size(value); got != len(want) {
		t.Fatalf("Size = %d, Marshal wrote %d", got, len(want))
	}
	decoded, err := binco.Decode(composite.NewMapSerializer(varnum.Int, composite.String), want)
	if err != nil {
		t.Fatalf("Decode with the ordinary map codec: %v", err)
	}
	if len(decoded) != len(value) {
		t.Fatalf("Decode returned %d entries, want %d", len(decoded), len(value))
	}
	for k, v := range value {
		if decoded[k] != v {
			t.Fatalf("Decode[%d] = %q, want %q", k, decoded[k], v)
		}
	}
}

// TestSortedMapRejectsShortBuffer checks that a too-small buffer panics with the
// usual sentinel instead of silently writing a truncated entry.
func TestSortedMapRejectsShortBuffer(t *testing.T) {
	value := map[int]int{1: 1, 2: 2}
	sorted := composite.NewSortedMapSerializer(varnum.Int, varnum.Int)
	bs := make([]byte, sorted.Size(value)-1)

	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
			t.Fatalf("panic value %v, want ErrBufferTooSmall", err)
		}
	}()
	sorted.Marshal(value, bs)
}

// shortCodec reports two bytes for every value but writes one. It is a broken
// codec, and the sorted map must report it instead of shifting every entry.
type shortCodec struct{}

func (shortCodec) Marshal(v int, bs []byte) (n int) {
	bs[0] = byte(v)
	return 1
}
func (shortCodec) Unmarshal(bs []byte) (v int, n int, err error) { return int(bs[0]), 1, nil }
func (shortCodec) Size(v int) (size int)                         { return 2 }
func (shortCodec) Skip(bs []byte) (n int, err error)             { return 2, nil }

// TestSortedMapRejectsInconsistentSize checks that a codec whose Marshal
// disagrees with its Size is reported rather than corrupting the payload.
func TestSortedMapRejectsInconsistentSize(t *testing.T) {
	sorted := composite.NewSortedMapSerializer(varnum.Int, shortCodec{})

	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, binco.ErrWrongSize) {
			t.Fatalf("panic value %v, want ErrWrongSize", err)
		}
	}()
	sorted.Marshal(map[int]int{1: 7}, make([]byte, 16))
}
