package codegen

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	genopts "github.com/malivvan/binco/codegen/config/gen"
	tpopts "github.com/malivvan/binco/codegen/config/type"
	"github.com/malivvan/binco/codegen/spec"
	cnvtr "github.com/malivvan/binco/codegen/spec/converter"
	"github.com/malivvan/binco/codegen/typename"
)

// SerializerFns answers the naming questions the templates ask: how a serializer, a typed serializer or a type identifier is named, and which codec a Go type maps to.
type SerializerFns struct {
	converter       cnvtr.TypeNameConverter
	builder         TypeBuilder
	foreignTypes    map[typename.FullName]struct{}
	serializerNames map[typename.FullName]string
	utilFns         TmplFns
	gopts           genopts.Options
}

func (f SerializerFns) lowerFirst(str string) string {
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func (f SerializerFns) upperFirst(str string) string {
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func (f SerializerFns) codecName(name typename.FullName) string {
	if serializerName, pst := f.serializerNames[name]; pst {
		return serializerName
	}
	var rname typename.RelativeName
	if _, pst := f.foreignTypes[name]; pst {
		rname = typename.RelativeName(name.TypeName())
	} else {
		rname = f.converter.ConvertToRelativeName(name)
	}
	return rname.WithoutSquares()
}

func (f SerializerFns) MapOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		keyType  = f.converter.ConvertToRelativeName(anonType.KeyType)
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "mapopts.WithLenSerializer[%v, %v](%v)", keyType, elemType,
			anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithLenValidator[%v, %v](%v)", keyType, elemType,
			anonType.LenVl)
	}
	if anonType.KeyVl != "" && anonType.KeyVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithKeyValidator[%v, %v](%v)", keyType, elemType,
			anonType.KeyVl)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "mapopts.WithValueValidator[%v, %v](%v)", keyType, elemType, anonType.ElemVl)
	}
	return b.String()
}

func (f SerializerFns) SliceOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "slopts.WithLenSerializer[%v](%v)", elemType, anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "slopts.WithLenValidator[%v](%v)", elemType, anonType.LenVl)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "slopts.WithElemValidator[%v](%v)", elemType, anonType.ElemVl)
	}
	return b.String()
}

func (f SerializerFns) ByteSliceOpts(anonType spec.AnonType) string {
	b := strings.Builder{}
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "bslopts.WithLenSerializer(%v)", anonType.LenSerializer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "bslopts.WithLenValidator(%v)", anonType.LenVl)
	}
	return b.String()
}

func (f SerializerFns) ArrayOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSerializer != "" && anonType.LenSerializer != "nil" {
		fmt.Fprintf(&b, "arropts.WithLenSerializer[%v](%v)", elemType, anonType.LenSerializer)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "arropts.WithElemValidator[%v](%v)", elemType, anonType.ElemVl)
	}
	return b.String()
}

func (f SerializerFns) StringOpts(anonData spec.AnonType) string {
	b := strings.Builder{}
	if anonData.LenSerializer != "" && anonData.LenSerializer != "nil" {
		fmt.Fprintf(&b, "stropts.WithLenSerializer(%v)", anonData.LenSerializer)
	}
	if anonData.LenVl != "" && anonData.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "stropts.WithLenValidator(%v)", anonData.LenVl)
	}
	return b.String()
}

func (f SerializerFns) FieldTmplPipe(structType spec.StructType,
	fieldType spec.FieldType, index int, opts genopts.Options) FieldTmplPipe {
	pipe := FieldTmplPipe{
		Val:         f.Val(structType),
		FieldsCount: len(structType.SerializedFields()),
		Field:       fieldType,
		Index:       index,
		Gops:        opts,
	}
	if len(structType.Sops.Fields) > 0 {
		pipe.Tops = structType.Sops.Fields[index].Type
	}
	return pipe
}

func (f SerializerFns) RelName(name typename.FullName) string {
	return string(f.converter.ConvertToRelativeName(name))
}

