package codegen_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/malivvan/binco/testdata"
)

// simpleStructFingerprint is the committed fingerprint of testdata.SimpleStruct.
//
// The wire format is positional: the meaning of a payload is the declaration
// order of the fields. A CI test that compares the generated constant with this
// committed value fails as soon as a field is added, removed, renamed, retyped
// or reordered, which is exactly when previously stored payloads stop decoding
// as they used to.
const simpleStructFingerprint = "03437121bdc167620d9f71555cda4157489041a513ae756be640d3851eaa580d"

// TestSchemaFingerprintIsStable covers D4.
func TestSchemaFingerprintIsStable(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[testdata.SimpleStruct]()); err != nil {
		t.Fatalf("AddStruct: %v", err)
	}
	src := generate(t, g)
	mustContain(t, src,
		"const SimpleStructSchemaFingerprint = \""+simpleStructFingerprint+"\"")
}

type orderAB struct {
	A int
	B string
}

type orderBA struct {
	B string
	A int
}

// TestSchemaFingerprintDetectsReordering covers the point of D4: swapping two
// fields of different types changes the wire meaning and must change the hash.
func TestSchemaFingerprintDetectsReordering(t *testing.T) {
	first := fingerprintOf(t, reflect.TypeFor[orderAB]())
	second := fingerprintOf(t, reflect.TypeFor[orderBA]())
	if first == second {
		t.Fatalf("reordering the fields produced the same fingerprint %q", first)
	}
	if len(first) != 64 {
		t.Fatalf("fingerprint %q is not a sha256 hex digest", first)
	}
}

func fingerprintOf(t *testing.T, typ reflect.Type) string {
	t.Helper()
	g := newGenerator(t)
	if err := g.AddStruct(typ); err != nil {
		t.Fatalf("AddStruct(%v): %v", typ, err)
	}
	src := generate(t, g)
	const marker = "SchemaFingerprint = \""
	i := strings.Index(src, marker)
	if i < 0 {
		t.Fatalf("no fingerprint in generated code")
	}
	rest := src[i+len(marker):]
	j := strings.Index(rest, "\"")
	if j < 0 {
		t.Fatalf("unterminated fingerprint in generated code")
	}
	return rest[:j]
}

// TestForeignUnexportedFieldIsRejected covers the second half of D4: generated
// code cannot reach an unexported field of another package, so the failure has
// to be reported during generation. time.Time is a struct from another package
// whose fields are all unexported, which is the case a local fixture would have
// to imitate.
func TestForeignUnexportedFieldIsRejected(t *testing.T) {
	g := newGenerator(t)
	if err := g.AddStruct(reflect.TypeFor[time.Time]()); err == nil {
		t.Fatal("AddStruct accepted a foreign struct with an unexported field")
	}
}
