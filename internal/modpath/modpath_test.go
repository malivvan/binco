package modpath_test

import (
	"testing"

	"github.com/malivvan/binco/internal/modpath"
)

func TestCheckPath(t *testing.T) {
	valid := []string{
		"github.com/user/project",
		"github.com/user/project/v2",
		"example.com/x",
		"gopkg.in/yaml.v2",
		"gopkg.in/yaml.v2-unstable",
		"gopkg.in/check.v0",
	}
	for _, path := range valid {
		if err := modpath.CheckPath(path); err != nil {
			t.Fatalf("CheckPath(%q) = %v, want nil", path, err)
		}
	}

	invalid := []string{
		"",
		"+++",                        // no dot in the first element, invalid character
		"fmt",                        // no dot in the first element
		"github.com/user/project/",   // trailing slash
		"github.com//user/project",   // double slash
		"github.com/user/project/v1", // /v1 is implicit
		"github.com/user/project/v2.0",
		"github.com/user/project/v02",
		"github.com/user/project/@v2",
		"GitHub.com/user/project", // upper case in the domain
		"github.com/user/.hidden", // leading dot in an element
		"github.com/user/con",     // reserved on Windows
		"github.com/user/PROGRA~1",
	}
	for _, path := range invalid {
		if err := modpath.CheckPath(path); err == nil {
			t.Fatalf("CheckPath(%q) returned nil, want an error", path)
		}
	}
}

func TestCheckImportPath(t *testing.T) {
	valid := []string{
		"fmt",
		"time",
		"github.com/user/project",
		"github.com/user/project/v1", // import paths have no version rules
		"gopkg.in/yaml.v2",
		"example.com/foo++", // '+' is allowed in import paths
		"example.com/x/y_z_1",
	}
	for _, path := range valid {
		if err := modpath.CheckImportPath(path); err != nil {
			t.Fatalf("CheckImportPath(%q) = %v, want nil", path, err)
		}
	}

	invalid := []string{
		"",
		"a b",
		"-example.com/x",
		"example.com/x/",
		"example.com//x",
		"example.com/x/...",
		"example.com/x@v1",
		"example.com/x\\y",
		"example.com/x/y_z~1", // looks like a Windows short name
	}
	for _, path := range invalid {
		if err := modpath.CheckImportPath(path); err == nil {
			t.Fatalf("CheckImportPath(%q) returned nil, want an error", path)
		}
	}
}
