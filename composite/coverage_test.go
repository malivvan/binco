package composite_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	mapopts "github.com/malivvan/binco/config/map"
	slopts "github.com/malivvan/binco/config/slice"
	stropts "github.com/malivvan/binco/config/string"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

// TestBlobDecodersReportATruncatedLength checks the error return every
// length-prefixed decoder has for input that ends before its length.
func TestBlobDecodersReportATruncatedLength(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	checkedBlob := composite.NewByteSliceSerializer(bslopts.WithLenValidator(never))
	checkedText := composite.NewStringSerializer(stropts.WithLenValidator(never))
	tests := []struct {
		name string
		call func() error
	}{
		{"String.Unmarshal", func() error { _, _, err := composite.String.Unmarshal(nil); return err }},
		{"String.Unmarshal long", func() error { _, _, err := composite.String.Unmarshal([]byte{5, 'a'}); return err }},
		{"ByteSlice.Unmarshal long", func() error { _, _, err := composite.ByteSlice.Unmarshal([]byte{5, 1}); return err }},
		{"SkipString", func() error { _, err := composite.SkipString(varnum.PositiveInt, nil); return err }},
		{"ByteSlice.Unmarshal", func() error { _, _, err := composite.ByteSlice.Unmarshal(nil); return err }},
		{"checkedText.Unmarshal", func() error { _, _, err := checkedText.Unmarshal(nil); return err }},
		{"checkedBlob.Unmarshal", func() error { _, _, err := checkedBlob.Unmarshal(nil); return err }},
		{"SkipByteSlice", func() error { _, err := composite.SkipByteSlice(varnum.PositiveInt, nil); return err }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, binco.ErrTruncated) {
				t.Fatalf("%s of an empty input = %v, want ErrTruncated", tc.name, err)
			}
		})
	}
}

// TestSortedMapUnmarshalAndSkip checks that the sorted codec decodes and skips
// through the ordinary map codec, so its payloads are symmetric.
func TestSortedMapUnmarshalAndSkip(t *testing.T) {
	value := map[int]string{1: "a", 2: "b"}
	sorted := composite.NewSortedMapSerializer(varnum.Int, composite.String)
	bs := binco.Encode(sorted, value)

	got, n, err := sorted.Unmarshal(bs)
	if err != nil || n != len(bs) || len(got) != len(value) {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}
	if n, err := sorted.Skip(bs); err != nil || n != len(bs) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}
}

// errorSerializer fails to decode, for the element-error path.
type errorSerializer struct{}

func (errorSerializer) Marshal(v int, bs []byte) (n int) { return 0 }
func (errorSerializer) Unmarshal(bs []byte) (v int, n int, err error) {
	return 0, 0, binco.ErrWrongFormat
}
func (errorSerializer) Size(v int) (size int)             { return 0 }
func (errorSerializer) Skip(bs []byte) (n int, err error) { return 0, binco.ErrWrongFormat }

// TestZeroSizeElementsFillInsteadOfLooping checks the shortcut for elements that
// consume nothing, and the cap that keeps a huge length from spinning.
func TestZeroSizeElementsFillInsteadOfLooping(t *testing.T) {
	slice := composite.NewSliceSerializer[struct{}](zeroSerializer{})
	payload := binco.Encode(varnum.PositiveInt, 3)

	v, n, err := slice.Unmarshal(payload)
	if err != nil || len(v) != 3 || n != len(payload) {
		t.Fatalf("Unmarshal = (%v, %d, %v)", v, n, err)
	}
	if n, err := slice.Skip(payload); err != nil || n != len(payload) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}

	tooMany := binco.Encode(varnum.PositiveInt, binco.MaxZeroSizeElements+1)
	if _, _, err := slice.Unmarshal(tooMany); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Unmarshal above the cap = %v, want ErrTooLargeLength", err)
	}
	if _, err := slice.Skip(tooMany); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Skip above the cap = %v, want ErrTooLargeLength", err)
	}
}

// TestMapStopsWhenNothingIsConsumed checks that a map with zero-size keys and
// values stops after the first pair instead of spinning on the declared length.
func TestMapStopsWhenNothingIsConsumed(t *testing.T) {
	m := composite.NewMapSerializer[struct{}, struct{}](zeroSerializer{}, zeroSerializer{})
	payload := binco.Encode(varnum.PositiveInt, 3)

	v, n, err := m.Unmarshal(payload)
	if err != nil || len(v) != 1 || n != len(payload) {
		t.Fatalf("Unmarshal = (%v, %d, %v), want one pair", v, n, err)
	}
	if n, err := m.Skip(payload); err != nil || n != len(payload) {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}
}

