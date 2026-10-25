// Package classifier answers what a reflect type is.
//
// The generator asks these questions before it accepts a type: a defined type
// must not be a struct, a struct must not be an interface, a typed serializer
// needs a defined basic type, a struct or a non-empty interface, and so on.
package classifier
