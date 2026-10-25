package codegen

import (
	"bytes"
	"regexp"
	"text/template"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/spec"
	"github.com/malivvan/binco/codegen/typename"
)

// TmplFns holds the helpers that build the signatures and the option lists of the generated code.
type TmplFns struct {
	tmpl *template.Template
}

// ModImportName is the package name of the API a template targets: binco or wirefmt.
func (f TmplFns) ModImportName(opts genopts.Options) string {
	if opts.Stream {
		return "wirefmt"
	}
	return "binco"
}

// SkipLastParam is the argument a Skip call passes on.
func (f TmplFns) SkipLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r"
	}
	return "bs[n:]"
}

// SkipSignatureLastParam is the last parameter of a Skip signature.
func (f TmplFns) SkipSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r wirefmt.Reader"
	}
	return "bs []byte"
}

// UnmarshalLastParam is the argument an Unmarshal call passes on.
func (f TmplFns) UnmarshalLastParam(first bool, opts genopts.Options) string {
	if opts.Stream {
		return "r"
	}
	if first {
		return "bs"
	}
	return "bs[n:]"
}

// UnmarshalSignatureLastParam is the last parameter of an Unmarshal signature.
func (f TmplFns) UnmarshalSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r wirefmt.Reader"
	}
	return "bs []byte"
}

// MarshalLastParam is the argument a Marshal call passes on.
func (f TmplFns) MarshalLastParam(first bool, opts genopts.Options) string {
	if opts.Stream {
		return "w"
	}
	if first {
		return "bs"
	}
	return "bs[n:]"
}

// MarshalSignatureLastParam is the last parameter of a Marshal signature.
func (f TmplFns) MarshalSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "w wirefmt.Writer"
	}
	return "bs []byte"
}

// Include renders another template into the current one.
func (f TmplFns) Include(name string, pipeline any) (str string, err error) {
	var buf bytes.Buffer
	if err = f.tmpl.ExecuteTemplate(&buf, name, pipeline); err != nil {
		return
	}
	str = buf.String()
	return
}

// TimeSerializer is the name of the time codec a template should use.
func (f TmplFns) TimeSerializer(topts tpopts.Options) string {
	return topts.TimeUnit.Serializer()
}

// ByteSliceStream reports whether byte slices need the streaming codec.
func (f TmplFns) ByteSliceStream(name typename.FullName,
	gopts genopts.Options) bool {
	return gopts.Stream && (name == "[]byte" || name == "[]uint8")
}

// Minus subtracts b from a, for templates that must compare indices.
func (f TmplFns) Minus(a int, b int) int {
	return a - b
}

// WithComma prefixes a non-empty option list with a comma and a space.
func (f TmplFns) WithComma(str string) string {
	if str == "" {
		return str
	}
	return ", " + str
}

// ArrayType reports whether a type name is an array.
func (f TmplFns) ArrayType(name typename.FullName) (ok bool) {
	return arrayRe.MatchString(string(name))
}

// PtrType reports whether a type name is a pointer.
func (f TmplFns) PtrType(name typename.FullName) (ok bool) {
	return ptrRe.MatchString(string(name))
}

// NilInterfaceMarshal is the statement that writes the reserved zero type
// identifier, which stands for a nil interface value.
func (f TmplFns) NilInterfaceMarshal(opts genopts.Options) string {
	if opts.Stream {
		return "err = w.WriteByte(0)\n\t\tif err != nil {\n\t\t\treturn\n\t\t}\n\t\tn = 1\n\t\treturn"
	}
	return "bs[0] = 0\n\t\treturn 1"
}

// NilInterfaceSize is the encoded size of a nil interface value.
func (f TmplFns) NilInterfaceSize() string {
	return "1"
}

// FieldTmplPipe is the data passed to the field templates.
type FieldTmplPipe struct {
	Val         string
	FieldsCount int
	Field       spec.FieldType
	Index       int
	Tops        tpopts.Options
	Gops        genopts.Options
}

// The patterns used to recognize pointer and array type names inside the templates.
const (
	PtrPattern = `(^\*)(.+$)`
	// ArrayPattern is anchored on purpose. The previous pattern, `[\d+](.+$)`,
	// was a character class: it matched "int32" and "Foo1Bar" as if they were
	// arrays, which would have misclassified field types as soon as a template
	// started using it.
	ArrayPattern = `^\*?\[\d+\]`
)

// ptrRe matches a pointer type name.
var (
	ptrRe   = regexp.MustCompile(PtrPattern)
	arrayRe = regexp.MustCompile(ArrayPattern)
)

// The names of the embedded templates.
const (
	definedTypeSerializerTmpl = "serializer_defined_type.tmpl"
	structSerializerTmpl      = "serializer_struct.tmpl"
	typedSerializerTmpl       = "serializer_typed.tmpl"
	interfaceSerializerTmpl   = "serializer_interface.tmpl"
	versionedSerializerTmpl   = "serializer_versioned.tmpl"
	packageTmpl               = "package.tmpl"
	importsTmpl               = "imports.tmpl"
	anonDefinitionsTmpl       = "anon_definitions.tmpl"
	dtmsDefinitionTmpl        = "dtm_definition.tmpl"
)
