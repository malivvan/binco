// Package typename models the names the generator works with.
//
// Three spellings of one type are in play at all times: the complete name that
// reflect reports, the full name that the templates parse, and the relative name
// that the generated file can use. This package converts between them and
// validates package and import names.
package typename
