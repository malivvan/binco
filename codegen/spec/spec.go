// Package spec holds the descriptions that the generator templates render.
//
// A builder fills these structures from reflect types; the templates only read
// them. Keeping the descriptions separate from both the reflection and the
// template layer is what lets the generator support new kinds of declarations
// without touching the templates.
package spec
