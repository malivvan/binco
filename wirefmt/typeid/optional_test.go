package typeid_test

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt/typeid"
)

// TestOptionalDTMSerializerNilMarker checks the reserved identifier an interface
// payload writes for a nil value: one zero byte, on the way in and on the way
// out.
func TestOptionalDTMSerializerNilMarker(t *testing.T) {
	bs := encode(t, typeid.OptionalDTMSerializer, binco.DTM(0))
	if len(bs) != 1 || bs[0] != 0 {
		t.Fatalf("Marshal(0) = %v, want [0]", bs)
	}

	dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(reader(bs...))
	if err != nil || n != 1 || dtm != 0 {
		t.Fatalf("Unmarshal([0]) = (%d, %d, %v)", dtm, n, err)
	}
	if n, err = typeid.OptionalDTMSerializer.Skip(reader(bs...)); err != nil || n != 1 {
		t.Fatalf("Skip([0]) = (%d, %v)", n, err)
	}
}

// TestOptionalDTMSerializerMatchesDTMSerializer checks that every non-zero
// identifier keeps the bytes and size of the plain DTM codec.
func TestOptionalDTMSerializerMatchesDTMSerializer(t *testing.T) {
	for _, v := range []binco.DTM{1, 7, 127, 128, 300, 1 << 20} {
		want := encode(t, typeid.DTMSerializer, v)
		got := encode(t, typeid.OptionalDTMSerializer, v)
		if !bytes.Equal(got, want) {
			t.Fatalf("identifier %d: %v, want %v", v, got, want)
		}
		dtm, n, err := typeid.OptionalDTMSerializer.Unmarshal(reader(got...))
		if err != nil || n != len(got) || dtm != v {
			t.Fatalf("identifier %d: Unmarshal = (%d, %d, %v)", v, dtm, n, err)
		}
		if n, err = typeid.OptionalDTMSerializer.Skip(reader(got...)); err != nil || n != len(want) {
			t.Fatalf("identifier %d: Skip = (%d, %v)", v, n, err)
		}
	}
}

// TestOptionalDTMSerializerRejectsBadInput checks the stream rules: an empty
// stream is io.EOF, a non-canonical identifier is rejected, and on a 32-bit host
// Skip agrees with Unmarshal about an identifier that does not fit binco.DTM.
func TestOptionalDTMSerializerRejectsBadInput(t *testing.T) {
	if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
	}
	if _, err := typeid.OptionalDTMSerializer.Skip(reader()); !errors.Is(err, io.EOF) {
		t.Fatalf("Skip of an empty stream = %v, want io.EOF", err)
	}
	if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(reader(0x80, 0x00)); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Unmarshal of a non-canonical identifier = %v, want ErrNonCanonical", err)
	}
	over := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(reader(over...)); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Unmarshal of MaxUint64 = %v, want ErrOverflow", err)
	}
	if _, err := typeid.OptionalDTMSerializer.Skip(reader(over...)); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Skip of MaxUint64 = %v, want ErrOverflow", err)
	}

	if strconv.IntSize == 32 {
		big := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}
		if _, _, err := typeid.OptionalDTMSerializer.Unmarshal(reader(big...)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("Unmarshal of MaxInt64 = %v, want ErrOverflow", err)
		}
		if _, err := typeid.OptionalDTMSerializer.Skip(reader(big...)); !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("Skip of MaxInt64 = %v, want ErrOverflow", err)
		}
	}
}

// TestOptionalDTMSerializerRejectsNegative checks the write side.
func TestOptionalDTMSerializerRejectsNegative(t *testing.T) {
	if _, err := typeid.OptionalDTMSerializer.Marshal(-1, &bytes.Buffer{}); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Marshal(-1) = %v, want ErrOverflow", err)
	}
	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, binco.ErrOverflow) {
			t.Fatalf("panic value %v, want ErrOverflow", err)
		}
	}()
	typeid.OptionalDTMSerializer.Size(-1)
}
