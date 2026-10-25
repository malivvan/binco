// Package codegen writes serializer code for user types.
//
// The generator inspects types through reflect and emits the code a person would
// write by hand: one method each for Marshal, Unmarshal, Size and Skip, with the
// codecs of this module inlined. The output depends on the options given to
// NewGenerator: the portable codecs by default, the zero-copy ones with
// WithUnsafe or WithNotUnsafe, and the streaming signatures with WithStream.
//
// Add types with AddDefinedType, AddStruct, AddInterface, AddVersioned and their
// companions, then call Generate and write the returned bytes to a file.
package codegen
