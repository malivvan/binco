// Package bslopts provides options for the streaming byte slice serializer.
package bslopts_test

import (
	"testing"

	"github.com/malivvan/binco"
	bslopts "github.com/malivvan/binco/wirefmt/config/byte_slice"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
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
		bslopts.WithLenSer(wscalar.Int),
		bslopts.WithLenValidator(vl),
		nil,
	}, &o)

	if o.LenSer != wscalar.Int {
		t.Fatal("WithLenSer was not applied")
	}
	if o.LenVl == nil {
		t.Fatal("WithLenValidator was not applied")
	}
	if err := o.LenVl.Validate(5); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 5 {
		t.Fatalf("validator received %d, want 5", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := bslopts.Options{}
	bslopts.Apply(nil, &o)
	if o.LenSer != nil || o.LenVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
