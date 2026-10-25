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
//
// The generator reports problems instead of panicking: a malformed validator or
// migration expression, an unexported field of a foreign struct, two types that
// would receive the same name and two types that would receive the same wire
// identifier all fail the Add* call or Generate with an error that wraps
// ErrCodeGenFailed or ErrNameCollision. Every generated struct also carries a
// SchemaFingerprint constant, a hash of its field names and types that makes a
// positional schema change visible.
//
// A Generator is not safe for concurrent use.
package codegen
