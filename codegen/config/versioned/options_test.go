package veropts_test

import (
	"reflect"
	"testing"

	veropts "github.com/malivvan/binco/codegen/config/versioned"
)

func TestOptions(t *testing.T) {
	v1 := reflect.TypeFor[int]()
	v2 := reflect.TypeFor[string]()
	o := veropts.Options{}
	if err := veropts.Apply(&o,
		veropts.WithVersion(v1, "Migrate"),
		veropts.WithCurrentVersion(v2),
		nil,
	); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(o.Versions) != 2 {
		t.Fatalf("Versions = %v", o.Versions)
	}
	if o.Versions[0].Type != v1 || o.Versions[0].Migration != "Migrate" {
		t.Fatalf("unexpected version %+v", o.Versions[0])
	}
	if o.Versions[1].Type != v2 || o.Versions[1].Migration != "" {
		t.Fatalf("unexpected version %+v", o.Versions[1])
	}
}

func TestValidate(t *testing.T) {
	o := veropts.Options{}
	if err := veropts.Apply(&o,
		veropts.WithCurrentVersion(reflect.TypeFor[int]()),
		veropts.WithCurrentVersion(reflect.TypeFor[string]()),
	); err == nil {
		t.Fatal("Apply accepted two current versions")
	}

	o = veropts.Options{}
	if err := veropts.Apply(&o,
		veropts.WithVersion(reflect.TypeFor[int](), "Migrate"),
		veropts.WithCurrentVersion(reflect.TypeFor[string]()),
	); err != nil {
		t.Fatalf("Apply: %v", err)
	}
}
