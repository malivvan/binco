// Package typeid prefixes streamed payloads with a type identifier.
//
// Reading the identifier first tells the receiver which serializer comes next,
// which is what makes a single connection able to carry messages of different
// types. Use UnmarshalData and SkipData once the identifier has been read.
package typeid
