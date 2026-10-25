// Command gen writes a serializer with per-field options. Run it through go generate.
package main

import (
	"log"
	"os"
	"reflect"

	"github.com/malivvan/binco/demos/gen_options/pkg"

	"github.com/malivvan/binco/codegen"
	fldopts "github.com/malivvan/binco/codegen/config/field"
	genopts "github.com/malivvan/binco/codegen/config/gen"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
)

func main() {
	g, err := codegen.NewGenerator(
		genopts.WithPkgPath("github.com/malivvan/binco/demos/gen_options/pkg"),

		genopts.WithSerName(reflect.TypeFor[pkg.Foo](), "CustomFoo"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = g.AddStruct(reflect.TypeFor[pkg.Foo](),

		stopts.WithField(
			fldopts.WithType(tpopts.WithNumEncoding(tpopts.NumEncodingRaw)),
		),

		stopts.WithField(
			fldopts.WithType(tpopts.WithLenValidator("ValidateLength")),
		),
	); err != nil {
		log.Fatal(err)
	}

	bs, err := g.Generate()
	if err != nil {
		if writeErr := os.WriteFile("failed.gen.go", bs, 0o644); writeErr != nil {
			log.Fatal(writeErr)
		}
		log.Fatal(err)
	}
	if err = os.WriteFile("./pkg/binco.gen.go", bs, 0o644); err != nil {
		log.Fatal(err)
	}
}
