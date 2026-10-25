package converter

import (
	"strings"
	"sync"
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

// TestProcessTypeRejectsAnUnknownKind checks the internal guard in both rewriters:
// a type the scanner should never produce is reported instead of rendered.
func TestProcessTypeRejectsAnUnknownKind(t *testing.T) {
	unknown := scanner.TypeInfo[typename.CompleteName]{Kind: scanner.KindUndefined}
	full := newFullNameOp(map[typename.Package]typename.PkgPath{}, new(sync.Mutex), genopts.New())
	if err := full.ProcessType(unknown, tpopts.Options{}); err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("fullNameOp.ProcessType = %v, want an unexpected-kind error", err)
	}

	relative := newRelativeNameOp(genopts.New())
	unknownRel := scanner.TypeInfo[typename.FullName]{Kind: scanner.KindUndefined}
	if err := relative.ProcessType(unknownRel, tpopts.Options{}); err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("relativeNameOp.ProcessType = %v, want an unexpected-kind error", err)
	}
}
