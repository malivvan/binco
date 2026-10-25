//go:build fuzz

package composite_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// maxFuzzLength bounds every decoded length. Without a bound, a fuzzed length
// prefix can ask these decoders to allocate an arbitrarily large container.
const maxFuzzLength = 1 << 12

var (
	fuzzString    wirefmt.Serializer[string]      = composite.NewValidStringSerializer(stropts.WithLenValidator(lengthLimit()))
	fuzzByteSlice wirefmt.Serializer[[]byte]      = composite.NewValidByteSliceSerializer(bslopts.WithLenValidator(lengthLimit()))
	fuzzSlice     wirefmt.Serializer[[]int]       = composite.NewValidSliceSerializer(varnum.Int, slopts.WithLenValidator[int](lengthLimit()))
	fuzzMap       wirefmt.Serializer[map[int]int] = composite.NewValidMapSerializer(varnum.Int, varnum.Int, mapopts.WithLenValidator[int, int](lengthLimit()))
)

func lengthLimit() binco.Validator[int] {
	return binco.ValidatorFn[int](func(length int) error {
		if length < 0 || length > maxFuzzLength {
			return fmt.Errorf("length %d exceeds %d", length, maxFuzzLength)
		}
		return nil
	})
}

func FuzzString(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0},
		{1, 'a'},
		{3, 'a', 'b', 'c'},
		{0x80, 0x01},
		{0x80},
		{0xff, 0xff},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "String", fuzzString, bs)
	})
}

func FuzzByteSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0},
		{1, 0xff},
		{3, 1, 2, 3},
		{0x80, 0x01},
		{0x80},
		{0xff, 0xff},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "ByteSlice", fuzzByteSlice, bs)
	})
}

func FuzzSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0},
		{1, 2},
		{2, 2, 4},
		{1, 0x80, 0x01},
		{1},
		{2, 2},
		{0x80, 0x01},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "Slice", fuzzSlice, bs)
	})
}

func FuzzMap(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0},
		{1, 2, 4},
		{2, 2, 4, 6, 8},
		{1, 2},
		{1, 2, 4, 6, 8},
		{0x80, 0x01},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "Map", fuzzMap, bs)
	})
}

func checkCodec[T any](t *testing.T, name string, ser wirefmt.Serializer[T],
	bs []byte,
) {
	t.Helper()
	v, n, err := ser.Unmarshal(bytes.NewReader(bs))
	if n < 0 || n > len(bs) {
		t.Fatalf("%s: Unmarshal consumed %d of %d bytes", name, n, len(bs))
	}
	if err != nil {
		return
	}
	size := ser.Size(v)
	if size <= 0 || size > n {
		t.Fatalf("%s: Size = %d, Unmarshal consumed %d", name, size, n)
	}
	skipN, skipErr := ser.Skip(bytes.NewReader(bs))
	if skipErr != nil {
		t.Fatalf("%s: Skip failed after Unmarshal succeeded: %v", name, skipErr)
	}
	if skipN != n {
		t.Fatalf("%s: Skip consumed %d, Unmarshal consumed %d", name, skipN, n)
	}
	var buf bytes.Buffer
	written, marshalErr := ser.Marshal(v, &buf)
	if marshalErr != nil {
		t.Fatalf("%s: Marshal after Unmarshal failed: %v", name, marshalErr)
	}
	if written != size {
		t.Fatalf("%s: Marshal wrote %d, Size reported %d", name, written, size)
	}
	v2, n2, err := ser.Unmarshal(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("%s: round-trip Unmarshal failed: %v", name, err)
	}
	if n2 != written {
		t.Fatalf("%s: round-trip consumed %d, Marshal wrote %d", name, n2, written)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Fatalf("%s: round-trip value %v != %v", name, v, v2)
	}
}
