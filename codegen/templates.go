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

func (f TmplFns) ModImportName(opts genopts.Options) string {
	if opts.Stream {
		return "wirefmt"
	}
	return "binco"
}

func (f TmplFns) SkipLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r"
	}
	return "bs[n:]"
}

func (f TmplFns) SkipSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r wirefmt.Reader"
	}
	return "bs []byte"
}

func (f TmplFns) UnmarshalLastParam(first bool, opts genopts.Options) string {
	if opts.Stream {
		return "r"
	}
	if first {
		return "bs"
	}
	return "bs[n:]"
}

func (f TmplFns) UnmarshalSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "r wirefmt.Reader"
	}
	return "bs []byte"
}

func (f TmplFns) MarshalLastParam(first bool, opts genopts.Options) string {
	if opts.Stream {
		return "w"
	}
	if first {
		return "bs"
	}
	return "bs[n:]"
}

func (f TmplFns) MarshalSignatureLastParam(opts genopts.Options) string {
	if opts.Stream {
		return "w wirefmt.Writer"
	}
	return "bs []byte"
}

func (f TmplFns) Include(name string, pipeline any) (str string, err error) {
	var buf bytes.Buffer
	if err = f.tmpl.ExecuteTemplate(&buf, name, pipeline); err != nil {
		return
	}
	str = buf.String()
	return
}

func (f TmplFns) TimeSerializer(topts tpopts.Options) string {
	return topts.TimeUnit.Serializer()
}

func (f TmplFns) ByteSliceStream(name typename.FullName,
	gopts genopts.Options) bool {
	return gopts.Stream && (name == "[]byte" || name == "[]uint8")
}

func (f TmplFns) Minus(a int, b int) int {
	return a - b
}

func (f TmplFns) WithComma(str string) string {
	if str == "" {
		return str
	}
	return ", " + str
}

func (f TmplFns) ArrayType(name typename.FullName) (ok bool) {
	return arrayRe.MatchString(string(name))
}

func (f TmplFns) PtrType(name typename.FullName) (ok bool) {
	return ptrRe.MatchString(string(name))
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
	PtrPattern   = `(^\*)(.+$)`
	ArrayPattern = `[\d+](.+$)`
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
