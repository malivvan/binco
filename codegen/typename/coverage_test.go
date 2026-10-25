package typename_test

import (
	"reflect"
	"testing"

	"github.com/malivvan/binco/codegen/typename"
)

// TestContainerElementErrors checks that a container whose element has no name
// reports the element's error instead of building a broken name.
func TestContainerElementErrors(t *testing.T) {
	if _, err := typename.TypeCompleteName(reflect.TypeFor[[2]**int]()); err == nil {
		t.Fatal("an array of a multi-pointer element was accepted")
	}
	if _, err := typename.TypeCompleteName(reflect.TypeFor[[]**int]()); err == nil {
		t.Fatal("a slice of a multi-pointer element was accepted")
	}
}
