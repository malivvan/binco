package ptrmap_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	"github.com/malivvan/binco/wirefmt/ptrmap"
	"github.com/malivvan/binco/wirefmt/varnum"
)

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

// TestSkipNilAndWriteErrors checks the nil skip and an identifier write that the
// writer refuses.
func TestSkipNilAndWriteErrors(t *testing.T) {
	pm, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	ints := ptrmap.NewPtrSerializer(pm, revPtrMap, varnum.Int)
	boom := errors.New("boom")

	// Skipping a nil pointer consumes just the flag byte.
	if n, err := ints.Skip(bytes.NewReader([]byte{byte(binco.PtrNil)})); err != nil || n != 1 {
		t.Fatalf("Skip(nil) = (%d, %v), want (1, nil)", n, err)
	}

	// The flag byte is written first, then the identifier.
	value := 1
	if _, err := ints.Marshal(&value, &failAfterWriter{limit: 1, err: boom}); !errors.Is(err, boom) {
		t.Fatalf("Marshal with a failing writer = %v, want %v", err, boom)
	}
}

// TestTypeConfusionOnTheStream checks that a reference is only resolved with the
// type it was defined with.
func TestTypeConfusionOnTheStream(t *testing.T) {
	pm, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	ints := ptrmap.NewPtrSerializer(pm, revPtrMap, varnum.Int)
	strs := ptrmap.NewPtrSerializer(pm, revPtrMap, composite.String)

	value := 1
	var buf bytes.Buffer
	if _, err := ints.Marshal(&value, &buf); err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if _, _, err := ints.Unmarshal(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Identifier 0 was defined as *int; reading it as *string is a mismatch.
	if _, _, err := strs.Unmarshal(bytes.NewReader([]byte{byte(binco.PtrMapping), 0})); !errors.Is(err, binco.ErrWrongType) {
		t.Fatalf("type confusion = %v, want ErrWrongType", err)
	}
}

// deepLink is a recursive type: its codec delegates the next pointer to the same
// pointer codec, so decoding recurses once per node.
type deepLink struct {
	next *deepLink
}

type deepLinkCodec struct {
	ptr wirefmt.Serializer[*deepLink]
}

func (c *deepLinkCodec) Marshal(v deepLink, w wirefmt.Writer) (n int, err error) {
	return c.ptr.Marshal(v.next, w)
}
func (c *deepLinkCodec) Unmarshal(r wirefmt.Reader) (v deepLink, n int, err error) {
	v.next, n, err = c.ptr.Unmarshal(r)
	return
}
func (c *deepLinkCodec) Size(v deepLink) (size int)               { return c.ptr.Size(v.next) }
func (c *deepLinkCodec) Skip(r wirefmt.Reader) (n int, err error) { return c.ptr.Skip(r) }

// TestReferenceDepthIsBounded checks the shared depth counter on the read and
// skip paths.
func TestReferenceDepthIsBounded(t *testing.T) {
	pm, revPtrMap := binco.NewPtrMap(), binco.NewReversePtrMap()
	c := &deepLinkCodec{}
	c.ptr = ptrmap.NewPtrSerializer(pm, revPtrMap, c)

	// Each level is a mapping reference with a fresh identifier, so the base
	// serializer decodes (and recurses) at every level.
	var payload []byte
	for id := range binco.MaxDepth + 2 {
		payload = append(payload, byte(binco.PtrMapping))
		var idBytes bytes.Buffer
		if _, err := varnum.PositiveInt.Marshal(id, &idBytes); err != nil {
			t.Fatalf("Marshal id %d: %v", id, err)
		}
		payload = append(payload, idBytes.Bytes()...)
	}

	if _, _, err := c.Unmarshal(bytes.NewReader(payload)); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Unmarshal = %v, want ErrTooDeep", err)
	}
	revPtrMap.Clear()
	if _, err := c.Skip(bytes.NewReader(payload)); !errors.Is(err, binco.ErrTooDeep) {
		t.Fatalf("Skip = %v, want ErrTooDeep", err)
	}

	// The counter is not left behind: a shallow value still skips.
	revPtrMap.Clear()
	if _, err := c.Skip(bytes.NewReader([]byte{byte(binco.PtrNil)})); err != nil {
		t.Fatalf("Skip after ErrTooDeep = %v", err)
	}
}
