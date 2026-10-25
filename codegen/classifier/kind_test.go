package classifier_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen/classifier"
	"github.com/malivvan/binco/testdata"
)

func TestDefined(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type

		defined       bool
		basic         bool
		structType    bool
		iface         bool
		nonEmptyIface bool
	}{
		{"defined int", reflect.TypeFor[testdata.Int](), true, true, false, false, false},
		{"defined slice", reflect.TypeFor[testdata.IntSlice](), true, true, false, false, false},
		{"defined string", reflect.TypeFor[testdata.Str](), true, true, false, false, false},
		{"defined struct", reflect.TypeFor[testdata.SimpleStruct](), true, false, true, false, false},
		{"defined interface", reflect.TypeFor[testdata.Universal](), true, false, false, true, true},

		{"builtin int", reflect.TypeFor[int](), false, false, false, false, false},
		{"builtin slice", reflect.TypeFor[[]int](), false, false, false, false, false},
		{"builtin map", reflect.TypeFor[map[string]int](), false, false, false, false, false},
		{"builtin pointer", reflect.TypeFor[*int](), false, false, false, false, false},
		{"anonymous struct", reflect.TypeFor[struct{ A int }](), false, false, false, false, false},
		{"anonymous interface", reflect.TypeFor[any](), false, false, false, false, false},
		{"channel", reflect.TypeFor[chan int](), false, false, false, false, false},
		{"func", reflect.TypeFor[func()](), false, false, false, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifier.DefinedType(tc.typ); got != tc.defined {
				t.Fatalf("DefinedType = %v, want %v", got, tc.defined)
			}
			if got := classifier.DefinedBasicType(tc.typ); got != tc.basic {
				t.Fatalf("DefinedBasicType = %v, want %v", got, tc.basic)
			}
			if got := classifier.DefinedStruct(tc.typ); got != tc.structType {
				t.Fatalf("DefinedStruct = %v, want %v", got, tc.structType)
			}
			if got := classifier.DefinedInterface(tc.typ); got != tc.iface {
				t.Fatalf("DefinedInterface = %v, want %v", got, tc.iface)
			}
			if got := classifier.DefinedNonEmptyInterface(tc.typ); got != tc.nonEmptyIface {
				t.Fatalf("DefinedNonEmptyInterface = %v, want %v", got, tc.nonEmptyIface)
			}
		})
	}
}

func TestKinds(t *testing.T) {
	tests := []struct {
		name      string
		typ       reflect.Type
		primitive bool
		array     bool
		slice     bool
		mapType   bool
		container bool
		pointer   bool
		iface     bool
		basic     bool
	}{
		{"bool", reflect.TypeFor[bool](), true, false, false, false, false, false, false, true},
		{"int8", reflect.TypeFor[int8](), true, false, false, false, false, false, false, true},
		{"uint64", reflect.TypeFor[uint64](), true, false, false, false, false, false, false, true},
		{"float32", reflect.TypeFor[float32](), true, false, false, false, false, false, false, true},
		{"string", reflect.TypeFor[string](), true, false, false, false, false, false, false, true},
		{"uintptr", reflect.TypeFor[uintptr](), false, false, false, false, false, false, false, false},
		{"complex64", reflect.TypeFor[complex64](), false, false, false, false, false, false, false, false},
		{"array", reflect.TypeFor[[3]int](), false, true, false, false, true, false, false, true},
		{"slice", reflect.TypeFor[[]int](), false, false, true, false, true, false, false, true},
		{"map", reflect.TypeFor[map[int]string](), false, false, false, true, true, false, false, true},
		{"pointer", reflect.TypeFor[*int](), false, false, false, false, false, true, false, true},
		{"interface", reflect.TypeFor[any](), false, false, false, false, false, false, true, false},
		{"struct", reflect.TypeFor[struct{}](), false, false, false, false, false, false, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifier.PrimitiveType(tc.typ); got != tc.primitive {
				t.Fatalf("PrimitiveType = %v, want %v", got, tc.primitive)
			}
			if got := classifier.ArrayType(tc.typ); got != tc.array {
				t.Fatalf("ArrayType = %v, want %v", got, tc.array)
			}
			if got := classifier.SliceType(tc.typ); got != tc.slice {
				t.Fatalf("SliceType = %v, want %v", got, tc.slice)
			}
			if got := classifier.MapType(tc.typ); got != tc.mapType {
				t.Fatalf("MapType = %v, want %v", got, tc.mapType)
			}
			if got := classifier.ContainerType(tc.typ); got != tc.container {
				t.Fatalf("ContainerType = %v, want %v", got, tc.container)
			}
			if got := classifier.PtrType(tc.typ); got != tc.pointer {
				t.Fatalf("PtrType = %v, want %v", got, tc.pointer)
			}
			if got := classifier.InterfaceType(tc.typ); got != tc.iface {
				t.Fatalf("InterfaceType = %v, want %v", got, tc.iface)
			}
			if got := classifier.BasicType(tc.typ); got != tc.basic {
				t.Fatalf("BasicType = %v, want %v", got, tc.basic)
			}
		})
	}
}
