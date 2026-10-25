package wirefmt_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wcomposite "github.com/malivvan/binco/wirefmt/composite"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
	wzerocopy "github.com/malivvan/binco/wirefmt/zerocopy"
)

// timeValue is the timestamp the time codecs round-trip.
var timeValue = time.Unix(1_700_000_000, 0)

// streamCase is one streaming codec exercised by the safety tests.
type streamCase struct {
	name string
	// lengthPrefixed marks codecs whose payload starts with a length.
	lengthPrefixed bool
	// fixedWidth marks codecs that read a known number of bytes without a
	// length, which is where a stalled reader is observable.
	fixedWidth bool
	encode     func(w wirefmt.Writer) (n int, err error)
	decode     func(r wirefmt.Reader) (n int, err error)
	skip       func(r wirefmt.Reader) (n int, err error)
}

// streamCases returns one representative value per streaming codec family.
func streamCases() []streamCase {
	sl := wcomposite.NewSliceSerializer[int32](wscalar.Int32)
	mp := wcomposite.NewMapSerializer[int, int](wvarnum.Int, wvarnum.Int)
	typed := wtypeid.NewSerializer(binco.DTM(7), wvarnum.Int32)
	return []streamCase{
		{
			name:           "composite.String",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return wcomposite.String.Marshal("hello world", w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wcomposite.String.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wcomposite.String.Skip(r) },
		},
		{
			name:           "composite.ByteSlice",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return wcomposite.ByteSlice.Marshal([]byte{1, 2, 3}, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wcomposite.ByteSlice.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wcomposite.ByteSlice.Skip(r) },
		},
		{
			name:           "composite slice of int",
			lengthPrefixed: true,
			encode: func(w wirefmt.Writer) (int, error) {
				return wcomposite.NewSliceSerializer[int](wvarnum.Int).Marshal([]int{1, 2, 300}, w)
			},
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wcomposite.NewSliceSerializer[int](wvarnum.Int).Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) {
				return wcomposite.NewSliceSerializer[int](wvarnum.Int).Skip(r)
			},
		},
		{
			name:           "composite slice of int32",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return sl.Marshal([]int32{1, -2}, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := sl.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return sl.Skip(r) },
		},
		{
			name:           "composite map",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return mp.Marshal(map[int]int{1: 2}, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := mp.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return mp.Skip(r) },
		},
		{
			name:           "zerocopy.String",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return wzerocopy.String.Marshal("hello world", w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wzerocopy.String.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wzerocopy.String.Skip(r) },
		},
		{
			name:           "typeid typed int32",
			lengthPrefixed: true,
			encode:         func(w wirefmt.Writer) (int, error) { return typed.Marshal(int32(1), w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := typed.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return typed.Skip(r) },
		},
		{
			name:       "scalar.Uint64",
			fixedWidth: true,
			encode:     func(w wirefmt.Writer) (int, error) { return wscalar.Uint64.Marshal(1<<40, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wscalar.Uint64.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wscalar.Uint64.Skip(r) },
		},
		{
			name:       "scalar.TimeUnix",
			fixedWidth: true,
			encode:     func(w wirefmt.Writer) (int, error) { return wscalar.TimeUnix.Marshal(timeValue, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wscalar.TimeUnix.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wscalar.TimeUnix.Skip(r) },
		},
		{
			name:       "zerocopy.Uint64",
			fixedWidth: true,
			encode:     func(w wirefmt.Writer) (int, error) { return wzerocopy.Uint64.Marshal(1<<40, w) },
			decode: func(r wirefmt.Reader) (int, error) {
				_, n, err := wzerocopy.Uint64.Unmarshal(r)
				return n, err
			},
			skip: func(r wirefmt.Reader) (int, error) { return wzerocopy.Uint64.Skip(r) },
		},
	}
}

// encodeCase returns the wire form of one case.
func encodeCase(t *testing.T, tc streamCase) []byte {
	t.Helper()
	var buf [64]byte
	w := wirefmt.NewSliceWriter(buf[:])
	n, err := tc.encode(w)
	if err != nil {
		t.Fatalf("%s encode: %v", tc.name, err)
	}
	return append([]byte(nil), buf[:n]...)
}

// stallReader returns (0, nil) forever: it never makes progress and never fails.
type stallReader struct{ reads int }

func (r *stallReader) Read([]byte) (int, error) {
	r.reads++
	return 0, nil
}

// errBoom is what failingReader returns after its data.
var errBoom = errors.New("boom")

// failingReader returns data, then errBoom instead of io.EOF.
type failingReader struct {
	data []byte
	pos  int
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, errBoom
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// TestHostileLengthsAreBounded feeds every length-prefixed codec a nine-byte
// varint of MaxInt64. Each must fail immediately, without allocating, because
// the stream holds no bytes the length could describe.
func TestHostileLengthsAreBounded(t *testing.T) {
	payload := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}
	for _, tc := range streamCases() {
		if !tc.lengthPrefixed {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			n, err := tc.decode(bytes.NewReader(payload))
			runtime.ReadMemStats(&after)

			if err == nil {
				t.Fatalf("%s decoded a declared length of MaxInt64", tc.name)
			}
			if n > len(payload) {
				t.Fatalf("%s consumed %d of %d bytes", tc.name, n, len(payload))
			}
			const maxAlloc = 8 << 20
			if allocated := after.TotalAlloc - before.TotalAlloc; allocated > maxAlloc {
				t.Fatalf("%s allocated %d bytes for a 9-byte payload", tc.name, allocated)
			}
		})
	}
}

// TestStalledReaderFails checks the readers that read a known number of bytes:
// a stream that keeps returning (0, nil) must report io.ErrNoProgress instead of
// blocking forever.
func TestStalledReaderFails(t *testing.T) {
	for _, tc := range streamCases() {
		if !tc.fixedWidth {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			r := &stallReader{}
			if _, err := tc.decode(wirefmt.NewReader(r)); !errors.Is(err, io.ErrNoProgress) {
				t.Fatalf("%s error = %v, want io.ErrNoProgress", tc.name, err)
			}
			if r.reads == 0 {
				t.Fatalf("%s never read from the stalled reader", tc.name)
			}
		})
	}
}

