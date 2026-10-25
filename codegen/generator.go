package codegen

import (
	"bytes"
	"embed"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"text/template"

	"github.com/malivvan/binco/codegen/classifier"
	genops "github.com/malivvan/binco/codegen/config/gen"
	intropts "github.com/malivvan/binco/codegen/config/interface"
	stopts "github.com/malivvan/binco/codegen/config/struct"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	veropts "github.com/malivvan/binco/codegen/config/versioned"
	"github.com/malivvan/binco/codegen/spec"
	bldr "github.com/malivvan/binco/codegen/spec/builder"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
	"golang.org/x/tools/imports"
)

// Generator collects the types to generate code for and renders the result.
type Generator struct {
	rootTmpl     *template.Template
	builder      bldr.TypeBuilder
	foreignTypes map[typename.FullName]struct{}
	anonTypes    map[spec.AnonSerName]spec.AnonType
	queue        []queued
	dtmGroups    [][]string
	scratch      []byte
	opts         genops.Options
}

func (g *Generator) checkSyntax(scratch []byte) (err error) {
	var (
		fs  = token.NewFileSet()
		src = string(scratch)
	)
	_, err = parser.ParseFile(fs, "", src, parser.AllErrors)
	return
}

func (g *Generator) isForeign(t reflect.Type) bool {
	return t.PkgPath() != string(g.opts.PkgPath)
}

func (g *Generator) markForeign(t reflect.Type, fullName typename.FullName) {
	if g.isForeign(t) {
		g.foreignTypes[fullName] = struct{}{}
	}
}

func (g *Generator) renderDTMs() (scratch []byte) {
	buf := bytes.Buffer{}
	start := 1
	for _, types := range g.dtmGroups {
		err := g.rootTmpl.ExecuteTemplate(&buf, dtmsDefinitionTmpl, dtmGroup{
			Types: types,
			Start: start,
		})
		if err != nil {
			panic(err)
		}
		start += len(types)
	}
	scratch = buf.Bytes()
	return
}

func (g *Generator) render(tmplName string, a any) (scratch []byte) {
	buf := bytes.NewBuffer(make([]byte, 0))
	err := g.rootTmpl.ExecuteTemplate(buf, tmplName, a)
	if err != nil {
		panic(err)
	}
	scratch = buf.Bytes()
	return
}

func (g *Generator) generateSerializers() (scratch []byte) {
	for _, item := range g.queue {
		scratch = append(scratch, g.render(item.fileName, item.typeData)...)
	}
	return
}

func (g *Generator) generateAnonymosDefinitions() (scratch []byte) {
	return g.render(anonDefinitionsTmpl, struct {
		Map map[spec.AnonSerName]spec.AnonType
		Ops genops.Options
	}{
		Map: g.anonTypes,
		Ops: g.opts,
	})
}

func (g *Generator) generateImports() (scratch []byte) {
	return g.render(importsTmpl, g.opts)
}

func (g *Generator) generatePackage() (scratch []byte) {
	return g.render(packageTmpl, g.opts)
}

func (g *Generator) addStructUnderlyingTime(t reflect.Type, sops stopts.Options) (
	err error,
) {
	uops := stopts.UnderlyingTimeOptions{}
	stopts.UnderlyingTimeApply(sops.UnderlyingTime, &uops)
	definedType, err := g.builder.BuildTimeType(t, uops, sops.Validator)
	if err != nil {
		return
	}
	g.markForeign(t, definedType.FullName)
	g.queue = append(g.queue, queued{definedTypeSerTmpl, definedType})
	return
}

func (g *Generator) Generate() (scratch []byte, err error) {
	tmp := g.generatePackage()
	tmp = append(tmp, g.generateImports()...)
	tmp = append(tmp, g.renderDTMs()...)
	tmp = append(tmp, g.generateAnonymosDefinitions()...)
	tmp = append(tmp, g.generateSerializers()...)
	scratch, err = imports.Process("", tmp, nil)
	if err != nil {
		err = ErrCodeGenFailed
		return tmp, err
	}
	err = g.checkSyntax(scratch)
	if err != nil {
		err = ErrCodeGenFailed
		return tmp, err
	}
	return
}

func (g *Generator) addDTM(types []string) {
	g.dtmGroups = append(g.dtmGroups, types)
}

func (g *Generator) RegisterVersioned(t reflect.Type,
	opts ...veropts.SetOption,
) (err error) {
	vops := veropts.Options{}
	if err = veropts.Apply(&vops, opts...); err != nil {
		return
	}
	var (
		types = make([]string, 0, len(vops.Versions))
	)
	for _, version := range vops.Versions {
		if classifier.DefinedStruct(version.Type) {
			err = g.AddStruct(version.Type)
		} else {
			err = g.AddDefinedType(version.Type)
		}
		if err != nil {
			return
		}
		err = g.AddTyped(version.Type)
		if err != nil {
			return
		}
		types = append(types, version.Type.Name())
	}
	g.addDTM(types)
	return g.AddVersioned(t, opts...)
}

