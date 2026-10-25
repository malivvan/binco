package converter

import (
	"fmt"

	"github.com/malivvan/binco/codegen/typename"
)

// NewTwoPathsSameAliasError reports two import paths that would be referred to
// by the same package name, which makes the generated file ambiguous.
func NewTwoPathsSameAliasError(pkgPath1, pkgPath2 typename.PkgPath,
	alias typename.Package) error {
	return fmt.Errorf("two pkgPath '%v' and '%v' have the same alias '%v'",
		pkgPath1, pkgPath2, alias)
}
