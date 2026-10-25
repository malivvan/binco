//go:build fuzz

package wirefmt_test

import (
	"bytes"
	"testing"

	"github.com/malivvan/binco/wirefmt"
)

// FuzzStreamDecode feeds arbitrary bytes to every streaming codec family.
//
// The properties are the ones that make an untrusted stream safe to decode: no
// panic, nothing consumed past the input, and Skip agreeing with Unmarshal.
func FuzzStreamDecode(f *testing.F) {
	for _, tc := range streamCases() {
		var buf [64]byte
		w := wirefmt.NewSliceWriter(buf[:])
		n, err := tc.encode(w)
		if err != nil {
			f.Fatalf("%s encode: %v", tc.name, err)
		}
		f.Add(append([]byte(nil), buf[:n]...))
	}
	f.Add([]byte{})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, tc := range streamCases() {
			n, err := tc.decode(bytes.NewReader(data))
			if n < 0 || n > len(data) {
				t.Fatalf("%s: Unmarshal consumed %d of %d bytes", tc.name, n, len(data))
			}
			sn, serr := tc.skip(bytes.NewReader(data))
			if sn < 0 || sn > len(data) {
				t.Fatalf("%s: Skip consumed %d of %d bytes", tc.name, sn, len(data))
			}
			if (err == nil) != (serr == nil) {
				t.Fatalf("%s: Unmarshal = %v, Skip = %v", tc.name, err, serr)
			}
			if err == nil && sn != n {
				t.Fatalf("%s: Unmarshal consumed %d, Skip %d", tc.name, n, sn)
			}
		}
	})
}
