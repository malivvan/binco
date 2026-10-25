package composite_test

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	bvarnum "github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
	"github.com/malivvan/binco/wirefmt/scalar"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

// zeroStream encodes a value that occupies no bytes at all.
type zeroStream struct{}

func (zeroStream) Marshal(v struct{}, w wirefmt.Writer) (n int, err error) { return 0, nil }
func (zeroStream) Unmarshal(r wirefmt.Reader) (struct{}, int, error)       { return struct{}{}, 0, nil }
func (zeroStream) Size(v struct{}) (size int)                              { return 0 }
func (zeroStream) Skip(r wirefmt.Reader) (n int, err error)                { return 0, nil }

// oneByteWriter accepts a single byte and then fails, which reaches the error
// return of the second write in a codec.
type oneByteWriter struct {
	buf bytes.Buffer
	err error
}

func (w *oneByteWriter) WriteByte(c byte) error {
	if w.buf.Len() >= 1 {
		return w.err
	}
	return w.buf.WriteByte(c)
}

func (w *oneByteWriter) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > 1 {
		return 0, w.err
	}
	return w.buf.Write(p)
}

func (w *oneByteWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }

// TestBlobDecoderErrors covers the length and body error returns of the
// streaming string and byte slice codecs.
func TestBlobDecoderErrors(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	checkedText := wcomposite.NewStringSerializer(stropts.WithLenValidator(never))
	checkedBlob := wcomposite.NewByteSliceSerializer(bslopts.WithLenValidator(never))

	// An empty stream fails at the length.
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"String", func() error { _, _, err := wcomposite.String.Unmarshal(reader()); return err }},
		{"ByteSlice", func() error { _, _, err := wcomposite.ByteSlice.Unmarshal(reader()); return err }},
		{"checked text", func() error { _, _, err := checkedText.Unmarshal(reader()); return err }},
		{"checked blob", func() error { _, _, err := checkedBlob.Unmarshal(reader()); return err }},
		{"ByteSlice.Skip", func() error { _, err := wcomposite.ByteSlice.Skip(reader()); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, io.EOF) {
				t.Fatalf("%s of an empty stream = %v, want io.EOF", tc.name, err)
			}
		})
	}

	// A body that never arrives is unexpected EOF.
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"String", func() error { _, _, err := wcomposite.String.Unmarshal(reader(5)); return err }},
		{"ByteSlice", func() error { _, _, err := wcomposite.ByteSlice.Unmarshal(reader(5)); return err }},
		{"checked text", func() error { _, _, err := checkedText.Unmarshal(reader(5)); return err }},
		{"checked blob", func() error { _, _, err := checkedBlob.Unmarshal(reader(5)); return err }},
		{"ByteSlice.Skip", func() error { _, err := wcomposite.ByteSlice.Skip(reader(5)); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("%s of a truncated body = %v, want io.ErrUnexpectedEOF", tc.name, err)
			}
		})
	}

	// A signed length codec lets a negative length through to the codec.
	negativeText := wcomposite.NewStringSerializer(stropts.WithLenSerializer(wvarnum.Int),
		stropts.WithLenValidator(never))
	if _, _, err := negativeText.Unmarshal(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("text with a negative length = %v, want ErrNegativeLength", err)
	}
	negativeBlob := wcomposite.NewByteSliceSerializer(bslopts.WithLenSerializer(wvarnum.Int),
		bslopts.WithLenValidator(never))
	if _, _, err := negativeBlob.Unmarshal(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("blob with a negative length = %v, want ErrNegativeLength", err)
	}
	if _, err := negativeBlob.Skip(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("blob Skip with a negative length = %v, want ErrNegativeLength", err)
	}
}

// streamChain is a recursive type: its codec delegates the next pointer to the
// same pointer codec, so decoding recurses once per node.
type streamChain struct {
	next *streamChain
}

type streamChainCodec struct {
	ptr wirefmt.Serializer[*streamChain]
}

func (c *streamChainCodec) Marshal(v streamChain, w wirefmt.Writer) (n int, err error) {
	return c.ptr.Marshal(v.next, w)
}
func (c *streamChainCodec) Unmarshal(r wirefmt.Reader) (v streamChain, n int, err error) {
	v.next, n, err = c.ptr.Unmarshal(r)
	return
}
func (c *streamChainCodec) Size(v streamChain) (size int)            { return c.ptr.Size(v.next) }
func (c *streamChainCodec) Skip(r wirefmt.Reader) (n int, err error) { return c.ptr.Skip(r) }

