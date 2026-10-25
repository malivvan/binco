package scanner_test

import (
	"testing"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

func TestParseMap(t *testing.T) {
	tests := []struct {
		name string
		in   typename.FullName
		ok   bool
		key  typename.FullName
		elem typename.FullName
	}{
		{"map", "map[string]int", true, "string", "int"},
		{"pointer to map", "*map[string][]int", true, "string", "[]int"},
		{"nested key", "map[[2]int]pkg.A", true, "[2]int", "pkg.A"},
		{"not a map", "pkg.Type", false, "", ""},
		{"unbalanced", "map[string", false, "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, key, elem, ok := scanner.ContainerType[typename.FullName](tc.in).ParseMap()
			if ok != tc.ok {
				t.Fatalf("ParseMap(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			}
			if !ok {
				return
			}
			if key != tc.key || elem != tc.elem {
				t.Fatalf("ParseMap(%q) = (%q, %q), want (%q, %q)",
					tc.in, key, elem, tc.key, tc.elem)
			}
			if info.KeyType != tc.key || info.ElemType != tc.elem {
				t.Fatalf("ParseMap(%q) info = %+v", tc.in, info)
			}
		})
	}
}

func TestScanNestedParams(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{}, "pkg.Type[[]int, map[string]int]", r,
		tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if r.events[0] != "type:Type" {
		t.Fatalf("first event = %q, want the outer type", r.events[0])
	}
	if len(r.infos) < 4 {
		t.Fatalf("parameters were not scanned: %d types, events %q", len(r.infos), events(r.events))
	}
}

func TestScanWithoutParamsSkipsNestedParams(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{WithoutParams: true},
		"pkg.Type[[]int, map[string]int]", r, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got, want := events(r.events), "type:Type"; got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
	if len(r.infos) != 1 {
		t.Fatalf("parameters were scanned: %d types", len(r.infos))
	}
}
