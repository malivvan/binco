// Package slopts provides options for the slice serializer.
package slopts_test

import (
	"testing"

	"github.com/malivvan/binco"
	slopts "github.com/malivvan/binco/config/slice"
	"github.com/malivvan/binco/varnum"
)

func TestApply(t *testing.T) {
	var (
		lenSeen, elemSeen int
		lenVl             = binco.ValidatorFn[int](func(n int) error {
			lenSeen = n
			return nil
		})
		elemVl = binco.ValidatorFn[int](func(v int) error {
			elemSeen = v
			return nil
		})
	)
	o := slopts.Options[int]{}
	slopts.Apply([]slopts.SetOption[int]{
		slopts.WithLenSer[int](varnum.PositiveInt),
		slopts.WithLenValidator[int](lenVl),
		slopts.WithElemValidator[int](elemVl),
		nil,
	}, &o)

	if o.LenSer != varnum.PositiveInt {
		t.Fatal("WithLenSer was not applied")
	}
	if o.LenVl == nil || o.ElemVl == nil {
		t.Fatalf("validators were not applied: %+v", o)
	}
	_ = o.LenVl.Validate(4)
	_ = o.ElemVl.Validate(9)
	if lenSeen != 4 || elemSeen != 9 {
		t.Fatalf("validators received (%d, %d)", lenSeen, elemSeen)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := slopts.Options[int]{}
	slopts.Apply(nil, &o)
	if o.LenSer != nil || o.LenVl != nil || o.ElemVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
