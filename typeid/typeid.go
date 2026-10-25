// Package typeid prefixes payloads with a type identifier.
//
// A payload alone carries no hint about what it holds. Adding a DTM (data type
// metadata) in front of it names the type, which is what makes polymorphic
// fields, versioned types and heterogeneous streams possible: the reader pulls
// the identifier off the wire and then knows which serializer to use.
package typeid
