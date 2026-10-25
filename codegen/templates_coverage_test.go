package codegen_test

import (
	"testing"

	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

// TestSkipSignatureParameter checks the byte-slice spelling used by the
// non-streaming templates.
func TestSkipSignatureParameter(t *testing.T) {
	if got := (codegen.TmplFns{}).SkipSignatureLastParam(genopts.Options{}); got != "bs []byte" {
		t.Fatalf("SkipSignatureLastParam = %q, want %q", got, "bs []byte")
	}
	if got := (codegen.TmplFns{}).SkipSignatureLastParam(genopts.Options{Stream: true}); got != "r wirefmt.Reader" {
		t.Fatalf("SkipSignatureLastParam(stream) = %q", got)
	}
}

// TestNilInterfaceSize checks the one-byte encoding of a nil interface.
func TestNilInterfaceSize(t *testing.T) {
	if got := (codegen.TmplFns{}).NilInterfaceSize(); got != "1" {
		t.Fatalf("NilInterfaceSize = %q, want %q", got, "1")
	}
}
