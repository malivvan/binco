package genopts

import (
	"errors"
	"fmt"
)

// ErrEmptyPkgPath reports a generator run without genopts.WithPath.
var ErrEmptyPkgPath = errors.New("option WithPath is not set")

// NewInvalidImportPathError reports a path WithImport cannot use.
func NewInvalidImportPathError(val string) error {
	return fmt.Errorf("invalid '%v' import path", val)
}

// NewInvalidAliasError reports an import alias that is not an identifier.
func NewInvalidAliasError(val string) error {
	return fmt.Errorf("invalid '%v' package alias", val)
}

// NewDuplicateImportPath reports the same path added twice.
func NewDuplicateImportPath(path ImportPath) error {
	return fmt.Errorf("duplicate '%v' import path in the generator options", path)
}

// NewDuplicateImportAlias reports two paths that share one alias.
func NewDuplicateImportAlias(alias Alias) error {
	return fmt.Errorf("duplicate '%v' package alias in the generator options",
		alias)
}
