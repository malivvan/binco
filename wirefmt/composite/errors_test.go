package composite_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco/wirefmt"
	"github.com/malivvan/binco/wirefmt/composite"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
)

// flakyWriter fails after the given number of successful writes, which lets a
// test fail a marshal at every step of a value.
type flakyWriter struct {
	left int
	err  error
}

func (w *flakyWriter) Write(p []byte) (int, error) {
	if w.left == 0 {
		return 0, w.err
	}
	w.left--
	return len(p), nil
}

func (w *flakyWriter) WriteByte(c byte) error {
	if w.left == 0 {
		return w.err
	}
	w.left--
	return nil
}

func (w *flakyWriter) WriteString(s string) (int, error) {
	if w.left == 0 {
		return 0, w.err
	}
	w.left--
	return len(s), nil
}

// countingWriter records how many writes a marshal performs.
type countingWriter struct{ writes int }

func (w *countingWriter) Write(p []byte) (int, error) { w.writes++; return len(p), nil }
func (w *countingWriter) WriteByte(c byte) error      { w.writes++; return nil }
func (w *countingWriter) WriteString(s string) (int, error) {
	w.writes++
	return len(s), nil
}

func TestWriterErrorsAtEveryStep(t *testing.T) {
	boom := errors.New("write failed")
	ptr := 7

	codecs := []struct {
		name    string
		marshal func(w wirefmt.Writer) error
	}{
		{"bool", func(w wirefmt.Writer) error { _, err := composite.Bool.Marshal(true, w); return err }},
		{"string", func(w wirefmt.Writer) error { _, err := composite.String.Marshal("hello", w); return err }},
		{"string empty", func(w wirefmt.Writer) error { _, err := composite.String.Marshal("", w); return err }},
		{"byte slice", func(w wirefmt.Writer) error {
			_, err := composite.ByteSlice.Marshal([]byte{1, 2, 3}, w)
			return err
		}},
		{"slice", func(w wirefmt.Writer) error {
			_, err := composite.NewSliceSer(wvarnum.Int).Marshal([]int{1, 2, 3}, w)
			return err
		}},
		{"map", func(w wirefmt.Writer) error {
			_, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Marshal(
				map[int]int{1: 2, 3: 4}, w)
			return err
		}},
		{"ptr", func(w wirefmt.Writer) error {
			_, err := composite.NewPtrSer(wvarnum.Int).Marshal(&ptr, w)
			return err
		}},
		{"valid slice", func(w wirefmt.Writer) error {
			_, err := composite.NewValidSliceSer(wvarnum.Int).Marshal([]int{1, 2}, w)
			return err
		}},
		{"valid map", func(w wirefmt.Writer) error {
			_, err := composite.NewValidMapSer(wvarnum.Int, wvarnum.Int).Marshal(
				map[int]int{1: 2}, w)
			return err
		}},
	}

	for _, tc := range codecs {
		counter := &countingWriter{}
		if err := tc.marshal(counter); err != nil {
			t.Fatalf("%v: unexpected error %v", tc.name, err)
		}
		if counter.writes == 0 {
			t.Fatalf("%v: marshal performed no writes", tc.name)
		}
		for fails := 0; fails < counter.writes; fails++ {
			w := &flakyWriter{left: fails, err: boom}
			if err := tc.marshal(w); !errors.Is(err, boom) {
				t.Fatalf("%v with %d of %d writes allowed: got %v, want %v",
					tc.name, fails, counter.writes, err, boom)
			}
		}
	}
}

func TestUnmarshalTruncatedStreams(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"bool", func() error { _, _, err := composite.Bool.Unmarshal(reader()); return err }},
		{"string length", func() error { _, _, err := composite.String.Unmarshal(reader()); return err }},
		{"string data", func() error { _, _, err := composite.String.Unmarshal(reader(3, 'a')); return err }},
		{"byte slice length", func() error {
			_, _, err := composite.ByteSlice.Unmarshal(reader())
			return err
		}},
		{"byte slice data", func() error {
			_, _, err := composite.ByteSlice.Unmarshal(reader(3, 1))
			return err
		}},
		{"slice length", func() error {
			_, _, err := composite.NewSliceSer(wvarnum.Int).Unmarshal(reader())
			return err
		}},
		{"slice elements", func() error {
			_, _, err := composite.NewSliceSer(wvarnum.Int).Unmarshal(reader(2, 2))
			return err
		}},
		{"map length", func() error {
			_, _, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Unmarshal(reader())
			return err
		}},
		{"map key", func() error {
			_, _, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Unmarshal(reader(1))
			return err
		}},
		{"map value", func() error {
			_, _, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Unmarshal(reader(1, 2))
			return err
		}},
		{"ptr flag", func() error {
			_, _, err := composite.NewPtrSer(wvarnum.Int).Unmarshal(reader())
			return err
		}},
		{"ptr data", func() error {
			_, _, err := composite.NewPtrSer(wvarnum.Int).Unmarshal(reader(0))
			return err
		}},
		{"valid slice", func() error {
			_, _, err := composite.NewValidSliceSer(wvarnum.Int).Unmarshal(reader(2, 2))
			return err
		}},
		{"valid map", func() error {
			_, _, err := composite.NewValidMapSer(wvarnum.Int, wvarnum.Int).Unmarshal(reader(1, 2))
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err == nil {
				t.Fatal("Unmarshal of a truncated stream returned no error")
			}
		})
	}
}

func TestSkipTruncatedStreams(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"bool", func() error { _, err := composite.SkipBool(reader()); return err }},
		{"string length", func() error {
			_, err := composite.String.Skip(reader())
			return err
		}},
		{"string data", func() error {
			_, err := composite.String.Skip(reader(3, 'a'))
			return err
		}},
		{"byte slice data", func() error {
			_, err := composite.ByteSlice.Skip(reader(3, 1))
			return err
		}},
		{"slice length", func() error {
			_, err := composite.NewSliceSer(wvarnum.Int).Skip(reader())
			return err
		}},
		{"slice elements", func() error {
			_, err := composite.NewSliceSer(wvarnum.Int).Skip(reader(2, 2))
			return err
		}},
		{"map length", func() error {
			_, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Skip(reader())
			return err
		}},
		{"map entries", func() error {
			_, err := composite.NewMapSer(wvarnum.Int, wvarnum.Int).Skip(reader(1, 2))
			return err
		}},
		{"ptr flag", func() error {
			_, err := composite.NewPtrSer(wvarnum.Int).Skip(reader())
			return err
		}},
		{"ptr data", func() error {
			_, err := composite.NewPtrSer(wvarnum.Int).Skip(reader(0))
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err == nil {
				t.Fatal("Skip of a truncated stream returned no error")
			}
		})
	}
}