// TestPointerDepthIsBoundedInEveryStreamMethod checks the recursion limit in all
// four methods, and that hitting it does not poison the serializer.
func TestPointerDepthIsBoundedInEveryStreamMethod(t *testing.T) {
	c := &streamChainCodec{}
	c.ptr = wcomposite.NewPtrSerializer[streamChain](c)
	depth := binco.MaxPointerDepth + 2

	root := &streamChain{}
	cur := root
	for range depth - 1 {
		cur.next = &streamChain{}
		cur = cur.next
	}

	var buf bytes.Buffer
	if _, err := c.Marshal(*root, &buf); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Marshal = %v, want ErrTooDeep", err)
	}
	func() {
		defer func() {
			err, ok := recover().(error)
			if !ok || !errors.Is(err, binco.ErrTooDeep) {
				t.Fatalf("panic value %v, want ErrTooDeep", err)
			}
		}()
		c.Size(*root)
	}()
	if _, _, err := c.Unmarshal(bytes.NewReader(make([]byte, depth))); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Unmarshal = %v, want ErrTooDeep", err)
	}
	if _, err := c.Skip(bytes.NewReader(make([]byte, depth))); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Skip = %v, want ErrTooDeep", err)
	}

	// The counter is back at zero, so a shallow value still works.
	shallow := []byte{byte(binco.PtrNil)}
	if _, _, err := c.Unmarshal(bytes.NewReader(shallow)); err != nil {
		t.Fatalf("Unmarshal after ErrTooDeep = %v", err)
	}
	if _, err := c.Skip(bytes.NewReader(shallow)); err != nil {
		t.Fatalf("Skip after ErrTooDeep = %v", err)
	}
	if got := c.Size(streamChain{}); got != 1 {
		t.Fatalf("Size after ErrTooDeep = %d, want 1", got)
	}
}

// TestPointerCodecPropagatesBaseErrors checks that a failing element codec
// surfaces its own error in every method.
func TestPointerCodecPropagatesBaseErrors(t *testing.T) {
	boom := errors.New("boom")
	value := 1

	// failingCodec fails to write the pointed-to value.
	marshalling := wcomposite.NewPtrSerializer(failingCodec{err: boom})
	if _, err := marshalling.Marshal(&value, &bytes.Buffer{}); !errors.Is(err, boom) {
		t.Fatalf("Marshal = %v, want %v", err, boom)
	}

	// failingElemSerializer fails to read and to skip the pointed-to value.
	reading := wcomposite.NewPtrSerializer(failingElemSerializer{err: boom})
	if _, _, err := reading.Unmarshal(reader(byte(binco.PtrNotNil))); !errors.Is(err, boom) {
		t.Fatalf("Unmarshal = %v, want %v", err, boom)
	}
	if _, err := reading.Skip(reader(byte(binco.PtrNotNil))); !errors.Is(err, boom) {
		t.Fatalf("Skip = %v, want %v", err, boom)
	}

	// A nil pointer still needs its flag byte, which can fail.
	if _, err := reading.Marshal(nil, failingWriter{err: boom}); !errors.Is(err, boom) {
		t.Fatalf("Marshal(nil) = %v, want %v", err, boom)
	}
}

// TestSortedMapStream covers the streaming sorted map in every method.
func TestSortedMapStream(t *testing.T) {
	value := map[int]string{1: "a", 2: "b"}
	sorted := wcomposite.NewSortedMapSerializer(wvarnum.Int, wcomposite.String)

	var buf bytes.Buffer
	n, err := sorted.Marshal(value, &buf)
	if err != nil || n != buf.Len() {
		t.Fatalf("Marshal = (%d, %v), buf has %d bytes", n, err, buf.Len())
	}
	if got := sorted.Size(value); got != buf.Len() {
		t.Fatalf("Size = %d, Marshal wrote %d", got, buf.Len())
	}
	got, n, err := sorted.Unmarshal(bytes.NewReader(buf.Bytes()))
	if err != nil || n != buf.Len() || len(got) != len(value) {
		t.Fatalf("Unmarshal = (%v, %d, %v)", got, n, err)
	}
	if n, err := sorted.Skip(bytes.NewReader(buf.Bytes())); err != nil || n != buf.Len() {
		t.Fatalf("Skip = (%d, %v)", n, err)
	}
}