// TestCheckedMapLengthErrors checks the validating map codec's own length
// handling: a failed length read and a negative length.
func TestCheckedMapLengthErrors(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	checked := composite.NewMapSerializer(varnum.Int, varnum.Int,
		mapopts.WithLenValidator[int, int](never))
	if _, _, err := checked.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Unmarshal of an empty input = %v, want ErrTruncated", err)
	}

	negative := composite.NewMapSerializer(varnum.Int, varnum.Int,
		mapopts.WithLenSerializer[int, int](varnum.Int),
		mapopts.WithLenValidator[int, int](never))
	if _, _, err := negative.Unmarshal(binco.Encode(varnum.Int, -1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}
}

// TestSliceLengthErrors checks the slice codec's negative and missing lengths,
// and the fixed-width fast path for an empty slice.
func TestSliceLengthErrors(t *testing.T) {
	if bs := binco.Encode(composite.NewSliceSerializer(scalar.Int32), []int32{}); len(bs) != 1 {
		t.Fatalf("empty fixed-width slice = %v, want just the length", bs)
	}

	negative := composite.NewSliceSerializer(varnum.Int, slopts.WithLenSerializer[int](varnum.Int))
	if _, _, err := negative.Unmarshal([]byte{0x01}); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}
	if _, err := negative.Skip([]byte{0x01}); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Skip of a negative length = %v, want ErrNegativeLength", err)
	}
	if _, _, err := negative.Unmarshal(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Unmarshal of an empty input = %v, want ErrTruncated", err)
	}
	if _, err := negative.Skip(nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Skip of an empty input = %v, want ErrTruncated", err)
	}

	// A fixed-width skip checks the element count before advancing.
	if _, err := composite.NewSliceSerializer(scalar.Int32).Skip([]byte{0x02}); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("Skip of a short fixed-width slice = %v, want ErrTruncated", err)
	}

	// An empty fixed-width slice decodes without reading anything.
	empty, n, err := composite.NewSliceSerializer(scalar.Int32).Unmarshal([]byte{0x00})
	if err != nil || len(empty) != 0 || n != 1 {
		t.Fatalf("Unmarshal of an empty slice = (%v, %d, %v)", empty, n, err)
	}
}

// TestUnmarshalValidSliceErrorPaths covers every early return of the validating
// slice decoder.
func TestUnmarshalValidSliceErrorPaths(t *testing.T) {
	payload := binco.Encode(varnum.PositiveInt, 1)
	// One element needs one byte after the length.
	withElement := append(binco.Encode(varnum.PositiveInt, 1), 0x02)
	reject := binco.ValidatorFn[int](func(int) error { return binco.ErrWrongFormat })

	if _, _, err := composite.UnmarshalValidSlice[int](nil, varnum.Int, varnum.PositiveInt, nil, nil); !errors.Is(err, binco.ErrTruncated) {
		t.Fatalf("truncated length = %v, want ErrTruncated", err)
	}
	if _, _, err := composite.UnmarshalValidSlice[int](binco.Encode(varnum.Int, -1), varnum.Int, varnum.Int, nil, nil); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("negative length = %v, want ErrNegativeLength", err)
	}
	if _, _, err := composite.UnmarshalValidSlice[int](payload, varnum.Int, varnum.PositiveInt, reject, nil); !errors.Is(err, binco.ErrWrongFormat) {
		t.Fatalf("length validator = %v, want ErrWrongFormat", err)
	}
	if _, _, err := composite.UnmarshalValidSlice[int](withElement, errorSerializer{}, varnum.PositiveInt, nil, nil); !errors.Is(err, binco.ErrWrongFormat) {
		t.Fatalf("element decoder = %v, want ErrWrongFormat", err)
	}
	if _, _, err := composite.UnmarshalValidSlice[int](withElement, varnum.Int, varnum.PositiveInt, nil, reject); !errors.Is(err, binco.ErrWrongFormat) {
		t.Fatalf("element validator = %v, want ErrWrongFormat", err)
	}

	// A clean decode reaches the end of the loop.
	valid, _, err := composite.UnmarshalValidSlice[int](withElement, varnum.Int, varnum.PositiveInt, nil, nil)
	if err != nil || len(valid) != 1 {
		t.Fatalf("valid decode = (%v, %v)", valid, err)
	}

	zeroPayload := binco.Encode(varnum.PositiveInt, 3)
	v, n, err := composite.UnmarshalValidSlice[struct{}](zeroPayload, zeroSerializer{}, varnum.PositiveInt, nil, nil)
	if err != nil || len(v) != 3 || n != len(zeroPayload) {
		t.Fatalf("zero-size elements = (%v, %d, %v)", v, n, err)
	}
	tooMany := binco.Encode(varnum.PositiveInt, binco.MaxZeroSizeElements+1)
	if _, _, err := composite.UnmarshalValidSlice[struct{}](tooMany, zeroSerializer{}, varnum.PositiveInt, nil, nil); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("zero-size elements above the cap = %v, want ErrTooLargeLength", err)
	}
}

