// Package stropts provides options for the string serializer.
package stropts_test

import (
	"testing"

	"github.com/malivvan/binco"
	stropts "github.com/malivvan/binco/config/string"
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
	o := stropts.Options{}
	stropts.Apply([]stropts.SetOption{
		stropts.WithLenSer(scalar.Int),
		stropts.WithLenValidator(vl),
		nil,
	}, &o)

	if o.LenSer != scalar.Int {
		t.Fatal("WithLenSer was not applied")
	}
	if o.LenVl == nil {
		t.Fatal("WithLenValidator was not applied")
	}
	if err := o.LenVl.Validate(6); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if validated != 6 {
		t.Fatalf("validator received %d, want 6", validated)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := stropts.Options{}
	stropts.Apply(nil, &o)
	if o.LenSer != nil || o.LenVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