// TestMapStopsWhenNothingIsConsumed checks the zero-progress shortcut in the map
// decoder and skipper.
func TestMapStopsWhenNothingIsConsumed(t *testing.T) {
	m := wcomposite.NewMapSerializer[struct{}, struct{}](zeroStream{}, zeroStream{})
	v, n, err := m.Unmarshal(reader(3))
	if err != nil || len(v) != 1 || n != 1 {
		t.Fatalf("Unmarshal = (%v, %d, %v), want one pair", v, n, err)
	}
	if n, err := m.Skip(reader(3)); err != nil || n != 1 {
		t.Fatalf("Skip = (%d, %v), want 1", n, err)
	}
}

// TestMapLengthAndValueErrors checks the validating map's length handling and
// the value-skip error path.
func TestMapLengthAndValueErrors(t *testing.T) {
	never := binco.ValidatorFn[int](func(int) error { return nil })
	checked := wcomposite.NewMapSerializer(wvarnum.Int, wvarnum.Int,
		mapopts.WithLenValidator[int, int](never))
	if _, _, err := checked.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
	}

	negative := wcomposite.NewMapSerializer(wvarnum.Int, wvarnum.Int,
		mapopts.WithLenSerializer[int, int](wvarnum.Int),
		mapopts.WithLenValidator[int, int](never))
	if _, _, err := negative.Unmarshal(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}

	// A map whose value serializer fails to skip reports the error.
	boom := errors.New("boom")
	badValue := wcomposite.NewMapSerializer(wvarnum.Int, failingElemSerializer{err: boom})
	if _, err := badValue.Skip(reader(1, 0, 0)); !errors.Is(err, boom) {
		t.Fatalf("Skip with a failing value = %v, want %v", err, boom)
	}
}

// TestSliceStreamBranches covers the streaming slice codec's fast paths and
// every error return.
func TestSliceStreamBranches(t *testing.T) {
	boom := errors.New("boom")
	fixed := wcomposite.NewSliceSerializer(scalar.Int32)

	var buf bytes.Buffer
	if n, err := fixed.Marshal([]int32{1, 2}, &buf); err != nil || n != buf.Len() {
		t.Fatalf("fixed Marshal = (%d, %v)", n, err)
	}
	if got, want := fixed.Size([]int32{1, 2}), buf.Len(); got != want {
		t.Fatalf("Size = %d, want %d", got, want)
	}
	if v, n, err := fixed.Unmarshal(bytes.NewReader(buf.Bytes())); err != nil || len(v) != 2 || n != buf.Len() {
		t.Fatalf("fixed Unmarshal = (%v, %d, %v)", v, n, err)
	}
	if n, err := fixed.Skip(bytes.NewReader(buf.Bytes())); err != nil || n != buf.Len() {
		t.Fatalf("fixed Skip = (%d, %v)", n, err)
	}

	// An empty fixed-width slice writes only its length.
	var empty bytes.Buffer
	if _, err := wcomposite.MarshalSlice([]int32{}, scalar.Int32, wvarnum.PositiveInt, &empty); err != nil {
		t.Fatalf("MarshalSlice of an empty slice: %v", err)
	}
	if v, _, err := fixed.Unmarshal(reader(0)); err != nil || len(v) != 0 {
		t.Fatalf("Unmarshal of an empty slice = (%v, %v)", v, err)
	}
	if n, err := fixed.Skip(reader(0)); err != nil || n != 1 {
		t.Fatalf("Skip of an empty slice = (%d, %v)", n, err)
	}

	// Truncated fixed-width bodies are unexpected EOF.
	if _, _, err := fixed.Unmarshal(reader(2, 1)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated Unmarshal = %v, want io.ErrUnexpectedEOF", err)
	}
	if _, err := fixed.Skip(reader(2, 1)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated Skip = %v, want io.ErrUnexpectedEOF", err)
	}

	// A length that cannot be multiplied by the element width is refused. The
	// threshold is MaxInt/8, which differs per architecture.
	hugeLen := int64(1) << 61
	if strconv.IntSize == 32 {
		hugeLen = 1 << 29
	}
	huge := binco.Encode(bvarnum.PositiveInt, int(hugeLen))
	if _, _, err := wcomposite.NewSliceSerializer(scalar.Int64).Unmarshal(bytes.NewReader(huge)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Unmarshal of a huge fixed length = %v, want ErrTooLargeLength", err)
	}
	if _, err := wcomposite.NewSliceSerializer(scalar.Int64).Skip(bytes.NewReader(huge)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("Skip of a huge fixed length = %v, want ErrTooLargeLength", err)
	}

	// Zero-size elements fill without looping, are capped, and exercise the
	// allocation hint.
	zero := wcomposite.NewSliceSerializer(zeroStream{})
	if v, _, err := zero.Unmarshal(reader(3)); err != nil || len(v) != 3 {
		t.Fatalf("zero-size Unmarshal = (%v, %v)", v, err)
	}
	if n, err := zero.Skip(reader(3)); err != nil || n != 1 {
		t.Fatalf("zero-size Skip = (%d, %v)", n, err)
	}
	capped := binco.Encode(bvarnum.PositiveInt, binco.MaxZeroSizeElements+1)
	if _, _, err := zero.Unmarshal(bytes.NewReader(capped)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("zero-size Unmarshal above the cap = %v, want ErrTooLargeLength", err)
	}
	if _, err := zero.Skip(bytes.NewReader(capped)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("zero-size Skip above the cap = %v, want ErrTooLargeLength", err)
	}
	if v, _, err := zero.Unmarshal(bytes.NewReader(binco.Encode(bvarnum.PositiveInt, 1025))); err != nil || len(v) != 1025 {
		t.Fatalf("zero-size Unmarshal above the hint = (%v, %v)", len(v), err)
	}

	// A signed length codec reaches the negative-length checks.
	negative := wcomposite.NewSliceSerializer(wvarnum.Int, slopts.WithLenSerializer[int](wvarnum.Int))
	if _, _, err := negative.Unmarshal(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Unmarshal of a negative length = %v, want ErrNegativeLength", err)
	}
	if _, err := negative.Skip(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Skip of a negative length = %v, want ErrNegativeLength", err)
	}
	if _, _, err := negative.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
	}
	if _, err := negative.Skip(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream = %v, want io.EOF", err)
	}

	// Element errors surface from Unmarshal and Skip.
	badElem := wcomposite.NewSliceSerializer(failingElemSerializer{err: boom})
	if _, _, err := badElem.Unmarshal(reader(1)); !errors.Is(err, boom) {
		t.Fatalf("Unmarshal with a failing element = %v, want %v", err, boom)
	}
	if _, err := badElem.Skip(reader(1)); !errors.Is(err, boom) {
		t.Fatalf("Skip with a failing element = %v, want %v", err, boom)
	}

	// A writer that fails after the length byte reaches the fixed-width write.
	if _, err := wcomposite.MarshalSlice([]int32{1}, scalar.Int32, wvarnum.PositiveInt,
		&oneByteWriter{err: boom}); !errors.Is(err, boom) {
		t.Fatalf("MarshalSlice with a failing writer = %v, want %v", err, boom)
	}
}