func (f SerializerFns) Elem(anonType spec.AnonType, opts genopts.Options) string {
	return f.Of(anonType.ElemType, tpopts.Options{}, opts)
}

func (f SerializerFns) Key(anonType spec.AnonType, opts genopts.Options) string {
	return f.Of(anonType.KeyType, tpopts.Options{}, opts)
}

func (f SerializerFns) Of(fullName typename.FullName, tops tpopts.Options,
	opts genopts.Options,
) string {
	anonType, ok, err := f.builder.BuildAnonType(fullName, tops)
	if err != nil {
		panic(fmt.Sprintf("can't get serializer for the %v, cause %v", fullName, err))
	}
	if ok {
		return string(anonType.SerializerName)
	}
	var (
		pkg            string
		serializerName string
		rules          map[string]string
	)

	switch opts.Mode {
	case genopts.ModeSafe:
		rules = portableRules(tops)
	case genopts.ModeNotUnsafe:
		rules = hybridRules(tops, opts)
	case genopts.ModeUnsafe:
		rules = nativeRules(tops, opts)
	}
	pkg, pst := rules[string(fullName)]
	if !pst {
		return f.Var(fullName)
	}

	switch fullName {
	case "int", "int64", "int32", "int16", "int8":
		serializerName = f.upperFirst(string(fullName))
		if tops.NumEncoding == tpopts.NumEncodingVarintPositive {
			serializerName = "Positive" + serializerName
		}
	case "uint", "uint64", "uint32", "uint16", "uint8", "float64", "float32", "byte":
		serializerName = f.upperFirst(string(fullName))
	case "bool", "string":
		serializerName = f.upperFirst(string(fullName))
	case "[]byte", "[]uint8":
		serializerName = "ByteSlice"
	case "time.Time":
		serializerName = rules["time_serializer"]
	default:
		panic(fmt.Sprintf("internal bug: unpredicted case %v", fullName))
	}
	return pkg + "." + serializerName
}

func (f SerializerFns) Val(tp any) string {
	return "v"
}

func (f SerializerFns) Type(fullName typename.FullName) string {
	return f.lowerFirst(f.Var(fullName))
}

func (f SerializerFns) DTMVar(fullName typename.FullName) string {
	return string(f.codecName(fullName)) + SuffixDTM
}

func (f SerializerFns) TypedVar(fullName typename.FullName) string {
	return string(f.codecName(fullName)) + SuffixTypedBinco
}

func (f SerializerFns) Var(name typename.FullName) string {
	return string(f.codecName(name)) + SuffixBinco
}

// NewSerializerFns prepares the naming helpers for one generator run.
func NewSerializerFns(converter cnvtr.TypeNameConverter,
	builder TypeBuilder,
	foreignTypes map[typename.FullName]struct{},
	utilFns TmplFns,
	gopts genopts.Options,
) SerializerFns {

	serializerNames := map[typename.FullName]string{}
	for t, serializerName := range gopts.SerializerNames {
		cname, err := typename.TypeCompleteName(t)
		if err != nil {
			panic(err)
		}
		fullName := converter.ConvertToFullName(cname)
		serializerNames[fullName] = serializerName
	}

	return SerializerFns{
		converter:       converter,
		builder:         builder,
		foreignTypes:    foreignTypes,
		serializerNames: serializerNames,
		utilFns:         utilFns,
		gopts:           gopts,
	}
}

// TypeBuilder turns reflect types into the specifications the templates consume.
type TypeBuilder interface {
	BuildDefinedType(t reflect.Type, tops tpopts.Options) (spec.DefinedType, error)
	BuildAnonType(name typename.FullName, topts tpopts.Options) (
		spec.AnonType, bool, error)
}

// The suffixes appended to a type name: SuffixBinco for a serializer, SuffixTypedBinco for a typed serializer and SuffixDTM for a type identifier.
const (
	SuffixBinco      = "Binco"
	SuffixTypedBinco = "TypedBinco"
	SuffixDTM        = "DTM"
)
