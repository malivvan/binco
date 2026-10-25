// Command gen writes the serializers of the pkg package. Run it through go generate.
package main

import (
	"log"
	"os"
	"reflect"

	"github.com/malivvan/binco/demos/gen/pkg"

	"github.com/malivvan/binco/codegen"
	genopts "github.com/malivvan/binco/codegen/config/gen"
)

func main() {

	g, err := codegen.NewGenerator(
		genopts.WithPkgPath("github.com/malivvan/binco/demos/gen/pkg"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = g.AddDefinedType(reflect.TypeFor[pkg.Int]()); err != nil {
		log.Fatal(err)
	}

	if err = g.AddStruct(reflect.TypeFor[pkg.Foo[pkg.Int]]()); err != nil {
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
