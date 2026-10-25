// Package genopts configures the code generator.
//
// WithPkgPath is the only required option; the rest tune the generated file:
package genopts

import (
	"go/token"
	"reflect"

	"github.com/malivvan/binco/codegen/typename"
	"golang.org/x/mod/module"
)

// ImportPath is a Go import path.
type ImportPath string

// Alias is the identifier a generated file uses for an import path.
type Alias string

// Options holds the configuration of one generator run.
type Options struct {
	PkgPath typename.PkgPath
	Package typename.Package

	Imports  []string
	SerNames map[reflect.Type]string

	Mode   Mode
	Stream bool

	importAliases map[ImportPath]Alias
}

// New returns Options with their maps initialized.
func New() Options {
	return Options{
		SerNames:      make(map[reflect.Type]string),
		importAliases: make(map[ImportPath]Alias),
	}
}

func (o Options) ImportAliases() map[ImportPath]Alias {
	return o.importAliases
}

// SetOption changes one field of Options and may fail validation.
type SetOption func(o *Options) error

// WithPkgPath sets the import path of the package the generated file belongs to, and derives its name.
func WithPkgPath(str string) SetOption {
	return func(o *Options) (err error) {
		o.PkgPath, err = typename.StrToPkgPath(str)
		if err != nil {
			return
		}
		if o.Package == "" {
			o.Package, err = typename.StrToPackage(o.PkgPath.Base())
		}
		return
	}
}

// WithPackage overrides the name of the generated package.
func WithPackage(str string) SetOption {
	return func(o *Options) (err error) {
		o.Package, err = typename.StrToPackage(str)
		return
	}
}

// WithImport adds an import path to the generated file.
func WithImport(importPath string) SetOption {
	return func(o *Options) (err error) {
		if err = module.CheckImportPath(importPath); err != nil {
			err = NewInvalidImportPathError(importPath)
			return
		}
		o.Imports = append(o.Imports, `"`+importPath+`"`)
		return
	}
}

// WithImportAlias adds an import path together with the alias used to refer to it.
func WithImportAlias(importPath, alias string) SetOption {
	return func(o *Options) (err error) {
		if err = module.CheckImportPath(importPath); err != nil {
			err = NewInvalidImportPathError(importPath)
			return
		}
		if !token.IsIdentifier(alias) {
			err = NewInvalidAliasError(alias)
			return
		}
		o.Imports = append(o.Imports, alias+" "+`"`+importPath+`"`)
		o.importAliases[ImportPath(importPath)] = Alias(alias)
		return
	}
}

// WithSerName overrides the name of the serializer generated for a type.
func WithSerName(t reflect.Type, serName string) SetOption {
	return func(o *Options) (err error) {
		o.SerNames[t] = serName
		return
	}
}

// WithUnsafe switches generation to the zero-copy codecs.
func WithUnsafe() SetOption {
	return func(o *Options) (err error) {
		o.Mode = ModeUnsafe
		return
	}
}

// WithNotUnsafe switches to the zero-copy codecs but keeps the copying string codec.
func WithNotUnsafe() SetOption {
	return func(o *Options) (err error) {
		o.Mode = ModeNotUnsafe
		return
	}
}

// WithStream generates codecs for readers and writers instead of byte slices.
func WithStream() SetOption {
	return func(o *Options) (err error) {
		o.Stream = true
		return
	}
}

func (o Options) Validate() (err error) {
	if o.PkgPath == "" {
		return ErrEmptyPkgPath
	}
	var aliases = make(map[Alias]struct{})
	for _, alias := range o.importAliases {
		if _, ok := aliases[alias]; ok {
			return NewDuplicateImportAlias(alias)
		}
		aliases[alias] = struct{}{}
	}
	return
}

func Apply(ops []SetOption, o *Options) (err error) {
	for i := range ops {
		if ops[i] != nil {
			err = ops[i](o)
			if err != nil {
				return
			}
		}
	}
	addImportAlias(o)
	return o.Validate()
}

func addImportAlias(o *Options) {
	o.importAliases[ImportPath(o.PkgPath)] = Alias(o.Package)
}
