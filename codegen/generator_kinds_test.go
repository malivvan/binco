package codegen_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen"
)

func TestGeneratorRejectsWrongKinds(t *testing.T) {
	tests := []struct {
		name string
		fn   func(g *codegen.Generator) error
	}{
		{"AddTyped of a primitive", func(g *codegen.Generator) error {
			return g.AddTyped(reflect.TypeFor[int]())
		}},
		{"AddInterface of a non-interface", func(g *codegen.Generator) error {
			return g.AddInterface(reflect.TypeFor[int]())
		}},
		{"AddVersioned of a non-defined type", func(g *codegen.Generator) error {
			return g.AddVersioned(reflect.TypeFor[int]())
		}},
		{"RegisterInterface of a non-interface", func(g *codegen.Generator) error {
			return g.RegisterInterface(reflect.TypeFor[int]())
		}},
		{"RegisterVersioned of a non-defined type", func(g *codegen.Generator) error {
			return g.RegisterVersioned(reflect.TypeFor[int]())
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newGenerator(t)
			if err := tc.fn(g); err == nil {
				t.Fatal("the generator accepted a type of the wrong kind")
			}
		})
	}
}
