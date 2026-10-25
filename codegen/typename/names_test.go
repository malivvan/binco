package typename_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen/typename"
	"github.com/malivvan/binco/testdata"
)

func TestTypeCompleteName(t *testing.T) {
	tests := []struct {
		t    reflect.Type
		want typename.CompleteName
	}{
		{reflect.TypeFor[int](), "int"},
		{reflect.TypeFor[*int](), "*int"},
		{reflect.TypeFor[[3]int](), "[3]int"},
		{reflect.TypeFor[*[3]int](), "*[3]int"},
		{reflect.TypeFor[[3]*int](), "[3]*int"},
		{reflect.TypeFor[[]int](), "[]int"},
		{reflect.TypeFor[*[]int](), "*[]int"},
		{reflect.TypeFor[map[int]string](), "map[int]string"},
		{reflect.TypeFor[map[*int]string](), "map[*int]string"},

		{reflect.TypeFor[testdata.SimpleStruct](),
			"github.com/malivvan/binco/testdata/testdata.SimpleStruct"},
		{reflect.TypeFor[*testdata.SimpleStruct](),
			"*github.com/malivvan/binco/testdata/testdata.SimpleStruct"},
	}
	for _, tc := range tests {
		got, err := typename.TypeCompleteName(tc.t)
		if err != nil {
			t.Fatalf("TypeCompleteName(%v): %v", tc.t, err)
		}
		if got != tc.want {
			t.Fatalf("TypeCompleteName(%v) = %q, want %q", tc.t, got, tc.want)
		}
	}
}

func TestTypeCompleteNameErrors(t *testing.T) {
	if _, err := typename.TypeCompleteName(reflect.TypeFor[**int]()); err == nil {
		t.Fatal("multi-pointer type returned no error")
	}
	if _, err := typename.TypeCompleteName(reflect.TypeFor[struct{}]()); err == nil {
		t.Fatal("struct type returned no error")
	}
	if got, want := typename.MustTypeCompleteName(reflect.TypeFor[string]()), typename.CompleteName("string"); got != want {
		t.Fatalf("MustTypeCompleteName = %q, want %q", got, want)
	}
}

func TestBaseTypeCompleteName(t *testing.T) {
	tests := []struct {
		t    reflect.Type
		want typename.CompleteName
	}{
		{reflect.TypeFor[testdata.Int](), "int"},
		{reflect.TypeFor[testdata.IntSlice](), "[]int"},
		{reflect.TypeFor[testdata.Str](), "string"},
	}
	for _, tc := range tests {
		got, err := typename.BaseTypeCompleteName(tc.t)
		if err != nil {
			t.Fatalf("BaseTypeCompleteName(%v): %v", tc.t, err)
		}
		if got != tc.want {
			t.Fatalf("BaseTypeCompleteName(%v) = %q, want %q", tc.t, got, tc.want)
		}
	}
	if _, err := typename.BaseTypeCompleteName(reflect.TypeFor[int]()); err != typename.ErrTypeMismatch {
		t.Fatalf("BaseTypeCompleteName(int) returned %v, want ErrTypeMismatch", err)
	}
	if _, err := typename.BaseTypeCompleteName(reflect.TypeFor[testdata.SimpleStruct]()); err == nil {
		t.Fatal("BaseTypeCompleteName(struct) returned no error")
	}
}

func TestFullName(t *testing.T) {
	tests := []struct {
		name     typename.FullName
		pkg      typename.Package
		typeName typename.TypeName
	}{
		{"pkg.Type", "pkg", "Type"},
		{"Type", "", "Type"},
		{"pkg123.Type", "pkg123", "Type"},
		{"pkg.Type[pkg.Type]", "pkg", "Type[pkg.Type]"},
		{"", "", ""},
	}
	for _, tc := range tests {
		if got := tc.name.Package(); got != tc.pkg {
			t.Fatalf("%q.Package() = %q, want %q", tc.name, got, tc.pkg)
		}
		if got := tc.name.TypeName(); got != tc.typeName {
			t.Fatalf("%q.TypeName() = %q, want %q", tc.name, got, tc.typeName)
		}
	}
	if got := typename.MakeFullName("pkg", "Type"); got != typename.FullName("pkg.Type") {
		t.Fatalf("MakeFullName = %q", got)
	}
	if got := typename.MakeFullName("", "Type"); got != typename.FullName("Type") {
		t.Fatalf("MakeFullName = %q", got)
	}
}

func TestRelativeName(t *testing.T) {
	if got := typename.RelativeName("pkg.Type").WithoutSquares(); got != "pkg.Type" {
		t.Fatalf("WithoutSquares = %q", got)
	}
	if got := typename.RelativeName("pkg.Type[Param1, Param2]").WithoutSquares(); got != "pkg.Type" {
		t.Fatalf("WithoutSquares = %q", got)
	}
}

func TestPkgNames(t *testing.T) {
	if pkg, err := typename.StrToPackage("pkg"); err != nil || pkg != typename.Package("pkg") {
		t.Fatalf("StrToPackage = (%q, %v)", pkg, err)
	}
	if _, err := typename.StrToPackage("+++"); err == nil {
		t.Fatal("StrToPackage accepted an invalid package")
	}
	path, err := typename.StrToPkgPath("github.com/user/project")
	if err != nil || path != typename.PkgPath("github.com/user/project") {
		t.Fatalf("StrToPkgPath = (%q, %v)", path, err)
	}
	if _, err = typename.StrToPkgPath("+++"); err == nil {
		t.Fatal("StrToPkgPath accepted an invalid path")
	}
	if got := path.Base(); got != "project" {
		t.Fatalf("Base = %q", got)
	}
}

type (
	arrayType [3]int
	sliceType []int
	mapType   map[int]string
	ptrType   *int
	doublePtr **int
)

func TestBaseTypeCompleteNameContainers(t *testing.T) {
	tests := []struct {
		typ  reflect.Type
		want typename.CompleteName
	}{
		{reflect.TypeFor[arrayType](), "[3]int"},
		{reflect.TypeFor[sliceType](), "[]int"},
		{reflect.TypeFor[mapType](), "map[int]string"},
		{reflect.TypeFor[ptrType](), "*int"},
	}
	for _, tc := range tests {
		got, err := typename.BaseTypeCompleteName(tc.typ)
		if err != nil {
			t.Fatalf("BaseTypeCompleteName(%v): %v", tc.typ, err)
		}
		if got != tc.want {
			t.Fatalf("BaseTypeCompleteName(%v) = %q, want %q", tc.typ, got, tc.want)
		}
	}

	if _, err := typename.BaseTypeCompleteName(reflect.TypeFor[doublePtr]()); err == nil {
		t.Fatal("a double pointer was accepted")
	}
}

func TestMustTypeCompleteNamePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustTypeCompleteName did not panic")
		}
	}()
	typename.MustTypeCompleteName(reflect.TypeFor[struct{}]())
}

func TestTypeCompleteNameMapErrors(t *testing.T) {

	if _, err := typename.TypeCompleteName(reflect.TypeFor[map[**int]string]()); err == nil {
		t.Fatal("a multi-pointer map key was accepted")
	}
	if _, err := typename.TypeCompleteName(reflect.TypeFor[map[int]**string]()); err == nil {
		t.Fatal("a multi-pointer map value was accepted")
	}
}
