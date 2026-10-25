package binco_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/internal/fixedwidth"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	"github.com/malivvan/binco/zerocopy"
)

// TestHostEndiannessMatchesStdlib pins the flag the bulk fast path depends on to
// the standard library's build-tag-based answer, so the two cannot drift.
func TestHostEndiannessMatchesStdlib(t *testing.T) {
	want := binary.NativeEndian.Uint16([]byte{1, 0}) == 1
	if binco.HostLittleEndian != want {
		t.Fatalf("HostLittleEndian = %v, want %v", binco.HostLittleEndian, want)
	}
}

// TestWireBytesAreLittleEndianOnEveryHost checks that the encoders do not depend
// on the host byte order: these are the bytes on a little-endian and on a
// big-endian machine alike.
func TestWireBytesAreLittleEndianOnEveryHost(t *testing.T) {
	cases := []struct {
		name string
		got  []byte
		want []byte
	}{
		{"scalar.Uint16", binco.Encode(scalar.Uint16, 0x1234), []byte{0x34, 0x12}},
		{"scalar.Uint32", binco.Encode(scalar.Uint32, 0x01020304), []byte{0x04, 0x03, 0x02, 0x01}},
		{"scalar.Uint64", binco.Encode(scalar.Uint64, 0x0102030405060708), []byte{8, 7, 6, 5, 4, 3, 2, 1}},
		{"scalar.Float64", binco.Encode(scalar.Float64, 1.0), []byte{0, 0, 0, 0, 0, 0, 0xF0, 0x3F}},
		{"zerocopy.Uint16", binco.Encode(zerocopy.Uint16, 0x1234), []byte{0x34, 0x12}},
		{"zerocopy.Uint32", binco.Encode(zerocopy.Uint32, 0x01020304), []byte{0x04, 0x03, 0x02, 0x01}},
		{"zerocopy.Uint64", binco.Encode(zerocopy.Uint64, 0x0102030405060708), []byte{8, 7, 6, 5, 4, 3, 2, 1}},
		{"zerocopy.Float64", binco.Encode(zerocopy.Float64, 1.0), []byte{0, 0, 0, 0, 0, 0, 0xF0, 0x3F}},
	}
	for _, tc := range cases {
		if !bytes.Equal(tc.got, tc.want) {
			t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

// littleEndianInt32 returns the wire form of one int32, independent of the host.
func littleEndianInt32(v int32) []byte {
	u := uint32(v)
	return []byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
}

// TestSliceBulkPathMatchesTheElementLoop checks the two slice paths against each
// other: the bulk copy is only taken on a little-endian host, and it must
// produce exactly what the per-element loop would.
func TestSliceBulkPathMatchesTheElementLoop(t *testing.T) {
	values := []int32{0, 1, -2, 0x01020304}

	// Reference: length prefix plus each element through the element codec.
	want := []byte{byte(len(values))}
	for _, v := range values {
		want = append(want, littleEndianInt32(v)...)
	}

	if got := binco.Encode(composite.NewSliceSerializer[int32](scalar.Int32), values); !bytes.Equal(got, want) {
		t.Fatalf("buffer slice = %v, want %v", got, want)
	}

	var buf [64]byte
	w := wirefmt.NewSliceWriter(buf[:])
	n, err := wcomposite.NewSliceSerializer[int32](wscalar.Int32).Marshal(values, w)
	if err != nil {
		t.Fatalf("stream Marshal: %v", err)
	}
	if got := buf[:n]; !bytes.Equal(got, want) {
		t.Fatalf("stream slice = %v, want %v", got, want)
	}

	back, err := binco.Decode(composite.NewSliceSerializer[int32](scalar.Int32), want)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for i := range values {
		if back[i] != values[i] {
			t.Fatalf("Unmarshal[%d] = %d, want %d", i, back[i], values[i])
		}
	}
}

// TestEmptySliceUsesTheSameBytesOnBothPaths checks the zero-length shortcut.
func TestEmptySliceUsesTheSameBytesOnBothPaths(t *testing.T) {
	got := binco.Encode(composite.NewSliceSerializer[int32](scalar.Int32), []int32{})
	if !bytes.Equal(got, []byte{0}) {
		t.Fatalf("empty slice = %v, want [0]", got)
	}
	back, err := binco.Decode(composite.NewSliceSerializer[int32](scalar.Int32), got)
	if err != nil || back == nil || len(back) != 0 {
		t.Fatalf("empty slice round trip = (%v, %v), want an empty non-nil slice", back, err)
	}
}

// TestFixedWidthGuardFollowsTheHost checks the gate on the bulk path: it is open
// exactly on a little-endian host, for a raw little-endian codec whose width
// matches the element type.
func TestFixedWidthGuardFollowsTheHost(t *testing.T) {
	width, ok := fixedwidth.Of[int32](scalar.Int32)
	if ok != binco.HostLittleEndian {
		t.Fatalf("fixedwidth.Of(scalar.Int32) ok = %v, want %v", ok, binco.HostLittleEndian)
	}
	if ok && width != 4 {
		t.Fatalf("fixedwidth.Of(scalar.Int32) width = %d, want 4", width)
	}
	if _, ok := fixedwidth.Of[int32](varnum.Int32); ok {
		t.Fatal("fixedwidth.Of accepted a varint codec")
	}
	if _, ok := fixedwidth.Of[int64](scalar.Int32); ok {
		t.Fatal("fixedwidth.Of accepted a codec whose width differs from the element type")
	}
}

// TestArrayCodecIsHostIndependent checks that the array codec inherits the
// element codec's little-endian bytes, on any host.
func TestArrayCodecIsHostIndependent(t *testing.T) {
	values := [3]int32{1, -2, 3}
	codec := zerocopy.NewArraySerializer[[3]int32, int32](scalar.Int32)

	want := []byte{3}
	for _, v := range values {
		want = append(want, littleEndianInt32(v)...)
	}
	got := binco.Encode(codec, values)
	if !bytes.Equal(got, want) {
		t.Fatalf("array = %v, want %v", got, want)
	}
	back, err := binco.Decode(codec, got)
	if err != nil || back != values {
		t.Fatalf("array round trip = (%v, %v), want %v", back, err, values)
	}
}
