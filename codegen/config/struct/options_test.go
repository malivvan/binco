package stopts_test

import (
	"testing"

	fldopts "github.com/malivvan/binco/codegen/config/field"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func TestFieldOptions(t *testing.T) {
	o := fldopts.Options{}
	fldopts.Apply(&o, fldopts.WithIgnore(), nil)
	if !o.Ignore {
		t.Fatal("WithIgnore was not applied")
	}

	o = fldopts.Options{}
	fldopts.Apply(&o, fldopts.WithType(
		tpopts.WithNumEncoding(tpopts.NumEncodingRaw),
		tpopts.WithLenValidator("Validate"),
	))
	if o.Type.NumEncoding != tpopts.NumEncodingRaw || o.Type.LenValidator != "Validate" {
		t.Fatalf("unexpected options %+v", o.Type)
	}
}

func TestStructOptions(t *testing.T) {
	o := stopts.Options{}
	stopts.Apply([]stopts.SetOption{
		stopts.WithValidator("Validate"),
		stopts.WithField(),
		stopts.WithField(fldopts.WithIgnore()),
		stopts.WithUnderlyingTime(
			stopts.WithUnderlyingTimeUnit(tpopts.TimeUnitMilli)),
		nil,
	}, &o)
	if o.Validator != "Validate" {
		t.Fatalf("Validator = %q", o.Validator)
	}
	if len(o.Fields) != 2 {
		t.Fatalf("Fields = %d, want 2", len(o.Fields))
	}
	if o.Fields[0].Ignore || !o.Fields[1].Ignore {
		t.Fatalf("unexpected fields %+v", o.Fields)
	}
	if len(o.UnderlyingTime) != 1 {
		t.Fatalf("UnderlyingTime = %d", len(o.UnderlyingTime))
	}
	uops := stopts.UnderlyingTimeOptions{}
	stopts.UnderlyingTimeApply(o.UnderlyingTime, &uops)
	if uops.TimeUnit != tpopts.TimeUnitMilli {
		t.Fatalf("TimeUnit = %v", uops.TimeUnit)
	}

	timeOnly := stopts.Options{}
	stopts.Apply([]stopts.SetOption{stopts.WithUnderlyingTime()}, &timeOnly)
	if timeOnly.UnderlyingTime == nil {
		t.Fatal("WithUnderlyingTime() did not enable the underlying time mode")
	}
	if len(timeOnly.UnderlyingTime) != 0 {
		t.Fatalf("UnderlyingTime = %d, want 0", len(timeOnly.UnderlyingTime))
	}
}
