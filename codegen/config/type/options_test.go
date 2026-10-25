package tpopts_test

import (
	"testing"

	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func TestNumEncoding(t *testing.T) {
	tests := []struct {
		enc  tpopts.NumEncoding
		pkg  string
		lenS string
	}{
		{tpopts.NumEncodingUndefined, "varnum", "varnum.PositiveInt"},
		{tpopts.NumEncodingVarint, "varnum", "varnum.Int"},
		{tpopts.NumEncodingVarintPositive, "varnum", "varnum.PositiveInt"},
		{tpopts.NumEncodingRaw, "scalar", "scalar.Int"},
	}
	for _, tc := range tests {
		if got := tc.enc.Package(); got != tc.pkg {
			t.Fatalf("%v.Package() = %q, want %q", tc.enc, got, tc.pkg)
		}
		if got := tc.enc.LenSerializer(); got != tc.lenS {
			t.Fatalf("%v.LenSerializer() = %q, want %q", tc.enc, got, tc.lenS)
		}
	}
	checkPanics(t, func() { tpopts.NumEncoding(99).Package() })
	checkPanics(t, func() { tpopts.NumEncoding(99).LenSerializer() })
}

func TestTimeUnit(t *testing.T) {
	tests := []struct {
		unit       tpopts.TimeUnit
		serializer string
	}{
		{tpopts.TimeUnitSec, "TimeUnix"},
		{tpopts.TimeUnitMilli, "TimeUnixMilli"},
		{tpopts.TimeUnitMicro, "TimeUnixMicro"},
		{tpopts.TimeUnitNano, "TimeUnixNano"},
		{tpopts.TimeUnitSecUTC, "TimeUnixUTC"},
		{tpopts.TimeUnitMilliUTC, "TimeUnixMilliUTC"},
		{tpopts.TimeUnitMicroUTC, "TimeUnixMicroUTC"},
		{tpopts.TimeUnitNanoUTC, "TimeUnixNanoUTC"},
	}
	for _, tc := range tests {
		if got := tc.unit.Serializer(); got != tc.serializer {
			t.Fatalf("%v.Serializer() = %q, want %q", tc.unit, got, tc.serializer)
		}
	}
	checkPanics(t, func() { tpopts.TimeUnit(99).Serializer() })
}

func TestOptions(t *testing.T) {
	o := tpopts.Options{}
	tpopts.Apply(&o,
		tpopts.WithNumEncoding(tpopts.NumEncodingRaw),
		tpopts.WithTimeUnit(tpopts.TimeUnitNano),
		tpopts.WithValidator("Validate"),
		tpopts.WithLenEncoding(tpopts.NumEncodingVarint),
		tpopts.WithLenValidator("ValidateLen"),
		tpopts.WithElemValidator("ValidateElem"),
		tpopts.WithKeyValidator("ValidateKey"),
		nil,
	)
	if o.NumEncoding != tpopts.NumEncodingRaw || o.TimeUnit != tpopts.TimeUnitNano ||
		o.Validator != "Validate" || o.LenEnc != tpopts.NumEncodingVarint ||
		o.LenValidator != "ValidateLen" || o.ElemValidator != "ValidateElem" ||
		o.KeyValidator != "ValidateKey" {
		t.Fatalf("unexpected options %+v", o)
	}
}

func checkPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("no panic")
		}
	}()
	fn()
}
