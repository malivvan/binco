// Package slopts provides options for the streaming slice serializer.
package slopts_test

import (
	"testing"

	"github.com/malivvan/binco"
	slopts "github.com/malivvan/binco/wirefmt/config/slice"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
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
		slopts.WithLenSerializer[int](wvarnum.PositiveInt),
		slopts.WithLenValidator[int](lenVl),
		slopts.WithElemValidator[int](elemVl),
		nil,
	}, &o)

	if o.LenSerializer != wvarnum.PositiveInt {
		t.Fatal("WithLenSerializer was not applied")
	}
	if o.LenVl == nil || o.ElemVl == nil {
		t.Fatalf("validators were not applied: %+v", o)
	}
	_ = o.LenVl.Validate(3)
	_ = o.ElemVl.Validate(8)
	if lenSeen != 3 || elemSeen != 8 {
		t.Fatalf("validators received (%d, %d)", lenSeen, elemSeen)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := slopts.Options[int]{}
	slopts.Apply(nil, &o)
	if o.LenSerializer != nil || o.LenVl != nil || o.ElemVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