// TestUnmarshalValidSliceStream covers every early return of the validating
// slice decoder.
func TestUnmarshalValidSliceStream(t *testing.T) {
	boom := errors.New("boom")
	reject := binco.ValidatorFn[int](func(int) error { return binco.ErrWrongFormat })

	// A clean decode.
	v, _, err := wcomposite.UnmarshalValidSlice[int](wvarnum.Int, wvarnum.PositiveInt, nil, nil, reader(1, 2))
	if err != nil || len(v) != 1 || v[0] != 1 {
		t.Fatalf("valid decode = (%v, %v)", v, err)
	}
	// A fixed-width element takes the bulk path when no element validator is set.
	if v, _, err := wcomposite.UnmarshalValidSlice[int32](scalar.Int32, wvarnum.PositiveInt, nil, nil,
		bytes.NewReader([]byte{1, 4, 0, 0, 0})); err != nil || len(v) != 1 || v[0] != 4 {
		t.Fatalf("fixed decode = (%v, %v)", v, err)
	}
	if _, _, err := wcomposite.UnmarshalValidSlice[int](wvarnum.Int, wvarnum.PositiveInt, reject, nil, reader(1, 2)); !errors.Is(err, binco.ErrWrongFormat) {
		t.Fatalf("length validator = %v, want ErrWrongFormat", err)
	}
	if _, _, err := wcomposite.UnmarshalValidSlice[int](wvarnum.Int, wvarnum.PositiveInt, nil, reject, reader(1, 2)); !errors.Is(err, binco.ErrWrongFormat) {
		t.Fatalf("element validator = %v, want ErrWrongFormat", err)
	}
	if _, _, err := wcomposite.UnmarshalValidSlice[int](failingElemSerializer{err: boom}, wvarnum.PositiveInt, nil, nil, reader(1)); !errors.Is(err, boom) {
		t.Fatalf("element decoder = %v, want %v", err, boom)
	}
	if _, _, err := wcomposite.UnmarshalValidSlice[int](wvarnum.Int, wvarnum.Int, nil, nil, reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("negative length = %v, want ErrNegativeLength", err)
	}
	if _, _, err := wcomposite.UnmarshalValidSlice[int](wvarnum.Int, wvarnum.PositiveInt, nil, nil, reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("empty stream = %v, want io.EOF", err)
	}

	zeroBytes := binco.Encode(bvarnum.PositiveInt, 3)
	if v, _, err := wcomposite.UnmarshalValidSlice[struct{}](zeroStream{}, wvarnum.PositiveInt, nil, nil,
		bytes.NewReader(zeroBytes)); err != nil || len(v) != 3 {
		t.Fatalf("zero-size elements = (%v, %v)", v, err)
	}
	capped := binco.Encode(bvarnum.PositiveInt, binco.MaxZeroSizeElements+1)
	if _, _, err := wcomposite.UnmarshalValidSlice[struct{}](zeroStream{}, wvarnum.PositiveInt, nil, nil,
		bytes.NewReader(capped)); !errors.Is(err, binco.ErrTooLargeLength) {
		t.Fatalf("zero-size elements above the cap = %v, want ErrTooLargeLength", err)
	}
}

