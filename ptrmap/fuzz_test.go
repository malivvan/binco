//go:build fuzz

package ptrmap_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/ptrmap"
	"github.com/malivvan/binco/varnum"
)

// FuzzPointer decodes arbitrary bytes as a pointer. A payload of nested
// references used to recurse until the stack died; the codec now stops with
// ErrTooDeep, so the fuzzer can explore the space safely.
func FuzzPointer(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1}, {2}, {2, 0}, {2, 0, 2}, {2, 0, 4},
		{1, 1, 1, 1, 1, 1, 1, 1},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		ptrMap := binco.NewPtrMap()
		revPtrMap := binco.NewReversePtrMap()
		serializer := ptrmap.Wrap(ptrMap, revPtrMap,
			ptrmap.NewPtrSerializer(ptrMap, revPtrMap, varnum.Int))

		_, n, err := serializer.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err != nil {
			return
		}
		if skipN, skipErr := serializer.Skip(bs); skipErr != nil || skipN != n {
			t.Fatalf("Skip = (%d, %v), Unmarshal consumed %d", skipN, skipErr, n)
		}
	})
}

// FuzzPointerSlice decodes arbitrary bytes as a slice of shared pointers.
func FuzzPointerSlice(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1, 2, 0}, {2, 2, 0, 2, 0}, {1, 1},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		ptrMap := binco.NewPtrMap()
		revPtrMap := binco.NewReversePtrMap()
		element := ptrmap.NewPtrSerializer(ptrMap, revPtrMap, varnum.Int)
		serializer := ptrmap.Wrap(ptrMap, revPtrMap, composite.NewSliceSerializer(element))

		_, n, err := serializer.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		_ = err
	})
}
