// Package bslopts provides options for the byte slice serializer.
package bslopts_test

import (
	"testing"

	"github.com/malivvan/binco"
	bslopts "github.com/malivvan/binco/config/byte_slice"
	"github.com/malivvan/binco/scalar"
)

func TestApply(t *testing.T) {
	var (
		validated int
		vl        = binco.ValidatorFn[int](func(n int) error {
			validated = n
			return nil
		})
	)
	o := bslopts.Options{}
	bslopts.Apply([]bslopts.SetOption{
		bslopts.WithLenSer(scalar.Int),
		bslopts.WithLenValidator(vl),
		nil,
	}, &o)

	if o.LenSer != scalar.Int {
		t.Fatal("WithLenSer was not applied")
	}
	if o.LenVl == nil {
		t.Fatal("WithLenValidator was not applied")
	}
	if err := o.LenVl.Validate(3); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 3 {
		t.Fatalf("validator received %d, want 3", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := bslopts.Options{}
	bslopts.Apply(nil, &o)
	if o.LenSer != nil || o.LenVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
