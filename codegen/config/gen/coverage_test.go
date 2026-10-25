package genopts_test

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"testing"

	"github.com/malivvan/binco"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

// TestWithImportAliasErrors checks the two ways an alias can be rejected.
func TestWithImportAliasErrors(t *testing.T) {
	o := genopts.New()
	if err := genopts.WithImportAlias("not a path", "x")(&o); err == nil {
		t.Fatal("WithImportAlias accepted an invalid import path")
	}
	if err := genopts.WithImportAlias("fmt", "not an alias")(&o); err == nil {
		t.Fatal("WithImportAlias accepted an invalid alias")
	}
}

// TestWithImportAlias checks that a valid path and alias are recorded in both
// the import block and the alias map.
func TestWithImportAlias(t *testing.T) {
	o := genopts.New()
	if err := genopts.WithImportAlias("example.com/other", "oth")(&o); err != nil {
		t.Fatalf("WithImportAlias: %v", err)
	}
	if len(o.Imports) != 1 || o.Imports[0] != `oth "example.com/other"` {
		t.Fatalf("Imports = %v", o.Imports)
	}
	if got := o.ImportAliases()["example.com/other"]; got != "oth" {
		t.Fatalf("ImportAliases = %q, want %q", got, "oth")
	}
}

// TestWithLogger checks the default logger and an explicit one.
func TestWithLogger(t *testing.T) {
	o := genopts.New()
	o.Logger = nil
	if err := genopts.WithLogger(nil)(&o); err != nil {
		t.Fatalf("WithLogger(nil): %v", err)
	}
	if o.Logger == nil {
		t.Fatal("WithLogger(nil) left no logger")
	}

	var buf bytes.Buffer
	logger := log.New(&buf, "prefix: ", 0)
	if err := genopts.WithLogger(logger)(&o); err != nil {
		t.Fatalf("WithLogger: %v", err)
	}
	if o.Logger != logger {
		t.Fatal("WithLogger did not store the logger")
	}
}

// TestWithDTM checks the positive-identifier rule and the stored value.
func TestWithDTM(t *testing.T) {
	tp := reflect.TypeFor[int]()
	o := genopts.New()
	if err := genopts.WithDTM(tp, 0)(&o); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("WithDTM(0) = %v, want ErrOverflow", err)
	}
	if err := genopts.WithDTM(tp, -1)(&o); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("WithDTM(-1) = %v, want ErrOverflow", err)
	}
	if err := genopts.WithDTM(tp, 7)(&o); err != nil {
		t.Fatalf("WithDTM(7): %v", err)
	}
	if got := o.DTMs[tp]; got != 7 {
		t.Fatalf("DTMs[int] = %d, want 7", got)
	}
}

// TestApplyInitializesDefaults checks that Apply fills in the maps and the
// logger, and that a failing option stops the run.
func TestApplyInitializesDefaults(t *testing.T) {
	o := genopts.New()
	o.DTMs = nil
	o.Logger = nil
	if err := genopts.WithPath("example.com/x")(&o); err != nil {
		t.Fatalf("WithPath: %v", err)
	}
	if err := genopts.Apply(nil, &o); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if o.DTMs == nil || o.Logger == nil {
		t.Fatalf("Apply left the defaults nil: %+v", o)
	}

	o = genopts.New()
	if err := genopts.Apply([]genopts.SetOption{genopts.WithDTM(reflect.TypeFor[int](), 0)}, &o); !errors.Is(err, binco.ErrOverflow) {
		t.Fatalf("Apply with a failing option = %v, want ErrOverflow", err)
	}
}

// TestApplyRejectsDuplicateAliases checks the validation step.
func TestApplyRejectsDuplicateAliases(t *testing.T) {
	o := genopts.New()
	err := genopts.Apply([]genopts.SetOption{
		genopts.WithPath("example.com/x"),
		genopts.WithImportAlias("fmt", "a"),
		genopts.WithImportAlias("io", "a"),
	}, &o)
	if err == nil {
		t.Fatal("Apply accepted two paths with the same alias")
	}
}
