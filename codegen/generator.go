package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"text/template"

	"github.com/malivvan/binco"
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
//
// A Generator is not safe for concurrent use: it records the types it renders
// and the packages it has seen. Create one per generation run.
type Generator struct {
	rootTmpl     *template.Template
	builder      bldr.TypeBuilder
	converter    cnvtr.TypeNameConverter
	foreignTypes map[typename.FullName]struct{}
	anonTypes    map[spec.AnonSerializerName]spec.AnonType
	queue        []queued
	dtmGroups    [][]dtmEntry
	dtmIDs       map[string]binco.DTM
	resolvedDTMs map[string]binco.DTM
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

// resolveDTMs assigns an identifier to every registered type.
//
// Pinned identifiers (genopts.WithDTM) are used as they are. When none is
// pinned the types are numbered in registration order, which keeps the previous
// behaviour but renumbers payloads when a type is added or removed; pin the ids
// in any format that has to outlive a build. Mixing pinned and automatic ids is
// rejected, because it makes the numbering depend on both.
func (g *Generator) resolveDTMs() ([][]dtmValue, error) {
	explicit := len(g.dtmIDs) > 0
	next := 1
	seenIDs := map[binco.DTM]string{}
	seenNames := map[string]string{}
	groups := make([][]dtmValue, 0, len(g.dtmGroups))
	resolved := make(map[string]binco.DTM, len(g.dtmIDs))
	for _, group := range g.dtmGroups {
		values := make([]dtmValue, 0, len(group))
		for _, e := range group {
			if other, dup := seenNames[e.Name]; dup {
				return nil, fmt.Errorf("%w: %s and %s would both declare %sDTM",
					binco.ErrDuplicateDTM, other, e.FullName, e.Name)
			}
			seenNames[e.Name] = string(e.FullName)
			id, pinned := g.dtmIDs[string(e.FullName)]
			if !pinned {
				if explicit {
					return nil, fmt.Errorf("%w: no identifier for %s; pin every registered type with genopts.WithDTM, or none",
						binco.ErrDuplicateDTM, e.FullName)
				}
				id = binco.DTM(next)
				next++
			}
			if id <= 0 {
				return nil, fmt.Errorf("%w: %s has a non-positive identifier %d",
					binco.ErrDuplicateDTM, e.FullName, id)
			}
			if other, dup := seenIDs[id]; dup {
				return nil, fmt.Errorf("%w: %s and %s share identifier %d",
					binco.ErrDuplicateDTM, other, e.FullName, id)
			}
			seenIDs[id] = string(e.FullName)
			resolved[string(e.FullName)] = id
			values = append(values, dtmValue{Name: e.Name, ID: id})
		}
		groups = append(groups, values)
	}
	g.resolvedDTMs = resolved
	return groups, nil
}

// DTMRegistry returns the identifier assigned to every registered type, keyed by
// full type name.
//
// With pinned identifiers the result is stable across builds, so it can be
// committed as a golden file and compared in CI.
func (g *Generator) DTMRegistry() (map[string]binco.DTM, error) {
	if _, err := g.resolveDTMs(); err != nil {
		return nil, err
	}
	out := make(map[string]binco.DTM, len(g.resolvedDTMs))
	for name, id := range g.resolvedDTMs {
		out[name] = id
	}
	return out, nil
}

// renderAll writes the whole file into one buffer.
//
// It used to render every section into its own buffer and append the results,
// which copied the output once per section.
func (g *Generator) renderAll(buf *bytes.Buffer) (err error) {
	if err = g.rootTmpl.ExecuteTemplate(buf, packageTmpl, g.opts); err != nil {
		return fmt.Errorf("package: %w", err)
	}
	if err = g.rootTmpl.ExecuteTemplate(buf, importsTmpl, g.opts); err != nil {
		return fmt.Errorf("imports: %w", err)
	}
	groups, err := g.resolveDTMs()
	if err != nil {
		return fmt.Errorf("type identifiers: %w", err)
	}
	for _, values := range groups {
		if err = g.rootTmpl.ExecuteTemplate(buf, dtmsDefinitionTmpl, values); err != nil {
			return fmt.Errorf("type identifiers: %w", err)
		}
	}
	if err = g.rootTmpl.ExecuteTemplate(buf, anonDefinitionsTmpl, struct {
		Map map[spec.AnonSerializerName]spec.AnonType
		Ops genops.Options
	}{
		Map: g.anonTypes,
		Ops: g.opts,
	}); err != nil {
		return fmt.Errorf("anonymous serializers: %w", err)
	}
	for _, item := range g.queue {
		if err = g.rootTmpl.ExecuteTemplate(buf, item.fileName, item.typeData); err != nil {
			return fmt.Errorf("serializers: %w", err)
		}
	}
	return nil
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
	g.queue = append(g.queue, queued{definedTypeSerializerTmpl, definedType})
	return
}

// Generate renders the collected types and returns the formatted source.
//
// Errors from the templates, from imports.Process and from the syntax check are
// wrapped with %w around ErrCodeGenFailed, so the cause survives; the previous
// implementation replaced it with the generic sentinel.
func (g *Generator) Generate() (scratch []byte, err error) {
	var buf bytes.Buffer
	if err = g.renderAll(&buf); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCodeGenFailed, err)
	}
	tmp := buf.Bytes()
	scratch, err = imports.Process("", tmp, nil)
	if err != nil {
		return tmp, fmt.Errorf("%w: %v", ErrCodeGenFailed, err)
	}
	if err = g.checkSyntax(scratch); err != nil {
		return tmp, fmt.Errorf("%w: %v", ErrCodeGenFailed, err)
	}
	return scratch, nil
}

