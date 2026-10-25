package codegen_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	"github.com/malivvan/binco/testdata"
)

// TestExplicitDTMIDs covers D2: pinned identifiers are emitted as literal
// values, so adding or removing a registration cannot renumber them.
func TestExplicitDTMIDs(t *testing.T) {
	g := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 42),
		genopts.WithDTM(reflect.TypeFor[testdata.DefinedImpl](), 7),
	)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	src := generate(t, g)
	mustContain(t, src,
		"StructImplDTM  binco.DTM = 42",
		"DefinedImplDTM binco.DTM = 7",
	)

	registry, err := g.DTMRegistry()
	if err != nil {
		t.Fatalf("DTMRegistry: %v", err)
	}
	if registry["testdata.StructImpl"] != 42 || registry["testdata.DefinedImpl"] != 7 {
		t.Fatalf("registry = %v", registry)
	}
}

// TestDTMRegistryIsOrderIndependent covers the stability half of D2: the same
// pinned identifiers produce the same registry whatever the registration order.
func TestDTMRegistryIsOrderIndependent(t *testing.T) {
	first := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 1),
		genopts.WithDTM(reflect.TypeFor[testdata.DefinedImpl](), 2),
	)
	if err := first.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}

	second := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 1),
		genopts.WithDTM(reflect.TypeFor[testdata.DefinedImpl](), 2),
	)
	if err := second.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]()),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}

	a, err := first.DTMRegistry()
	if err != nil {
		t.Fatalf("DTMRegistry: %v", err)
	}
	b, err := second.DTMRegistry()
	if err != nil {
		t.Fatalf("DTMRegistry: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("registries differ:\n%v\n%v", a, b)
	}
}

// TestDuplicateDTMIDFails covers the collision half of D2.
func TestDuplicateDTMIDFails(t *testing.T) {
	g := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 5),
		genopts.WithDTM(reflect.TypeFor[testdata.DefinedImpl](), 5),
	)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	_, err := g.Generate()
	if !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("Generate returned %v, want ErrDuplicateDTM", err)
	}
	if !errors.Is(err, codegen.ErrCodeGenFailed) {
		t.Fatalf("Generate returned %v, want an error wrapping ErrCodeGenFailed", err)
	}
}

// TestMixedDTMIDModesFail covers the third D2 rule: a mix of pinned and
// automatic identifiers is rejected instead of silently depending on order.
func TestMixedDTMIDModesFail(t *testing.T) {
	g := newGenerator(t,
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), 5),
	)
	if err := g.RegisterInterface(reflect.TypeFor[testdata.Universal](),
		intropts.WithStructImpl(reflect.TypeFor[testdata.StructImpl]()),
		intropts.WithDefinedTypeImpl(reflect.TypeFor[testdata.DefinedImpl]())); err != nil {
		t.Fatalf("RegisterInterface: %v", err)
	}
	_, err := g.Generate()
	if !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("Generate returned %v, want ErrDuplicateDTM", err)
	}
	if !strings.Contains(err.Error(), "testdata.DefinedImpl") {
		t.Fatalf("error %q does not name the type without an identifier", err)
	}
}

// TestNegativeDTMIsRejected covers the last D2 note: a negative identifier used
// to be encoded as a huge unsigned value.
func TestNegativeDTMIsRejected(t *testing.T) {
	_, err := codegen.NewGenerator(
		genopts.WithPath("github.com/malivvan/binco/testdata"),
		genopts.WithDTM(reflect.TypeFor[testdata.StructImpl](), -1),
	)
	if err == nil {
		t.Fatal("NewGenerator accepted a negative identifier")
	}
}
