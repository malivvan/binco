// Package stropts provides options for the streaming string serializer.
package stropts_test

import (
	"testing"

	"github.com/malivvan/binco"
	stropts "github.com/malivvan/binco/wirefmt/config/string"
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
	o := stropts.Options{}
	stropts.Apply([]stropts.SetOption{
		stropts.WithLenSer(wscalar.Int),
		stropts.WithLenValidator(vl),
		nil,
	}, &o)

	if o.LenSer != wscalar.Int {
		t.Fatal("WithLenSer was not applied")
	}
	if o.LenVl == nil {
		t.Fatal("WithLenValidator was not applied")
	}
	if err := o.LenVl.Validate(9); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 9 {
		t.Fatalf("validator received %d, want 9", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := stropts.Options{}
	stropts.Apply(nil, &o)
	if o.LenSer != nil || o.LenVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
