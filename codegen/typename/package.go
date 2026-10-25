package typename

import (
	"go/token"
	"path/filepath"

	"github.com/malivvan/binco/internal/modpath"
)

// Package is a Go package name.
type Package string

// StrToPackage checks that str can be used as a package name.
func StrToPackage(str string) (pkg Package, err error) {
	if !token.IsIdentifier(str) {
		err = NewInvalidPackageError(str)
		return
	}
	pkg = Package(str)
	return
}

// PkgPath is a Go import path.
type PkgPath string

func (p PkgPath) Base() string {
	return filepath.Base(string(p))
}

// StrToPkgPath checks that str is a valid import path.
func StrToPkgPath(str string) (path PkgPath, err error) {
	if err = modpath.CheckPath(str); err != nil {
		err = NewInvalidPkgPathError(str)
		return
	}
	path = PkgPath(str)
	return
}
