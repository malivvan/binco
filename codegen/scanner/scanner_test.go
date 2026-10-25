package scanner_test

import (
	"testing"

	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/scanner"
	"github.com/malivvan/binco/codegen/typename"
)

type recorder[T scanner.QualifiedName] struct {
	events []string
	infos  []scanner.TypeInfo[T]
	tops   []tpopts.Options
}

func (r *recorder[T]) ProcessType(t scanner.TypeInfo[T], tops tpopts.Options) error {
	r.events = append(r.events, "type:"+string(t.Name))
	r.infos = append(r.infos, t)
	r.tops = append(r.tops, tops)
	return nil
}

func (r *recorder[T]) ProcessLeftSquare()  { r.events = append(r.events, "[") }
func (r *recorder[T]) ProcessComma()       { r.events = append(r.events, ",") }
func (r *recorder[T]) ProcessRightSquare() { r.events = append(r.events, "]") }

func events(events []string) string {
	out := ""
	for i, e := range events {
		if i > 0 {
			out += " "
		}
		out += e
	}
	return out
}

func TestScan(t *testing.T) {
	tests := []struct {
		name       string
		scan       typename.FullName
		wantEvents string
	}{
		{"primitive", "int", "type:int"},
		{"defined", "pkg.Type", "type:Type"},
		{"pointer", "*pkg.Type", "type:Type"},
		{"array", "[3]int", "type:[3]int type:int"},
		{"slice", "[]pkg.Type", "type:[]pkg.Type type:Type"},
		{"map", "map[string]int", "type:map[string]int [ type:string ] type:int"},
		{"nested map", "map[string][]int",
			"type:map[string][]int [ type:string ] type:[]int type:int"},
		{"generic", "pkg.Type[pkg.Param]", "type:Type [ type:Param ]"},
		{"two params", "pkg.Type[pkg.A, pkg.B]", "type:Type [ type:A , type:B ]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &recorder[typename.FullName]{}
			if err := scanner.Scan(scanner.Config{}, tc.scan, r, tpopts.Options{}); err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if got := events(r.events); got != tc.wantEvents {
				t.Fatalf("events = %q, want %q", got, tc.wantEvents)
			}
		})
	}
}

func TestScanTypeInfo(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{},
		"*github.com/user/project/types.Foo", r, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	info := r.infos[0]
	if info.Kind != scanner.KindDefined {
		t.Fatalf("Kind = %v, want KindDefined", info.Kind)
	}
	if info.Stars != "*" {
		t.Fatalf("Stars = %q, want *", info.Stars)
	}
	if info.PkgPath != "github.com/user/project" {
		t.Fatalf("PkgPath = %q", info.PkgPath)
	}
	if info.Package != "types" {
		t.Fatalf("Package = %q", info.Package)
	}
	if info.Name != "Foo" {
		t.Fatalf("Name = %q", info.Name)
	}
	if info.Position != scanner.PositionUndefined {
		t.Fatalf("Position = %v", info.Position)
	}
}

func TestScanContainerInfo(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{}, "map[string][]int", r, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	m := r.infos[0]
	if m.Kind != scanner.KindMap || m.KeyType != "string" || m.ElemType != "[]int" {
		t.Fatalf("unexpected map info %+v", m)
	}
	if r.infos[1].Position != scanner.PositionKey {
		t.Fatalf("key position = %v", r.infos[1].Position)
	}
	if r.infos[2].Position != scanner.PositionElem {
		t.Fatalf("elem position = %v", r.infos[2].Position)
	}
	if r.infos[2].Kind != scanner.KindSlice || r.infos[2].ElemType != "int" {
		t.Fatalf("unexpected slice info %+v", r.infos[2])
	}
}

func TestScanArrayInfo(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{}, "*[4][]byte", r, tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	a := r.infos[0]
	if a.Kind != scanner.KindArray || a.Stars != "*" || a.ArrLength != "4" || a.ElemType != "[]byte" {
		t.Fatalf("unexpected array info %+v", a)
	}
}

func TestScanWithoutParams(t *testing.T) {
	r := &recorder[typename.FullName]{}
	if err := scanner.Scan(scanner.Config{WithoutParams: true}, "pkg.Type[pkg.A]", r,
		tpopts.Options{}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got, want := events(r.events), "type:Type"; got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestScanErrors(t *testing.T) {
	r := &recorder[typename.FullName]{}
	err := scanner.Scan(scanner.Config{}, "struct{}", r, tpopts.Options{})
	if err == nil {
		t.Fatal("Scan of an unsupported name returned no error")
	}
	if want := scanner.NewUnsupportedQualifiedNameError(typename.FullName("struct{}")); err.Error() != want.Error() {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestScanToptsPropagation(t *testing.T) {
	r := &recorder[typename.FullName]{}
	tops := tpopts.Options{LenValidator: "Validate"}
	if err := scanner.Scan(scanner.Config{}, "[]int", r, tops); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if r.tops[0].LenValidator != "Validate" {
		t.Fatalf("options were not passed to the container")
	}
	if r.tops[1].LenValidator != "" {
		t.Fatalf("options were passed to the element")
	}
}

func TestParseContainerType(t *testing.T) {
	if _, _, _, kind, ok := scanner.ParseContainerType(typename.FullName("[3]int")); !ok || kind != scanner.KindArray {
		t.Fatalf("array: ok=%v kind=%v", ok, kind)
	}
	if _, _, _, kind, ok := scanner.ParseContainerType(typename.FullName("[]int")); !ok || kind != scanner.KindSlice {
		t.Fatalf("slice: ok=%v kind=%v", ok, kind)
	}
	if _, _, _, kind, ok := scanner.ParseContainerType(typename.FullName("map[int]string")); !ok || kind != scanner.KindMap {
		t.Fatalf("map: ok=%v kind=%v", ok, kind)
	}
	if _, _, _, _, ok := scanner.ParseContainerType(typename.FullName("int")); ok {
		t.Fatal("a primitive was parsed as a container")
	}
}
