// Package arropts provides options for the streaming array serializer.
package arropts_test

import (
	"testing"

	"github.com/malivvan/binco"
	arropts "github.com/malivvan/binco/wirefmt/config/array"
	wvarnum "github.com/malivvan/binco/wirefmt/varnum"
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
		arropts.WithLenSerializer[int](wvarnum.PositiveInt),
		arropts.WithElemValidator[int](vl),
		nil,
	}, &o)

	if o.LenSerializer != wvarnum.PositiveInt {
		t.Fatal("WithLenSerializer was not applied")
	}
	if o.ElemVl == nil {
		t.Fatal("WithElemValidator was not applied")
	}
	if err := o.ElemVl.Validate(2); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 2 {
		t.Fatalf("validator received %d, want 2", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := arropts.Options[int]{}
	arropts.Apply(nil, &o)
	if o.LenSerializer != nil || o.ElemVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
