// Package genopts configures the code generator.
//
// WithPath is the only required option; the rest tune the generated file:
package genopts

import (
	"go/token"
	"io"
	"log"
	"reflect"
	"strconv"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/codegen/typename"
	"github.com/malivvan/binco/internal/modpath"
)

// ImportPath is a Go import path.
type ImportPath string

// Alias is the identifier a generated file uses for an import path.
type Alias string

// Options holds the configuration of one generator run.
type Options struct {
	PkgPath typename.PkgPath
	Package typename.Package

	Imports         []string
	SerializerNames map[reflect.Type]string
	// DTMs pins the wire identifier of a type that is registered as a versioned
	// type or as an interface implementation. Without it the generator numbers
	// the types in registration order, so adding or removing one renumbers the
	// others and old payloads decode as the wrong type.
	DTMs map[reflect.Type]binco.DTM

	Mode   Mode
	Stream bool

	// Logger receives the warnings the generator produces, such as a type
	// parameter without an explicit import alias. Library code must not write to
	// the standard logger of the process.
	Logger *log.Logger

	importAliases map[ImportPath]Alias
}

// New returns Options with their maps initialized.
func New() Options {
	return Options{
		SerializerNames: make(map[reflect.Type]string),
		DTMs:            make(map[reflect.Type]binco.DTM),
		Logger:          log.New(io.Discard, "", 0),
		importAliases:   make(map[ImportPath]Alias),
	}
}

// ImportAliases returns the alias of every configured import path.
func (o Options) ImportAliases() map[ImportPath]Alias {
	return o.importAliases
}

// SetOption changes one field of Options and may fail validation.
type SetOption func(o *Options) error

// WithPath sets the import path of the package the generated file belongs to, and derives its name.
func WithPath(str string) SetOption {
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
		if err = modpath.CheckImportPath(importPath); err != nil {
			err = NewInvalidImportPathError(importPath)
			return
		}
		o.Imports = append(o.Imports, strconv.Quote(importPath))
		return
	}
}

// WithImportAlias adds an import path together with the alias used to refer to it.
func WithImportAlias(importPath, alias string) SetOption {
	return func(o *Options) (err error) {
		if err = modpath.CheckImportPath(importPath); err != nil {
			err = NewInvalidImportPathError(importPath)
			return
		}
		if !token.IsIdentifier(alias) {
			err = NewInvalidAliasError(alias)
			return
		}
		o.Imports = append(o.Imports, alias+" "+strconv.Quote(importPath))
		o.importAliases[ImportPath(importPath)] = Alias(alias)
		return
	}
}

// WithSerializer overrides the name of the serializer generated for a type.
func WithSerializer(t reflect.Type, serializerName string) SetOption {
	return func(o *Options) (err error) {
		o.SerializerNames[t] = serializerName
		return
	}
}

// WithLogger sets the destination of the generator's warnings. The default
// discards them.
func WithLogger(logger *log.Logger) SetOption {
	return func(o *Options) (err error) {
		if logger == nil {
			logger = log.New(io.Discard, "", 0)
		}
		o.Logger = logger
		return
	}
}

// WithDTM pins the wire identifier of a type that is registered as a versioned
// type or as an interface implementation.
//
// Identifiers are written to the payload and cannot change without breaking
// stored data, so pin them instead of relying on registration order. If any
// type is pinned, every type that receives an identifier must be pinned.
func WithDTM(t reflect.Type, dtm binco.DTM) SetOption {
	return func(o *Options) (err error) {
		if dtm <= 0 {
			return binco.ErrOverflow
		}
		o.DTMs[t] = dtm
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

// Validate reports options that cannot produce a working generator.
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

// Apply runs opts in order and validates the result.
func Apply(ops []SetOption, o *Options) (err error) {
	for i := range ops {
		if ops[i] != nil {
			err = ops[i](o)
			if err != nil {
				return
			}
		}
	}
	if o.DTMs == nil {
		o.DTMs = make(map[reflect.Type]binco.DTM)
	}
	if o.Logger == nil {
		o.Logger = log.New(io.Discard, "", 0)
	}
	addImportAlias(o)
	return o.Validate()
}

func addImportAlias(o *Options) {
	o.importAliases[ImportPath(o.PkgPath)] = Alias(o.Package)
}