// failAfterWriter accepts limit bytes and then returns err.
type failAfterWriter struct {
	limit int
	n     int
	err   error
}

func (w *failAfterWriter) WriteByte(c byte) error {
	if w.n >= w.limit {
		return w.err
	}
	w.n++
	return nil
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.n+len(p) > w.limit {
		return 0, w.err
	}
	w.n += len(p)
	return len(p), nil
}

func (w *failAfterWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }

// failingLenStream fails to write a length.
type failingLenStream struct{ err error }

func (s failingLenStream) Marshal(v int, w wirefmt.Writer) (n int, err error) { return 0, s.err }
func (s failingLenStream) Unmarshal(r wirefmt.Reader) (int, int, error)       { return 0, 0, nil }
func (s failingLenStream) Size(v int) (size int)                              { return 1 }
func (s failingLenStream) Skip(r wirefmt.Reader) (n int, err error)           { return 1, nil }

// TestSortedMapWriteErrors covers every write and encode failure of the sorted
// map encoder.
func TestSortedMapWriteErrors(t *testing.T) {
	boom := errors.New("boom")
	value := map[int]string{1: "a"}

	// A failing key codec is reported instead of writing a shifted payload.
	badKey := wcomposite.NewSortedMapSerializer(failingCodec{err: boom}, wcomposite.String)
	if _, err := badKey.Marshal(value, &bytes.Buffer{}); !errors.Is(err, boom) {
		t.Fatalf("failing key codec = %v, want %v", err, boom)
	}

	// A failing length codec is reported by the first write.
	badLen := wcomposite.NewSortedMapSerializer(wvarnum.Int, wcomposite.String,
		mapopts.WithLenSerializer[int, string](failingLenStream{err: boom}))
	if _, err := badLen.Marshal(value, &bytes.Buffer{}); !errors.Is(err, boom) {
		t.Fatalf("failing length codec = %v, want %v", err, boom)
	}

	// The length byte is followed by the key, then the value.
	sorted := wcomposite.NewSortedMapSerializer(wvarnum.Int, wcomposite.String)
	if _, err := sorted.Marshal(value, &failAfterWriter{limit: 1, err: boom}); !errors.Is(err, boom) {
		t.Fatalf("failing key write = %v, want %v", err, boom)
	}
	if _, err := sorted.Marshal(value, &failAfterWriter{limit: 2, err: boom}); !errors.Is(err, boom) {
		t.Fatalf("failing value write = %v, want %v", err, boom)
	}
}

// TestMapSkipNegativeLength checks the skipper's negative-length rule.
func TestMapSkipNegativeLength(t *testing.T) {
	m := wcomposite.NewMapSerializer(wvarnum.Int, wvarnum.Int,
		mapopts.WithLenSerializer[int, int](wvarnum.Int))
	if _, err := m.Skip(reader(1)); !errors.Is(err, binco.ErrNegativeLength) {
		t.Fatalf("Skip of a negative length = %v, want ErrNegativeLength", err)
	}
}
