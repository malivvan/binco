package converter_test

import (
	"strings"
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	"github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

func newConverter(t *testing.T, opts ...genopts.SetOption) converter.TypeNameConverter {
	t.Helper()
	o := genopts.New()
	if err := genopts.Apply(opts, &o); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return converter.NewTypeNameConverter(o)
}

func TestConvertToFullName(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithPackage("project"),
	)
	tests := []struct {
		cname typename.CompleteName
		want  typename.FullName
	}{
		{"int", "int"},
		{"[]int", "[]int"},
		{"map[string]int", "map[string]int"},
		{"github.com/user/project/types/types.Foo", "types.Foo"},
		{"*github.com/user/project/types/types.Foo", "*types.Foo"},
		{"github.com/other/pkg/pkg.Bar", "pkg.Bar"},

		{"github.com/user/project/types/types.Foo[github.com/other/pkg.Bar]",
			"types.Foo[pkg.Bar]"},
	}
	for _, tc := range tests {
		if got := c.ConvertToFullName(tc.cname); got != tc.want {
			t.Fatalf("ConvertToFullName(%q) = %q, want %q", tc.cname, got, tc.want)
		}
	}
}

func TestConvertToFullNameWithAlias(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithImportAlias("github.com/other/pkg", "other"),
	)

	if got := c.ConvertToFullName("github.com/other/pkg/pkg.Bar"); got != "other.Bar" {
		t.Fatalf("ConvertToFullName = %q, want other.Bar", got)
	}
}

func TestConvertToFullNameTwoPathsSameAlias(t *testing.T) {
	c := newConverter(t, genopts.WithPath("github.com/user/project"))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("conversion of two paths with the same package name did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "same alias") {
			t.Fatalf("unexpected panic value %v", r)
		}
	}()
	c.ConvertToFullName("a.com/x.T1")
	c.ConvertToFullName("b.com/x.T2")
}

func TestConvertToRelativeName(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithPackage("project"),
	)
	tests := []struct {
		fname typename.FullName
		want  typename.RelativeName
	}{
		{"int", "int"},
		{"project.Foo", "Foo"},
		{"types.Foo", "types.Foo"},
		{"*[]project.Foo", "*[]Foo"},
		{"map[string]types.Foo", "map[string]types.Foo"},
	}
	for _, tc := range tests {
		if got := c.ConvertToRelativeName(tc.fname); got != tc.want {
			t.Fatalf("ConvertToRelativeName(%q) = %q, want %q", tc.fname, got, tc.want)
		}
	}
}

func TestConvertGenericsAndContainers(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithPackage("project"),
	)

	full := c.ConvertToFullName(
		"github.com/user/project/types/types.Pair[github.com/other/pkg.A, github.com/other/pkg.B]")
	if want := typename.FullName("types.Pair[pkg.A,pkg.B]"); full != want {
		t.Fatalf("ConvertToFullName = %q, want %q", full, want)
	}

	for _, tc := range []struct {
		cname typename.CompleteName
		want  typename.FullName
	}{
		{"map[string]github.com/user/project/types/types.Foo", "map[string]types.Foo"},
		{"[3]github.com/user/project/types/types.Foo", "[3]types.Foo"},
		{"[]github.com/user/project/types/types.Foo", "[]types.Foo"},
	} {
		if got := c.ConvertToFullName(tc.cname); got != tc.want {
			t.Fatalf("ConvertToFullName(%q) = %q, want %q", tc.cname, got, tc.want)
		}
	}

	if got := c.ConvertToRelativeName("types.Pair[other.A, other.B]"); got !=
		"types.Pair[other.A,other.B]" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
	if got := c.ConvertToRelativeName("map[string]types.Foo"); got != "map[string]types.Foo" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
	if got := c.ConvertToRelativeName("[3]types.Foo"); got != "[3]types.Foo" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
}

func TestConvertToRelativeNamePanics(t *testing.T) {
	c := newConverter(t, genopts.WithPath("github.com/user/project"))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("conversion of an invalid name did not panic")
		} else if msg, ok := r.(string); !ok || !strings.Contains(msg, "RelativeName") {
			t.Fatalf("unexpected panic value %v", r)
		}
	}()
	c.ConvertToRelativeName("struct{}")
}
