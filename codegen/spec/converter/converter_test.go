package converter_test

import (
	"errors"
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

// fullName converts and fails the test on error.
func fullName(t *testing.T, c converter.TypeNameConverter, name typename.CompleteName) typename.FullName {
	t.Helper()
	got, err := c.ConvertToFullName(name)
	if err != nil {
		t.Fatalf("ConvertToFullName(%q): %v", name, err)
	}
	return got
}

// relName converts and fails the test on error.
func relName(t *testing.T, c converter.TypeNameConverter, name typename.FullName) typename.RelativeName {
	t.Helper()
	got, err := c.ConvertToRelativeName(name)
	if err != nil {
		t.Fatalf("ConvertToRelativeName(%q): %v", name, err)
	}
	return got
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
		if got := fullName(t, c, tc.cname); got != tc.want {
			t.Fatalf("ConvertToFullName(%q) = %q, want %q", tc.cname, got, tc.want)
		}
	}
}

func TestConvertToFullNameWithAlias(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithImportAlias("github.com/other/pkg", "other"),
	)

	if got := fullName(t, c, "github.com/other/pkg/pkg.Bar"); got != "other.Bar" {
		t.Fatalf("ConvertToFullName = %q, want other.Bar", got)
	}
}

// TestConvertToFullNameTwoPathsSameAlias covers D5: the converter used to panic
// on a name it could not rewrite; it now returns an error that Generate wraps.
func TestConvertToFullNameTwoPathsSameAlias(t *testing.T) {
	c := newConverter(t, genopts.WithPath("github.com/user/project"))
	if _, err := c.ConvertToFullName("a.com/x.T1"); err != nil {
		t.Fatalf("first conversion: %v", err)
	}
	_, err := c.ConvertToFullName("b.com/x.T2")
	if err == nil {
		t.Fatal("conversion of two paths with the same package name returned no error")
	}
	if !strings.Contains(err.Error(), "same alias") {
		t.Fatalf("unexpected error %v", err)
	}
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
		if got := relName(t, c, tc.fname); got != tc.want {
			t.Fatalf("ConvertToRelativeName(%q) = %q, want %q", tc.fname, got, tc.want)
		}
	}
}

func TestConvertGenericsAndContainers(t *testing.T) {
	c := newConverter(t,
		genopts.WithPath("github.com/user/project"),
		genopts.WithPackage("project"),
	)

	full := fullName(t, c,
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
		if got := fullName(t, c, tc.cname); got != tc.want {
			t.Fatalf("ConvertToFullName(%q) = %q, want %q", tc.cname, got, tc.want)
		}
	}

	if got := relName(t, c, "types.Pair[other.A, other.B]"); got !=
		"types.Pair[other.A,other.B]" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
	if got := relName(t, c, "map[string]types.Foo"); got != "map[string]types.Foo" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
	if got := relName(t, c, "[3]types.Foo"); got != "[3]types.Foo" {
		t.Fatalf("ConvertToRelativeName = %q", got)
	}
}

// TestConvertToRelativeNameReturnsError covers the other half of D5.
func TestConvertToRelativeNameReturnsError(t *testing.T) {
	c := newConverter(t, genopts.WithPath("github.com/user/project"))
	if _, err := c.ConvertToRelativeName("struct{}"); err == nil {
		t.Fatal("conversion of an invalid name returned no error")
	} else if !errors.Is(err, err) {
		t.Fatalf("unexpected error %v", err)
	}
}
