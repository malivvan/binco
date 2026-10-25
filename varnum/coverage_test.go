package varnum_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/varnum"
)

// TestPositiveIntSizeRejectsNegative checks that Size refuses a value Marshal
// would refuse: a positive codec cannot encode a negative number.
func TestPositiveIntSizeRejectsNegative(t *testing.T) {
	codecs := []struct {
		name string
		call func()
	}{
		{"PositiveInt8", func() { varnum.PositiveInt8.Size(-1) }},
		{"PositiveInt16", func() { varnum.PositiveInt16.Size(-1) }},
		{"PositiveInt32", func() { varnum.PositiveInt32.Size(-1) }},
		{"PositiveInt64", func() { varnum.PositiveInt64.Size(-1) }},
		{"PositiveInt", func() { varnum.PositiveInt.Size(-1) }},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				err, ok := recover().(error)
				if !ok || !errors.Is(err, binco.ErrOverflow) {
					t.Fatalf("panic value %v, want ErrOverflow", err)
				}
			}()
			tc.call()
		})
	}
}

// TestPositiveIntUnmarshalRejectsTruncatedInput checks the error return of every
// positive codec.
func TestPositiveIntUnmarshalRejectsTruncatedInput(t *testing.T) {
	codecs := []struct {
		name string
		call func() error
	}{
		{"PositiveInt8", func() error { _, _, err := varnum.PositiveInt8.Unmarshal(nil); return err }},
		{"PositiveInt16", func() error { _, _, err := varnum.PositiveInt16.Unmarshal(nil); return err }},
		{"PositiveInt32", func() error { _, _, err := varnum.PositiveInt32.Unmarshal(nil); return err }},
		{"PositiveInt64", func() error { _, _, err := varnum.PositiveInt64.Unmarshal(nil); return err }},
		{"PositiveInt", func() error { _, _, err := varnum.PositiveInt.Unmarshal(nil); return err }},
		{"Uint", func() error { _, _, err := varnum.Uint.Unmarshal(nil); return err }},
		{"Int", func() error { _, _, err := varnum.Int.Unmarshal(nil); return err }},
	}
	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, binco.ErrTruncated) {
				t.Fatalf("Unmarshal(nil) = %v, want ErrTruncated", err)
			}
		})
	}
}

// TestSingleByteMarshalRejectsEmptyBuffer checks the one-byte fast path of the
// encoder, which checks the buffer itself instead of going through CheckWrite.
func TestSingleByteMarshalRejectsEmptyBuffer(t *testing.T) {
	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
			t.Fatalf("panic value %v, want ErrBufferTooSmall", err)
		}
	}()
	varnum.PositiveInt8.Marshal(1, nil)
}

// TestPositiveIntRejectsOutOfRangeInput checks the range check each positive
// codec applies before narrowing the decoded value.
func TestPositiveIntRejectsOutOfRangeInput(t *testing.T) {
	codecs := []struct {
		name    string
		payload []byte
		call    func([]byte) error
	}{
		{"PositiveInt16", []byte{0x80, 0x80, 0x02}, func(bs []byte) error { // 32768
			_, _, err := varnum.PositiveInt16.Unmarshal(bs)
			return err
		}},
		{"PositiveInt32", []byte{0x80, 0x80, 0x80, 0x80, 0x08}, func(bs []byte) error { // 1<<31
			_, _, err := varnum.PositiveInt32.Unmarshal(bs)
			return err
		}},
		{"PositiveInt64", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, func(bs []byte) error { // 1<<63
			_, _, err := varnum.PositiveInt64.Unmarshal(bs)
			return err
		}},
		{"PositiveInt", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, func(bs []byte) error {
			_, _, err := varnum.PositiveInt.Unmarshal(bs)
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
