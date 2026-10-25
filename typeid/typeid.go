// Package typeid prefixes payloads with a type identifier.
//
// A payload alone carries no hint about what it holds. Adding a DTM (data type
// metadata) in front of it names the type, which is what makes polymorphic
// fields, versioned types and heterogeneous streams possible: the reader pulls
// the identifier off the wire and then knows which serializer to use.
//
// Identifiers are positive. DTMSerializer rejects a zero or negative value with
// binco.ErrOverflow, which would otherwise be written as a huge unsigned number.
// The code generator emits the identifiers as literal constants; pin them with
// genopts.WithDTM when payloads have to outlive a build.
package typeid