func (g *Generator) RegisterInterface(t reflect.Type,
	opts ...intropts.SetRegisterOption,
) (err error) {
	rops := intropts.RegisterOptions{}
	if opts != nil {
		intropts.ApplyRegister(&rops, opts...)
	}
	var (
		l     = len(rops.StructImpls) + len(rops.DefinedTypeImpls)
		iops  = make([]intropts.SetOption, 0, l)
		types = make([]string, 0, l)
	)
	for _, impl := range rops.StructImpls {
		err = g.AddStruct(impl.Type, impl.Opts...)
		if err != nil {
			return
		}
		err = g.AddTyped(impl.Type)
		if err != nil {
			return
		}
		iops = append(iops, intropts.WithImpl(impl.Type))
		types = append(types, impl.Type.Name())
	}
	for _, impl := range rops.DefinedTypeImpls {
		err = g.AddDefinedType(impl.Type, impl.Opts...)
		if err != nil {
			return
		}
		err = g.AddTyped(impl.Type)
		if err != nil {
			return
		}
		iops = append(iops, intropts.WithImpl(impl.Type))
		types = append(types, impl.Type.Name())
	}
	if rops.Marshaller {
		iops = append(iops, intropts.WithMarshaller())
	}
	g.addDTM(types)
	return g.AddInterface(t, iops...)
}

func (g *Generator) AddVersioned(t reflect.Type, opts ...veropts.SetOption) (
	err error,
) {
	vops := veropts.Options{}
	if err = veropts.Apply(&vops, opts...); err != nil {
		return
	}
	versionedType, err := g.builder.BuildVersionedType(t, vops)
	if err != nil {
		return
	}
	g.markForeign(t, versionedType.FullName)
	g.queue = append(g.queue, queued{versionedSerTmpl, versionedType})
	return nil
}

func (g *Generator) AddInterface(t reflect.Type, opts ...intropts.SetOption) (
	err error,
) {
	iops := intropts.Options{}
	intropts.Apply(&iops, opts...)
	interfaceType, err := g.builder.BuildInterfaceType(t, iops)
	if err != nil {
		return
	}
	g.markForeign(t, interfaceType.FullName)
	g.queue = append(g.queue, queued{interfaceSerTmpl, interfaceType})
	return
}

func (g *Generator) AddTyped(t reflect.Type, opts ...tpopts.SetOption) (
	err error,
) {
	tops := tpopts.Options{}
	tpopts.Apply(&tops, opts...)
	tp, err := g.builder.BuildTyped(t)
	if err != nil {
		return
	}
	g.queue = append(g.queue, queued{typedSerTmpl, tp})
	return
}

func (g *Generator) AddStruct(t reflect.Type, opts ...stopts.SetOption) (
	err error,
) {
	sops := stopts.Options{}
	stopts.Apply(opts, &sops)
	if sops.UnderlyingTime != nil {
		return g.addStructUnderlyingTime(t, sops)
	}
	structType, err := g.builder.BuildStructType(t, sops)
	if err != nil {
		return
	}
	for i, fd := range structType.SerializedFields() {
		var tops tpopts.Options
		if len(sops.Fields) > 0 {
			tops = sops.Fields[i].Type
		}
		if err = g.builder.CollectAnonTypes(fd.FullName, g.anonTypes, tops); err != nil {
			return
		}
	}
	g.markForeign(t, structType.FullName)
	g.queue = append(g.queue, queued{structSerTmpl, structType})
	return
}

func (g *Generator) AddDefinedType(t reflect.Type, opts ...tpopts.SetOption) (
	err error,
) {
	tops := tpopts.Options{}
	tpopts.Apply(&tops, opts...)
	definedType, err := g.builder.BuildDefinedType(t, tops)
	if err != nil {
		return
	}
	g.markForeign(t, definedType.FullName)
	if err = g.builder.CollectAnonTypes(definedType.UnderlyingTypeName,
		g.anonTypes, tops); err != nil {
		return
	}
	g.queue = append(g.queue, queued{definedTypeSerTmpl, definedType})
	return
}

// NewGenerator returns a Generator configured by opts. Only WithPkgPath is required.
func NewGenerator(opts ...genops.SetOption) (g *Generator, err error) {
	settings := genops.New()
	if err = genops.Apply(opts, &settings); err != nil {
		return
	}
	var (
		foreignTypes = map[typename.FullName]struct{}{}
		converter    = cnvtr.NewTypeNameConverter(settings)
		builder      = bldr.NewTypeBuilder(converter, settings)
	)
	rootTmpl := template.New("base")
	registerFns(builder, converter, foreignTypes, rootTmpl, settings)
	err = registerTemplates(rootTmpl)
	if err != nil {
		panic(err)
	}
	g = &Generator{
		rootTmpl:     rootTmpl,
		builder:      builder,
		foreignTypes: foreignTypes,
		anonTypes:    map[spec.AnonSerName]spec.AnonType{},
		queue:        []queued{},
		scratch:      []byte{},
		opts:         settings,
	}
	return
}

// queued is one planned template execution: a template name and its data.
type queued struct {
	fileName string
	typeData any
}

// dtmGroup is one group of type identifiers that share a numbering block.
type dtmGroup struct {
	Types []string
	Start int
}

// templatesFS holds the template files compiled into the binary.
//
//go:embed templates/*.tmpl
var templatesFS embed.FS

// registerFns installs the template helper functions.
func registerFns(builder bldr.TypeBuilder, converter cnvtr.TypeNameConverter,
	foreignTypes map[typename.FullName]struct{},
	tmpl *template.Template,
	gopts genops.Options,
) {
	var (
		tmplFns = TmplFns{tmpl: tmpl}
		serFns  = NewSerFns(converter, builder, foreignTypes, tmplFns, gopts)
	)
	m := map[string]any{}
	m["Tmpl"] = func() TmplFns {
		return tmplFns
	}
	m["Ser"] = func() SerFns {
		return serFns
	}
	tmpl.Funcs(m)
}

// registerTemplates parses the embedded templates.
func registerTemplates(tmpl *template.Template) (err error) {
	subFS, _ := fs.Sub(templatesFS, "templates")
	_, err = tmpl.ParseFS(subFS, "*.tmpl")
	return
}
