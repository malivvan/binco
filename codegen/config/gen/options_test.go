package genopts_test

import (
	"reflect"
	"testing"

	genopts "github.com/malivvan/binco/codegen/config/gen"
)

func TestNew(t *testing.T) {
	o := genopts.New()
	if o.SerializerNames == nil || o.ImportAliases() == nil {
		t.Fatal("New did not initialize the internal maps")
	}
}

func TestOptions(t *testing.T) {
	o := genopts.New()
	if err := genopts.WithPath("github.com/user/project")(&o); err != nil {
		t.Fatalf("WithPath: %v", err)
	}
	if o.PkgPath != "github.com/user/project" || o.Package != "project" {
		t.Fatalf("unexpected options %+v", o)
	}
	if err := genopts.WithPackage("custom")(&o); err != nil {
		t.Fatalf("WithPackage: %v", err)
	}
	if o.Package != "custom" {
		t.Fatalf("Package = %q", o.Package)
	}
	if err := genopts.WithPath("not a path")(&o); err == nil {
		t.Fatal("WithPath accepted an invalid path")
	}
	if err := genopts.WithPackage("not a package")(&o); err == nil {
		t.Fatal("WithPackage accepted an invalid name")
	}

	o = genopts.New()
	if err := genopts.WithImport("fmt")(&o); err != nil {
		t.Fatalf("WithImport: %v", err)
	}
	if len(o.Imports) != 1 || o.Imports[0] != `"fmt"` {
		t.Fatalf("Imports = %v", o.Imports)
	}
	if err := genopts.WithImport("a b")(&o); err == nil {
		t.Fatal("WithImport accepted an invalid path")
	}
	if err := genopts.WithImportAlias("io", "i")(&o); err != nil {
		t.Fatalf("WithImportAlias: %v", err)
	}
	if o.Imports[1] != `i "io"` {
		t.Fatalf("Imports = %v", o.Imports)
	}
	if got := o.ImportAliases()["io"]; got != "i" {
		t.Fatalf("ImportAliases[io] = %q", got)
	}
	if err := genopts.WithImportAlias("io", "not an alias")(&o); err == nil {
		t.Fatal("WithImportAlias accepted an invalid alias")
	}

	intType := reflect.TypeFor[int]()
	if err := genopts.WithSerializer(intType, "IntSerializer")(&o); err != nil {
		t.Fatalf("WithSerializer: %v", err)
	}
	if o.SerializerNames[intType] != "IntSerializer" {
		t.Fatalf("SerializerNames = %v", o.SerializerNames)
	}

	if err := genopts.WithUnsafe()(&o); err != nil || o.Mode != genopts.ModeUnsafe {
		t.Fatalf("WithUnsafe = (%v, %v)", o.Mode, err)
	}
	if err := genopts.WithNotUnsafe()(&o); err != nil || o.Mode != genopts.ModeNotUnsafe {
		t.Fatalf("WithNotUnsafe = (%v, %v)", o.Mode, err)
	}
	if err := genopts.WithStream()(&o); err != nil || !o.Stream {
		t.Fatalf("WithStream = (%v, %v)", o.Stream, err)
	}
}

func TestValidate(t *testing.T) {
	o := genopts.New()
	if err := o.Validate(); err != genopts.ErrEmptyPkgPath {
		t.Fatalf("Validate = %v, want ErrEmptyPkgPath", err)
	}
	o.PkgPath = "path"
	if err := o.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestApply(t *testing.T) {
	o := genopts.New()
	ops := []genopts.SetOption{
		genopts.WithPath("github.com/user/project"),
		genopts.WithStream(),
		nil,
	}
	if err := genopts.Apply(ops, &o); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if o.PkgPath != "github.com/user/project" || !o.Stream {
		t.Fatalf("unexpected options %+v", o)
	}

	if got := o.ImportAliases()[genopts.ImportPath(o.PkgPath)]; got != genopts.Alias(o.Package) {
		t.Fatalf("alias = %q", got)
	}

	o = genopts.New()
	ops = []genopts.SetOption{
		genopts.WithPath("github.com/user/project"),
		genopts.WithImportAlias("fmt", "project"),
	}
	if err := genopts.Apply(ops, &o); err == nil {
		t.Fatal("Apply accepted a duplicate alias")
	}

	o = genopts.New()
	if err := genopts.Apply([]genopts.SetOption{
		genopts.WithPath("github.com/user/project"),
		genopts.WithImport("a b"),
	}, &o); err == nil {
		t.Fatal("Apply ignored an invalid import path")
	}
}

func TestErrors(t *testing.T) {
	if got, want := genopts.NewInvalidImportPathError("x").Error(),
		"invalid 'x' import path"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got, want := genopts.NewInvalidAliasError("x").Error(),
		"invalid 'x' package alias"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got, want := genopts.NewDuplicateImportAlias("x").Error(),
		"duplicate 'x' package alias in the generator options"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got, want := genopts.NewDuplicateImportPath("x").Error(),
		"duplicate 'x' import path in the generator options"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
