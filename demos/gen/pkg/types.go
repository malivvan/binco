// Package pkg declares the types the generator produces serializers for.
package pkg

type Int int

type Foo[T any] struct {
	S string
	T T
}
