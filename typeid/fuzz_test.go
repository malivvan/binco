//go:build fuzz

package typeid_test

import (
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/typeid"
)

// FuzzDTM decodes arbitrary bytes as a type identifier. Identifiers are
// positive, and an overlong or zero value must be rejected rather than wrapped.
func FuzzDTM(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1}, {3}, {0x80}, {0x80, 0x01},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, bs []byte) {
		dtm, n, err := typeid.DTMSerializer.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err == nil && dtm <= 0 {
			t.Fatalf("Unmarshal returned the non-positive identifier %d", dtm)
		}
		if err != nil {
			return
		}
		if skipN, skipErr := typeid.DTMSerializer.Skip(bs); skipErr != nil || skipN != n {
			t.Fatalf("Skip = (%d, %v), Unmarshal consumed %d", skipN, skipErr, n)
		}
		size := typeid.DTMSerializer.Size(dtm)
		out := make([]byte, size)
		if written := typeid.DTMSerializer.Marshal(dtm, out); written != size {
			t.Fatalf("Marshal wrote %d, Size reported %d", written, size)
		}
		got, _, err := typeid.DTMSerializer.Unmarshal(out)
		if err != nil || got != dtm {
			t.Fatalf("round trip = (%d, %v), want %d", got, err, dtm)
		}
	})
}

// FuzzTyped decodes arbitrary bytes as a typed payload and checks that a
// foreign identifier is rejected instead of being decoded as data.
func FuzzTyped(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {1}, {7, 2}, {7}, {8, 2}, {7, 0x80},
	} {
		f.Add(seed)
	}
	serializer := typeid.NewSerializer[binco.DTM](7, typeid.DTMSerializer)
	f.Fuzz(func(t *testing.T, bs []byte) {
		_, n, err := serializer.Unmarshal(bs)
		if n < 0 || n > len(bs) {
			t.Fatalf("Unmarshal consumed %d of %d bytes", n, len(bs))
		}
		if err == nil {
			if skipN, skipErr := serializer.Skip(bs); skipErr != nil || skipN != n {
				t.Fatalf("Skip = (%d, %v), Unmarshal consumed %d", skipN, skipErr, n)
			}
		}
	})
}
