package binco_test

import (
	"errors"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/composite"
	"github.com/malivvan/binco/scalar"
	"github.com/malivvan/binco/varnum"
)

// TestMarshalPanicsWithOneSentinel covers the D5 note that Marshal panicked
// sometimes with ErrBufferTooSmall and sometimes with a raw runtime index
// error, so a caller could not write one recover for both.
func TestMarshalPanicsWithOneSentinel(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"varnum.Int", func() { varnum.Int.Marshal(300, nil) }},
		{"scalar.Int64", func() { scalar.Int64.Marshal(1, nil) }},
		{"composite.Bool", func() { composite.Bool.Marshal(true, nil) }},
		{"composite.String", func() { composite.String.Marshal("abcd", make([]byte, 2)) }},
		{"composite.ByteSlice", func() { composite.ByteSlice.Marshal([]byte{1, 2}, make([]byte, 1)) }},
		{"composite.Slice", func() {
			composite.NewSliceSerializer(varnum.Int).Marshal([]int{1, 2, 3}, make([]byte, 2))
		}},
		{"composite.Map", func() {
			composite.NewMapSerializer(varnum.Int, varnum.Int).Marshal(map[int]int{1: 1}, nil)
		}},
		{"composite.Ptr", func() {
			v := 1
			composite.NewPtrSerializer(varnum.Int).Marshal(&v, nil)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("Marshal did not panic")
				}
				err, ok := r.(error)
				if !ok || !errors.Is(err, binco.ErrBufferTooSmall) {
					t.Fatalf("panic value %v, want ErrBufferTooSmall", r)
				}
			}()
			tc.fn()
		})
	}
}
