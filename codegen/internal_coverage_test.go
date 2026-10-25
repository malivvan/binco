package codegen

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/malivvan/binco"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/testdata"
	other "github.com/malivvan/binco/testdata/other"
)

// conflictingAliasOptions maps two import paths to one alias, which the
// validation of the public options rejects but which makes the converter report
// a conflict.
func conflictingAliasOptions(t *testing.T) genopts.Options {
	t.Helper()
	settings := genopts.New()
	if err := genopts.WithPath("github.com/malivvan/binco/testdata")(&settings); err != nil {
		t.Fatalf("WithPath: %v", err)
	}
	for _, path := range []string{
		"github.com/malivvan/binco/testdata",
		"github.com/malivvan/binco/testdata/other",
	} {
		if err := genopts.WithImportAlias(path, "dup")(&settings); err != nil {
			t.Fatalf("WithImportAlias(%q): %v", path, err)
		}
	}
	return settings
}

// brokenTemplates returns a template set in which one named template fails.
func brokenTemplates(t *testing.T, failName string) *template.Template {
	t.Helper()
	root := template.New("base").Funcs(template.FuncMap{
		"fail": func() (string, error) { return "", errors.New("template failure") },
	})
	for _, name := range []string{
		packageTmpl, importsTmpl, anonDefinitionsTmpl, dtmsDefinitionTmpl,
		structSerializerTmpl, definedTypeSerializerTmpl, typedSerializerTmpl,
		interfaceSerializerTmpl, versionedSerializerTmpl,
	} {
		body := "ok"
		if name == failName {
			body = "{{fail}}"
		}
		template.Must(root.New(name).Parse(body))
	}
	return root
}

// TestRenderAllReportsTemplateErrors checks that a failing template is reported
// with the section it belongs to.
func TestRenderAllReportsTemplateErrors(t *testing.T) {
	cases := []struct{ name, want string }{
		{packageTmpl, "package"},
		{importsTmpl, "imports"},
		{dtmsDefinitionTmpl, "type identifiers"},
		{anonDefinitionsTmpl, "anonymous serializers"},
		{structSerializerTmpl, "serializers"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGenerator(genopts.WithPath("github.com/malivvan/binco/testdata"))
			if err != nil {
				t.Fatalf("NewGenerator: %v", err)
			}
			if err = g.AddStruct(reflect.TypeFor[testdata.SimpleStruct]()); err != nil {
				t.Fatalf("AddStruct: %v", err)
			}
			g.addDTM([]dtmEntry{{Name: "SimpleStruct", FullName: "testdata.SimpleStruct"}})
			g.rootTmpl = brokenTemplates(t, tc.name)

			var buf bytes.Buffer
			if err = g.renderAll(&buf); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("renderAll with a failing %s template = %v", tc.name, err)
			}
		})
	}
}

// TestResolveDTMsRejectsNonPositiveIDs checks the invariant that a wire
// identifier must be positive, which the public options already enforce.
func TestResolveDTMsRejectsNonPositiveIDs(t *testing.T) {
	g, err := NewGenerator(genopts.WithPath("github.com/malivvan/binco/testdata"))
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	g.dtmGroups = [][]dtmEntry{{{Name: "SimpleStruct", FullName: "testdata.SimpleStruct"}}}
	g.dtmIDs = map[string]binco.DTM{"testdata.SimpleStruct": 0}
	if _, err := g.resolveDTMs(); !errors.Is(err, binco.ErrDuplicateDTM) {
		t.Fatalf("resolveDTMs with identifier 0 = %v, want ErrDuplicateDTM", err)
	}
}

// TestTypeDTMEntryRejectsUnnamedTypes checks the name conversion of a registered
// type: a type without a name cannot receive an identifier.
func TestTypeDTMEntryRejectsUnnamedTypes(t *testing.T) {
	g, err := NewGenerator(genopts.WithPath("github.com/malivvan/binco/testdata"))
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	if _, err := g.typeDTMEntry(reflect.TypeFor[struct{ X int }]()); err == nil {
		t.Fatal("typeDTMEntry accepted an unnamed type")
	}
}

// TestTypeDTMEntryReportsConversionErrors checks the second conversion of an
// identifier: a name the converter cannot rewrite is reported.
func TestTypeDTMEntryReportsConversionErrors(t *testing.T) {
	g, err := NewGenerator(genopts.WithPath("github.com/malivvan/binco/testdata"))
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	settings := conflictingAliasOptions(t)
	g.converter = cnvtr.NewTypeNameConverter(settings)

	if _, err := g.typeDTMEntry(reflect.TypeFor[testdata.SameName]()); err != nil {
		t.Fatalf("first typeDTMEntry: %v", err)
	}
	if _, err := g.typeDTMEntry(reflect.TypeFor[other.SameName]()); err == nil {
		t.Fatal("typeDTMEntry ignored a failing conversion")
	}
}

// TestDTMIDsForReportsConversionErrors checks the pinned-identifier resolution.
func TestDTMIDsForReportsConversionErrors(t *testing.T) {
	settings := conflictingAliasOptions(t)
	settings.DTMs = map[reflect.Type]binco.DTM{
		reflect.TypeFor[testdata.SameName](): 1,
		reflect.TypeFor[other.SameName]():    2,
	}
	if _, err := dtmIDsFor(settings, cnvtr.NewTypeNameConverter(settings)); err == nil {
		t.Fatal("dtmIDsFor ignored a failing conversion")
	}
}
