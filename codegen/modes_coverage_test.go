package codegen

import (
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

// TestNativeRulesHonoursPerTypeOptions checks the unsafe-mode table: a pinned
// number encoding, a pinned time unit, and the streaming byte-slice codec.
func TestNativeRulesHonoursPerTypeOptions(t *testing.T) {
	tops := tpopts.Options{
		NumEncoding: tpopts.NumEncodingRaw,
		TimeUnit:    tpopts.TimeUnitNano,
	}
	rules := nativeRules(tops, genopts.Options{Stream: true})
	if got := rules["int64"]; got != tpopts.NumEncodingRaw.Package() {
		t.Fatalf("int64 rule = %q, want %q", got, tpopts.NumEncodingRaw.Package())
	}
	if got := rules[TimeSerializerRule]; got != tpopts.TimeUnitNano.Serializer() {
		t.Fatalf("time rule = %q, want %q", got, tpopts.TimeUnitNano.Serializer())
	}
	if got := rules["[]byte"]; got != CompositePkg {
		t.Fatalf("[]byte rule = %q, want %q", got, CompositePkg)
	}
}
