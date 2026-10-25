// Package mapopts provides options for the streaming map serializer.
package mapopts_test

import (
	"testing"

	"github.com/malivvan/binco"
	mapopts "github.com/malivvan/binco/wirefmt/config/map"
	wscalar "github.com/malivvan/binco/wirefmt/scalar"
)

func TestApply(t *testing.T) {
	var (
		lenSeen, keySeen, valSeen int
		lenVl                     = binco.ValidatorFn[int](func(n int) error {
			lenSeen = n
			return nil
		})
		keyVl = binco.ValidatorFn[string](func(k string) error {
			keySeen = len(k)
			return nil
		})
		valVl = binco.ValidatorFn[int](func(v int) error {
			valSeen = v
			return nil
		})
	)
	o := mapopts.Options[string, int]{}
	mapopts.Apply([]mapopts.SetOption[string, int]{
		mapopts.WithLenSerializer[string, int](wscalar.Int),
		mapopts.WithLenValidator[string, int](lenVl),
		mapopts.WithKeyValidator[string, int](keyVl),
		mapopts.WithValueValidator[string, int](valVl),
		nil,
	}, &o)

	if o.LenSerializer != wscalar.Int {
		t.Fatal("WithLenSerializer was not applied")
	}
	if o.LenVl == nil || o.KeyVl == nil || o.ValueVl == nil {
		t.Fatalf("validators were not applied: %+v", o)
	}
	_ = o.LenVl.Validate(1)
	_ = o.KeyVl.Validate("ab")
	_ = o.ValueVl.Validate(4)
	if lenSeen != 1 || keySeen != 2 || valSeen != 4 {
		t.Fatalf("validators received (%d, %d, %d)", lenSeen, keySeen, valSeen)
	}
}

func TestApplyEmpty(t *testing.T) {
	o := mapopts.Options[int, int]{}
	mapopts.Apply(nil, &o)
	if o.LenSerializer != nil || o.LenVl != nil || o.KeyVl != nil || o.ValueVl != nil {
		t.Fatalf("empty Apply changed the options: %+v", o)
	}
}