// chainNode is a recursive type: its codec delegates the next pointer to the
// same pointer codec, so decoding recurses once per node.
type chainNode struct {
	next *chainNode
}

type chainCodec struct {
	ptr binco.Serializer[*chainNode]
}

func (c *chainCodec) Marshal(v chainNode, bs []byte) (n int) { return c.ptr.Marshal(v.next, bs) }
func (c *chainCodec) Unmarshal(bs []byte) (v chainNode, n int, err error) {
	v.next, n, err = c.ptr.Unmarshal(bs)
	return
}
func (c *chainCodec) Size(v chainNode) (size int)       { return c.ptr.Size(v.next) }
func (c *chainCodec) Skip(bs []byte) (n int, err error) { return c.ptr.Skip(bs) }

func newChainCodec() *chainCodec {
	c := &chainCodec{}
	c.ptr = composite.NewPtrSerializer[chainNode](c)
	return c
}

// deepChain returns a list with n nodes.
func deepChain(n int) *chainNode {
	root := &chainNode{}
	cur := root
	for range n - 1 {
		cur.next = &chainNode{}
		cur = cur.next
	}
	return root
}

// TestPointerDepthIsBoundedInEveryMethod checks the recursion limit of the plain
// pointer codec in all four methods, and that hitting it does not poison the
// serializer for later values.
func TestPointerDepthIsBoundedInEveryMethod(t *testing.T) {
	c := newChainCodec()
	depth := binco.MaxPointerDepth + 2
	deep := deepChain(depth)

	t.Run("Marshal", func(t *testing.T) {
		defer func() {
			err, ok := recover().(error)
			if !ok || !errors.Is(err, binco.ErrTooDeep) {
				t.Fatalf("panic value %v, want ErrTooDeep", err)
			}
		}()
		c.Marshal(*deep, make([]byte, depth))
	})
	t.Run("Size", func(t *testing.T) {
		defer func() {
			err, ok := recover().(error)
			if !ok || !errors.Is(err, binco.ErrTooDeep) {
				t.Fatalf("panic value %v, want ErrTooDeep", err)
			}
		}()
		c.Size(*deep)
	})
	t.Run("Unmarshal", func(t *testing.T) {
		if _, _, err := c.Unmarshal(make([]byte, depth)); !errors.Is(err, binco.ErrTooDeep) {
			t.Fatalf("Unmarshal = %v, want ErrTooDeep", err)
		}
	})
	t.Run("Skip", func(t *testing.T) {
		if _, err := c.Skip(make([]byte, depth)); !errors.Is(err, binco.ErrTooDeep) {
			t.Fatalf("Skip = %v, want ErrTooDeep", err)
		}
	})

	// The counter is back at zero, so a shallow value still works.
	shallow := []byte{byte(binco.PtrNil)}
	if _, n, err := c.Unmarshal(shallow); err != nil || n != 1 {
		t.Fatalf("Unmarshal after ErrTooDeep = (%d, %v)", n, err)
	}
	if n, err := c.Skip(shallow); err != nil || n != 1 {
		t.Fatalf("Skip after ErrTooDeep = (%d, %v)", n, err)
	}
	if got := c.Size(chainNode{}); err2safe(got) {
		t.Fatalf("Size after ErrTooDeep = %d, want 1", got)
	}
}

// err2safe reports whether a size is not the one-byte nil encoding.
func err2safe(size int) bool { return size != 1 }
