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

// SerFns answers the naming questions the templates ask: how a serializer, a typed serializer or a type identifier is named, and which codec a Go type maps to.
type SerFns struct {
	converter    cnvtr.TypeNameConverter
	builder      TypeBuilder
	foreignTypes map[typename.FullName]struct{}
	serNames     map[typename.FullName]string
	utilFns      TmplFns
	gopts        genopts.Options
}

func (f SerFns) lowerFirst(str string) string {
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func (f SerFns) upperFirst(str string) string {
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func (f SerFns) codecName(name typename.FullName) string {
	if serName, pst := f.serNames[name]; pst {
		return serName
	}
	var rname typename.RelativeName
	if _, pst := f.foreignTypes[name]; pst {
		rname = typename.RelativeName(name.TypeName())
	} else {
		rname = f.converter.ConvertToRelativeName(name)
	}
	return rname.WithoutSquares()
}

func (f SerFns) MapOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		keyType  = f.converter.ConvertToRelativeName(anonType.KeyType)
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSer != "" && anonType.LenSer != "nil" {
		fmt.Fprintf(&b, "mapopts.WithLenSer[%v, %v](%v)", keyType, elemType,
			anonType.LenSer)
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

func (f SerFns) SliceOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSer != "" && anonType.LenSer != "nil" {
		fmt.Fprintf(&b, "slopts.WithLenSer[%v](%v)", elemType, anonType.LenSer)
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

func (f SerFns) ByteSliceOpts(anonType spec.AnonType) string {
	b := strings.Builder{}
	if anonType.LenSer != "" && anonType.LenSer != "nil" {
		fmt.Fprintf(&b, "bslopts.WithLenSer(%v)", anonType.LenSer)
	}
	if anonType.LenVl != "" && anonType.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "bslopts.WithLenValidator(%v)", anonType.LenVl)
	}
	return b.String()
}

func (f SerFns) ArrayOpts(anonType spec.AnonType) string {
	var (
		b        = strings.Builder{}
		elemType = f.converter.ConvertToRelativeName(anonType.ElemType)
	)
	if anonType.LenSer != "" && anonType.LenSer != "nil" {
		fmt.Fprintf(&b, "arropts.WithLenSer[%v](%v)", elemType, anonType.LenSer)
	}
	if anonType.ElemVl != "" && anonType.ElemVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "arropts.WithElemValidator[%v](%v)", elemType, anonType.ElemVl)
	}
	return b.String()
}

func (f SerFns) StringOpts(anonData spec.AnonType) string {
	b := strings.Builder{}
	if anonData.LenSer != "" && anonData.LenSer != "nil" {
		fmt.Fprintf(&b, "stropts.WithLenSer(%v)", anonData.LenSer)
	}
	if anonData.LenVl != "" && anonData.LenVl != "nil" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "stropts.WithLenValidator(%v)", anonData.LenVl)
	}
	return b.String()
}

func (f SerFns) FieldTmplPipe(structType spec.StructType,
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

func (f SerFns) RelName(name typename.FullName) string {
	return string(f.converter.ConvertToRelativeName(name))
}

func (f SerFns) Elem(anonType spec.AnonType, opts genopts.Options) string {
	return f.Of(anonType.ElemType, tpopts.Options{}, opts)
}

func (f SerFns) Key(anonType spec.AnonType, opts genopts.Options) string {
	return f.Of(anonType.KeyType, tpopts.Options{}, opts)
}

func (f SerFns) Of(fullName typename.FullName, tops tpopts.Options,
	opts genopts.Options,
) string {
	anonType, ok, err := f.builder.BuildAnonType(fullName, tops)
	if err != nil {
		panic(fmt.Sprintf("can't get serializer for the %v, cause %v", fullName, err))
	}
	if ok {
		return string(anonType.SerName)
	}
	var (
		pkg     string
		serName string
		rules   map[string]string
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
		serName = f.upperFirst(string(fullName))
		if tops.NumEncoding == tpopts.NumEncodingVarintPositive {
			serName = "Positive" + serName
		}
	case "uint", "uint64", "uint32", "uint16", "uint8", "float64", "float32", "byte":
		serName = f.upperFirst(string(fullName))
	case "bool", "string":
		serName = f.upperFirst(string(fullName))
	case "[]byte", "[]uint8":
		serName = "ByteSlice"
	case "time.Time":
		serName = rules["time_ser"]
	default:
		panic(fmt.Sprintf("internal bug: unpredicted case %v", fullName))
	}
	return pkg + "." + serName
}

func (f SerFns) Val(tp any) string {
	return "v"
}

func (f SerFns) Type(fullName typename.FullName) string {
	return f.lowerFirst(f.Var(fullName))
}

func (f SerFns) DTMVar(fullName typename.FullName) string {
	return string(f.codecName(fullName)) + SuffixDTM
}

func (f SerFns) TypedVar(fullName typename.FullName) string {
	return string(f.codecName(fullName)) + SuffixTypedBinco
}

func (f SerFns) Var(name typename.FullName) string {
	return string(f.codecName(name)) + SuffixBinco
}

// NewSerFns prepares the naming helpers for one generator run.
func NewSerFns(converter cnvtr.TypeNameConverter,
	builder TypeBuilder,
	foreignTypes map[typename.FullName]struct{},
	utilFns TmplFns,
	gopts genopts.Options,
) SerFns {

	serNames := map[typename.FullName]string{}
	for t, serName := range gopts.SerNames {
		cname, err := typename.TypeCompleteName(t)
		if err != nil {
			panic(err)
		}
		fullName := converter.ConvertToFullName(cname)
		serNames[fullName] = serName
	}

	return SerFns{
		converter:    converter,
		builder:      builder,
		foreignTypes: foreignTypes,
		serNames:     serNames,
		utilFns:      utilFns,
		gopts:        gopts,
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
