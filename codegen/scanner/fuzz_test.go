//go:build fuzz

package scanner_test

import (
	"errors"
	"testing"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

// maxFuzzName keeps regexp and recursion work proportional to a small input.
const maxFuzzName = 1 << 12

func FuzzParseDefinedType(f *testing.F) {
	addNameSeeds(f)
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > maxFuzzName {
			t.Skip()
		}
		scanner.ParseDefinedType[typename.FullName](typename.FullName(name))
	})
}

func FuzzParseContainerType(f *testing.F) {
	addNameSeeds(f)
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > maxFuzzName {
			t.Skip()
		}
		scanner.ParseContainerType[typename.FullName](typename.FullName(name))
	})
}

func FuzzParsePrimitive(f *testing.F) {
	addNameSeeds(f)
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > maxFuzzName {
			t.Skip()
		}
		scanner.ParsePrimitiveType[typename.FullName](typename.FullName(name))
	})
}

func FuzzScan(f *testing.F) {
	addNameSeeds(f)
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > maxFuzzName {
			t.Skip()
		}
		for _, cfg := range []scanner.Config{{}, {WithoutParams: true}} {
			op := &limitedOp{limit: 256}
			// A limit error is expected for deeply nested or very long names;
			// the target only asserts that scanning never panics.
			_ = scanner.Scan(cfg, typename.FullName(name), op, tpopts.Options{})
		}
	})
}

func addNameSeeds(f *testing.F) {
	for _, seed := range []string{
		"",
		"int",
		"string",
		"pkg.Type",
		"*pkg.Type",
		"[3]int",
		"[]pkg.Type",
		"map[string]int",
		"map[string][]int",
		"map[[2]int]pkg.A",
		"pkg.Type[pkg.Param]",
		"pkg.Type[pkg.A, pkg.B]",
		"*[4][]byte",
		"pkg.Type[",
		"map[",
		"[]]]",
	} {
		f.Add(seed)
	}
}

var errScanLimit = errors.New("scan event limit reached")

// limitedOp stops the recursive scanner before a pathological name can recurse
// without bound. Only ProcessType can return an error, which unwinds Scan.
type limitedOp struct {
	events int
	limit  int
}

func (o *limitedOp) ProcessType(scanner.TypeInfo[typename.FullName],
	tpopts.Options,
) error {
	o.events++
	if o.events > o.limit {
		return errScanLimit
	}
	return nil
}

func (o *limitedOp) ProcessLeftSquare()  {}
func (o *limitedOp) ProcessComma()       {}
func (o *limitedOp) ProcessRightSquare() {}
