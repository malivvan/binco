package varnum_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt/varnum"
)

// TestPositiveIntRejectsNegativeValues checks that the positive codecs refuse to
// write a negative number instead of encoding it as a huge unsigned value.
func TestPositiveIntRejectsNegativeValues(t *testing.T) {
	calls := []struct {
		name string
		call func() (int, error)
	}{
		{"PositiveInt8", func() (int, error) { return varnum.PositiveInt8.Marshal(-1, &bytes.Buffer{}) }},
		{"PositiveInt16", func() (int, error) { return varnum.PositiveInt16.Marshal(-1, &bytes.Buffer{}) }},
		{"PositiveInt32", func() (int, error) { return varnum.PositiveInt32.Marshal(-1, &bytes.Buffer{}) }},
		{"PositiveInt64", func() (int, error) { return varnum.PositiveInt64.Marshal(-1, &bytes.Buffer{}) }},
		{"PositiveInt", func() (int, error) { return varnum.PositiveInt.Marshal(-1, &bytes.Buffer{}) }},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.call(); !errors.Is(err, binco.ErrOverflow) {
				t.Fatalf("Marshal(-1) = %v, want ErrOverflow", err)
			}
		})
	}
}

// TestPositiveIntUnmarshalRejectsAnEmptyStream checks the error return of every
// positive codec.
func TestPositiveIntUnmarshalRejectsAnEmptyStream(t *testing.T) {
	codecs := []struct {
		name string
		call func() error
	}{
		{"PositiveInt8", func() error { _, _, err := varnum.PositiveInt8.Unmarshal(bytes.NewReader(nil)); return err }},
		{"PositiveInt16", func() error { _, _, err := varnum.PositiveInt16.Unmarshal(bytes.NewReader(nil)); return err }},
		{"PositiveInt32", func() error { _, _, err := varnum.PositiveInt32.Unmarshal(bytes.NewReader(nil)); return err }},
		{"PositiveInt64", func() error { _, _, err := varnum.PositiveInt64.Unmarshal(bytes.NewReader(nil)); return err }},
		{"PositiveInt", func() error { _, _, err := varnum.PositiveInt.Unmarshal(bytes.NewReader(nil)); return err }},
		{"Uint", func() error { _, _, err := varnum.Uint.Unmarshal(bytes.NewReader(nil)); return err }},
		{"Int", func() error { _, _, err := varnum.Int.Unmarshal(bytes.NewReader(nil)); return err }},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, io.EOF) {
				t.Fatalf("Unmarshal of an empty stream = %v, want io.EOF", err)
			}
		})
	}
}

// TestPositiveIntRejectsOutOfRangeInput checks the range check each positive
// codec applies before narrowing the decoded value.
func TestPositiveIntRejectsOutOfRangeInput(t *testing.T) {
	codecs := []struct {
		name    string
		payload []byte
		call    func([]byte) error
	}{
		{"PositiveInt8", []byte{0xc8, 0x01}, func(bs []byte) error { // 200
			_, _, err := varnum.PositiveInt8.Unmarshal(bytes.NewReader(bs))
			return err
		}},
		{"PositiveInt16", []byte{0x80, 0x80, 0x02}, func(bs []byte) error { // 32768
			_, _, err := varnum.PositiveInt16.Unmarshal(bytes.NewReader(bs))
			return err
		}},
		{"PositiveInt32", []byte{0x80, 0x80, 0x80, 0x80, 0x08}, func(bs []byte) error { // 1<<31
			_, _, err := varnum.PositiveInt32.Unmarshal(bytes.NewReader(bs))
			return err
		}},
		{"PositiveInt64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, func(bs []byte) error { // 1<<63
			_, _, err := varnum.PositiveInt64.Unmarshal(bytes.NewReader(bs))
			return err
		}},
		{"PositiveInt", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, func(bs []byte) error {
			_, _, err := varnum.PositiveInt.Unmarshal(bytes.NewReader(bs))
			return err
		}},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(tc.payload); !errors.Is(err, binco.ErrOverflow) {
				t.Fatalf("Unmarshal(%v) = %v, want ErrOverflow", tc.payload, err)
			}
		})
	}
}

// TestNonCanonicalVarintsAreRejected checks the shortest-form rule on the read
// and the skip path.
func TestNonCanonicalVarintsAreRejected(t *testing.T) {
	// 0x80 0x00 is zero written in two bytes.
	payload := []byte{0x80, 0x00}
	if _, _, err := varnum.PositiveInt.Unmarshal(bytes.NewReader(payload)); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Unmarshal = %v, want ErrNonCanonical", err)
	}
	// Uint64.Skip uses the dedicated skip loop; the positive codecs delegate to
	// Unmarshal, so both paths are covered.
	if _, err := varnum.Uint64.Skip(bytes.NewReader(payload)); !errors.Is(err, binco.ErrNonCanonical) {
		t.Fatalf("Skip = %v, want ErrNonCanonical", err)
	}
}