// TestReaderErrorIsPropagated checks that an I/O error in the middle of a value
// reaches the caller unchanged, for every codec family.
func TestReaderErrorIsPropagated(t *testing.T) {
	for _, tc := range streamCases() {
		t.Run(tc.name, func(t *testing.T) {
			full := encodeCase(t, tc)
			// Half the payload is never enough to finish the value.
			prefix := full[:len(full)/2]
			if _, err := tc.decode(wirefmt.NewReader(&failingReader{data: prefix})); !errors.Is(err, errBoom) {
				t.Fatalf("%s error = %v, want %v", tc.name, err, errBoom)
			}
		})
	}
}

// TestTruncatedPrefixesNeverPanic cuts every encoded value at every offset.
// Each prefix must fail cleanly, must not be read past its end, and Skip must
// report the same outcome as Unmarshal.
func TestTruncatedPrefixesNeverPanic(t *testing.T) {
	for _, tc := range streamCases() {
		t.Run(tc.name, func(t *testing.T) {
			full := encodeCase(t, tc)
			for cut := 0; cut < len(full); cut++ {
				n, err := tc.decode(bytes.NewReader(full[:cut]))
				sn, serr := tc.skip(bytes.NewReader(full[:cut]))

				if n < 0 || n > cut {
					t.Fatalf("cut %d: Unmarshal consumed %d bytes", cut, n)
				}
				if sn < 0 || sn > cut {
					t.Fatalf("cut %d: Skip consumed %d bytes", cut, sn)
				}
				if (err == nil) != (serr == nil) {
					t.Fatalf("cut %d: Unmarshal = %v, Skip = %v", cut, err, serr)
				}
				if err == nil && sn != n {
					t.Fatalf("cut %d: Unmarshal consumed %d, Skip %d", cut, n, sn)
				}
			}
		})
	}
}

// TestDecodeReadsNothingPastTheValue checks that a shared stream keeps the next
// message intact, both through the single-byte adapter and through bufio.
func TestDecodeReadsNothingPastTheValue(t *testing.T) {
	const sentinel = 0xAB
	for _, tc := range streamCases() {
		t.Run(tc.name, func(t *testing.T) {
			full := encodeCase(t, tc)

			// Through the one-byte adapter: the underlying reader must still
			// hold the sentinel.
			data := append(append([]byte(nil), full...), sentinel)
			raw := bytes.NewReader(data)
			if _, err := tc.decode(wirefmt.NewReader(raw)); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if raw.Len() != 1 {
				t.Fatalf("decode left %d bytes, want 1", raw.Len())
			}
			if b, err := raw.ReadByte(); err != nil || b != sentinel {
				t.Fatalf("next byte = (%#x, %v), want %#x", b, err, sentinel)
			}

			// Through bufio: the buffered reader must not have consumed it.
			br := bufio.NewReader(bytes.NewReader(data))
			if _, err := tc.decode(br); err != nil {
				t.Fatalf("decode via bufio: %v", err)
			}
			if b, err := br.ReadByte(); err != nil || b != sentinel {
				t.Fatalf("bufio next byte = (%#x, %v), want %#x", b, err, sentinel)
			}
		})
	}
}
