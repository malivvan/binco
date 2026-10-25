package scalar_test

import (
	"errors"
	"testing"
	"time"

	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/scalar"
)

func TestWriterFailsAtEveryByte(t *testing.T) {
	boom := errors.New("write failed")
	ts := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)

	tests := []struct {
		name  string
		bytes int
		fn    func(w wirefmt.Writer) (int, error)
	}{
		{"uint64", 8, func(w wirefmt.Writer) (int, error) { return scalar.Uint64.Marshal(1<<40, w) }},
		{"int64", 8, func(w wirefmt.Writer) (int, error) { return scalar.Int64.Marshal(-1, w) }},
		{"float64", 8, func(w wirefmt.Writer) (int, error) { return scalar.Float64.Marshal(1.5, w) }},
		{"time", 8, func(w wirefmt.Writer) (int, error) { return scalar.TimeUnix.Marshal(ts, w) }},
		{"uint32", 4, func(w wirefmt.Writer) (int, error) { return scalar.Uint32.Marshal(1<<20, w) }},
		{"int32", 4, func(w wirefmt.Writer) (int, error) { return scalar.Int32.Marshal(-1, w) }},
		{"float32", 4, func(w wirefmt.Writer) (int, error) { return scalar.Float32.Marshal(1.5, w) }},
		{"uint16", 2, func(w wirefmt.Writer) (int, error) { return scalar.Uint16.Marshal(300, w) }},
		{"int16", 2, func(w wirefmt.Writer) (int, error) { return scalar.Int16.Marshal(-1, w) }},
		{"uint8", 1, func(w wirefmt.Writer) (int, error) { return scalar.Uint8.Marshal(1, w) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for remaining := 0; remaining < tc.bytes; remaining++ {
				w := &flakyWriter{remaining: remaining, err: boom}
				if _, err := tc.fn(w); !errors.Is(err, boom) {
					t.Fatalf("%d of %d bytes written: got %v, want %v",
						remaining, tc.bytes, err, boom)
				}
			}
		})
	}
}

func TestReaderFailsMidValue(t *testing.T) {
	one := func() wirefmt.Reader {
		return wirefmt.NewReader(&oneByteReader{})
	}
	tests := []struct {
		name string
		fn   func(r wirefmt.Reader) (int, error)
	}{
		{"uint16", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Uint16.Unmarshal(r)
			return n, err
		}},
		{"float32", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Float32.Unmarshal(r)
			return n, err
		}},
		{"float64", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.Float64.Unmarshal(r)
			return n, err
		}},
		{"time", func(r wirefmt.Reader) (int, error) {
			_, n, err := scalar.TimeUnix.Unmarshal(r)
			return n, err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.fn(one()); err == nil {
				t.Fatal("Unmarshal of a short stream returned no error")
			}
		})
	}
}

// oneByteReader returns a single zero byte and then EOF.
type oneByteReader struct{ read bool }

func (r *oneByteReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, errors.New("eof")
	}
	r.read = true
	p[0] = 0
	return 1, nil
}

func (r *oneByteReader) ReadByte() (byte, error) {
	if r.read {
		return 0, errors.New("eof")
	}
	r.read = true
	return 0, nil
}
