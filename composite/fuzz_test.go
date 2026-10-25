//go:build fuzz

package composite_test

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	mapopts "github.com/malivvan/binco/config/map"
	slopts "github.com/malivvan/binco/config/slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/varnum"
)

// maxFuzzLength bounds every decoded length in this package. Without a bound a
// fuzzed length prefix asks the decoders for an arbitrarily large container.
// The codecs themselves now cap allocations by the input, so this limit only
// keeps the fuzzer focused on the format.
const maxFuzzLength = 1 << 12

func TestMain(m *testing.M) {
	// S2 was an amplification bug: five bytes used to allocate gigabytes. A
	// memory limit turns a regression into a failing test instead of a killed
	// runner.
	debug.SetMemoryLimit(256 << 20)
	os.Exit(m.Run())
}

var (
	fuzzString    binco.Serializer[string]      = composite.NewStringSerializer(stropts.WithLenValidator(lengthLimit()))
	fuzzByteSlice binco.Serializer[[]byte]      = composite.NewByteSliceSerializer(bslopts.WithLenValidator(lengthLimit()))
	fuzzSlice     binco.Serializer[[]int]       = composite.NewSliceSerializer(varnum.Int, slopts.WithLenValidator[int](lengthLimit()))
	fuzzMap       binco.Serializer[map[int]int] = composite.NewMapSerializer(varnum.Int, varnum.Int, mapopts.WithLenValidator[int, int](lengthLimit()))
	fuzzPtr       binco.Serializer[*int]        = composite.NewPtrSerializer(varnum.Int)
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
		{}, {0}, {1, 'a'}, {3, 'a', 'b', 'c'},
		{0x80, 0x01}, {0x80}, {0xff, 0xff},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "String", fuzzString, bs)
	})
}

func FuzzByteSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 0xff}, {3, 1, 2, 3}, {0x80, 0x01}, {0x80},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "ByteSlice", fuzzByteSlice, bs)
	})
}

func FuzzSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 2}, {2, 2, 4}, {1, 0x80, 0x01}, {1}, {2, 2},
		{0x80, 0x01}, {0x80, 0x80, 0x80, 0x80, 0x01},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "Slice", fuzzSlice, bs)
	})
}

func FuzzMap(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 2, 4}, {2, 2, 4, 6, 8}, {1, 2},
		{0x80, 0x01}, {0x80, 0x80, 0x80, 0x80, 0x01},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "Map", fuzzMap, bs)
	})
}

func FuzzPtr(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {0, 2}, {1}, {2}, {0, 0, 0}, {0xff},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		checkCodec(t, "Ptr", fuzzPtr, bs)
	})
}

// checkCodec asserts the invariants every codec promises: decoding arbitrary
// bytes never panics, the consumed count stays in range, Skip agrees with
// Unmarshal, and a value that decoded round-trips.
func checkCodec[T any](t *testing.T, name string, ser binco.Serializer[T], bs []byte) {
	t.Helper()
	v, n, err := ser.Unmarshal(bs)
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
	skipN, skipErr := ser.Skip(bs)
	if skipErr != nil {
		t.Fatalf("%s: Skip failed after Unmarshal succeeded: %v", name, skipErr)
	}
	if skipN != n {
		t.Fatalf("%s: Skip consumed %d, Unmarshal consumed %d", name, skipN, n)
	}
	out := make([]byte, size)
	if written := ser.Marshal(v, out); written != size {
		t.Fatalf("%s: Marshal wrote %d, Size reported %d", name, written, size)
	}
	v2, n2, err := ser.Unmarshal(out)
	if err != nil {
		t.Fatalf("%s: round-trip Unmarshal failed: %v", name, err)
	}
	if n2 != len(out) {
		t.Fatalf("%s: round-trip consumed %d, Marshal wrote %d", name, n2, len(out))
	}
	if !reflect.DeepEqual(v, v2) {
		t.Fatalf("%s: round-trip value %v != %v", name, v, v2)
	}
}
