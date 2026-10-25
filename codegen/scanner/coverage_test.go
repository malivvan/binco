package scanner_test

import (
	"errors"
	"strings"
	"testing"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

// failingOp fails when it meets a type whose name contains failOn, or on the
// callAt-th call when callAt is not zero.
type failingOp struct {
	failOn string
	callAt int
	calls  int
	err    error
}

func (o *failingOp) ProcessType(t scanner.TypeInfo[typename.CompleteName], tops tpopts.Options) error {
	o.calls++
	if o.callAt > 0 && o.calls == o.callAt {
		return o.err
	}
	if o.failOn != "" && strings.Contains(string(t.Name), o.failOn) {
		return o.err
	}
	return nil
}
func (o *failingOp) ProcessLeftSquare()  {}
func (o *failingOp) ProcessComma()       {}
func (o *failingOp) ProcessRightSquare() {}

// TestScanPropagatesOpErrors checks that an error from any level of a name stops
// the scan and is returned unchanged.
func TestScanPropagatesOpErrors(t *testing.T) {
	boom := errors.New("boom")
	cases := []struct {
		name   string
		cfg    scanner.Config
		scan   typename.CompleteName
		failOn string
		callAt int
	}{
		{"container", scanner.Config{}, "[]example.com/p.Foo", "Foo", 0},
		// A map reports its key and its element in one type name, so the key
		// error is reached by failing on the second type the scan reports.
		{"container key", scanner.Config{}, "map[example.com/p.Key]example.com/p.Value", "", 2},
		{"defined", scanner.Config{}, "example.com/p.Foo", "Foo", 0},
		{"type parameter", scanner.Config{}, "example.com/p.Gen[example.com/p.Param]", "Param", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op := &failingOp{failOn: tc.failOn, callAt: tc.callAt, err: boom}
			if err := scanner.Scan(tc.cfg, tc.scan, op, tpopts.Options{}); !errors.Is(err, boom) {
				t.Fatalf("Scan(%q) = %v, want %v", tc.scan, err, boom)
			}
		})
	}
}

// countingOp records how many types it saw.
type countingOp struct{ types int }

func (o *countingOp) ProcessType(t scanner.TypeInfo[typename.CompleteName], tops tpopts.Options) error {
	o.types++
	return nil
}
func (o *countingOp) ProcessLeftSquare()  {}
func (o *countingOp) ProcessComma()       {}
func (o *countingOp) ProcessRightSquare() {}

// TestWithoutParamsSkipsTypeParameters checks the config flag that ignores the
// parameters of a generic name.
func TestWithoutParamsSkipsTypeParameters(t *testing.T) {
	with := &countingOp{}
	if err := scanner.Scan(scanner.Config{}, "example.com/p.Gen[example.com/p.Param]", with, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	without := &countingOp{}
	if err := scanner.Scan(scanner.Config{WithoutParams: true}, "example.com/p.Gen[example.com/p.Param]", without, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if without.types >= with.types {
		t.Fatalf("WithoutParams saw %d types, with parameters %d", without.types, with.types)
	}
}