func (g *Generator) addDTM(entries []dtmEntry) {
	g.dtmGroups = append(g.dtmGroups, entries)
}

// typeDTMEntry names one type that receives a wire identifier.
func (g *Generator) typeDTMEntry(t reflect.Type) (entry dtmEntry, err error) {
	cname, err := typename.TypeCompleteName(t)
	if err != nil {
		return dtmEntry{}, err
	}
	fullName, err := g.converter.ConvertToFullName(cname)
	if err != nil {
		return dtmEntry{}, err
	}
	return dtmEntry{
		Name:     t.Name(),
		FullName: fullName,
	}, nil
}

func (g *Generator) RegisterVersioned(t reflect.Type,
	opts ...veropts.SetOption,
) (err error) {
	vops := veropts.Options{}
	if err = veropts.Apply(&vops, opts...); err != nil {
		return
	}
	var (
		entries = make([]dtmEntry, 0, len(vops.Versions))
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
		entry, err := g.typeDTMEntry(version.Type)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
	}
	g.addDTM(entries)
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
		l       = len(rops.StructImpls) + len(rops.DefinedTypeImpls)
		iops    = make([]intropts.SetOption, 0, l)
		entries = make([]dtmEntry, 0, l)
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
		entry, err := g.typeDTMEntry(impl.Type)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
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
		entry, err := g.typeDTMEntry(impl.Type)
		if err != nil {
			return err
		}
		entries = append(entries, entry)
	}
	if rops.Marshaller {
		iops = append(iops, intropts.WithMarshaller())
	}
	g.addDTM(entries)
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
	g.queue = append(g.queue, queued{versionedSerializerTmpl, versionedType})
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
	g.queue = append(g.queue, queued{interfaceSerializerTmpl, interfaceType})
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
	g.queue = append(g.queue, queued{typedSerializerTmpl, tp})
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
	for _, fd := range structType.SerializedFields() {
		var tops tpopts.Options
		if len(sops.Fields) > 0 && fd.Index < len(sops.Fields) {
			tops = sops.Fields[fd.Index].Type
		}
		if err = g.builder.CollectAnonTypes(fd.FullName, g.anonTypes, tops); err != nil {
			return
		}
	}
	g.markForeign(t, structType.FullName)
	g.queue = append(g.queue, queued{structSerializerTmpl, structType})
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
	g.queue = append(g.queue, queued{definedTypeSerializerTmpl, definedType})
	return
}

// NewGenerator returns a Generator configured by opts. Only WithPath is required.
func NewGenerator(opts ...genops.SetOption) (g *Generator, err error) {
	settings := genops.New()
	if err = genops.Apply(opts, &settings); err != nil {
		return
	}
	var (
		foreignTypes = map[typename.FullName]struct{}{}
		converter    = cnvtr.NewTypeNameConverter(settings)
		builder      = bldr.NewTypeBuilder(converter, settings)
		dtmIDs       = make(map[string]binco.DTM, len(settings.DTMs))
	)
	for t, id := range settings.DTMs {
		cname, cerr := typename.TypeCompleteName(t)
		if cerr != nil {
			return nil, cerr
		}
		fullName, ferr := converter.ConvertToFullName(cname)
		if ferr != nil {
			return nil, ferr
		}
		dtmIDs[string(fullName)] = id
	}
	rootTmpl := template.New("base")
	if err = registerFns(builder, converter, foreignTypes, rootTmpl, settings); err != nil {
		return nil, err
	}
	if err = registerTemplates(rootTmpl); err != nil {
		return nil, err
	}
	g = &Generator{
		rootTmpl:     rootTmpl,
		builder:      builder,
		converter:    converter,
		foreignTypes: foreignTypes,
		anonTypes:    map[spec.AnonSerializerName]spec.AnonType{},
		queue:        []queued{},
		dtmIDs:       dtmIDs,
		opts:         settings,
	}
	return
}

// queued is one planned template execution: a template name and its data.
type queued struct {
	fileName string
	typeData any
}

// dtmEntry is one type that receives a wire identifier.
type dtmEntry struct {
	Name     string
	FullName typename.FullName
}

// dtmValue is one rendered identifier: the constant name and its literal value.
type dtmValue struct {
	Name string
	ID   binco.DTM
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
) error {
	tmplFns := TmplFns{tmpl: tmpl}
	serializerFns, err := NewSerializerFns(converter, builder, foreignTypes, tmplFns, gopts)
	if err != nil {
		return err
	}
	m := map[string]any{}
	m["Tmpl"] = func() TmplFns {
		return tmplFns
	}
	m["Serializer"] = func() SerializerFns {
		return serializerFns
	}
	tmpl.Funcs(m)
	return nil
}

// registerTemplates parses the embedded templates.
func registerTemplates(tmpl *template.Template) (err error) {
	subFS, _ := fs.Sub(templatesFS, "templates")
	_, err = tmpl.ParseFS(subFS, "*.tmpl")
	return
}
