// Package arropts provides options for the array serializer.
package arropts_test

import (
	"testing"

	"github.com/malivvan/binco"
	arropts "github.com/malivvan/binco/config/array"
	"github.com/malivvan/binco/varnum"
)

func TestApply(t *testing.T) {
	var (
		validated int
		vl        = binco.ValidatorFn[int](func(n int) error {
			validated = n
			return nil
		})
	)
	o := arropts.Options[int]{}
	arropts.Apply([]arropts.SetOption[int]{
		arropts.WithLenSerializer[int](varnum.PositiveInt),
		arropts.WithElemValidator[int](vl),
		nil,
	}, &o)

	if o.LenSerializer != varnum.PositiveInt {
		t.Fatal("WithLenSerializer was not applied")
	}
	if o.ElemVl == nil {
		t.Fatal("WithElemValidator was not applied")
	}
	if err := o.ElemVl.Validate(5); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 5 {
		t.Fatalf("validator received %d, want 5", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := arropts.Options[int]{}
	arropts.Apply(nil, &o)
	if o.LenSerializer != nil || o.ElemVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
